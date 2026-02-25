package service

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mylxsw/namecheap-domain-probe/backend/internal/model"
)

// ProbeService handles domain probe operations
type ProbeService struct {
	tasks       map[string]*model.ProbeTask
	results     map[string][]model.ProbeItem
	runningTasks map[string]bool // track tasks that are being started
	mu          sync.RWMutex
	outdir      string
}

// NewProbeService creates a new probe service
func NewProbeService(outdir string) *ProbeService {
	return &ProbeService{
		tasks:        make(map[string]*model.ProbeTask),
		results:      make(map[string][]model.ProbeItem),
		runningTasks: make(map[string]bool),
		outdir:       outdir,
	}
}

// CreateTask creates a new probe task
func (s *ProbeService) CreateTask(req *model.CreateProbeRequest) (*model.ProbeTask, error) {
	task := &model.ProbeTask{
		ID:            uuid.New().String(),
		Word:          req.Word,
		Status:        model.ProbeStatusPending,
		TldMode:       model.TldMode(req.TldMode),
		RatePerMin:    req.RatePerMin,
		MaxBatch:      req.MaxBatch,
		CacheTTLHours: req.CacheTTLHours,
		Total:         0,
		Completed:     0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if task.TldMode == "" {
		task.TldMode = model.TldModeAll
	}
	if task.RatePerMin <= 0 {
		task.RatePerMin = 45.0
	}
	if task.MaxBatch <= 0 {
		task.MaxBatch = 50
	}
	if task.CacheTTLHours <= 0 {
		task.CacheTTLHours = 24.0
	}

	s.mu.Lock()
	s.tasks[task.ID] = task
	s.mu.Unlock()

	return task, nil
}

// GetTask gets a task by ID
func (s *ProbeService) GetTask(id string) (*model.ProbeTask, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	task, ok := s.tasks[id]
	return task, ok
}

// GetResults gets results for a task
func (s *ProbeService) GetResults(id string) ([]model.ProbeItem, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	results, ok := s.results[id]
	return results, ok
}

// GetResultsOrLoad returns in-memory results, and falls back to rebuilding
// from results.jsonl if memory cache is empty.
func (s *ProbeService) GetResultsOrLoad(id string) ([]model.ProbeItem, bool, error) {
	results, ok := s.GetResults(id)
	if ok && len(results) > 0 {
		return results, true, nil
	}

	task, taskOK := s.GetTask(id)
	if !taskOK {
		return results, ok, nil
	}

	jsonlPath := task.ResultsJsonl
	if jsonlPath == "" && task.Outdir != "" {
		jsonlPath = filepath.Join(task.Outdir, "results.jsonl")
	}
	if jsonlPath == "" {
		jsonlPath = filepath.Join(s.outdir, task.ID, "results.jsonl")
	}

	rebuilt, err := s.loadResultsFromJSONL(jsonlPath)
	if err != nil {
		return results, ok, err
	}
	if len(rebuilt) == 0 {
		return rebuilt, true, nil
	}

	s.mu.Lock()
	s.results[id] = rebuilt
	s.mu.Unlock()

	return rebuilt, true, nil
}

func (s *ProbeService) loadResultsFromJSONL(path string) ([]model.ProbeItem, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	var items []model.ProbeItem
	seen := make(map[string]bool)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var env struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal([]byte(line), &env); err == nil && env.Type == "progress" && len(env.Data) > 0 {
			var item model.ProbeItem
			if err := json.Unmarshal(env.Data, &item); err == nil && item.Domain != "" && !seen[item.Domain] {
				seen[item.Domain] = true
				items = append(items, item)
				continue
			}
		}

		var direct model.ProbeItem
		if err := json.Unmarshal([]byte(line), &direct); err == nil && direct.Domain != "" && !seen[direct.Domain] {
			seen[direct.Domain] = true
			items = append(items, direct)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

// RunTask runs a probe task
func (s *ProbeService) RunTask(ctx context.Context, taskID string, progressChan chan<- interface{}) error {
	task, ok := s.GetTask(taskID)
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}

	// Check if task is already running or completed
	s.mu.Lock()
	if s.runningTasks[taskID] {
		s.mu.Unlock()
		return fmt.Errorf("task is already running: %s", taskID)
	}
	if task.Status == model.ProbeStatusRunning || task.Status == model.ProbeStatusCompleted {
		s.mu.Unlock()
		return fmt.Errorf("task is already %s: %s", task.Status, taskID)
	}

	// Mark task as running
	s.runningTasks[taskID] = true
	task.Status = model.ProbeStatusRunning
	task.UpdatedAt = time.Now()
	s.mu.Unlock()

	// Clean up runningTasks when done
	defer func() {
		s.mu.Lock()
		delete(s.runningTasks, taskID)
		s.mu.Unlock()
	}()

	// Create output directory for this task
	taskOutdir := filepath.Join(s.outdir, task.ID)
	if err := os.MkdirAll(taskOutdir, 0755); err != nil {
		s.markTaskFailed(taskID, err.Error())
		return err
	}

	cfg, err := FromEnv()
	if err != nil {
		s.markTaskFailed(taskID, err.Error())
		return err
	}

	client := NewNamecheapClient(cfg, task.RatePerMin, 30.0)
	defer client.Close()

	cacheDir := filepath.Join(taskOutdir, ".cache")
	tldCache := filepath.Join(cacheDir, "tlds.json")
	pricingCache := filepath.Join(cacheDir, "pricing_register_1y.json")
	jsonlPath := filepath.Join(taskOutdir, "results.jsonl")

	// Create or truncate JSONL file
	if _, err := os.Create(jsonlPath); err != nil {
		s.markTaskFailed(taskID, err.Error())
		return err
	}

	// Append to JSONL
	appendJsonl := func(obj interface{}) error {
		f, err := os.OpenFile(jsonlPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		defer f.Close()

		w := bufio.NewWriter(f)
		data, _ := json.Marshal(obj)
		w.Write(data)
		w.WriteByte('\n')
		return w.Flush()
	}

	startedAt := time.Now()

	sendPhase := func(phase, message string) {
		log.Printf("[probe:%s] %s: %s", taskID[:8], phase, message)
		if progressChan != nil {
			progressChan <- model.PhaseEvent{
				Type:    "phase",
				Phase:   phase,
				Message: message,
			}
		}
	}

	// Load TLDs
	sendPhase("loading_tlds", "正在从 Namecheap API 加载 TLD 列表...")
	tlds, err := s.loadOrRefreshTlds(ctx, tldCache, task.CacheTTLHours, func() ([]map[string]string, error) {
		return client.GetTldList(ctx)
	})
	if err != nil {
		s.markTaskFailed(taskID, err.Error())
		sendPhase("error", fmt.Sprintf("加载 TLD 列表失败: %s", err.Error()))
		return err
	}
	sendPhase("loading_tlds", fmt.Sprintf("已加载 %d 个 TLD", len(tlds)))

	// Load pricing
	sendPhase("loading_pricing", "正在从 Namecheap API 加载定价信息...")
	pricing, err := s.loadOrRefreshPricing(ctx, pricingCache, task.CacheTTLHours, func() (map[string]map[string]string, error) {
		return client.GetPricingRegister1y(ctx)
	})
	if err != nil {
		s.markTaskFailed(taskID, err.Error())
		sendPhase("error", fmt.Sprintf("加载定价信息失败: %s", err.Error()))
		return err
	}
	sendPhase("loading_pricing", fmt.Sprintf("已加载 %d 个 TLD 的定价信息", len(pricing)))

	// Filter TLDs
	sendPhase("filtering", "正在筛选 TLD...")
	var tldNames []string
	tldMeta := make(map[string]map[string]string)
	seen := make(map[string]bool)
	for _, t := range tlds {
		name := strings.ToLower(strings.TrimSpace(t["Name"]))
		if name == "" {
			continue
		}
		if task.TldMode == model.TldModeApiRegisterableOnly {
			if strings.ToLower(strings.TrimSpace(t["IsApiRegisterable"])) != "true" {
				continue
			}
		}
		if task.TldMode == model.TldModeMainstreamOnly {
			if !s.isMainstreamTld(name, pricing) {
				continue
			}
		}
		if !seen[name] {
			seen[name] = true
			tldNames = append(tldNames, name)
			tldMeta[name] = t
		}
	}

	// Sort TLD names for deterministic order
	sort.Strings(tldNames)

	total := len(tldNames)
	sendPhase("filtering", fmt.Sprintf("筛选后共 %d 个 TLD 需要探测 (模式: %s)", total, task.TldMode))

	if total == 0 {
		log.Printf("[probe:%s] WARNING: 0 TLDs after filtering. TLDs from API: %d, Pricing entries: %d", taskID[:8], len(tlds), len(pricing))
		if len(pricing) == 0 {
			log.Printf("[probe:%s] Pricing is empty - this is likely the cause. Check Namecheap API key permissions.", taskID[:8])
		}
	}

	// Update task total
	s.mu.Lock()
	task.Total = total
	task.UpdatedAt = time.Now()
	s.mu.Unlock()

	var yielded []model.ProbeItem
	done := 0

	// Process in batches
	for i := 0; i < total; i += task.MaxBatch {
		select {
		case <-ctx.Done():
			s.markTaskFailed(taskID, ctx.Err().Error())
			return ctx.Err()
		default:
		}

		end := i + task.MaxBatch
		if end > total {
			end = total
		}
		batchTlds := tldNames[i:end]

		batchDomains := make([]string, len(batchTlds))
		for j, tld := range batchTlds {
			batchDomains[j] = fmt.Sprintf("%s.%s", task.Word, tld)
		}

		batchNum := i/task.MaxBatch + 1
		totalBatches := (total + task.MaxBatch - 1) / task.MaxBatch
		sendPhase("probing", fmt.Sprintf("正在探测第 %d/%d 批域名: %s ... %s", batchNum, totalBatches, batchDomains[0], batchDomains[len(batchDomains)-1]))

		results, err := client.DomainsCheck(ctx, batchDomains)
		if err != nil {
			// Mark entire batch as error
			for _, tld := range batchTlds {
				d := fmt.Sprintf("%s.%s", task.Word, tld)
				errMsg := err.Error()
				item := model.ProbeItem{
					Domain: d,
					Tld:    tld,
					Error:  &errMsg,
					Raw:    map[string]interface{}{"error": errMsg},
				}
				if priceInfo, ok := pricing[tld]; ok {
					item.BaseRegisterPrice = chooseBasePrice(priceInfo)
				}

				yielded = append(yielded, item)
				done++

				event := model.ProgressEvent{
					Type:  "progress",
					Index: done,
					Total: total,
					Data:  item,
				}
				appendJsonl(event)
				if progressChan != nil {
					progressChan <- event
				}
			}
			continue
		}

		// Map results by domain
		byDomain := make(map[string]model.DomainCheckResult)
		for _, r := range results {
			byDomain[strings.ToLower(r.Domain)] = r
		}

		for _, tld := range batchTlds {
			d := fmt.Sprintf("%s.%s", task.Word, tld)
			raw, ok := byDomain[strings.ToLower(d)]
			if !ok {
				raw = model.DomainCheckResult{Domain: d}
			}

			available := toBool(raw.Available)
			isPremium := toBool(raw.IsPremiumName)
			premiumReg := toFloat(raw.PremiumRegistrationPrice)
			icannFee := toFloat(raw.IcannFee)
			eapFee := toFloat(raw.EapFee)
			unsupportedByTLDMeta := false
			if meta, ok := tldMeta[tld]; ok {
				unsupportedByTLDMeta = strings.ToLower(strings.TrimSpace(meta["IsApiRegisterable"])) != "true"
			}

			var currency *string
			var basePrice *float64
			if priceInfo, ok := pricing[tld]; ok {
				if c := strings.TrimSpace(priceInfo["Currency"]); c != "" {
					currency = &c
				}
				basePrice = chooseBasePrice(priceInfo)
			}

			// If base price is missing from pricing API, treat "0" premium fee from domains.check
			// as unknown instead of a real zero-dollar price.
			if basePrice == nil && premiumReg != nil && *premiumReg == 0 {
				premiumReg = nil
			}

			totalPrice := calcTotalPrice(available, isPremium, basePrice, premiumReg, icannFee, eapFee)

			// Convert raw to map
			rawMap := map[string]interface{}{
				"Domain":                   raw.Domain,
				"Available":                raw.Available,
				"ErrorNo":                  raw.ErrorNo,
				"Description":              raw.Description,
				"IsPremiumName":            raw.IsPremiumName,
				"PremiumRegistrationPrice": raw.PremiumRegistrationPrice,
				"IcannFee":                 raw.IcannFee,
				"EapFee":                   raw.EapFee,
			}

			var errMsg *string
			if unsupportedByTLDMeta {
				msg := "Unsupported TLD (not API registerable)"
				errMsg = &msg
				available = nil
				totalPrice = nil
			} else if strings.TrimSpace(raw.ErrorNo) != "" && strings.TrimSpace(raw.ErrorNo) != "0" {
				msg := strings.TrimSpace(raw.Description)
				if msg == "" {
					msg = "Domain check failed"
				}
				errMsg = &msg
				totalPrice = nil
			}

			item := model.ProbeItem{
				Domain:               d,
				Tld:                  tld,
				Available:            available,
				IsPremium:            isPremium,
				Currency:             currency,
				BaseRegisterPrice:    basePrice,
				PremiumRegisterPrice: premiumReg,
				IcannFee:             icannFee,
				EapFee:               eapFee,
				TotalPrice:           totalPrice,
				Error:                errMsg,
				Raw:                  rawMap,
			}

			yielded = append(yielded, item)
			done++

			// Update task progress
			s.mu.Lock()
			task.Completed = done
			task.UpdatedAt = time.Now()
			s.mu.Unlock()

			event := model.ProgressEvent{
				Type:  "progress",
				Index: done,
				Total: total,
				Data:  item,
			}
			appendJsonl(event)
			if progressChan != nil {
				progressChan <- event
			}
		}
	}

	// Sort results by price
	sort.Slice(yielded, func(i, j int) bool {
		ii, ji := yielded[i], yielded[j]
		// Items with no price go to the end
		if ii.TotalPrice == nil && ji.TotalPrice == nil {
			return ii.Domain < ji.Domain
		}
		if ii.TotalPrice == nil {
			return false
		}
		if ji.TotalPrice == nil {
			return true
		}
		if *ii.TotalPrice != *ji.TotalPrice {
			return *ii.TotalPrice < *ji.TotalPrice
		}
		return ii.Domain < ji.Domain
	})

	csvPath := filepath.Join(taskOutdir, "results.csv")
	mdPath := filepath.Join(taskOutdir, "report.md")

	s.writeCsv(csvPath, yielded)
	s.writeMd(mdPath, task.Word, yielded)

	elapsed := time.Since(startedAt).Seconds()
	finishedAt := time.Now()

	// Update task with completion info
	s.mu.Lock()
	task.Status = model.ProbeStatusCompleted
	task.Completed = done
	task.Outdir = taskOutdir
	task.ReportMd = mdPath
	task.ResultsCsv = csvPath
	task.ResultsJsonl = jsonlPath
	task.ElapsedSeconds = math.Round(elapsed*1000) / 1000
	task.FinishedAt = &finishedAt
	task.UpdatedAt = time.Now()
	s.results[task.ID] = yielded
	s.mu.Unlock()

	summary := model.SummaryEvent{
		Type:           "summary",
		Word:           task.Word,
		Total:          total,
		Outdir:         taskOutdir,
		ReportMd:       mdPath,
		ResultsCsv:     csvPath,
		ResultsJsonl:   jsonlPath,
		FinishedAt:     finishedAt.Format("2006-01-02 15:04:05"),
		ElapsedSeconds: math.Round(elapsed*1000) / 1000,
	}
	appendJsonl(summary)
	if progressChan != nil {
		progressChan <- summary
	}

	return nil
}

// markTaskFailed marks a task as failed
func (s *ProbeService) markTaskFailed(taskID string, errMsg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if task, ok := s.tasks[taskID]; ok {
		task.Status = model.ProbeStatusFailed
		task.Error = errMsg
		task.UpdatedAt = time.Now()
		now := time.Now()
		task.FinishedAt = &now
	}
}

// loadOrRefreshTlds loads TLD list from cache or fetches it
func (s *ProbeService) loadOrRefreshTlds(ctx context.Context, path string, ttlHours float64, fetcher func() ([]map[string]string, error)) ([]map[string]string, error) {
	info, err := os.Stat(path)
	if err == nil {
		ageHours := time.Since(info.ModTime()).Hours()
		if ageHours <= ttlHours {
			data, err := os.ReadFile(path)
			if err == nil {
				var result []map[string]string
				if err := json.Unmarshal(data, &result); err == nil {
					return result, nil
				}
			}
		}
	}

	data, err := fetcher()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		return nil, err
	}

	return data, nil
}

// loadOrRefreshPricing loads pricing data from cache or fetches it
func (s *ProbeService) loadOrRefreshPricing(ctx context.Context, path string, ttlHours float64, fetcher func() (map[string]map[string]string, error)) (map[string]map[string]string, error) {
	info, err := os.Stat(path)
	if err == nil {
		ageHours := time.Since(info.ModTime()).Hours()
		if ageHours <= ttlHours {
			data, err := os.ReadFile(path)
			if err == nil {
				var result map[string]map[string]string
				if err := json.Unmarshal(data, &result); err == nil {
					return result, nil
				}
			}
		}
	}

	data, err := fetcher()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		return nil, err
	}

	return data, nil
}

// isValidTld checks if a TLD is valid
func (s *ProbeService) isValidTld(name string, pricing map[string]map[string]string) bool {
	if name == "" {
		return false
	}
	// Exclude numeric TLDs
	if _, err := strconv.Atoi(name); err == nil {
		return false
	}
	// Exclude TLDs starting with digit
	if len(name) > 0 && name[0] >= '0' && name[0] <= '9' {
		return false
	}
	// Must have pricing info
	if _, ok := pricing[name]; !ok {
		return false
	}
	return true
}

// isMainstreamTld checks if a TLD is "mainstream" (has ICANN fee)
func (s *ProbeService) isMainstreamTld(name string, pricing map[string]map[string]string) bool {
	priceInfo, ok := pricing[name]
	if !ok {
		return false
	}
	_, hasAdditionalCost := priceInfo["AdditionalCost"]
	return hasAdditionalCost
}

// writeCsv writes results to CSV file
func (s *ProbeService) writeCsv(path string, items []model.ProbeItem) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	w.Write([]string{
		"domain", "tld", "available", "is_premium", "currency",
		"base_register_price", "premium_register_price", "icann_fee",
		"eap_fee", "total_price", "error",
	})

	for _, it := range items {
		available := ""
		if it.Available != nil {
			available = strconv.FormatBool(*it.Available)
		}
		isPremium := ""
		if it.IsPremium != nil {
			isPremium = strconv.FormatBool(*it.IsPremium)
		}
		currency := ""
		if it.Currency != nil {
			currency = *it.Currency
		}
		errStr := ""
		if it.Error != nil {
			errStr = *it.Error
		}

		basePrice := ""
		if it.BaseRegisterPrice != nil {
			basePrice = fmt.Sprintf("%.4f", *it.BaseRegisterPrice)
		}
		premiumPrice := ""
		if it.PremiumRegisterPrice != nil {
			premiumPrice = fmt.Sprintf("%.4f", *it.PremiumRegisterPrice)
		}
		icannFee := ""
		if it.IcannFee != nil {
			icannFee = fmt.Sprintf("%.4f", *it.IcannFee)
		}
		eapFee := ""
		if it.EapFee != nil {
			eapFee = fmt.Sprintf("%.4f", *it.EapFee)
		}
		totalPrice := ""
		if it.TotalPrice != nil {
			totalPrice = fmt.Sprintf("%.4f", *it.TotalPrice)
		}

		w.Write([]string{
			it.Domain, it.Tld, available, isPremium, currency,
			basePrice, premiumPrice, icannFee, eapFee, totalPrice, errStr,
		})
	}

	w.Flush()
	return nil
}

// writeMd writes results to Markdown file
func (s *ProbeService) writeMd(path string, word string, items []model.ProbeItem) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	var available, availablePriced, unavailable, errored []model.ProbeItem
	for _, x := range items {
		if x.Error != nil {
			errored = append(errored, x)
		} else if x.Available != nil && *x.Available {
			available = append(available, x)
			if x.TotalPrice != nil {
				availablePriced = append(availablePriced, x)
			}
		} else if x.Available != nil && !*x.Available {
			unavailable = append(unavailable, x)
		}
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("# Namecheap 域名检测报告：%s\n", word))
	lines = append(lines, "## 概览\n")
	lines = append(lines, fmt.Sprintf("- 总 TLD 数：%d", len(items)))
	lines = append(lines, fmt.Sprintf("- 可注册（Available=true）：%d", len(available)))
	lines = append(lines, fmt.Sprintf("- 可注册且有价格：%d", len(availablePriced)))
	lines = append(lines, fmt.Sprintf("- 不可用（Available=false）：%d", len(unavailable)))
	lines = append(lines, fmt.Sprintf("- 出错：%d\n", len(errored)))

	lines = append(lines, "## 可注册域名（按总价升序）\n")
	lines = append(lines, "| 排名 | 域名 | 总价(1年) | 货币 | 基础价 | 溢价注册价 | ICANN | EAP | Premium | 备注 |")
	lines = append(lines, "|---:|---|---:|---|---:|---:|---:|---:|---|---|")

	rank := 0
	for _, it := range items {
		if it.TotalPrice == nil {
			continue
		}
		rank++

		currency := ""
		if it.Currency != nil {
			currency = *it.Currency
		}
		basePrice := ""
		if it.BaseRegisterPrice != nil {
			basePrice = fmt.Sprintf("%.4f", *it.BaseRegisterPrice)
		}
		premiumPrice := ""
		if it.PremiumRegisterPrice != nil {
			premiumPrice = fmt.Sprintf("%.4f", *it.PremiumRegisterPrice)
		}
		icannFee := ""
		if it.IcannFee != nil {
			icannFee = fmt.Sprintf("%.4f", *it.IcannFee)
		}
		eapFee := ""
		if it.EapFee != nil {
			eapFee = fmt.Sprintf("%.4f", *it.EapFee)
		}
		isPremium := ""
		if it.IsPremium != nil {
			if *it.IsPremium {
				isPremium = "yes"
			} else {
				isPremium = "no"
			}
		}
		errStr := ""
		if it.Error != nil {
			errStr = *it.Error
		}

		lines = append(lines, fmt.Sprintf("| %d | %s | %.4f | %s | %s | %s | %s | %s | %s | %s |",
			rank, it.Domain, *it.TotalPrice, currency, basePrice, premiumPrice,
			icannFee, eapFee, isPremium, errStr))
	}

	lines = append(lines, "\n## 不可注册域名（Available=false）\n")
	lines = append(lines, "| 域名 | Premium | 备注 |")
	lines = append(lines, "|---|---|---|")
	for _, it := range items {
		if it.Available == nil || *it.Available {
			continue
		}
		isPremium := ""
		if it.IsPremium != nil {
			if *it.IsPremium {
				isPremium = "yes"
			} else {
				isPremium = "no"
			}
		}
		errStr := ""
		if it.Error != nil {
			errStr = *it.Error
		}
		lines = append(lines, fmt.Sprintf("| %s | %s | %s |", it.Domain, isPremium, errStr))
	}

	if len(errored) > 0 {
		lines = append(lines, "\n## 出错项\n")
		lines = append(lines, "| 域名 | 错误 |")
		lines = append(lines, "|---|---|")
		for _, it := range errored {
			errStr := ""
			if it.Error != nil {
				errStr = *it.Error
			}
			lines = append(lines, fmt.Sprintf("| %s | %s |", it.Domain, errStr))
		}
	}

	return os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644)
}

// Helper functions

// toBool converts a value to bool pointer
func toBool(v interface{}) *bool {
	if v == nil {
		return nil
	}
	if b, ok := v.(bool); ok {
		return &b
	}
	s := strings.TrimSpace(strings.ToLower(fmt.Sprintf("%v", v)))
	if s == "true" || s == "1" || s == "yes" {
		b := true
		return &b
	}
	if s == "false" || s == "0" || s == "no" {
		b := false
		return &b
	}
	return nil
}

// toFloat converts a value to float64 pointer
func toFloat(v interface{}) *float64 {
	if v == nil {
		return nil
	}
	s := strings.TrimSpace(fmt.Sprintf("%v", v))
	if s == "" {
		return nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &f
}

// chooseBasePrice picks the best available register price field.
// Namecheap pricing commonly provides YourPrice/Price/RegularPrice.
func chooseBasePrice(priceInfo map[string]string) *float64 {
	candidates := []string{"YourPrice", "Price", "RegularPrice"}
	for _, k := range candidates {
		if v, ok := priceInfo[k]; ok {
			if p := toFloat(v); p != nil && *p > 0 {
				return p
			}
		}
	}
	return nil
}

// safeAdd safely adds multiple float values
func safeAdd(xs ...*float64) *float64 {
	var acc float64
	anyv := false
	for _, x := range xs {
		if x == nil {
			continue
		}
		anyv = true
		acc += *x
	}
	if !anyv {
		return nil
	}
	return &acc
}

// calcTotalPrice calculates the total price
func calcTotalPrice(available, isPremium *bool, basePrice, premiumPrice, icannFee, eapFee *float64) *float64 {
	if available == nil || !*available {
		return nil
	}
	if isPremium != nil && *isPremium {
		if premiumPrice == nil || *premiumPrice <= 0 {
			return nil
		}
		return safeAdd(premiumPrice, icannFee, eapFee)
	}
	if basePrice == nil || *basePrice <= 0 {
		return nil
	}
	return safeAdd(basePrice, icannFee, eapFee)
}

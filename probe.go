package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// TldMode represents the TLD filtering mode
type TldMode string

const (
	TldModeAll                  TldMode = "all"
	TldModeApiRegisterableOnly  TldMode = "api-registerable-only"
	TldModeMainstreamOnly       TldMode = "mainstream-only"
)

// ProbeItem represents a single domain probe result
type ProbeItem struct {
	Domain               string                 `json:"domain"`
	Tld                  string                 `json:"tld"`
	Available            *bool                  `json:"available"`
	IsPremium            *bool                  `json:"is_premium"`
	Currency             *string                `json:"currency"`
	BaseRegisterPrice    *float64               `json:"base_register_price"`
	PremiumRegisterPrice *float64               `json:"premium_register_price"`
	IcannFee             *float64               `json:"icann_fee"`
	EapFee               *float64               `json:"eap_fee"`
	TotalPrice           *float64               `json:"total_price"`
	Error                *string                `json:"error"`
	Raw                  map[string]interface{} `json:"raw"`
}

// ProgressEvent represents a progress event
type ProgressEvent struct {
	Type  string    `json:"type"`
	Index int       `json:"index"`
	Total int       `json:"total"`
	Data  ProbeItem `json:"data"`
}

// SummaryEvent represents the final summary
type SummaryEvent struct {
	Type           string  `json:"type"`
	Word           string  `json:"word"`
	Total          int     `json:"total"`
	Outdir         string  `json:"outdir"`
	ReportMd       string  `json:"report_md"`
	ResultsCsv     string  `json:"results_csv"`
	ResultsJsonl   string  `json:"results_jsonl"`
	FinishedAt     string  `json:"finished_at"`
	ElapsedSeconds float64 `json:"elapsed_seconds"`
}

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
		return safeAdd(premiumPrice, icannFee, eapFee)
	}
	return safeAdd(basePrice, icannFee, eapFee)
}

// loadOrRefreshTlds loads TLD list from cache or fetches it
func loadOrRefreshTlds(ctx context.Context, path string, ttlHours float64, fetcher func() ([]map[string]string, error)) ([]map[string]string, error) {
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
func loadOrRefreshPricing(ctx context.Context, path string, ttlHours float64, fetcher func() (map[string]map[string]string, error)) (map[string]map[string]string, error) {
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

// loadOrRefreshCache loads data from cache or fetches it
func loadOrRefreshCache(ctx context.Context, path string, ttlHours float64, fetcher func() (interface{}, error)) (interface{}, error) {
	info, err := os.Stat(path)
	if err == nil {
		ageHours := time.Since(info.ModTime()).Hours()
		if ageHours <= ttlHours {
			data, err := os.ReadFile(path)
			if err == nil {
				var result interface{}
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
func isValidTld(name string, pricing map[string]map[string]string) bool {
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
func isMainstreamTld(name string, pricing map[string]map[string]string) bool {
	priceInfo, ok := pricing[name]
	if !ok {
		return false
	}
	_, hasAdditionalCost := priceInfo["AdditionalCost"]
	return hasAdditionalCost
}

// ProbeWord probes a word against all TLDs
func ProbeWord(ctx context.Context, word string, outdir string, tldMode TldMode, cacheTTLHours, ratePerMin float64, maxBatch int, outputChan chan<- interface{}) error {
	cfg, err := FromEnv()
	if err != nil {
		return err
	}

	client := NewNamecheapClient(cfg, ratePerMin, 30.0)
	defer client.Close()

	if err := os.MkdirAll(outdir, 0755); err != nil {
		return err
	}

	cacheDir := filepath.Join(outdir, ".cache")
	tldCache := filepath.Join(cacheDir, "tlds.json")
	pricingCache := filepath.Join(cacheDir, "pricing_register_1y.json")
	jsonlPath := filepath.Join(outdir, fmt.Sprintf("%s_results.jsonl", word))

	// Create or truncate JSONL file
	if _, err := os.Create(jsonlPath); err != nil {
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

	// Load TLDs
	tlds, err := loadOrRefreshTlds(ctx, tldCache, cacheTTLHours, func() ([]map[string]string, error) {
		return client.GetTldList(ctx)
	})
	if err != nil {
		return err
	}

	// Load pricing
	pricing, err := loadOrRefreshPricing(ctx, pricingCache, cacheTTLHours, func() (map[string]map[string]string, error) {
		return client.GetPricingRegister1y(ctx)
	})
	if err != nil {
		return err
	}

	// Filter TLDs
	var tldNames []string
	seen := make(map[string]bool)
	for _, t := range tlds {
		name := strings.ToLower(strings.TrimSpace(t["Name"]))
		if !isValidTld(name, pricing) {
			continue
		}
		if tldMode == TldModeApiRegisterableOnly {
			if strings.ToLower(strings.TrimSpace(t["IsApiRegisterable"])) != "true" {
				continue
			}
		}
		if tldMode == TldModeMainstreamOnly {
			if !isMainstreamTld(name, pricing) {
				continue
			}
		}
		if !seen[name] {
			seen[name] = true
			tldNames = append(tldNames, name)
		}
	}

	// Sort TLD names for deterministic order
	sort.Strings(tldNames)

	total := len(tldNames)
	var yielded []ProbeItem
	done := 0

	// Process in batches
	for i := 0; i < total; i += maxBatch {
		end := i + maxBatch
		if end > total {
			end = total
		}
		batchTlds := tldNames[i:end]

		batchDomains := make([]string, len(batchTlds))
		for j, tld := range batchTlds {
			batchDomains[j] = fmt.Sprintf("%s.%s", word, tld)
		}

		results, err := client.DomainsCheck(ctx, batchDomains)
		if err != nil {
			// Mark entire batch as error
			for _, tld := range batchTlds {
				d := fmt.Sprintf("%s.%s", word, tld)
				errMsg := err.Error()
				item := ProbeItem{
					Domain: d,
					Tld:    tld,
					Error:  &errMsg,
					Raw:    map[string]interface{}{"error": errMsg},
				}
				if priceInfo, ok := pricing[tld]; ok {
					item.BaseRegisterPrice = toFloat(priceInfo["Price"])
				}

				yielded = append(yielded, item)
				done++

				event := ProgressEvent{
					Type:  "progress",
					Index: done,
					Total: total,
					Data:  item,
				}
				appendJsonl(event)
				outputChan <- event
			}
			continue
		}

		// Map results by domain
		byDomain := make(map[string]DomainCheckResult)
		for _, r := range results {
			byDomain[strings.ToLower(r.Domain)] = r
		}

		for _, tld := range batchTlds {
			d := fmt.Sprintf("%s.%s", word, tld)
			raw, ok := byDomain[strings.ToLower(d)]
			if !ok {
				raw = DomainCheckResult{Domain: d}
			}

			available := toBool(raw.Available)
			isPremium := toBool(raw.IsPremiumName)
			premiumReg := toFloat(raw.PremiumRegistrationPrice)
			icannFee := toFloat(raw.IcannFee)
			eapFee := toFloat(raw.EapFee)

			var currency *string
			var basePrice *float64
			if priceInfo, ok := pricing[tld]; ok {
				c := priceInfo["Currency"]
				currency = &c
				basePrice = toFloat(priceInfo["Price"])
			}

			totalPrice := calcTotalPrice(available, isPremium, basePrice, premiumReg, icannFee, eapFee)

			// Convert raw to map
			rawMap := map[string]interface{}{
				"Domain":                   raw.Domain,
				"Available":                raw.Available,
				"IsPremiumName":            raw.IsPremiumName,
				"PremiumRegistrationPrice": raw.PremiumRegistrationPrice,
				"IcannFee":                 raw.IcannFee,
				"EapFee":                   raw.EapFee,
			}

			item := ProbeItem{
				Domain:               d,
				Tld:                tld,
				Available:          available,
				IsPremium:          isPremium,
				Currency:           currency,
				BaseRegisterPrice:  basePrice,
				PremiumRegisterPrice: premiumReg,
				IcannFee:           icannFee,
				EapFee:             eapFee,
				TotalPrice:         totalPrice,
				Raw:                rawMap,
			}

			yielded = append(yielded, item)
			done++

			event := ProgressEvent{
				Type:  "progress",
				Index: done,
				Total: total,
				Data:  item,
			}
			appendJsonl(event)
			outputChan <- event
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

	csvPath := filepath.Join(outdir, fmt.Sprintf("%s_results.csv", word))
	mdPath := filepath.Join(outdir, fmt.Sprintf("%s_report.md", word))

	writeCsv(csvPath, yielded)
	writeMd(mdPath, word, yielded)

	elapsed := time.Since(startedAt).Seconds()
	finishedAt := time.Now().Format("2006-01-02 15:04:05")

	summary := SummaryEvent{
		Type:           "summary",
		Word:           word,
		Total:          total,
		Outdir:         outdir,
		ReportMd:       mdPath,
		ResultsCsv:     csvPath,
		ResultsJsonl:   jsonlPath,
		FinishedAt:     finishedAt,
		ElapsedSeconds: math.Round(elapsed*1000) / 1000,
	}
	appendJsonl(summary)
	outputChan <- summary

	return nil
}

// writeCsv writes results to CSV file
func writeCsv(path string, items []ProbeItem) error {
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
func writeMd(path string, word string, items []ProbeItem) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	var available, availablePriced, unavailable, errored []ProbeItem
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

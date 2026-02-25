package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	_ "modernc.org/sqlite"

	"github.com/mylxsw/namecheap-domain-probe/backend/internal/model"
)

func (s *ProbeService) initStore(path string) error {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return err
	}

	if _, err := db.Exec(`
PRAGMA journal_mode=WAL;
PRAGMA synchronous=NORMAL;
CREATE TABLE IF NOT EXISTS probe_tasks (
  id TEXT PRIMARY KEY,
  word TEXT NOT NULL,
  status TEXT NOT NULL,
  tld_mode TEXT NOT NULL,
  rate_per_min REAL NOT NULL,
  max_batch INTEGER NOT NULL,
  cache_ttl_hours REAL NOT NULL,
  total INTEGER NOT NULL,
  completed INTEGER NOT NULL,
  error TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  finished_at TEXT,
  outdir TEXT,
  report_md TEXT,
  results_csv TEXT,
  results_jsonl TEXT,
  elapsed_seconds REAL
);
CREATE TABLE IF NOT EXISTS probe_results (
  task_id TEXT NOT NULL,
  domain TEXT NOT NULL,
  tld TEXT NOT NULL,
  available INTEGER,
  is_premium INTEGER,
  currency TEXT,
  base_register_price REAL,
  premium_register_price REAL,
  icann_fee REAL,
  eap_fee REAL,
  total_price REAL,
  error TEXT,
  raw_json TEXT,
  PRIMARY KEY(task_id, domain)
);
CREATE INDEX IF NOT EXISTS idx_probe_results_task_id ON probe_results(task_id);
`); err != nil {
		_ = db.Close()
		return err
	}

	s.db = db
	return nil
}

func (s *ProbeService) upsertTask(task *model.ProbeTask) error {
	if s.db == nil || task == nil {
		return nil
	}
	_, err := s.db.ExecContext(context.Background(), `
INSERT INTO probe_tasks (
  id, word, status, tld_mode, rate_per_min, max_batch, cache_ttl_hours,
  total, completed, error, created_at, updated_at, finished_at, outdir,
  report_md, results_csv, results_jsonl, elapsed_seconds
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
  word=excluded.word,
  status=excluded.status,
  tld_mode=excluded.tld_mode,
  rate_per_min=excluded.rate_per_min,
  max_batch=excluded.max_batch,
  cache_ttl_hours=excluded.cache_ttl_hours,
  total=excluded.total,
  completed=excluded.completed,
  error=excluded.error,
  created_at=excluded.created_at,
  updated_at=excluded.updated_at,
  finished_at=excluded.finished_at,
  outdir=excluded.outdir,
  report_md=excluded.report_md,
  results_csv=excluded.results_csv,
  results_jsonl=excluded.results_jsonl,
  elapsed_seconds=excluded.elapsed_seconds
`,
		task.ID,
		task.Word,
		string(task.Status),
		string(task.TldMode),
		task.RatePerMin,
		task.MaxBatch,
		task.CacheTTLHours,
		task.Total,
		task.Completed,
		nullableString(task.Error),
		task.CreatedAt.Format(time.RFC3339Nano),
		task.UpdatedAt.Format(time.RFC3339Nano),
		nullableTime(task.FinishedAt),
		nullableString(task.Outdir),
		nullableString(task.ReportMd),
		nullableString(task.ResultsCsv),
		nullableString(task.ResultsJsonl),
		task.ElapsedSeconds,
	)
	return err
}

func (s *ProbeService) loadTask(id string) (*model.ProbeTask, bool, error) {
	if s.db == nil {
		return nil, false, nil
	}

	row := s.db.QueryRowContext(context.Background(), `
SELECT id, word, status, tld_mode, rate_per_min, max_batch, cache_ttl_hours,
       total, completed, error, created_at, updated_at, finished_at, outdir,
       report_md, results_csv, results_jsonl, elapsed_seconds
FROM probe_tasks WHERE id = ?`, id)

	var (
		task                                  model.ProbeTask
		status, tldMode, createdAt, updatedAt string
		finishedAt, errText, outdir           sql.NullString
		reportMD, resultsCSV, resultsJSONL    sql.NullString
	)
	if err := row.Scan(
		&task.ID, &task.Word, &status, &tldMode, &task.RatePerMin, &task.MaxBatch, &task.CacheTTLHours,
		&task.Total, &task.Completed, &errText, &createdAt, &updatedAt, &finishedAt, &outdir,
		&reportMD, &resultsCSV, &resultsJSONL, &task.ElapsedSeconds,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, false, nil
		}
		return nil, false, err
	}

	task.Status = model.ProbeStatus(status)
	task.TldMode = model.TldMode(tldMode)
	task.CreatedAt = parseTimeOrZero(createdAt)
	task.UpdatedAt = parseTimeOrZero(updatedAt)
	if finishedAt.Valid {
		t := parseTimeOrZero(finishedAt.String)
		task.FinishedAt = &t
	}
	if errText.Valid {
		task.Error = errText.String
	}
	if outdir.Valid {
		task.Outdir = outdir.String
	}
	if reportMD.Valid {
		task.ReportMd = reportMD.String
	}
	if resultsCSV.Valid {
		task.ResultsCsv = resultsCSV.String
	}
	if resultsJSONL.Valid {
		task.ResultsJsonl = resultsJSONL.String
	}

	return &task, true, nil
}

func (s *ProbeService) upsertResult(taskID string, item *model.ProbeItem) error {
	if s.db == nil || item == nil {
		return nil
	}
	raw, _ := json.Marshal(item.Raw)
	_, err := s.db.ExecContext(context.Background(), `
INSERT INTO probe_results (
  task_id, domain, tld, available, is_premium, currency,
  base_register_price, premium_register_price, icann_fee, eap_fee, total_price, error, raw_json
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(task_id, domain) DO UPDATE SET
  tld=excluded.tld,
  available=excluded.available,
  is_premium=excluded.is_premium,
  currency=excluded.currency,
  base_register_price=excluded.base_register_price,
  premium_register_price=excluded.premium_register_price,
  icann_fee=excluded.icann_fee,
  eap_fee=excluded.eap_fee,
  total_price=excluded.total_price,
  error=excluded.error,
  raw_json=excluded.raw_json
`,
		taskID,
		item.Domain,
		item.Tld,
		nullableBool(item.Available),
		nullableBool(item.IsPremium),
		nullableStringPtr(item.Currency),
		nullableFloat(item.BaseRegisterPrice),
		nullableFloat(item.PremiumRegisterPrice),
		nullableFloat(item.IcannFee),
		nullableFloat(item.EapFee),
		nullableFloat(item.TotalPrice),
		nullableStringPtr(item.Error),
		string(raw),
	)
	return err
}

func (s *ProbeService) loadResults(taskID string) ([]model.ProbeItem, error) {
	if s.db == nil {
		return nil, nil
	}

	rows, err := s.db.QueryContext(context.Background(), `
SELECT domain, tld, available, is_premium, currency, base_register_price,
       premium_register_price, icann_fee, eap_fee, total_price, error, raw_json
FROM probe_results
WHERE task_id = ?
ORDER BY domain ASC`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []model.ProbeItem
	for rows.Next() {
		var (
			item                                                  model.ProbeItem
			available, isPremium                                  sql.NullInt64
			currency, errText, rawJSON                            sql.NullString
			basePrice, premiumPrice, icannFee, eapFee, totalPrice sql.NullFloat64
		)
		if err := rows.Scan(
			&item.Domain, &item.Tld, &available, &isPremium, &currency,
			&basePrice, &premiumPrice, &icannFee, &eapFee, &totalPrice, &errText, &rawJSON,
		); err != nil {
			return nil, err
		}

		item.Available = nullIntToBoolPtr(available)
		item.IsPremium = nullIntToBoolPtr(isPremium)
		item.Currency = nullStringPtr(currency)
		item.BaseRegisterPrice = nullFloatPtr(basePrice)
		item.PremiumRegisterPrice = nullFloatPtr(premiumPrice)
		item.IcannFee = nullFloatPtr(icannFee)
		item.EapFee = nullFloatPtr(eapFee)
		item.TotalPrice = nullFloatPtr(totalPrice)
		item.Error = nullStringPtr(errText)
		if rawJSON.Valid && rawJSON.String != "" {
			_ = json.Unmarshal([]byte(rawJSON.String), &item.Raw)
		}

		results = append(results, item)
	}

	return results, rows.Err()
}

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func nullableStringPtr(s *string) interface{} {
	if s == nil || *s == "" {
		return nil
	}
	return *s
}

func nullableTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return t.Format(time.RFC3339Nano)
}

func nullableFloat(f *float64) interface{} {
	if f == nil {
		return nil
	}
	return *f
}

func nullableBool(b *bool) interface{} {
	if b == nil {
		return nil
	}
	if *b {
		return int64(1)
	}
	return int64(0)
}

func parseTimeOrZero(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		// Keep compatibility for old format.
		t2, err2 := time.Parse("2006-01-02 15:04:05", s)
		if err2 != nil {
			return time.Time{}
		}
		return t2
	}
	return t
}

func nullStringPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	v := ns.String
	return &v
}

func nullFloatPtr(nf sql.NullFloat64) *float64 {
	if !nf.Valid {
		return nil
	}
	v := nf.Float64
	return &v
}

func nullIntToBoolPtr(ni sql.NullInt64) *bool {
	if !ni.Valid {
		return nil
	}
	b := ni.Int64 != 0
	return &b
}

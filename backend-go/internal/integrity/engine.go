// Package integrity 🩺 فحص سلامة البيانات (دازاي 11-02 — شبكة الأمان ٢، تعديلات بند ٨).
//
// النظام صار بي بيانات حقيقية قبل لا يكمل، فلازم نعرف حالها قبل أي ترتيب:
// كل فحص يعد صفوف تكسر قاعدة يفترضها الكود (يتيمة، حالة مستحيلة، مبالغ ما
// تطابق، مكرر…). **قراءة بس**: كل فحص بمعاملة READ ONLY ومهلة، ويترجع
// (ROLLBACK) — ما يغيّر ولا صف. النتائج تنحفظ يومياً حتى نشوف إذا خلل يكبر.
package integrity

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// Check فحص واحد. Count يرجّع عمود واحد n، وSample يرجّع id وinfo (حد ٢٠).
type Check struct {
	Key      string
	Domain   string
	Severity string // critical | high | medium | low
	Title    string
	Why      string
	Fix      string
	Count    string
	Sample   string
}

// Result نتيجة فحص بدورة وحدة.
type Result struct {
	Key      string `json:"key"`
	Domain   string `json:"domain"`
	Severity string `json:"severity"`
	Title    string `json:"title"`
	Why      string `json:"why"`
	Fix      string `json:"fix"`
	Count    int64  `json:"count"`
	Error    string `json:"error,omitempty"`
	Ms       int64  `json:"ms"`
}

// SampleRow صف مثال على الخلل.
type SampleRow struct {
	ID   string `db:"id" json:"id"`
	Info string `db:"info" json:"info"`
}

const checkTimeout = 20 * time.Second

func readOnly(db *sqlx.DB, fn func(*sqlx.Tx) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), checkTimeout+5*time.Second)
	defer cancel()
	tx, err := db.BeginTxx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }() // قراءة بس — ما ينحفظ شي أبداً
	if _, err := tx.Exec(`SET LOCAL statement_timeout = '20s'`); err != nil {
		return err
	}
	return fn(tx)
}

// RunAll يشغّل كل الفحوصات. فحص يفشل ما يوقف الباقي — ينكتب خطأه.
func RunAll(db *sqlx.DB) []Result {
	out := make([]Result, 0, len(Checks))
	for _, c := range Checks {
		r := Result{Key: c.Key, Domain: c.Domain, Severity: c.Severity, Title: c.Title, Why: c.Why, Fix: c.Fix}
		start := time.Now()
		err := readOnly(db, func(tx *sqlx.Tx) error { return tx.Get(&r.Count, c.Count) })
		r.Ms = time.Since(start).Milliseconds()
		if err != nil {
			r.Error = shortErr(err)
		}
		out = append(out, r)
	}
	return out
}

// Sample أمثلة لفحص واحد.
func Sample(db *sqlx.DB, key string) ([]SampleRow, bool, error) {
	for _, c := range Checks {
		if c.Key != key {
			continue
		}
		rows := []SampleRow{}
		err := readOnly(db, func(tx *sqlx.Tx) error { return tx.Select(&rows, c.Sample) })
		return rows, true, err
	}
	return nil, false, nil
}

func shortErr(err error) string {
	s := err.Error()
	if i := strings.Index(s, "\n"); i > 0 {
		s = s[:i]
	}
	if len(s) > 200 {
		s = s[:200]
	}
	return s
}

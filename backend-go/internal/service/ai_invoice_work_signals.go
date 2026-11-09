package service

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ═══ ماتركس — مطابقة الفاتورة مع الشغل ═══
//
// نقارن أرقام منظّمة بس (بلا قراءة نص حر):
//
//	١. DEVICE_COUNT: مجموع أعداد بنود التنفيذ بالفاتورة (totalDeviceCount)
//	   مقابل عدد الأجهزة المسجّل بالحجز (Booking.deviceCount — إجباري
//	   للخدمات requiresDeviceInfo). فرق = |الفرق| ≥ ٢ **و** الأكبر ≥ ١.٥×
//	   الأصغر.
//	٢. QUOTED_PRICE: صافي الفاتورة مقابل المبلغ المقدّر بالحجز
//	   (Booking.quotedPrice). فرق = الصافي ≥ ١.٥× المقدّر أو ≤ ٠.٥× منه.
//	   الفاتورة المجانية (isFree) ما تنقارن بالسعر.
//
// 🔴 لا غرامة ولا نقاط ولا لوم: الحكم «يستاهل مراجعة قبل الاعتماد».

const (
	invoiceMismatchMinDiff    = 2
	invoiceMismatchRatio      = 1.5
	invoicePriceHighRatio     = 1.5
	invoicePriceLowRatio      = 0.5
	invoiceFindingDeviceCount = "DEVICE_COUNT"
	invoiceFindingQuotedPrice = "QUOTED_PRICE"
)

// compareInvoiceWork يرجّع الفروقات الواضحة بس.
func compareInvoiceWork(f repository.InvoiceWorkFacts) []model.InvoiceWorkFinding {
	out := []model.InvoiceWorkFinding{}
	if f.BookedDeviceCount != nil && *f.BookedDeviceCount > 0 && f.InvoiceDeviceCount > 0 {
		inv, booked := f.InvoiceDeviceCount, *f.BookedDeviceCount
		lo, hi := inv, booked
		if lo > hi {
			lo, hi = hi, lo
		}
		if hi-lo >= invoiceMismatchMinDiff && float64(hi) >= invoiceMismatchRatio*float64(lo) {
			out = append(out, model.InvoiceWorkFinding{Kind: invoiceFindingDeviceCount, Invoiced: float64(inv), Booked: float64(booked)})
		}
	}
	if !f.IsFree && f.QuotedPrice != nil && *f.QuotedPrice > 0 && f.InvoiceNetTotal > 0 {
		q := *f.QuotedPrice
		if f.InvoiceNetTotal >= invoicePriceHighRatio*q || f.InvoiceNetTotal <= invoicePriceLowRatio*q {
			out = append(out, model.InvoiceWorkFinding{Kind: invoiceFindingQuotedPrice, Invoiced: f.InvoiceNetTotal, Booked: q})
		}
	}
	return out
}

type invoiceWorkAiLookup interface {
	InvoiceWorkFacts(invoiceID string) (*repository.InvoiceWorkFacts, error)
}

// recordInvoiceWorkMismatch يسجّل إشارة لو اكو فرق واضح. أي فشل ينطبع
// باللوق وبس — ما يوقف إنشاء الفاتورة.
func recordInvoiceWorkMismatch(ai AiSignalRecorder, invoiceID, leaderID string) {
	if ai == nil {
		return
	}
	lk, ok := ai.(invoiceWorkAiLookup)
	if !ok {
		return
	}
	f, err := lk.InvoiceWorkFacts(invoiceID)
	if err != nil {
		log.Printf("[ai] تعذر قراءة أرقام الفاتورة %s: %v", invoiceID, err)
		return
	}
	if f == nil || len(compareInvoiceWork(*f)) == 0 {
		return
	}
	emp := leaderID
	if _, err := ai.RecordSignal(model.AiSignal{
		Kind:       model.AiSignalInvoiceWorkMismatch,
		EntityType: "LEADER_INVOICE",
		EntityID:   invoiceID,
		EmployeeID: &emp,
	}); err != nil {
		log.Printf("[ai] تعذر تسجيل إشارة مطابقة الفاتورة %s: %v", invoiceID, err)
	}
}

// CollectForInvoiceWorkMismatch — EntityID = معرّف فاتورة الليدر.
func (s *AiEvidenceService) CollectForInvoiceWorkMismatch(signal model.AiSignal) (*model.AiEvidence, error) {
	f, err := s.db.InvoiceWorkFacts(signal.EntityID)
	if err != nil {
		return nil, fmt.Errorf("فاتورة الليدر مو مقروءة: %w", err)
	}
	gaps := []string{}
	ev := model.InvoiceWorkMismatchEvidence{InvoiceID: signal.EntityID, Findings: []model.InvoiceWorkFinding{}}
	if f == nil {
		gaps = append(gaps, "الفاتورة مو مربوطة بحجز وقت الجمع")
	} else {
		ev.AccountingCode = f.AccountingCode
		ev.BookingCode = f.BookingCode
		ev.InvoiceDeviceCount = f.InvoiceDeviceCount
		ev.BookedDeviceCount = f.BookedDeviceCount
		ev.InvoiceNetTotal = f.InvoiceNetTotal
		ev.QuotedPrice = f.QuotedPrice
		ev.Findings = compareInvoiceWork(*f)
		if len(ev.Findings) == 0 {
			gaps = append(gaps, "وقت الجمع ما بقى فرق واضح (ممكن انعدّلت الفاتورة أو الحجز)")
		}
	}
	facts, _ := json.Marshal(ev)
	gapsJSON, _ := json.Marshal(gaps)
	return s.db.SaveEvidence(signal.ID, facts, gapsJSON)
}

// judgeInvoiceWorkMismatch — WATCH لفرق سعر لحاله، WARN لو عدد الأجهزة
// ما يطابق (رقم ملموس) أو الفرقين سوة.
func (RulesJudge) judgeInvoiceWorkMismatch(sig model.AiSignal, ev model.InvoiceWorkMismatchEvidence) (*model.AiVerdict, error) {
	v := &model.AiVerdict{
		Source:     model.AiSourceRules,
		Headline:   "فاتورة ما تطابق بيانات الحجز — يستاهل مراجعة قبل الاعتماد",
		Severity:   model.AiSeverityWatch,
		Confidence: 50,
	}
	parts := []string{}
	hasDevice := false
	for _, f := range ev.Findings {
		switch f.Kind {
		case invoiceFindingDeviceCount:
			hasDevice = true
			parts = append(parts, fmt.Sprintf("الفاتورة تحسب %.0f جهاز والحجز مسجّل بيه %.0f", f.Invoiced, f.Booked))
		case invoiceFindingQuotedPrice:
			parts = append(parts, fmt.Sprintf("صافي الفاتورة %.0f مقابل مبلغ مقدّر بالحجز %.0f", f.Invoiced, f.Booked))
		}
	}
	if hasDevice || len(ev.Findings) >= 2 {
		v.Severity = model.AiSeverityWarn
		v.Confidence = 65
	}
	reason := "ماكو فرق واضح وقت الجمع."
	if len(parts) > 0 {
		reason = strings.Join(parts, "؛ ") + "."
	}
	if ev.AccountingCode != "" {
		reason += " الفاتورة: " + ev.AccountingCode + "."
	}
	if ev.BookingCode != "" {
		reason += " الحجز: " + ev.BookingCode + "."
	}
	reason += " ممكن الزبون زاد أو نقّص الشغل بالموقع — سبب مشروع."
	suggestion := "يستاهل مراجعة قبل الاعتماد: طابق العدد والمبلغ ويا الليدر والحجز."
	v.Reasoning = &reason
	v.Suggestion = &suggestion
	return v, nil
}

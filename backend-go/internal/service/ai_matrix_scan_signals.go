package service

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ═══ ماتركس — زبون قرب يزعل، تسعير شاذ، الورق المتأخر، الحضور مقابل الشغل ═══
//
// 🔴 نفس القواعد: لا غرامة ولا نقاط ولا خصم تلقائي. الأدلة أكواد وأرقام
// وتواريخ بس. أي فشل ينطبع باللوق وبس. بيانات ناقصة = صمت.
//
// منع التكرار: الفحوصات الدورية تحط occurredAt على بداية أسبوع ISO
// (الزبون والورق) أو بداية اليوم (الحضور)، والفهرس الفريد
// (kind, entityType, entityId, occurredAt) يرفض نفس الحالة مرة ثانية.
// وفوقها ClaimDailyMarker حتى الفحص نفسه يشتغل مرة باليوم.

const (
	riskMinFactors        = 2
	riskPostponeMin       = 2
	riskLowRatingMax      = 2
	priceOutlierMinSample = 8
	priceOutlierHigh      = 2.0
	priceOutlierLow       = 0.4
	paperworkWarnCount    = 3
	attendanceGapMinutes  = 360
	matrixScanHour        = 8 // ٨ الصبح بغداد

	riskFactorPostponed     = "POSTPONED"
	riskFactorOverdue       = "OVERDUE"
	riskFactorOpenComplaint = "OPEN_COMPLAINT"
	riskFactorLowRating     = "LOW_RATING"

	gapPresentNoWork    = "PRESENT_NO_WORK"
	gapWorkNoAttendance = "WORK_NO_ATTENDANCE"
)

// ── دوال صافية ──

// customerRiskFactors يحوّل الأرقام الخام لعوامل خطر فعلية.
func customerRiskFactors(r repository.CustomerRiskRow) []model.CustomerRiskFactor {
	out := []model.CustomerRiskFactor{}
	if r.MaxPostpone >= riskPostponeMin {
		out = append(out, model.CustomerRiskFactor{Kind: riskFactorPostponed, Value: r.MaxPostpone, Ref: r.PostponeCode})
	}
	if r.OverdueDays >= 1 && r.OverdueCode != "" {
		out = append(out, model.CustomerRiskFactor{Kind: riskFactorOverdue, Value: r.OverdueDays, Ref: r.OverdueCode})
	}
	if r.OpenComplaints > 0 {
		out = append(out, model.CustomerRiskFactor{Kind: riskFactorOpenComplaint, Value: r.OpenComplaints})
	}
	if r.LowRatings > 0 && r.MinRating > 0 && r.MinRating <= riskLowRatingMax {
		out = append(out, model.CustomerRiskFactor{Kind: riskFactorLowRating, Value: r.MinRating})
	}
	return out
}

// classifyPriceOutlier: HIGH لو فوق ٢× الوسيط، LOW لو تحت ٠.٤×، وإلا "".
// أقل من ٨ عيّنات أو فاتورة مجانية/صفرية = "" (صمت).
func classifyPriceOutlier(net, median float64, samples int, isFree bool) string {
	if isFree || net <= 0 || median <= 0 || samples < priceOutlierMinSample {
		return ""
	}
	if net > priceOutlierHigh*median {
		return "HIGH"
	}
	if net < priceOutlierLow*median {
		return "LOW"
	}
	return ""
}

// groupLatePaperwork يجمّع الحجوزات لكل ليدر.
func groupLatePaperwork(rows []repository.LatePaperworkRow) map[string][]model.LatePaperworkItem {
	out := map[string][]model.LatePaperworkItem{}
	for _, r := range rows {
		if r.LeaderID == "" {
			continue
		}
		out[r.LeaderID] = append(out[r.LeaderID], r.LatePaperworkItem)
	}
	return out
}

// baghdadMidnight يرجّع منتصف ليل يوم بغداد (YYYY-MM-DD) كلحظة زمنية.
func baghdadMidnight(day string) time.Time {
	t, err := time.ParseInLocation("2006-01-02", day, debriefLoc)
	if err != nil {
		return time.Time{}
	}
	return t
}

func marshalPayload(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return []byte(`{}`)
	}
	return b
}

// ── الخدمة الدورية ──

type matrixScanRepo interface {
	ClaimDailyMarker(metricKey, day string) (bool, error)
	RecordSignal(model.AiSignal) (*model.AiSignal, error)
	CustomerRiskRows() ([]repository.CustomerRiskRow, error)
	LatePaperworkRows(leaderID string) ([]repository.LatePaperworkRow, error)
	AttendanceWorkGaps(day string, minMinutes int, employeeID string) ([]repository.AttendanceGapRow, error)
}

// MatrixScanService الفحوصات اليومية الثلاث — تسجّل إشارات بس، والبرين
// يجمع الأدلة ويحكم بمساره العادي.
type MatrixScanService struct {
	repo matrixScanRepo
}

func NewMatrixScanService(repo *repository.AiRepository) *MatrixScanService {
	return &MatrixScanService{repo: repo}
}

// RunDailyIfDue مرة باليوم بعد ٨ الصبح بغداد. كل فحص مستقل: فشل واحد
// ما يوقف الباقي.
func (s *MatrixScanService) RunDailyIfDue() error {
	now := time.Now().In(debriefLoc)
	if now.Hour() < matrixScanHour {
		return nil
	}
	today := now.Format("2006-01-02")
	claimed, err := s.repo.ClaimDailyMarker("DAILY_MATRIX_SCANS", today)
	if err != nil || !claimed {
		return err
	}
	week := baghdadMidnight(isoWeekMonday(now))
	if err := s.scanCustomersAtRisk(week); err != nil {
		log.Printf("[ai] فحص الزبائن القريبين يزعلون: %v", err)
	}
	if err := s.scanLatePaperwork(week); err != nil {
		log.Printf("[ai] فحص الورق المتأخر: %v", err)
	}
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")
	if err := s.scanAttendanceWorkGap(yesterday); err != nil {
		log.Printf("[ai] فحص الحضور مقابل الشغل: %v", err)
	}
	return nil
}

func (s *MatrixScanService) record(sig model.AiSignal) {
	if _, err := s.repo.RecordSignal(sig); err != nil {
		log.Printf("[ai] تعذر تسجيل إشارة %s للكيان %s: %v", sig.Kind, sig.EntityID, err)
	}
}

func (s *MatrixScanService) scanCustomersAtRisk(week time.Time) error {
	rows, err := s.repo.CustomerRiskRows()
	if err != nil {
		return err
	}
	for _, r := range rows {
		f := customerRiskFactors(r)
		if len(f) < riskMinFactors {
			continue
		}
		var emp *string
		if r.LeaderID.Valid && r.LeaderID.String != "" {
			id := r.LeaderID.String
			emp = &id
		}
		s.record(model.AiSignal{
			Kind: model.AiSignalCustomerAtRisk, EntityType: "CUSTOMER", EntityID: r.CustomerID,
			EmployeeID: emp, OccurredAt: week,
		})
	}
	return nil
}

func (s *MatrixScanService) scanLatePaperwork(week time.Time) error {
	rows, err := s.repo.LatePaperworkRows("")
	if err != nil {
		return err
	}
	weekKey := week.Format("2006-01-02")
	for leader, items := range groupLatePaperwork(rows) {
		id := leader
		s.record(model.AiSignal{
			Kind: model.AiSignalLatePaperwork, EntityType: "EMPLOYEE", EntityID: leader,
			EmployeeID: &id, OccurredAt: week,
			Payload: marshalPayload(model.LatePaperworkEvidence{WeekStart: weekKey, LateCount: len(items), Bookings: items}),
		})
	}
	return nil
}

func (s *MatrixScanService) scanAttendanceWorkGap(day string) error {
	at := baghdadMidnight(day)
	if at.IsZero() {
		return fmt.Errorf("يوم غلط %q", day)
	}
	rows, err := s.repo.AttendanceWorkGaps(day, attendanceGapMinutes, "")
	if err != nil {
		return err
	}
	for _, r := range rows {
		id := r.EmployeeID
		s.record(model.AiSignal{
			Kind: model.AiSignalAttendanceWorkGap, EntityType: "EMPLOYEE", EntityID: r.EmployeeID,
			EmployeeID: &id, OccurredAt: at,
			Payload: marshalPayload(model.AttendanceWorkGapEvidence{
				Day: day, Kind: r.Kind, AttendedMinutes: r.AttendedMinutes, BookingCodes: []string(r.BookingCodes),
			}),
		})
	}
	return nil
}

// ── تسعير شاذ: لحظة إنشاء الفاتورة ──

type priceOutlierAiLookup interface {
	PriceOutlierBaseline(invoiceID string) (*repository.PriceOutlierFacts, error)
}

// recordPriceOutlier أي فشل ينطبع باللوق وبس — ما يوقف الفاتورة.
func recordPriceOutlier(ai AiSignalRecorder, invoiceID, leaderID string) {
	if ai == nil {
		return
	}
	lk, ok := ai.(priceOutlierAiLookup)
	if !ok {
		return
	}
	f, err := lk.PriceOutlierBaseline(invoiceID)
	if err != nil {
		log.Printf("[ai] تعذر حساب وسيط سعر الفاتورة %s: %v", invoiceID, err)
		return
	}
	if f == nil || classifyPriceOutlier(f.NetTotal, f.Median, f.Samples, f.IsFree) == "" {
		return
	}
	emp := leaderID
	if _, err := ai.RecordSignal(model.AiSignal{
		Kind: model.AiSignalPriceOutlier, EntityType: "LEADER_INVOICE", EntityID: invoiceID, EmployeeID: &emp,
	}); err != nil {
		log.Printf("[ai] تعذر تسجيل إشارة تسعير شاذ %s: %v", invoiceID, err)
	}
}

// ── جامعو الأدلة ──

func (s *AiEvidenceService) saveFacts(signalID string, ev any, gaps []string) (*model.AiEvidence, error) {
	facts, _ := json.Marshal(ev)
	gapsJSON, _ := json.Marshal(gaps)
	return s.db.SaveEvidence(signalID, facts, gapsJSON)
}

// CollectForCustomerAtRisk — EntityID = معرّف الزبون. يعيد الحساب وقت الجمع.
func (s *AiEvidenceService) CollectForCustomerAtRisk(signal model.AiSignal) (*model.AiEvidence, error) {
	row, err := s.db.CustomerRiskRowFor(signal.EntityID)
	if err != nil {
		return nil, fmt.Errorf("أرقام الزبون مو مقروءة: %w", err)
	}
	ev := model.CustomerAtRiskEvidence{Factors: []model.CustomerRiskFactor{}}
	gaps := []string{}
	if row == nil {
		gaps = append(gaps, "وقت الجمع الزبون ما عنده حجز فعّال أو منجز حديثاً")
	} else {
		ev.CustomerCode = model.Customer{CustomerCode: row.CustomerCode}.FormatCode()
		ev.LatestBookingCode = row.LatestBookingCode
		ev.Factors = customerRiskFactors(*row)
		if len(ev.Factors) < riskMinFactors {
			gaps = append(gaps, "وقت الجمع العوامل نزلت تحت اثنين (ممكن انحلّت)")
		}
	}
	return s.saveFacts(signal.ID, ev, gaps)
}

// CollectForPriceOutlier — EntityID = معرّف فاتورة الليدر.
func (s *AiEvidenceService) CollectForPriceOutlier(signal model.AiSignal) (*model.AiEvidence, error) {
	f, err := s.db.PriceOutlierBaseline(signal.EntityID)
	if err != nil {
		return nil, fmt.Errorf("فاتورة الليدر مو مقروءة: %w", err)
	}
	ev := model.PriceOutlierEvidence{InvoiceID: signal.EntityID}
	gaps := []string{}
	if f == nil {
		gaps = append(gaps, "الفاتورة مو مربوطة بحجز له خدمة وقت الجمع")
	} else {
		ev.AccountingCode, ev.BookingCode, ev.ServiceID = f.AccountingCode, f.BookingCode, f.ServiceID
		ev.NetTotal, ev.Median, ev.Samples = f.NetTotal, f.Median, f.Samples
		if f.Median > 0 {
			ev.Ratio = float64(int(f.NetTotal/f.Median*100+0.5)) / 100
		}
		ev.Direction = classifyPriceOutlier(f.NetTotal, f.Median, f.Samples, f.IsFree)
		if ev.Direction == "" {
			gaps = append(gaps, "وقت الجمع الصافي رجع ضمن المعتاد (ممكن انعدّلت الفاتورة)")
		}
	}
	return s.saveFacts(signal.ID, ev, gaps)
}

// CollectForLatePaperwork — الأدلة لقطة وقت الفحص (Payload)، ونعلّم
// الحجوزات الي انكمل ورقها بعدين كفجوة.
func (s *AiEvidenceService) CollectForLatePaperwork(signal model.AiSignal) (*model.AiEvidence, error) {
	var ev model.LatePaperworkEvidence
	gaps := []string{}
	if err := json.Unmarshal(signal.Payload, &ev); err != nil || ev.LateCount == 0 {
		gaps = append(gaps, "لقطة الفحص مفقودة")
	}
	if ev.Bookings == nil {
		ev.Bookings = []model.LatePaperworkItem{}
	}
	if len(ev.Bookings) > 0 {
		if rows, err := s.db.LatePaperworkRows(signal.EntityID); err == nil {
			still := map[string]bool{}
			for _, r := range rows {
				still[r.BookingCode] = true
			}
			done := 0
			for _, b := range ev.Bookings {
				if !still[b.BookingCode] {
					done++
				}
			}
			if done > 0 {
				gaps = append(gaps, fmt.Sprintf("%d حجز انكمل ورقه بعد الفحص", done))
			}
		}
	}
	return s.saveFacts(signal.ID, ev, gaps)
}

// CollectForAttendanceWorkGap — لقطة وقت الفحص (Payload).
func (s *AiEvidenceService) CollectForAttendanceWorkGap(signal model.AiSignal) (*model.AiEvidence, error) {
	var ev model.AttendanceWorkGapEvidence
	gaps := []string{"ممكن نقص تسجيل (بصمة منسية أو تكليف مو مسجّل) — مو دليل تقصير"}
	if err := json.Unmarshal(signal.Payload, &ev); err != nil || ev.Kind == "" {
		gaps = append(gaps, "لقطة الفحص مفقودة")
	}
	if ev.BookingCodes == nil {
		ev.BookingCodes = []string{}
	}
	return s.saveFacts(signal.ID, ev, gaps)
}

// ── القواعد ──

// judgeCustomerAtRisk — WARN من عاملين، CRITICAL من ثلاث. بلا لوم.
func (RulesJudge) judgeCustomerAtRisk(sig model.AiSignal, ev model.CustomerAtRiskEvidence) (*model.AiVerdict, error) {
	v := &model.AiVerdict{
		Source:     model.AiSourceRules,
		Headline:   "زبون قرب يزعل — يستاهل مراجعة",
		Severity:   model.AiSeverityWatch,
		Confidence: 50,
	}
	n := len(ev.Factors)
	if n >= riskMinFactors {
		v.Severity, v.Confidence = model.AiSeverityWarn, 65
	}
	if n >= 3 {
		v.Severity, v.Confidence = model.AiSeverityCritical, 75
	}
	parts := []string{}
	for _, f := range ev.Factors {
		switch f.Kind {
		case riskFactorPostponed:
			parts = append(parts, fmt.Sprintf("حجزه %s انأجّل %d مرات", f.Ref, f.Value))
		case riskFactorOverdue:
			parts = append(parts, fmt.Sprintf("حجزه %s متأخر %d يوم عن موعده", f.Ref, f.Value))
		case riskFactorOpenComplaint:
			parts = append(parts, fmt.Sprintf("%d شكوى مفتوحة من أكثر من ٣ أيام", f.Value))
		case riskFactorLowRating:
			parts = append(parts, fmt.Sprintf("تقييمه نزل لـ%d من ٥ خلال ٦٠ يوم", f.Value))
		}
	}
	reason := fmt.Sprintf("%d عوامل خطر سوية: %s.", n, strings.Join(parts, "؛ "))
	if ev.CustomerCode != "" {
		reason += " الزبون: " + ev.CustomerCode + "."
	}
	reason += " الأسباب ممكن تكون من الزبون نفسه أو من المواد — مو حكماً على أحد."
	suggestion := "يستاهل مراجعة: مهندس الجودة يتصل بالزبون ويطمّنه قبل ما يزعل."
	v.Reasoning, v.Suggestion = &reason, &suggestion
	return v, nil
}

// judgePriceOutlier — WATCH دائماً.
func (RulesJudge) judgePriceOutlier(sig model.AiSignal, ev model.PriceOutlierEvidence) (*model.AiVerdict, error) {
	v := &model.AiVerdict{
		Source:     model.AiSourceRules,
		Headline:   "تسعير شاذ عن المعتاد لنفس الخدمة — يستاهل مراجعة",
		Severity:   model.AiSeverityWatch,
		Confidence: 50,
	}
	word := "أعلى"
	if ev.Direction == "LOW" {
		word = "أقل"
	}
	reason := fmt.Sprintf("صافي الفاتورة %.0f %s بوضوح من وسيط %.0f لفواتير نفس الخدمة (%d عيّنة، النسبة %.2f×).",
		ev.NetTotal, word, ev.Median, ev.Samples, ev.Ratio)
	if ev.AccountingCode != "" {
		reason += " الفاتورة: " + ev.AccountingCode + "."
	}
	if ev.BookingCode != "" {
		reason += " الحجز: " + ev.BookingCode + "."
	}
	reason += " ممكن حجم الشغل مختلف فعلاً — سبب مشروع."
	suggestion := "يستاهل مراجعة: قارن المبلغ بحجم الشغل وبنود الفاتورة قبل الاعتماد."
	v.Reasoning, v.Suggestion = &reason, &suggestion
	return v, nil
}

// judgeLatePaperwork — WATCH لحجز أو اثنين، WARN من ثلاثة. تذكير مو عقوبة.
func (RulesJudge) judgeLatePaperwork(sig model.AiSignal, ev model.LatePaperworkEvidence) (*model.AiVerdict, error) {
	v := &model.AiVerdict{
		Source:     model.AiSourceRules,
		Headline:   "ورق حجوزات منجزة متأخر — تذكير",
		Severity:   model.AiSeverityWatch,
		Confidence: 60,
	}
	if ev.LateCount >= paperworkWarnCount {
		v.Severity, v.Confidence = model.AiSeverityWarn, 70
	}
	noInv, noRep := 0, 0
	codes := []string{}
	for _, b := range ev.Bookings {
		if b.MissingInvoice {
			noInv++
		}
		if b.MissingReport {
			noRep++
		}
		codes = append(codes, b.BookingCode)
	}
	reason := fmt.Sprintf("%d حجز منجز من أكثر من ٤٨ ساعة ورقه ناقص (%d بلا فاتورة، %d بلا تقرير): %s.",
		ev.LateCount, noInv, noRep, strings.Join(codes, "، "))
	reason += " ممكن الورق عالق بسبب مو من الليدر (مواد، موافقة، نظام)."
	suggestion := "تذكير لطيف للليدر يكمّل الفاتورة والتقرير — مو عقوبة."
	v.Reasoning, v.Suggestion = &reason, &suggestion
	return v, nil
}

// judgeAttendanceWorkGap — INFO، وWATCH لو شغل بلا حضور (تأثيره على الراتب).
func (RulesJudge) judgeAttendanceWorkGap(sig model.AiSignal, ev model.AttendanceWorkGapEvidence) (*model.AiVerdict, error) {
	v := &model.AiVerdict{
		Source:     model.AiSourceRules,
		Severity:   model.AiSeverityInfo,
		Confidence: 40,
	}
	var reason string
	switch ev.Kind {
	case gapWorkNoAttendance:
		v.Severity = model.AiSeverityWatch
		v.Headline = "شغل مسجّل بلا حضور بنفس اليوم — ممكن نقص تسجيل"
		reason = fmt.Sprintf("يوم %s: الموظف مسجّل على حجز بدا أو خلص (%s) وماكو ولا جلسة حضور.",
			ev.Day, strings.Join(ev.BookingCodes, "، "))
	default:
		v.Headline = "حضور بلا شغل مسجّل بنفس اليوم — ممكن نقص تسجيل"
		reason = fmt.Sprintf("يوم %s: حضور %d دقيقة وماكو ولا حجز أو مهمة باسمه.", ev.Day, ev.AttendedMinutes)
	}
	reason += " هذا ممكن يكون نقص تسجيل (بصمة منسية أو تكليف مو مسجّل أو شغل داخلي) — مو اتهام."
	suggestion := "يستاهل مراجعة خفيفة: تأكد من التسجيل ويا الموظف أو المنسّق."
	v.Reasoning, v.Suggestion = &reason, &suggestion
	return v, nil
}

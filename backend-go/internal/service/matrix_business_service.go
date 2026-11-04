package service

import (
	"fmt"
	"math"
	"strings"
	"time"

	"staffmange-api/internal/repository"
)

// ═══ أرقام الشركة بعين ماتركس ═══
// إيراد كل شهر، وهالشهر مقابل الماضي، وتوقع نهاية الشهر ويا أساس حسابه،
// والزبون الفعلي مقابل المستفسر. كل رقم من استعلام واحد، والتوقع يبين
// منين جه — «رقم بلا أساس أسوأ من ماكو رقم».

type BusinessForecast struct {
	Expected     float64 `json:"expected"`
	ExpectedJobs int     `json:"expectedJobs"`
	Basis        string  `json:"basis"`
	Insufficient bool    `json:"insufficient"`
	// المدى المتوقع (من غلط طريقة الوتيرة بالأشهر الماضية).
	Low  float64 `json:"low,omitempty"`
	High float64 `json:"high,omitempty"`
	// Accuracy دقة توقع الشهر الماضي — «توقعنا X بيوم كذا، وطلع Y».
	Accuracy string `json:"accuracy,omitempty"`
	// (ع) 10-15: التعلّم — التصحيح من أغلاط آخر ٣ أشهر، والثقة الحقيقية من نسبة الصح
	RawExpected        float64                      `json:"rawExpected"`
	RawExpectedJobs    int                          `json:"rawExpectedJobs"`
	CalibrationPct     float64                      `json:"calibrationPct"`
	JobsCalibrationPct float64                      `json:"jobsCalibrationPct"`
	Confidence         *int                         `json:"confidence"`
	JobsConfidence     *int                         `json:"jobsConfidence"`
	History            []repository.ForecastOutcome `json:"history"`
}

type BusinessView struct {
	Months    []repository.MonthRevenue    `json:"months"`
	MTD       *repository.MonthToDate      `json:"mtd"`
	Forecast  BusinessForecast             `json:"forecast"`
	Customers *repository.CustomerSplit    `json:"customers"`
	Inquirers []repository.InquiryCustomer `json:"inquirers"`
	// حجوزات منجزة هالشهر بلا فاتورة — (ع): «من اضغط يوديني وين المشكلة».
	Uninvoiced []repository.UninvoicedItem `json:"uninvoiced"`
}

type MatrixBusinessService struct {
	repo *repository.MatrixBusinessRepository
}

func NewMatrixBusinessService(repo *repository.MatrixBusinessRepository) *MatrixBusinessService {
	return &MatrixBusinessService{repo: repo}
}

func (s *MatrixBusinessService) View() (*BusinessView, error) {
	months, err := s.repo.Monthly(6)
	if err != nil {
		return nil, err
	}
	mtd, err := s.repo.MTD()
	if err != nil {
		return nil, err
	}
	split, err := s.repo.Split()
	if err != nil {
		return nil, err
	}
	inq, err := s.repo.RepeatInquirers(30)
	if err != nil {
		return nil, err
	}
	v := &BusinessView{Months: months, MTD: mtd, Customers: split, Inquirers: inq, Uninvoiced: []repository.UninvoicedItem{}}
	if un, err := s.repo.Uninvoiced(50); err == nil {
		v.Uninvoiced = un
	}
	v.Forecast = s.forecast(time.Now(), mtd, months)
	// سجل الأحكام والثقة يطلعن حتى لو البيانات ما تكفي لتوقع هالشهر
	if v.Forecast.History == nil {
		hist, _ := s.repo.Outcomes(24)
		v.Forecast.History = hist
		v.Forecast.Confidence, v.Forecast.JobsConfidence = realConfidence(hist, "REVENUE"), realConfidence(hist, "JOBS")
	}
	return v, nil
}

// forecast توقع نهاية الشهر من طريقتين، وكل وحدة تعوّض نقص الثانية:
//   - **المجدول**: المنجز + قيمة الحجوزات المفتوحة الي موعدها هالشهر (نسبة إنجاز
//     خدمتها × معدل المقبوض لخدمتها). دقيقة بآخر الشهر، بس بأوله ما تشوف
//     الشغل الي بعده ما انحجز.
//   - **الوتيرة**: المنجز ÷ «شكد چان الإيراد واصل لحد هاليوم» بآخر ٣ أشهر.
//     تشوف الشغل الي راح يجي، بس ما تعرف شنو بالجدول.
//
// الوزن حسب كم مضى من الشهر: بأوله نعتمد الوتيرة، وبآخره المجدول. والمدى من
// غلط الوتيرة لو جربناها على الأشهر الماضية نفسها.
func (s *MatrixBusinessService) forecast(now time.Time, mtd *repository.MonthToDate, months []repository.MonthRevenue) BusinessForecast {
	pl, err := s.repo.Pipeline()
	if err != nil || pl.Samples < 10 || !pl.AvgValue.Valid || !pl.Rate.Valid {
		return BusinessForecast{Insufficient: true, Expected: mtd.Revenue, ExpectedJobs: mtd.Bookings,
			Basis: "البيانات ما تكفي للتوقع بعد (أقل من ١٠ حجوزات منجزة بفلوس بآخر ٩٠ يوم) — الرقم هو المقبوض لحد هسه بس."}
	}
	pipeVal := pl.Value.Float64
	expJobs := int(math.Round(pl.ExpJobs.Float64))
	// (ع) 10-13: التوقع من الحجوزات بس. دفعات المشاريع المقبوضة هالشهر تنضاف
	// كما هي بلا ضرب بوتيرة — إدخال دفعات قديمة بالجملة چان يطلّع التوقع بالمليارات.
	bookRev := mtd.Revenue - mtd.ProjectRevenue
	projRev := mtd.ProjectRevenue
	pipeline := bookRev + pipeVal

	t := now.In(debriefLoc)
	daysIn := time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, debriefLoc).Day()
	elapsed := float64(t.Day()) / float64(daysIn)

	var pace *float64
	var spread float64
	paceNote := ""
	if pm, err := s.repo.Pace(); err == nil {
		var tot, upto float64
		fr := []float64{}
		for _, m := range pm {
			if m.Total > 0 {
				tot += m.Total
				upto += m.Upto
				fr = append(fr, m.Upto/m.Total)
			}
		}
		if len(fr) >= 2 && upto > 0 && bookRev > 0 {
			f := upto / tot
			p := bookRev / f
			pace = &p
			// جرّب الطريقة على كل شهر سابق: لو چنا نتوقع بهاليوم، شكد چان الغلط؟
			for i, fi := range fr {
				others, n := 0.0, 0
				for j, fj := range fr {
					if j != i {
						others += fj
						n++
					}
				}
				if n > 0 && others > 0 {
					if e := math.Abs(fi/(others/float64(n)) - 1); e > spread {
						spread = e
					}
				}
			}
			paceNote = fmt.Sprintf(" الوتيرة: بآخر %d أشهر چان الإيراد لحد يوم %d يوصل %.0f٪ من الشهر، فالمقبوض (%s) يعني ~%s.",
				len(fr), t.Day(), f*100, iqdShort(bookRev), iqdShort(p))
		}
	}

	expected := pipeline
	if pace != nil && *pace > pipeline {
		// الوتيرة تشوف شغل بعده ما انحجز — نعطيها وزن يقل كل ما يمشي الشهر.
		expected = elapsed*pipeline + (1-elapsed)**pace
	}
	bookExpected := expected
	expected += projRev
	projNote := ""
	if projRev > 0 {
		projNote = fmt.Sprintf(" + دفعات مشاريع انسجّلت هالشهر %s (تنضاف كما هي، ما تنحسب بالوتيرة) = %s.", iqdShort(projRev), iqdShort(expected))
	}
	fc := BusinessForecast{
		Expected:     math.Round(expected),
		ExpectedJobs: mtd.Bookings + expJobs,
		Basis: fmt.Sprintf("الحجوزات: المقبوض لحد هسه %s + %d حجز مفتوح موعده هالشهر، متوقع ينجز منها ~%d (نسبة إنجاز كل خدمة × معدل المقبوض بيها بآخر ٩٠ يوم) = %s.%s توقع الحجوزات %s.%s",
			iqdShort(bookRev), pl.Jobs, expJobs, iqdShort(pipeline), paceNote, iqdShort(bookExpected), projNote),
	}
	// المدى: الشغل الي بعده ما صار (المتوقع − المقبوض) بيه غلط بقدر غلط الوتيرة
	// بالأشهر الماضية (أقل شي ١٠٪، أكثر شي ٥٠٪). المقبوض ثابت ما بيه غلط.
	if rest := expected - mtd.Revenue; rest > 0 {
		e := math.Min(math.Max(spread, 0.10), 0.50)
		fc.Low = math.Round(expected - rest*e)
		fc.High = math.Round(expected + rest*e)
	}
	// ═══ التعلّم: نصحّح الجزء الي بعده ما صار حسب ميل ماتركس بأخطائه ═══
	hist, _ := s.repo.Outcomes(24)
	fc.History = hist
	fc.RawExpected, fc.RawExpectedJobs = fc.Expected, fc.ExpectedJobs
	if e, ok := calibration(hist, "REVENUE"); ok {
		fc.CalibrationPct = e
		fc.Expected = math.Round(correctForecast(fc.Expected, mtd.Revenue, e))
		if fc.Low > 0 {
			fc.Low = math.Round(correctForecast(fc.Low, mtd.Revenue, e))
			fc.High = math.Round(correctForecast(fc.High, mtd.Revenue, e))
		}
	}
	if e, ok := calibration(hist, "JOBS"); ok {
		fc.JobsCalibrationPct = e
		fc.ExpectedJobs = int(math.Round(correctForecast(float64(fc.ExpectedJobs), float64(mtd.Bookings), e)))
	}
	fc.Confidence, fc.JobsConfidence = realConfidence(hist, "REVENUE"), realConfidence(hist, "JOBS")
	// السجل يحفظ التوقع الخام (قبل التصحيح) حتى الحكم يقيس الطريقة نفسها ويصحّحها
	_ = s.repo.LogForecast(fc.RawExpected, pipeline, pace, mtd.Revenue, fc.RawExpectedJobs, mtd.Bookings)
	if c, err := s.repo.LastMonthForecast(); err == nil && c != nil {
		for _, m := range months {
			if m.Month == c.Month && m.Revenue > 0 {
				diff := (c.Expected - m.Revenue) / m.Revenue * 100
				fc.Accuracy = fmt.Sprintf("الشهر الماضي (%s): توقعنا بيوم %s %s، والنتيجة طلعت %s — الفرق %+.0f٪.",
					c.Month, c.Day, iqdShort(c.Expected), iqdShort(m.Revenue), diff)
			}
		}
	}
	return fc
}

// iqdShort مبلغ مختصر: 12.5 مليون / 850 ألف.
func iqdShort(v float64) string {
	switch {
	case math.Abs(v) >= 1e6:
		return fmt.Sprintf("%.1f مليون", v/1e6)
	case math.Abs(v) >= 1e3:
		return fmt.Sprintf("%.0f ألف", v/1e3)
	}
	return fmt.Sprintf("%.0f د.ع", v)
}

// InquiryWarning جملة تحذير لو الزبون مستفسر متكرر — فاضية لو لا.
func (s *MatrixBusinessService) InquiryWarning(customerID string) string {
	archived, done := s.repo.InquiryHistory(customerID)
	if done == 0 && archived >= 2 {
		return fmt.Sprintf("🤖 ماتركس — انتبه: هالزبون سجّلنا إله %d حجز قبل وكلها انحذفت بلا تنفيذ (استفسار). تأكد إنه جاد قبل ما تثبّت.", archived)
	}
	return ""
}

// ═══ الحكم والتعلّم — قرار (ع) 10-15 ═══
// «التوقع الي يجيبه صحيح ينبهني إنه توقع صح، والي خطأ يسجله خطأ حتى يتعلم منه».

const forecastTolerance = 10.0 // ±١٠٪ = صح

// judgeForecast الخطأ٪ (موجب = بالغ) والحكم.
func judgeForecast(forecast, actual float64) (float64, string) {
	if actual <= 0 {
		if forecast <= 0 {
			return 0, "CORRECT"
		}
		return 100, "WRONG"
	}
	e := (forecast - actual) / actual * 100
	if math.Abs(e) <= forecastTolerance {
		return e, "CORRECT"
	}
	return e, "WRONG"
}

// calibration معدّل خطأ آخر ٣ أحكام لنفس النوع (محصور ±٣٠٪) — يحتاج حكمين على الأقل.
func calibration(hist []repository.ForecastOutcome, kind string) (float64, bool) {
	sum, n := 0.0, 0
	for _, o := range hist {
		if o.Kind == kind && n < 3 {
			sum += o.ErrorPct
			n++
		}
	}
	if n < 2 {
		return 0, false
	}
	e := sum / float64(n)
	return math.Max(-30, math.Min(30, e)), true
}

// correctForecast يصحّح الجزء الي بعده ما صار بس (المقبوض/المنجز ثابت).
func correctForecast(expected, done, errPct float64) float64 {
	rest := expected - done
	if rest <= 0 {
		return expected
	}
	return done + rest/(1+errPct/100)
}

// realConfidence نسبة الصح من آخر ٦ أحكام — nil إذا أقل من حكمين («لسه يتعلّم»).
func realConfidence(hist []repository.ForecastOutcome, kind string) *int {
	ok, n := 0, 0
	for _, o := range hist {
		if o.Kind == kind && n < 6 {
			n++
			if o.Verdict == "CORRECT" {
				ok++
			}
		}
	}
	if n < 2 {
		return nil
	}
	v := int(math.Round(float64(ok) / float64(n) * 100))
	return &v
}

// EvaluateForecasts يحكم على كل شهر خلص وبيه توقع محفوظ. notify للمالك.
func (s *MatrixBusinessService) EvaluateForecasts(notify func(string)) {
	snaps, err := s.repo.MidMonthSnapshots()
	if err != nil || len(snaps) == 0 {
		return
	}
	months, err := s.repo.Monthly(13)
	if err != nil {
		return
	}
	actual := map[string]repository.MonthRevenue{}
	for _, m := range months {
		actual[m.Month] = m
	}
	for _, sn := range snaps {
		// شهر بلا ولا حجز منجز = النتيجة صفر (مو «ما ينحكم عليه»)
		a := actual[sn.Month]
		day := sn.Day.Format("2006-01-02")
		e, v := judgeForecast(sn.Expected, a.Revenue)
		_ = s.repo.SaveOutcome(repository.ForecastOutcome{Month: sn.Month, Kind: "REVENUE", Forecast: sn.Expected, Actual: a.Revenue, ErrorPct: e, Verdict: v, Day: day})
		msgs := []string{verdictLine("الإيراد", iqdShort(sn.Expected), iqdShort(a.Revenue), e, v)}
		if sn.ExpectedJobs != nil {
			ej, vj := judgeForecast(float64(*sn.ExpectedJobs), float64(a.Bookings))
			_ = s.repo.SaveOutcome(repository.ForecastOutcome{Month: sn.Month, Kind: "JOBS", Forecast: float64(*sn.ExpectedJobs), Actual: float64(a.Bookings), ErrorPct: ej, Verdict: vj, Day: day})
			msgs = append(msgs, verdictLine("الحجوزات", fmt.Sprintf("%d حجز", *sn.ExpectedJobs), fmt.Sprintf("%d", a.Bookings), ej, vj))
		}
		// الإشعار للشهر الماضي بس — الأشهر الأقدم تنحكم بصمت (أول نشر ما يغرّق المالك)
		if notify != nil && sn.Month == time.Now().In(debriefLoc).AddDate(0, -1, 0).Format("2006-01") {
			notify(fmt.Sprintf("📒 توقعات ماتركس لشهر %s (توقع يوم %s): %s", sn.Month, day, strings.Join(msgs, " · ")))
		}
	}
}

func verdictLine(what, forecast, actual string, e float64, v string) string {
	if v == "CORRECT" {
		return fmt.Sprintf("✓ %s صح: توقّع %s وطلع %s (%+.0f٪)", what, forecast, actual, e)
	}
	return fmt.Sprintf("✗ %s غلط: توقّع %s وطلع %s (%+.0f٪) — راح يصحّح توقعاته الجاية", what, forecast, actual, e)
}

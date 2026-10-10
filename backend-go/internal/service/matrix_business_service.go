package service

import (
	"fmt"
	"math"
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
	pipeline := mtd.Revenue + pipeVal

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
		if len(fr) >= 2 && upto > 0 && mtd.Revenue > 0 {
			f := upto / tot
			p := mtd.Revenue / f
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
				len(fr), t.Day(), f*100, iqdShort(mtd.Revenue), iqdShort(p))
		}
	}

	expected := pipeline
	if pace != nil && *pace > pipeline {
		// الوتيرة تشوف شغل بعده ما انحجز — نعطيها وزن يقل كل ما يمشي الشهر.
		expected = elapsed*pipeline + (1-elapsed)**pace
	}
	fc := BusinessForecast{
		Expected:     math.Round(expected),
		ExpectedJobs: mtd.Bookings + expJobs,
		Basis: fmt.Sprintf("المقبوض لحد هسه %s + %d حجز مفتوح موعده هالشهر، متوقع ينجز منها ~%d (نسبة إنجاز كل خدمة × معدل المقبوض بيها بآخر ٩٠ يوم) = %s.%s",
			iqdShort(mtd.Revenue), pl.Jobs, expJobs, iqdShort(pipeline), paceNote),
	}
	// المدى: الشغل الي بعده ما صار (المتوقع − المقبوض) بيه غلط بقدر غلط الوتيرة
	// بالأشهر الماضية (أقل شي ١٠٪، أكثر شي ٥٠٪). المقبوض ثابت ما بيه غلط.
	if rest := expected - mtd.Revenue; rest > 0 {
		e := math.Min(math.Max(spread, 0.10), 0.50)
		fc.Low = math.Round(expected - rest*e)
		fc.High = math.Round(expected + rest*e)
	}
	_ = s.repo.LogForecast(fc.Expected, pipeline, pace, mtd.Revenue)
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

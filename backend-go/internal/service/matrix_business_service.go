package service

import (
	"fmt"
	"math"

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
	fb, err := s.repo.Forecast()
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
	if !fb.CompletionRate.Valid || !fb.AvgInvoice.Valid || fb.Samples < 10 {
		v.Forecast = BusinessForecast{Insufficient: true, Expected: mtd.Revenue, ExpectedJobs: mtd.Bookings,
			Basis: "البيانات ما تكفي للتوقع بعد (أقل من ١٠ فواتير بآخر ٩٠ يوم) — الرقم هو المنجز لحد هسه بس."}
		return v, nil
	}
	jobs := int(math.Round(float64(fb.PendingThisMonth) * fb.CompletionRate.Float64))
	v.Forecast = BusinessForecast{
		ExpectedJobs: mtd.Bookings + jobs,
		Expected:     mtd.Revenue + float64(jobs)*fb.AvgInvoice.Float64,
		Basis: fmt.Sprintf("المنجز لحد هسه (%d حجز) + %d حجز مثبّت باقي هالشهر × نسبة إنجاز المثبّت بآخر ٩٠ يوم (%.0f٪) × معدل الفاتورة (%.0f د.ع من %d فاتورة).",
			mtd.Bookings, fb.PendingThisMonth, fb.CompletionRate.Float64*100, fb.AvgInvoice.Float64, fb.Samples),
	}
	return v, nil
}

// InquiryWarning جملة تحذير لو الزبون مستفسر متكرر — فاضية لو لا.
func (s *MatrixBusinessService) InquiryWarning(customerID string) string {
	archived, done := s.repo.InquiryHistory(customerID)
	if done == 0 && archived >= 2 {
		return fmt.Sprintf("🤖 ماتركس — انتبه: هالزبون سجّلنا إله %d حجز قبل وكلها انحذفت بلا تنفيذ (استفسار). تأكد إنه جاد قبل ما تثبّت.", archived)
	}
	return ""
}

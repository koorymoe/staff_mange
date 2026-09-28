package service

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ═══ ماتركس — فرص البيع ═══
//
// قواعد بس، من بيانات موجودة:
//
//	CROSS_SELL: زبون عنده حجز مكتمل بعائلة خدمة وما حجز أبداً (بأي حالة
//	            غير ملغاة) بالعائلة المكمّلة — حسب خريطة صريحة تحت.
//	MAINTENANCE_DUE: (خدمات العوائل بس) آخر إنجاز لنفس الخدمة صار قبل ١١–١٣ شهر وما اكو حجز
//	            لنفس الخدمة بعده — صيانة سنوية.
//
// الضمان: ماكو أي بيانات ضمان على الحجوزات/المشاريع/أجهزة الزبائن (بس
// ItAsset.warrantyUntil لأجهزة الشركة الداخلية، و SolarSystem.warrantyPrice
// سعر مو تاريخ) — فهذا السبب **ما ينحسب**.
//
// شاشة داخلية للمبيعات: الاسم والهاتف ينعرضون، وما يطلعون لأي نموذج خارجي.

const (
	OpportunityCrossSell      = "CROSS_SELL"
	OpportunityMaintenanceDue = "MAINTENANCE_DUE"

	maintenanceDueMinMonths = 11
	maintenanceDueMaxMonths = 13
)

// عوائل الخدمات — تنعرف من اسم الخدمة (جدول الخدمات، مو نص زبون) أو
// serviceKind.
const (
	famCameras = "CAMERAS"
	famGPS     = "GPS"
	famNetwork = "NETWORK"
	famSolar   = "SOLAR"
	famAlarm   = "ALARM"
)

var familyLabels = map[string]string{
	famCameras: "كاميرات",
	famGPS:     "GPS / داش كام",
	famNetwork: "شبكات",
	famSolar:   "طاقة شمسية",
	famAlarm:   "إنذار",
}

// crossSellMap: عندك هذي ← نقترح هذي.
var crossSellMap = map[string][]string{
	famCameras: {famNetwork, famGPS, famAlarm},
	famGPS:     {famCameras},
	famNetwork: {famCameras},
	famAlarm:   {famCameras},
	famSolar:   {famCameras},
}

// serviceFamily يصنّف خدمة. "" = بلا عائلة (ما تدخل بالبيع المتقاطع).
func serviceFamily(name string, kind *string) string {
	if kind != nil && (*kind == "GPS" || *kind == "DASHCAM") {
		return famGPS
	}
	n := strings.ToLower(name)
	switch {
	case strings.Contains(n, "كاميرات"), strings.Contains(n, "كاميرا"):
		return famCameras
	case strings.Contains(n, "gps"), strings.Contains(n, "داش"):
		return famGPS
	case strings.Contains(n, "شبكات"), strings.Contains(n, "شبكة"):
		return famNetwork
	case strings.Contains(n, "شمسية"), strings.Contains(n, "شمسي"):
		return famSolar
	case strings.Contains(n, "انذار"), strings.Contains(n, "إنذار"):
		return famAlarm
	}
	return ""
}

// SalesOpportunity فرصة وحدة.
type SalesOpportunity struct {
	Reason               string    `json:"reason"`
	ReasonLabel          string    `json:"reasonLabel"`
	CustomerID           string    `json:"customerId"`
	CustomerCode         string    `json:"customerCode"`
	CustomerName         string    `json:"customerName"`
	CustomerPhone        string    `json:"customerPhone"`
	BasedOnService       string    `json:"basedOnService"`
	SuggestedService     string    `json:"suggestedService"`
	SuggestedServiceID   string    `json:"suggestedServiceId,omitempty"`
	LastBookingCompleted time.Time `json:"lastBookingCompleted"`
}

type SalesOpportunityService struct {
	aiRepo    *repository.AiRepository
	notifRepo *repository.NotificationRepository
}

func NewSalesOpportunityService(aiRepo *repository.AiRepository, notifRepo *repository.NotificationRepository) *SalesOpportunityService {
	return &SalesOpportunityService{aiRepo: aiRepo, notifRepo: notifRepo}
}

// isMaintenanceDue آخر إنجاز بين ١١ و١٣ شهر قبل now.
func isMaintenanceDue(last, now time.Time) bool {
	return !last.After(now.AddDate(0, -maintenanceDueMinMonths, 0)) && last.After(now.AddDate(0, -maintenanceDueMaxMonths, 0))
}

// buildOpportunities المنطق الصرف — مفصول للاختبار.
func buildOpportunities(hist []repository.CustomerFamilyHistory, booked map[string]map[string]bool, services []model.Service, now time.Time) []SalesOpportunity {
	// أول خدمة (بالاسم) لكل عائلة — تنقترح وتنربط بالحجز.
	famService := map[string]model.Service{}
	famOf := map[string]string{}
	for _, s := range services {
		f := serviceFamily(s.Name, s.ServiceKind)
		famOf[s.ID] = f
		if f == "" {
			continue
		}
		if _, ok := famService[f]; !ok {
			famService[f] = s
		}
	}
	out := []SalesOpportunity{}
	// بيع متقاطع: مرة وحدة لكل (زبون، عائلة مقترحة) — بآخر إنجاز.
	type key struct{ cid, fam string }
	cross := map[key]SalesOpportunity{}
	for _, h := range hist {
		code := model.Customer{CustomerCode: h.CustomerCode}.FormatCode()
		base := SalesOpportunity{
			CustomerID: h.CustomerID, CustomerCode: code, CustomerName: h.CustomerName,
			CustomerPhone: h.CustomerPhone, BasedOnService: h.ServiceName, LastBookingCompleted: h.LastCompleted,
		}
		// الصيانة السنوية للمنظومات بس (العوائل فوق) — مو صباغة ولبخ.
		if famOf[h.ServiceID] != "" && !h.LaterBookingOf && isMaintenanceDue(h.LastCompleted, now) {
			o := base
			o.Reason, o.ReasonLabel = OpportunityMaintenanceDue, "صيانة سنوية مستحقة"
			o.SuggestedService, o.SuggestedServiceID = "صيانة "+h.ServiceName, h.ServiceID
			out = append(out, o)
		}
		fam := famOf[h.ServiceID]
		for _, target := range crossSellMap[fam] {
			if customerHasFamily(booked[h.CustomerID], famOf, target) {
				continue
			}
			k := key{h.CustomerID, target}
			if prev, ok := cross[k]; ok && !h.LastCompleted.After(prev.LastBookingCompleted) {
				continue
			}
			o := base
			o.Reason, o.ReasonLabel = OpportunityCrossSell, "خدمة مكمّلة"
			o.SuggestedService = familyLabels[target]
			if s, ok := famService[target]; ok {
				o.SuggestedService, o.SuggestedServiceID = s.Name, s.ID
			}
			cross[k] = o
		}
	}
	for _, o := range cross {
		out = append(out, o)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Reason != out[j].Reason {
			return out[i].Reason > out[j].Reason // الصيانة أول
		}
		return out[i].LastBookingCompleted.After(out[j].LastBookingCompleted)
	})
	return out
}

func customerHasFamily(booked map[string]bool, famOf map[string]string, fam string) bool {
	for sid := range booked {
		if famOf[sid] == fam {
			return true
		}
	}
	return false
}

// List كل الفرص الحالية.
func (s *SalesOpportunityService) List() ([]SalesOpportunity, error) {
	hist, err := s.aiRepo.CustomerServiceHistory()
	if err != nil {
		return nil, err
	}
	booked, err := s.aiRepo.CustomerBookedServiceIDs()
	if err != nil {
		return nil, err
	}
	services, err := s.aiRepo.ActiveServices()
	if err != nil {
		return nil, err
	}
	return buildOpportunities(hist, booked, services, time.Now()), nil
}

// RunWeeklyIfDue ملخص أسبوعي (أعداد بس، بلا أسماء) للمالك ومدير النظام.
func (s *SalesOpportunityService) RunWeeklyIfDue() error {
	now := time.Now().In(debriefLoc)
	claimed, err := s.aiRepo.ClaimDailyMarker("WEEKLY_SALES_OPPORTUNITIES_SENT", isoWeekMonday(now))
	if err != nil || !claimed {
		return err
	}
	ops, err := s.List()
	if err != nil || len(ops) == 0 {
		return err
	}
	counts := map[string]int{}
	for _, o := range ops {
		counts[o.Reason]++
	}
	msg := fmt.Sprintf("💡 ماتركس — فرص البيع هذا الأسبوع: %d خدمة مكمّلة، %d صيانة سنوية مستحقة. التفاصيل بشاشة «فرص البيع».",
		counts[OpportunityCrossSell], counts[OpportunityMaintenanceDue])
	if err := s.notifRepo.CreateForRole("OWNER", "AI_SALES_OPPORTUNITIES", msg); err != nil {
		return err
	}
	return s.notifRepo.CreateForRole("ADMIN", "AI_SALES_OPPORTUNITIES", msg)
}

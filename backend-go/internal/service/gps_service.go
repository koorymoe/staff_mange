package service

import (
	"errors"
	"fmt"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type GpsService struct {
	repo *repository.GpsRepository
	// monitor: صندوق المراقب. اختياري.
	monitor MonitorFeed
}

// SetMonitorFeed يربط صندوق المراقب بعد البناء.
func (s *GpsService) SetMonitorFeed(m MonitorFeed) { s.monitor = m }

func NewGpsService(repo *repository.GpsRepository) *GpsService {
	return &GpsService{repo: repo}
}

func (s *GpsService) ListCustomers() ([]model.GpsCustomer, error) {
	return s.repo.ListCustomers()
}

func (s *GpsService) CreateCustomer(req model.UpsertGpsCustomerRequest) (*model.GpsCustomer, error) {
	return s.repo.CreateCustomer(req)
}

func (s *GpsService) UpdateCustomer(id string, req model.UpsertGpsCustomerRequest) (*model.GpsCustomer, error) {
	return s.repo.UpdateCustomer(id, req)
}

func (s *GpsService) ListSims() ([]model.SimCard, error) {
	return s.repo.ListSims()
}

func (s *GpsService) CreateSim(req model.UpsertSimCardRequest) (*model.SimCard, error) {
	return s.repo.CreateSim(req)
}

func (s *GpsService) UpdateSim(id string, req model.UpsertSimCardRequest) (*model.SimCard, error) {
	return s.repo.UpdateSim(id, req)
}

func (s *GpsService) ListDevices() ([]model.GpsDeviceRequest, error) {
	return s.repo.ListDevices()
}

func (s *GpsService) CreateDevice(req model.UpsertGpsDeviceRequest) (*model.GpsDeviceRequest, error) {
	return s.repo.CreateDevice(req)
}

// UpdateDevice تحديث طلب الجهاز — ولو إداري الجي بي اس اختار شريحة
// وقت ترحيل الزبون من المبيعات، نربطها بالزبون بنفس العملية.
//
// الربط يصير أول شي وبشرط status='AVAILABLE' داخل التحديث نفسه، حتى
// لو إداريين اثنين اختاروا نفس الشريحة بنفس اللحظة واحد بس يفوز
// والثاني ياخذ خطأ واضح بدل ما تنربط الشريحة لزبونين.
func (s *GpsService) UpdateDevice(id string, req model.UpsertGpsDeviceRequest) (*model.GpsDeviceRequest, error) {
	if req.SimCardID != nil && *req.SimCardID != "" {
		current, err := s.repo.FindDevice(id)
		if err != nil {
			return nil, err
		}
		// ما نعيد الربط إذا نفس الشريحة مربوطة أصلاً بهذا الطلب
		alreadyLinked := current.SimCardID != nil && *current.SimCardID == *req.SimCardID
		if !alreadyLinked {
			customerID := current.CustomerID
			if req.CustomerID != nil && *req.CustomerID != "" {
				customerID = *req.CustomerID
			}
			if customerID == "" {
				return nil, errors.New("ما نعرف الزبون حتى نربطله الشريحة")
			}
			if _, err := s.repo.AssignSimToCustomer(*req.SimCardID, customerID); err != nil {
				return nil, errors.New("الشريحة مو متوفرة — يمكن انربطت لزبون ثاني قبل شوي، اختر وحدة غيرها")
			}
		}
	}
	// نلتقط الحالة قبل التعديل حتى نعرف **الانتقال** مو الحالة:
	// بدونه كل تعديل على جهاز مسلّم يعيد إضافته لصندوق المراقب.
	before, _ := s.repo.FindDevice(id)
	saved, err := s.repo.UpdateDevice(id, req)
	if err != nil {
		return nil, err
	}
	// التسليم هو نهاية شغل إداري الجي بي اس: الجهاز راح للزبون
	// والاشتراك بدأ. هاي اللحظة الي تنشاف.
	if s.monitor != nil && saved != nil && saved.IsDelivered && (before == nil || !before.IsDelivered) {
		gps := "بلا رقم"
		if saved.GpsNumber != nil {
			gps = *saved.GpsNumber
		}
		sub := "بلا تاريخ"
		if saved.SubscriptionEnd != nil {
			sub = saved.SubscriptionEnd.Format("2006-01-02")
		}
		// الاشتراك بالعربي — المراقب ما يقرا YEARLY
		subType := map[string]string{
			"THREE_MONTHS": "ثلاثة أشهر", "SIX_MONTHS": "ستة أشهر", "YEARLY": "سنوي",
		}[saved.SubscriptionType]
		if subType == "" {
			subType = saved.SubscriptionType
		}
		s.monitor.Stage(model.MonitorStageGpsDeviceDone, "GPS_DEVICE", saved.ID,
			"جهاز جي بي اس انسلّم — "+gps,
			fmt.Sprintf("نوع الاشتراك: %s • ينتهي: %s", subType, sub),
			"GPS_ADMIN", saved.AdminID)
	}
	return saved, nil
}

func (s *GpsService) ListRenewals() ([]model.GpsRenewalRequest, error) {
	return s.repo.ListRenewals()
}

func (s *GpsService) CreateRenewal(req model.UpsertGpsRenewalRequest) (*model.GpsRenewalRequest, error) {
	return s.repo.CreateRenewal(req)
}

func (s *GpsService) UpdateRenewal(id string, req model.UpsertGpsRenewalRequest) (*model.GpsRenewalRequest, error) {
	return s.repo.UpdateRenewal(id, req)
}

func (s *GpsService) ListMaintenance() ([]model.GpsMaintenanceRequest, error) {
	return s.repo.ListMaintenance()
}

func (s *GpsService) CreateMaintenance(req model.UpsertGpsMaintenanceRequest) (*model.GpsMaintenanceRequest, error) {
	return s.repo.CreateMaintenance(req)
}

func (s *GpsService) UpdateMaintenance(id string, req model.UpsertGpsMaintenanceRequest) (*model.GpsMaintenanceRequest, error) {
	return s.repo.UpdateMaintenance(id, req)
}

func (s *GpsService) ListSettings() ([]model.GpsSubscriptionPrice, error) {
	return s.repo.ListSettings()
}

func (s *GpsService) UpsertSetting(req model.UpsertGpsSubscriptionPriceRequest) (*model.GpsSubscriptionPrice, error) {
	id := "default"
	if req.ID != nil && *req.ID != "" {
		id = *req.ID
	}
	return s.repo.UpsertSetting(id, req.SubscriptionType, req.Price)
}

func (s *GpsService) Stats() (*model.GpsStats, error) {
	return s.repo.Stats()
}

// ── دورة حياة الشريحة ومتابعة التجديد ────────────────────────────────────────

func (s *GpsService) ListAvailableSims() ([]model.SimCard, error) {
	return s.repo.ListAvailableSims()
}

func (s *GpsService) AssignSim(simID, customerID string) (*model.SimCard, error) {
	if simID == "" || customerID == "" {
		return nil, errors.New("لازم تحدد الشريحة والزبون")
	}
	return s.repo.AssignSimToCustomer(simID, customerID)
}

func (s *GpsService) ReleaseSim(simID string) (*model.SimCard, error) {
	return s.repo.ReleaseSim(simID)
}

func (s *GpsService) BurnSim(simID string) (*model.SimCard, error) {
	return s.repo.MarkSimBurned(simID)
}

// SubscriptionFollowUps قائمة متابعة الاشتراكات المنتهية. نزامن حالة
// "تحتاج حرق" قبل ما نرجّع القائمة حتى مسؤول الجي بي اس يشوف الوضع الصح
// بدون مهمة مجدولة منفصلة.
func (s *GpsService) SubscriptionFollowUps() ([]model.GpsSubscriptionFollowUpRow, error) {
	if err := s.repo.SyncSimsNeedingBurn(); err != nil {
		return nil, err
	}
	return s.repo.ListSubscriptionFollowUps()
}

func (s *GpsService) CreateFollowUp(deviceRequestID string, calledByID *string, req model.CreateFollowUpRequest) (*model.GpsRenewalFollowUp, error) {
	if _, ok := model.FollowUpOutcomeLabels[req.Outcome]; !ok {
		return nil, errors.New("نتيجة الاتصال غير صحيحة")
	}
	return s.repo.CreateFollowUp(deviceRequestID, calledByID, req)
}

func (s *GpsService) ListFollowUps(deviceRequestID string) ([]model.GpsRenewalFollowUp, error) {
	return s.repo.ListFollowUps(deviceRequestID)
}

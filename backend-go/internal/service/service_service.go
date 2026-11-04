package service

import (
	"errors"
	"strings"

	"github.com/google/uuid"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type ServiceCatalogService struct {
	repo *repository.ServiceRepository
}

func NewServiceCatalogService(repo *repository.ServiceRepository) *ServiceCatalogService {
	return &ServiceCatalogService{repo: repo}
}

// SetManagerHandlesPaperwork يأشّر إن ورق هالخدمة (تقرير + فاتورة) على
// مسؤولها مو على الفني.
func (s *ServiceCatalogService) SetManagerHandlesPaperwork(serviceID string, on bool) error {
	if serviceID == "" {
		return errors.New("معرّف الخدمة مطلوب")
	}
	return s.repo.SetManagerHandlesPaperwork(serviceID, on)
}

// SetKind يأشّر نوع الخدمة (GPS / DASHCAM) أو يمسح التأشير.
//
// ⚠️ القيمة تنتحقّق هنا: نص حر على العمود يخلي «gps» و«Gps» و«جي بي
// اس» ثلاث قيم مختلفة، والترشيح يفشل بهدوء بعد شهر.
func (s *ServiceCatalogService) SetKind(serviceID, kind string) error {
	if serviceID == "" {
		return errors.New("معرّف الخدمة مطلوب")
	}
	if kind != "" && kind != model.ServiceInvoiceGps && kind != model.ServiceInvoiceDashcam {
		return errors.New("نوع الخدمة لازم يكون جي بي اس أو داش كام")
	}
	return s.repo.SetKind(serviceID, kind)
}

func (s *ServiceCatalogService) List() ([]model.Service, error) {
	return s.repo.List()
}

func (s *ServiceCatalogService) Create(req model.CreateServiceRequest) (*model.Service, error) {
	if req.Name == "" {
		return nil, errors.New("اسم الخدمة مطلوب")
	}

	// الشعبة مو تفصيل شكلي: هي الي تحدد مهارات هذي الخدمة تنعرض لأي كادر.
	// نتحقق منها هنا حتى ما تنحفظ قيمة غلط تخلي الخدمة ما تطلع لولا شعبة.
	division := model.DivisionEngineering
	if req.Division != nil && *req.Division != "" {
		if *req.Division != model.DivisionEngineering && *req.Division != model.DivisionDecor {
			return nil, errors.New("شعبة الخدمة لازم تكون هندسية أو ديكور")
		}
		division = *req.Division
	}

	svc := &model.Service{ID: uuid.NewString(), Name: req.Name, Category: req.Category, Division: division}
	if err := s.repo.Create(svc); err != nil {
		return nil, err
	}
	svc.Skills = []model.Skill{}
	return svc, nil
}

func (s *ServiceCatalogService) CreateSkill(serviceID string, req model.CreateSkillRequest) (*model.Skill, error) {
	if req.Name == "" {
		return nil, errors.New("اسم المهارة مطلوب")
	}
	sk := &model.Skill{ID: uuid.NewString(), Name: req.Name, ServiceID: serviceID}
	if err := s.repo.CreateSkill(sk); err != nil {
		return nil, err
	}
	return sk, nil
}

func (s *ServiceCatalogService) Delete(id string) error {
	err := s.repo.Delete(id)
	if err != nil && strings.Contains(err.Error(), "violates foreign key constraint") {
		return errors.New("لا يمكن حذف هذه الخدمة لوجود حجوزات مرتبطة بها")
	}
	return err
}

// AllSkills كل المهارات بقائمة مسطّحة — تحتاجها برامج التدريب.
func (s *ServiceCatalogService) AllSkills() ([]repository.SkillWithService, error) {
	return s.repo.AllSkills()
}

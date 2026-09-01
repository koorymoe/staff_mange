package service

import (
	"errors"
	"strconv"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type ComplaintService struct {
	repo *repository.ComplaintRepository
}

func NewComplaintService(repo *repository.ComplaintRepository) *ComplaintService {
	return &ComplaintService{repo: repo}
}

func (s *ComplaintService) List() ([]model.Complaint, error) {
	return s.repo.List()
}

func (s *ComplaintService) StatsByCustomer() ([]model.ComplaintCustomerStat, error) {
	return s.repo.StatsByCustomer()
}

func (s *ComplaintService) Create(req model.CreateComplaintRequest) (*model.Complaint, error) {
	if req.CustomerID == "" || req.CreatedByEmployeeID == "" {
		return nil, errors.New("customerId و createdByEmployeeId مطلوبين")
	}
	if req.Type == "" {
		req.Type = "OTHER"
	}
	if _, ok := model.ComplaintTypeLabels[req.Type]; !ok {
		return nil, errors.New("نوع شكوى غير معروف")
	}
	c, err := s.repo.Create(req.CustomerID, req.BookingID, req.Type, req.Description, req.CreatedByEmployeeID, req.RelatedEmployeeID)
	if err != nil {
		return nil, err
	}
	s.repo.AddEvent(c.ID, model.EventCreated, nil, req.CreatedByEmployeeID)
	return c, nil
}

func (s *ComplaintService) Update(id string, req model.UpdateComplaintRequest) (*model.Complaint, error) {
	return s.repo.Update(id, req.Status, req.AssignedToEmployeeID, req.Resolution)
}

// SetContacted تأشير الاتصال بالزبون — منو اتصل ومتى ينتخزنون،
// ومعاهن تقييم الزبون لو سأله مهندس الجودة بنفس المكالمة.
func (s *ComplaintService) SetContacted(
	id string, req model.SetContactedRequest, byID string,
) (*model.Complaint, error) {
	c, err := s.repo.SetContacted(id, req.Contacted, byID, req.Rating)
	if err != nil {
		return nil, err
	}
	if req.Contacted {
		s.repo.AddEvent(id, model.EventContacted, nil, byID)
		if req.Rating != nil {
			d := "تقييم الزبون: " + strconv.Itoa(*req.Rating) + "/5"
			s.repo.AddEvent(id, model.EventRated, &d, byID)
		}
	} else {
		d := "شال تأشير التواصل"
		s.repo.AddEvent(id, model.EventContacted, &d, byID)
	}
	return c, nil
}

// Audit حكم المدقق على شغل مهندس الجودة.
//
// ⚠️ الصلاحية تتقرّر بالمسار (المالك والمراقب والمدير بس) — مهندس
// الجودة ما يدقّق نفسه، هو الي انتقيّم شغله.
func (s *ComplaintService) Audit(
	id string, req model.AuditComplaintRequest, byID string,
) (*model.Complaint, error) {
	c, err := s.repo.Audit(id, req.Verdict, req.Note, byID)
	if err != nil {
		return nil, err
	}
	d := model.AuditVerdictLabels[req.Verdict]
	if req.Note != nil && *req.Note != "" {
		d += " — " + *req.Note
	}
	s.repo.AddEvent(id, model.EventAudited, &d, byID)
	return c, nil
}

// Events سجل إجراءات الشكوى.
func (s *ComplaintService) Events(id string) ([]model.ComplaintEvent, error) {
	return s.repo.Events(id)
}

// SetNotes ملاحظات الزبون على الشكوى.
func (s *ComplaintService) SetNotes(id, notes, byID string) (*model.Complaint, error) {
	c, err := s.repo.SetNotes(id, notes)
	if err != nil {
		return nil, err
	}
	s.repo.AddEvent(id, model.EventNoted, nil, byID)
	return c, nil
}

func (s *ComplaintService) Resolve(id string, req model.ResolveComplaintRequest, byID string) (*model.Complaint, error) {
	c, err := s.repo.Resolve(id, req.Resolution)
	if err != nil {
		return nil, err
	}
	s.repo.AddEvent(id, model.EventResolved, req.Resolution, byID)
	return c, nil
}

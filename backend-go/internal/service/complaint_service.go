package service

import (
	"errors"
	"strconv"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type ComplaintService struct {
	repo *repository.ComplaintRepository
	// ai: مسجّل إشارات ماتركس (شكاوى متكررة) — اختياري.
	ai AiSignalRecorder
	// notifyVip: إشعار فوري للمالك والمدير بشكوى شخصية مهمة (ع) 10-15
	notifyVip func(message string)
	notifyNew func(assigneeID *string, message string)
}

// SetNewNotifier (ع) 10-15: كل شكوى جديدة → إشعار لمهندس الجودة المكلّف + المراقب والمدير.
func (s *ComplaintService) SetNewNotifier(f func(assigneeID *string, message string)) {
	s.notifyNew = f
}

// SetVipNotifier يربط إشعار شكاوى الشخصيات المهمة.
func (s *ComplaintService) SetVipNotifier(f func(string)) { s.notifyVip = f }

// SetAiRecorder يربط مسجّل إشارات ماتركس بعد البناء.
func (s *ComplaintService) SetAiRecorder(a AiSignalRecorder) { s.ai = a }

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
	// ماتركس ٩: شكاوى متكررة على نفس الليدر — ما يوقف الشكوى أبداً.
	recordRepeatComplaint(s.ai, c.ID)
	if s.notifyNew != nil {
		who := "زبون"
		if c.Customer != nil {
			who = c.Customer.Name
		}
		to := "ما أكو مهندس جودة فعّال — كلّف أحد"
		if c.AssignedToEmployee != nil {
			to = "انكلّف بيها " + c.AssignedToEmployee.Name
		}
		s.notifyNew(c.AssignedToEmployeeID, "🎧 شكوى جديدة: "+who+" — "+model.ComplaintTypeLabels[c.Type]+" · "+to+" · تواصل ويّا الزبون")
	}
	if c.IsVip && s.notifyVip != nil {
		who := "زبون"
		if c.Customer != nil {
			who = c.Customer.Name
		}
		if c.VipNote != nil {
			who += " (" + *c.VipNote + ")"
		}
		label := model.ComplaintTypeLabels[c.Type]
		s.notifyVip("👑 شكوى من شخصية مهمة: " + who + " — " + label + ". مهلتها ٢٤ ساعة.")
	}
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

// NoAnswer «الزبون ما رد» — يأجّل المكالمة، والثالثة تسقط الشكوى.
func (s *ComplaintService) NoAnswer(id, byID string) (*model.Complaint, error) {
	c, err := s.repo.NoAnswer(id)
	if err != nil {
		return nil, err
	}
	d := "الزبون ما رد — محاولة " + strconv.Itoa(c.NoAnswerCount) + " من " + strconv.Itoa(repository.ComplaintNoAnswerLimit)
	if c.DroppedNoAnswerAt != nil {
		d += " — الشكوى سقطت"
	}
	s.repo.AddEvent(id, model.EventNoAnswer, &d, byID)
	return c, nil
}

// DropNoAnswer تسدّ الشكوى فوراً «الزبون ما يرد» — وتنسجّل بسجلها.
func (s *ComplaintService) DropNoAnswer(id, byID string) (*model.Complaint, error) {
	c, err := s.repo.DropNoAnswer(id)
	if err != nil {
		return nil, err
	}
	d := "الزبون ما يرد — الشكوى انسدّت مباشرة"
	s.repo.AddEvent(id, model.EventNoAnswer, &d, byID)
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

func (s *ComplaintService) LinkBooking(id, bookingID, byID string) (*model.Complaint, error) {
	code, err := s.repo.LinkBooking(id, bookingID)
	if err != nil {
		return nil, errors.New("تعذر ربط الحجز — لازم يكون حجز لنفس زبون الشكوى")
	}
	d := "انسجل حجز شكوى " + code
	s.repo.AddEvent(id, model.EventBooked, &d, byID)
	return s.repo.Find(id)
}

func (s *ComplaintService) Resolve(id string, req model.ResolveComplaintRequest, byID string) (*model.Complaint, error) {
	c, err := s.repo.Resolve(id, req.Resolution)
	if err != nil {
		return nil, err
	}
	s.repo.AddEvent(id, model.EventResolved, req.Resolution, byID)
	return c, nil
}

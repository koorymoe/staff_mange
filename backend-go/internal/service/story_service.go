package service

import (
	"encoding/json"
	"log"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// StoryService عقل محرّك القصص.
//
// ⚠️ **الحركة لا تسبق نجاح العملية**: `Emit` تنستدعى **بعد** ما
// ينجح الإجراء الإداري ويترسّخ بجدوله. والقصة **ما تنشئ ولا تعدّل
// أي حقيقة إدارية** — هي عرض لحدث صار، مو الحدث نفسه.
type StoryService struct {
	repo      *repository.StoryRepository
	employees *repository.EmployeeRepository
}

func NewStoryService(repo *repository.StoryRepository, employees *repository.EmployeeRepository) *StoryService {
	return &StoryService{repo: repo, employees: employees}
}

// Emit ينشئ قصة من حدث رسمي.
//
// ⚠️⚠️ **ما يرجّع خطأ يوقف المستدعي**: فشل القصة **ما يلغي الخصم**
// ولا الفاتورة ولا أي إجراء. بس **ما ينبلع بصمت** — ينسجّل بالسجل،
// لأن الدرس الي كلّفنا: تحذير ناقص أسوأ من ماكو تحذير.
func (s *StoryService) Emit(req model.EmitStoryRequest) {
	if s == nil || s.repo == nil {
		return
	}
	if req.EventID == "" || req.RecipientID == "" {
		log.Printf("[story] طلب ناقص: eventId=%q recipient=%q", req.EventID, req.RecipientID)
		return
	}

	// ① التجميع: عنده قصة مفتوحة بنفس المجموعة؟ ندمج بدل ما نزاحم.
	if req.GroupKey != nil && *req.GroupKey != "" {
		line, _ := req.Payload["title"].(string)
		if line == "" {
			line = model.StoryEventLabel[req.EventKind]
		}
		merged, err := s.repo.MergeIntoGroup(req.RecipientID, *req.GroupKey, line)
		if err != nil {
			log.Printf("[story] دمج المجموعة %s فشل: %v", *req.GroupKey, err)
		} else if merged {
			return
		}
	}

	// ② السقف اليومي: بعده القصة تبقى محفوظة بس **هادئة** — تنعرض
	// بالصندوق بلا مشهد جسدي. ما تنلغى ولا تضيع.
	physical := true
	if n, err := s.repo.CountPhysicalToday(req.RecipientID); err != nil {
		// ⚠️ فشل العدّ ← نهدّي المشهد بدل ما نزاحم الموظف. الخطأ
		// يضيّق ما يوسّع — نفس قاعدة `canSeeAllBookings`.
		log.Printf("[story] عدّ مشاهد اليوم فشل لـ%s: %v", req.RecipientID, err)
		physical = false
	} else if n >= model.StoryDailyPhysicalCap {
		physical = false
	}

	payload := req.Payload
	if payload == nil {
		payload = map[string]any{}
	}
	payload["mergedCount"] = 1
	raw, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[story] payload غير قابل للتحويل: %v", err)
		raw = []byte(`{}`)
	}

	// ⚠️ **الاسم منسوخ نصاً وقت الإنشاء**: لو انحذف الحساب بكرة،
	// السطر يبقى مقروءاً بدل ما يصير «مجهول». نفس نمط `byName`
	// المطبَّق بـComplaintEvent وCoordinationAlert وDesignAsset.
	recipientName := req.RecipientName
	if recipientName == "" && s.employees != nil {
		if e, err := s.employees.FindByID(req.RecipientID); err == nil && e != nil {
			recipientName = e.Name
		}
	}
	// ونفس الشي للمرسِل: لو انمرّر معرّفه بلا اسم، نكمّله من الجدول
	// **مرة وحدة هنا** بدل ما كل مستدعٍ يجيبه بنفسه.
	senderName := req.SenderName
	if senderName == "" && req.SenderID != nil && *req.SenderID != "" && s.employees != nil {
		if e, err := s.employees.FindByID(*req.SenderID); err == nil && e != nil {
			senderName = e.Name
		}
	}
	recipientID := req.RecipientID
	story := model.StoryInstance{
		EventID:             req.EventID,
		EventKind:           req.EventKind,
		StoryType:           req.EventKind,
		SenderEmployeeID:    req.SenderID,
		SenderName:          senderName,
		RecipientEmployeeID: &recipientID,
		RecipientRef:        req.RecipientID,
		RecipientName:       recipientName,
		Priority:            model.StoryPriority[req.EventKind],
		Physical:            physical,
		GroupKey:            req.GroupKey,
		Payload:             raw,
	}
	created, err := s.repo.Enqueue(story)
	if err != nil {
		log.Printf("[story] تعذر إنشاء قصة %s لـ%s: %v", req.EventKind, req.RecipientID, err)
		return
	}
	if !created {
		return // انسجّلت قبل — الفهرس الفريد منعها، وهذا الي نريده
	}
	log.Printf("[story] قصة %s → %s (جسدية=%v)", req.EventKind, req.RecipientID, physical)
}

// Next القصة الي دورها الآن لهذا الموظف، ومعها مشهدها الجاهز.
//
// ⚠️ **المشهد يجي من الخادم**: الواجهة تنفّذ ما تبني. مصدر واحد
// للمشهد يعني تعديل خطوة يوصل كل الأجهزة بلا نشر جديد.
func (s *StoryService) Next(employeeID string) (*model.StoryWithScene, error) {
	story, err := s.repo.NextForEmployee(employeeID)
	if err != nil || story == nil {
		return nil, err
	}
	return &model.StoryWithScene{
		StoryInstance: *story,
		Scene:         model.SceneFor(story.EventKind),
		Label:         model.StoryEventLabel[story.EventKind],
	}, nil
}

// Claim يحجز القصة لهذي النافذة — نافذة وحدة تشغّل المشهد.
func (s *StoryService) Claim(id, employeeID string) (bool, error) {
	return s.repo.Claim(id, employeeID)
}

// Advance ينقل قصة الموظف نفسه لمرحلة أبعد.
func (s *StoryService) Advance(id, employeeID, status string, step int) error {
	return s.repo.Advance(id, employeeID, status, step)
}

// Inbox صندوق قصص الموظف — سجل مقروء حتى بعد ما ينتهي المشهد.
func (s *StoryService) Inbox(employeeID string, limit int) ([]model.StoryInstance, error) {
	return s.repo.ListForEmployee(employeeID, limit)
}

// PendingCount عدد الي ينتظره.
func (s *StoryService) PendingCount(employeeID string) (int, error) {
	return s.repo.PendingCount(employeeID)
}

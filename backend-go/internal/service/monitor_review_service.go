package service

import (
	"fmt"
	"log"
	"time"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
	"strings"
	"unicode/utf8"
)

// MonitorFeed يخلي خدمة الحجوزات وخدمة الفواتير تدزّ للمراقب بدون ما
// تعتمد على مستودعه — نفس أسلوب AssignmentBalanceChecker، حتى ما يصير
// اعتماد دائري ولا يوقف الشغل الأصلي لو انكسر شي بالمراقبة.
type MonitorFeed interface {
	BookingStage(stage string, b *model.Booking, ownerRole string, ownerEmployeeID *string)
	InvoiceStage(stage, invoiceID, title, summary, ownerRole string, ownerEmployeeID *string)
}

// MonitorReviewService صندوق المراقب — إضافة وقرار.
type MonitorReviewService struct {
	repo          *repository.MonitorReviewRepository
	employees     *repository.EmployeeRepository
	notifications *repository.NotificationRepository
}

func NewMonitorReviewService(
	repo *repository.MonitorReviewRepository,
	employees *repository.EmployeeRepository,
	notifications *repository.NotificationRepository,
) *MonitorReviewService {
	return &MonitorReviewService{repo: repo, employees: employees, notifications: notifications}
}

func (s *MonitorReviewService) List(stage, status, ownerRole string, limit int) ([]model.MonitorReview, error) {
	return s.repo.List(stage, status, ownerRole, limit)
}

func (s *MonitorReviewService) Counts() ([]model.MonitorInboxCount, error) {
	return s.repo.Counts()
}

// BookingStage يدزّ حجز لمحطة مراقبة.
//
// الخطأ ينتسجّل وما ينرجّع: فشل صف مراقبة ما يصير يمنع تثبيت حجز.
func (s *MonitorReviewService) BookingStage(stage string, b *model.Booking, ownerRole string, ownerEmployeeID *string) {
	if b == nil {
		return
	}
	when := "بلا موعد"
	if b.ScheduledAt != nil {
		when = b.ScheduledAt.Add(3 * time.Hour).Format("2006-01-02 15:04") // بتوقيت بغداد
	}
	customer := ""
	if b.Customer != nil {
		customer = b.Customer.Name
	}
	err := s.repo.Enqueue(model.EnqueueMonitorReview{
		Stage:           stage,
		EntityType:      "BOOKING",
		EntityID:        b.ID,
		Title:           fmt.Sprintf("حجز %s — %s", b.Code, customer),
		Summary:         fmt.Sprintf("الموعد: %s • الحالة: %s", when, b.Status),
		OwnerRole:       ownerRole,
		OwnerEmployeeID: ownerEmployeeID,
	})
	if err != nil {
		log.Printf("[monitor] تعذر إضافة صف مراقبة للحجز %s: %v", b.ID, err)
	}
}

// InvoiceStage يدزّ فاتورة لمحطة مراقبة.
func (s *MonitorReviewService) InvoiceStage(stage, invoiceID, title, summary, ownerRole string, ownerEmployeeID *string) {
	err := s.repo.Enqueue(model.EnqueueMonitorReview{
		Stage:           stage,
		EntityType:      "LEADER_INVOICE",
		EntityID:        invoiceID,
		Title:           title,
		Summary:         summary,
		OwnerRole:       ownerRole,
		OwnerEmployeeID: ownerEmployeeID,
	})
	if err != nil {
		log.Printf("[monitor] تعذر إضافة صف مراقبة للفاتورة %s: %v", invoiceID, err)
	}
}

// Decide قرار المراقب.
//
// «عندي ملاحظة» بدون نص ما تفيد أحد — الموظف الي توصله ما راح يعرف
// شنو يصلّح. فالنص إجباري بالتأشير، وبالحروف مو البايتات.
func (s *MonitorReviewService) Decide(id, monitorID string, req model.DecideMonitorReviewRequest) (*model.MonitorReview, error) {
	if req.Flag && utf8.RuneCountInString(strings.TrimSpace(req.Note)) < 5 {
		return nil, fmt.Errorf("اكتب الملاحظة — الموظف لازم يعرف شنو يصلّح")
	}
	row, err := s.repo.Decide(id, monitorID, req)
	if err != nil {
		return nil, err
	}
	if row.Status == model.MonitorStatusFlagged {
		s.notifyFlag(row)
	}
	return row, nil
}

// notifyFlag يوصّل الملاحظة لصاحب الشغل وللإدارة.
//
// بدون هذا، التأشير يبقى بشاشة المراقب بس — يعني نفس المشكلة القديمة
// بس بالاتجاه المعاكس.
func (s *MonitorReviewService) notifyFlag(row *model.MonitorReview) {
	msg := fmt.Sprintf("👁️ ملاحظة من المراقب على «%s» (%s): %s",
		row.Title, model.MonitorStageLabel(row.Stage), derefText(row.Note))

	if row.OwnerEmployeeID != nil {
		if err := s.notifications.Create(*row.OwnerEmployeeID, "monitor_flag", msg); err != nil {
			log.Printf("[monitor] تعذر إرسال إشعار الملاحظة للموظف: %v", err)
		}
	}
	// الإدارة تنبلغ دائماً: ملاحظة المراقب الي ما يشوفها المدير ما
	// تنفّذ، وتبقى بس تدوين.
	_ = s.notifications.CreateForRole("ADMIN", "monitor_flag", msg)
}

func derefText(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

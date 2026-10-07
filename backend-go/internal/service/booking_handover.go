package service

import (
	"errors"
	"fmt"
	"strings"

	"staffmange-api/internal/model"
)

// ═══ ترحيل الحجز للتقني (قرار (ع) 10-07) ═══
// الإداري يتواصل ويا الزبون، وإذا المشكلة ما يعرفون الفنيين يحلّوها يرحّل
// الحجز لتقني أو مسؤول خدمة. من هنا الحجز برقبة التقني: يتواصل ويا الزبون
// مرة ثانية، يكتب الكشف، ويعالج بنفسه — أو يطلب طاقم من الإداريين.

var ErrNotHandoverOwner = errors.New("هذا الحجز مو مرحّل إلك")


func (s *BookingService) Handover(id, toID, reason, byID string) (*model.Booking, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return nil, errors.New("اكتب شنو مشكلة الزبون وليش ترحّله للتقني")
	}
	b, err := s.repo.FindByID(id)
	if err != nil || b == nil {
		return nil, errors.New("الحجز مو موجود")
	}
	if b.Status == "COMPLETED" || b.Status == "CANCELLED" {
		return nil, errors.New("الحجز مخلّص أو ملغي — ما ينرحّل")
	}
	if err := s.ensureNotProjectLocked(id); err != nil {
		return nil, err
	}
	if b.ConfirmationContactedAt == nil {
		return nil, errors.New("تواصل ويا الزبون أول واضغط «تواصلت ويا الزبون»، بعدين رحّله")
	}
	if !s.employees.IsHandoverTarget(toID) {
		return nil, errors.New("اختار تقني أو مسؤول خدمة (مو فني ميداني)")
	}
	if err := s.repo.Handover(id, toID, byID, reason); err != nil {
		return nil, err
	}
	if s.notifications != nil {
		_ = s.notifications.Create(toID, "booking_handover",
			fmt.Sprintf("🛠️ انرحّلك الحجز %s — صار برقبتك. تواصل ويا الزبون، اكتب الكشف، وعالج. المشكلة: %s", b.Code, reason))
	}
	return s.repo.FindByID(id)
}

// handoverOwned يرجّع الحجز إذا مرحّل لهالموظف وبعده مفتوح.
func (s *BookingService) handoverOwned(id, empID string) (*model.Booking, error) {
	b, err := s.repo.FindByID(id)
	if err != nil || b == nil {
		return nil, errors.New("الحجز مو موجود")
	}
	if b.HandoverToID == nil || *b.HandoverToID != empID {
		return nil, ErrNotHandoverOwner
	}
	if b.Status == "COMPLETED" || b.Status == "CANCELLED" {
		return nil, errors.New("الحجز مخلّص أو ملغي")
	}
	return b, nil
}

func (s *BookingService) TechContacted(id, empID string) (*model.Booking, error) {
	if _, err := s.handoverOwned(id, empID); err != nil {
		return nil, err
	}
	if err := s.repo.TechContacted(id); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

func (s *BookingService) TechDiagnose(id, empID, diagnosis string) (*model.Booking, error) {
	diagnosis = strings.TrimSpace(diagnosis)
	if len([]rune(diagnosis)) < 10 {
		return nil, errors.New("اكتب الكشف بالتفصيل: شنو المشكلة وشنو الحل")
	}
	if _, err := s.handoverOwned(id, empID); err != nil {
		return nil, err
	}
	if err := s.repo.TechDiagnose(id, diagnosis); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

func (s *BookingService) TechRequestCrew(id, empID, note string) (*model.Booking, error) {
	b, err := s.handoverOwned(id, empID)
	if err != nil {
		return nil, err
	}
	if b.TechDiagnosedAt == nil {
		return nil, errors.New("اكتب الكشف أول حتى الطاقم يعرف شنو المشكلة")
	}
	if err := s.repo.TechRequestCrew(id); err != nil {
		return nil, err
	}
	if s.notifications != nil {
		msg := fmt.Sprintf("👷 التقني يطلب طاقم للحجز %s. الكشف: %s", b.Code, *b.TechDiagnosis)
		if n := strings.TrimSpace(note); n != "" {
			msg += " — " + n
		}
		_ = s.notifications.CreateForRole("HR_COORDINATOR", "booking_tech_crew", msg)
	}
	return s.repo.FindByID(id)
}

func (s *BookingService) TechResolve(id, empID, notes string) (*model.Booking, error) {
	notes = strings.TrimSpace(notes)
	if notes == "" {
		return nil, errors.New("اكتب شلون انحلّت المشكلة")
	}
	b, err := s.handoverOwned(id, empID)
	if err != nil {
		return nil, err
	}
	if b.TechDiagnosedAt == nil {
		return nil, errors.New("اكتب الكشف أول")
	}
	if err := s.repo.TechResolve(id, notes); err != nil {
		return nil, err
	}
	// مثل أي حجز يخلص: الجودة تتصل بالزبون تتأكد انحلّت.
	if s.qualityFollowUps != nil {
		_ = s.qualityFollowUps.CreateForBooking(b.ID, b.CustomerID)
	}
	return s.repo.FindByID(id)
}

func (s *BookingService) HandoverCandidates() ([]model.EmployeeBrief, error) {
	return s.employees.HandoverCandidates()
}

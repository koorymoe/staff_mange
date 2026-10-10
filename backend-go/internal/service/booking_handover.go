package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

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
	b, err := s.handoverOwned(id, empID)
	if err != nil {
		return nil, err
	}
	if b.TechDecision == nil || *b.TechDecision != "VISIT" {
		return nil, errors.New("حدد أول: انحلّت بالتلفون لو تحتاج زيارة")
	}
	if b.TechVisitedAt == nil {
		return nil, errors.New("سجّل وصولك للزبون أول، بعدين اكتب الكشف")
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

// TechDecide بعد التواصل (قرار (ع) 10-07): PHONE = انحلّت بالتلفون (يكتب شلون
// ويتسكّر الحجز)، VISIT = يحتاج كشف/زيارة ويحدد موعد. visitAt بتوقيت بغداد
// «2006-01-02T15:04» أو RFC3339.
func (s *BookingService) TechDecide(id, empID, decision, text, visitAt string) (*model.Booking, error) {
	b, err := s.handoverOwned(id, empID)
	if err != nil {
		return nil, err
	}
	if b.TechContactedAt == nil {
		return nil, errors.New("تواصل ويا الزبون أول")
	}
	text = strings.TrimSpace(text)
	switch decision {
	case "PHONE":
		if len([]rune(text)) < 10 {
			return nil, errors.New("اكتب شلون انحلّت بالتلفون (شنو المشكلة وشنو سوّيت)")
		}
		if err := s.repo.TechDecide(id, "PHONE", nil); err != nil {
			return nil, err
		}
		if err := s.repo.TechDiagnose(id, text); err != nil {
			return nil, err
		}
		if err := s.repo.TechResolve(id, "انحلّت بالتلفون: "+text); err != nil {
			return nil, err
		}
		if s.qualityFollowUps != nil {
			_ = s.qualityFollowUps.CreateForBooking(b.ID, b.CustomerID)
		}
		return s.repo.FindByID(id)
	case "VISIT":
		at, err := parseVisitAt(visitAt)
		if err != nil {
			return nil, err
		}
		if at.Before(time.Now().Add(-time.Hour)) {
			return nil, errors.New("الموعد لازم يكون من هسه وطالع")
		}
		if err := s.repo.TechDecide(id, "VISIT", &at); err != nil {
			return nil, err
		}
		if s.notifications != nil {
			_ = s.notifications.CreateForRole("HR_COORDINATOR", "booking_tech_visit",
				fmt.Sprintf("🔍 التقني حدد زيارة كشف للحجز %s يوم %s", b.Code, at.In(debriefLoc).Format("2006-01-02 15:04")))
		}
		return s.repo.FindByID(id)
	}
	return nil, errors.New("اختار: انحلّت بالتلفون لو تحتاج زيارة")
}

func parseVisitAt(v string) (time.Time, error) {
	v = strings.TrimSpace(v)
	if t, err := time.Parse(time.RFC3339, v); err == nil {
		return t, nil
	}
	if t, err := time.ParseInLocation("2006-01-02T15:04", v, debriefLoc); err == nil {
		return t, nil
	}
	return time.Time{}, errors.New("حدد موعد الزيارة (اليوم والساعة)")
}

// TechVisited التقني وصل للزبون بموعد الزيارة.
func (s *BookingService) TechVisited(id, empID string) (*model.Booking, error) {
	b, err := s.handoverOwned(id, empID)
	if err != nil {
		return nil, err
	}
	if b.TechDecision == nil || *b.TechDecision != "VISIT" {
		return nil, errors.New("ماكو زيارة محددة لهالحجز")
	}
	if err := s.repo.TechVisited(id); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

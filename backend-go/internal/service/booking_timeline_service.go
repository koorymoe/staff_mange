package service

import (
	"fmt"
	"sort"
	"time"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ═══ الخط الزمني للحجز ═══
//
// يجمع قصة الحجز من سبعة جداول ويرتّبها بالوقت، ويحسب التأخيرات.
//
// ⚠️ ماكو بيانات جديدة تنجمع — كلها موجودة أصلاً بس متفرقة، والي
// يريد يعرف «شنو صار بهذا الحجز؟» چان لازم يفتح سبع شاشات ويرتّب
// بمخه.
type BookingTimelineService struct {
	bookings *repository.BookingRepository
	progress *repository.BookingProgressRepository
	invoices *repository.LeaderInvoiceRepository
	quality  *repository.QualityFollowUpRepository
	monitor  *repository.MonitorReviewRepository
}

func NewBookingTimelineService(
	bookings *repository.BookingRepository,
	progress *repository.BookingProgressRepository,
	invoices *repository.LeaderInvoiceRepository,
	quality *repository.QualityFollowUpRepository,
	monitor *repository.MonitorReviewRepository,
) *BookingTimelineService {
	return &BookingTimelineService{bookings: bookings, progress: progress,
		invoices: invoices, quality: quality, monitor: monitor}
}

func (s *BookingTimelineService) Build(bookingID string) (*model.BookingTimeline, error) {
	b, err := s.bookings.FindByID(bookingID)
	if err != nil || b == nil {
		return nil, fmt.Errorf("الحجز مو موجود")
	}

	out := &model.BookingTimeline{BookingID: b.ID, Code: b.Code}
	add := func(at *time.Time, kind, title, detail, actor string) {
		// ⚠️ الحدث بلا وقت ما ينضاف: ترتيبه مستحيل، وحطّه بوقت وهمي
		// (الآن، أو الصفر) يكذب على القارئ.
		if at == nil || at.IsZero() {
			return
		}
		out.Events = append(out.Events, model.TimelineEvent{
			At: *at, Kind: kind, Title: title, Detail: detail, Actor: actor,
		})
	}
	name := func(p *string) string {
		if p == nil {
			return ""
		}
		return *p
	}

	// ═══ أحداث الحجز نفسه ═══
	created := b.CreatedAt
	add(&created, model.TimelineCreated, "انسجّل الحجز", "", name(b.CreatedByName))
	add(b.ConfirmationContactedAt, model.TimelineContacted, "تواصل مع الزبون", "", "")
	add(b.ConfirmedAt, model.TimelineConfirmed, "انثبّت الحجز", "", name(b.ConfirmedByName))
	add(b.StartedAt, model.TimelineStarted, "بدأ الشغل", "", "")
	if b.WorkStoppedAt != nil {
		add(b.WorkStoppedAt, model.TimelineStopped, "توقف العمل", name(b.WorkStopReason), "")
	}
	add(b.CompletedAt, model.TimelineCompleted, "انجز الحجز", name(b.CompletionNotes), "")
	if b.CancelledAt != nil {
		add(b.CancelledAt, model.TimelineCancelled, "انلغى الحجز", name(b.CancelReason), name(b.CancelledByName))
	}
	if b.LastPostponedAt != nil {
		add(b.LastPostponedAt, model.TimelinePostponed,
			fmt.Sprintf("انأجّل (%d مرة)", b.PostponeCount), name(b.PostponeReason), "")
	}
	if b.WaitingSince != nil {
		add(b.WaitingSince, model.TimelineWaiting,
			fmt.Sprintf("الزبون ما رد (%d محاولة)", b.ContactAttempts), name(b.WaitingNote), "")
	}

	// أول تكليف كادر — منه ينحسب «تأخر التكليف»
	var firstAssign *time.Time
	if assigns, err := s.bookings.ListAssignments(b.ID); err == nil {
		for i := range assigns {
			at := assigns[i].CreatedAt
			if firstAssign == nil || at.Before(*firstAssign) {
				c := at
				firstAssign = &c
			}
		}
		if firstAssign != nil {
			add(firstAssign, model.TimelineAssigned,
				fmt.Sprintf("انكلّف الكادر (%d)", len(assigns)), "", "")
		}
	}

	// ═══ تغييرات الموعد ═══
	if logs, err := s.bookings.ScheduleLog(b.ID); err == nil {
		for i := range logs {
			actor := ""
			if logs[i].ChangedBy != nil {
				actor = logs[i].ChangedBy.Name
			}
			at := logs[i].CreatedAt
			add(&at, model.TimelineSchedule, "انتغيّر الموعد", "", actor)
		}
	}

	// ═══ الإنجاز الجزئي ═══
	if reports, err := s.progress.Reports(b.ID); err == nil {
		for i := range reports {
			at := reports[i].CreatedAt
			add(&at, model.TimelinePartial,
				fmt.Sprintf("إنجاز جزئي — اليوم %d", reports[i].DayNumber), "", "")
		}
	}

	// ═══ الفواتير ═══
	var firstInvoiceAt, firstApprovedAt *time.Time
	invoiceIDs := []string{}
	if invs, err := s.invoices.ListByBooking(b.ID); err == nil {
		for i := range invs {
			inv := invs[i]
			invoiceIDs = append(invoiceIDs, inv.ID)
			at := inv.CreatedAt
			if firstInvoiceAt == nil {
				c := at
				firstInvoiceAt = &c
			}
			add(&at, model.TimelineInvoiced, "انصدرت فاتورة",
				fmt.Sprintf("%s — %.0f د.ع", inv.AccountingCode, inv.NetTotal), "")
			if inv.ApprovedAt != nil && firstApprovedAt == nil {
				c := *inv.ApprovedAt
				firstApprovedAt = &c
			}
			if inv.ApprovedAt != nil {
				detail := ""
				if inv.ExternalInvoiceNumber != nil {
					detail = "رقم محاسبي: " + *inv.ExternalInvoiceNumber
				}
				add(inv.ApprovedAt, model.TimelineApproved, "انعتمدت الفاتورة", detail, "")
			}
		}
	}

	// ═══ متابعة الجودة ═══
	if q, err := s.quality.ByBooking(b.ID); err == nil && q != nil {
		add(q.ContactedAt, model.TimelineQuality, "متابعة جودة — تواصل مع الزبون", "", "")
		add(q.InspectedAt, model.TimelineQuality, "كشف الجودة", "", "")
	}

	// ═══ صفوف المراقب ═══
	// ⚠️ صفوف الحجز مفتاحها الحجز، وصفوف الفواتير مفتاحها **الفاتورة**
	// — فنسأل عن الاثنين بأنواعهم، وإلا صفوف الفواتير ما تطلع أبداً.
	addMonitor := func(rows []model.MonitorReview) {
		for i := range rows {
			r := rows[i]
			detail := model.MonitorStageLabel(r.Stage)
			if r.Status != model.MonitorStatusPending && r.Note != nil {
				detail += " — " + *r.Note
			}
			actor := ""
			if r.ReviewedBy != nil {
				actor = r.ReviewedBy.Name
			}
			at := r.CreatedAt
			title := "وصل صندوق المراقب"
			if r.Status == model.MonitorStatusFlagged {
				title = "⚠️ المراقب أشّر ملاحظة"
			} else if r.Status == model.MonitorStatusOK {
				title = "✓ المراقب دقّق: سليم"
			}
			add(&at, model.TimelineMonitor, title, detail, actor)
		}
	}
	if rows, err := s.monitor.ByEntity("BOOKING", []string{b.ID}); err == nil {
		addMonitor(rows)
	}
	if len(invoiceIDs) > 0 {
		if rows, err := s.monitor.ByEntity("LEADER_INVOICE", invoiceIDs); err == nil {
			addMonitor(rows)
		}
	}

	sort.Slice(out.Events, func(i, j int) bool { return out.Events[i].At.Before(out.Events[j].At) })

	out.Delays = buildDelays(b, firstAssign, firstInvoiceAt, firstApprovedAt)
	return out, nil
}

// buildDelays يحسب التأخيرات الستة.
//
// ⚠️ كل قياس يرجع nil لما ما ينطبق — مو صفر. حجز ما وصل الفوترة ما
// إله «تأخر فوترة»، وعرضه صفراً يعني إنه انفوتر فوراً (كذبة تخلي
// موظفاً يطلع ممتازاً بشي ما سوّاه أصلاً).
func buildDelays(b *model.Booking, firstAssign, firstInvoice, firstApproved *time.Time) []model.DelayMetric {
	mins := func(from, to *time.Time) *int {
		if from == nil || to == nil {
			return nil
		}
		d := int(to.Sub(*from).Minutes())
		// ⚠️ السالب يصير فعلاً: الفني يطلع قبل موعده. «قبل الموعد» مو
		// تأخير، فنصفّرها بدل ما نعرض رقماً سالباً يربك القارئ.
		if d < 0 {
			d = 0
		}
		return &d
	}
	build := func(key, label, owner string, v *int, threshold int) model.DelayMetric {
		m := model.DelayMetric{Key: key, Label: label, Owner: owner, Minutes: v, ThresholdMinutes: threshold}
		m.Breached = v != nil && *v > threshold
		return m
	}
	createdAt := b.CreatedAt
	return []model.DelayMetric{
		build("CONFIRM", "تأخر التثبيت", "إداري الحجوزات",
			mins(&createdAt, b.ConfirmedAt), model.DelayConfirmMinutes),
		build("ASSIGN", "تأخر التكليف", "إداري الحجوزات",
			mins(b.ConfirmedAt, firstAssign), model.DelayAssignMinutes),
		build("DEPART", "تأخر الخروج للزبون", "الفني / الليدر",
			mins(b.ScheduledAt, b.StartedAt), model.DelayDepartMinutes),
		build("EXECUTE", "مدة التنفيذ", "الفني / الليدر",
			mins(b.StartedAt, b.CompletedAt), model.DelayExecuteMinutes),
		build("INVOICE", "تأخر الفوترة", "الليدر",
			mins(b.CompletedAt, firstInvoice), model.DelayInvoiceMinutes),
		build("AUDIT", "تأخر التدقيق", "المحاسب",
			mins(firstInvoice, firstApproved), model.DelayAuditMinutes),
	}
}

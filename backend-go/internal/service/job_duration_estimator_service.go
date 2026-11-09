package service

import (
	"fmt"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// minSamplesForEstimate هو أقل عدد عيّنات تاريخية حقيقية قبل ما يجرؤ النظام يعطي أي
// تقدير — تحت هذا العدد "ما عندنا رأي" فعلاً، ما نخمّن رقم. هذا الرقم نفسه ثابت تقني
// (حجم عيّنة كافي إحصائياً) وليس "دقيقة لكل كاميرا" مفروضة من الإدارة.
const MinSamplesForEstimate = 5

// minSamplesForEstimate يُبقي على الاسم الداخلي القديم للتوافق مع بقية الملف.
const minSamplesForEstimate = MinSamplesForEstimate

// overrunFlagThresholdPercent هو حد التنبيه لو المدة الفعلية تجاوزت المتوقع بهذه
// النسبة — قابل للتعديل، الشركة ما حددت رقم دقيق، هذا افتراضي محافظ.
const overrunFlagThresholdPercent = 50.0

// JobDurationEstimatorService يحسب متوسط وتيرة العمل الحقيقية (دقيقة عمل لكل عنصر
// لكل عامل واحد) من عيّنات مدة العمل التاريخية الحقيقية المسجّلة تلقائياً — لا يوجد
// أي رقم مفروض يدوياً بالكود، النظام "يتعلّم" من job's البيانات الفعلية فقط.
type JobDurationEstimatorService struct {
	samples       *repository.JobDurationSampleRepository
	notifications *repository.NotificationRepository
	complaints    *repository.ComplaintRepository
}

func NewJobDurationEstimatorService(
	samples *repository.JobDurationSampleRepository,
	notifications *repository.NotificationRepository,
	complaints *repository.ComplaintRepository,
) *JobDurationEstimatorService {
	return &JobDurationEstimatorService{samples: samples, notifications: notifications, complaints: complaints}
}

// EstimateExpectedMinutes يرجّع التقدير المتعلَّم لمدة عمل جديدة (بالدقائق) لنفس
// (systemName, jobType)، بناءً على متوسط "دقيقة عمل لكل عنصر لكل عامل" المحسوب من
// كل العيّنات التاريخية المطابقة. لو العدد أقل من minSamplesForEstimate يرجع
// expectedMinutes = nil (بدون رأي فعلاً، مو تخمين).
func (s *JobDurationEstimatorService) EstimateExpectedMinutes(
	systemName, jobType string, itemCount, crewSize int,
) (expectedMinutes *float64, sampleCount int, err error) {
	samples, err := s.samples.ListBySystemAndJobType(systemName, jobType)
	if err != nil {
		return nil, 0, err
	}
	return estimateExpectedMinutesFromSamples(samples, itemCount, crewSize)
}

// estimateExpectedMinutesFromSamples هو المنطق الحسابي الصرف (بدون أي استعلام قاعدة
// بيانات) — مفصول هكذا حتى يكون قابل للاختبار مباشرة بدون DB وهمية. يحسب متوسط
// "دقيقة عمل لكل عنصر لكل عامل" (learnedPacePerItemPerWorker) عبر كل العيّنات
// الصالحة، ثم يطبّقه على الوظيفة الجديدة: expected = (pace × itemCount) / crewSize.
func estimateExpectedMinutesFromSamples(samples []model.JobDurationSample, itemCount, crewSize int) (*float64, int, error) {
	sampleCount := len(samples)
	if sampleCount < minSamplesForEstimate {
		return nil, sampleCount, nil
	}
	if itemCount <= 0 || crewSize <= 0 {
		return nil, sampleCount, nil
	}

	var paceSum float64
	var paceCount int
	for _, sample := range samples {
		if sample.ItemCount <= 0 || sample.CrewSize <= 0 {
			continue // عيّنة غير صالحة رياضياً (قسمة على صفر) — تُتجاهل بدل ما تكسر المتوسط
		}
		pace := float64(sample.DurationMinutes) * float64(sample.CrewSize) / float64(sample.ItemCount)
		paceSum += pace
		paceCount++
	}
	if paceCount < minSamplesForEstimate {
		return nil, sampleCount, nil
	}

	learnedPacePerItemPerWorker := paceSum / float64(paceCount)
	expected := (learnedPacePerItemPerWorker * float64(itemCount)) / float64(crewSize)
	return &expected, sampleCount, nil
}

// RecordSample يحفظ عيّنة زمنية حقيقية جديدة — تُستدعى تلقائياً عند إنجاز عمل حقيقي
// (فاتورة ليدر لحجز تركيب مكتمل / تذكرة صيانة جهاز مُسلَّمة)، مع تخطي التسجيل صراحة
// (وطباعة سبب واضح) لو ما توفرت بيانات زمنية حقيقية كافية لحساب مدة صحيحة.
func (s *JobDurationEstimatorService) RecordSample(sample model.JobDurationSample) error {
	if sample.DurationMinutes < 0 {
		fmt.Printf("job_duration_estimator: تخطي تسجيل العيّنة (systemName=%s jobType=%s) — مدة سالبة/غير منطقية\n", sample.SystemName, sample.JobType)
		return nil
	}
	if sample.ItemCount <= 0 || sample.CrewSize <= 0 {
		fmt.Printf("job_duration_estimator: تخطي تسجيل العيّنة (systemName=%s jobType=%s) — itemCount/crewSize غير صالح\n", sample.SystemName, sample.JobType)
		return nil
	}
	return s.samples.Create(sample)
}

// CheckOverrunAndNotify يقارن مدة عمل فعلية منتهية بالتقدير المتعلَّم (لو موجود
// أصلاً وقت بدء هذا العمل — لو العدد كان أقل من الحد الأدنى وقتها نتجاهل بصمت)،
// ولو المدة الفعلية تجاوزت المتوقع بأكثر من overrunFlagThresholdPercent% ينشئ
// إشعاراً لكل الإدارة (ADMIN) يذكر اسم الموظف والمهمة والوقت المتوقع مقابل الفعلي،
// مع أي شكوى زبون مرتبطة بنفس الحجز إن وجدت.
func (s *JobDurationEstimatorService) CheckOverrunAndNotify(
	systemName, jobType string,
	itemCount, crewSize, actualDurationMinutes int,
	employeeName, jobLabel string,
	bookingID *string,
) error {
	expected, sampleCount, err := s.EstimateExpectedMinutes(systemName, jobType, itemCount, crewSize)
	if err != nil {
		return err
	}
	if expected == nil || sampleCount < minSamplesForEstimate {
		return nil // ما عندنا تقدير موثوق بعد — ما نطلق تنبيه بلا أساس
	}
	if *expected <= 0 {
		return nil
	}
	overrunPercent := (float64(actualDurationMinutes) - *expected) / *expected * 100
	if overrunPercent <= overrunFlagThresholdPercent {
		return nil
	}

	complaintNote := ""
	if bookingID != nil && s.complaints != nil {
		allComplaints, cerr := s.complaints.List()
		if cerr == nil {
			for _, c := range allComplaints {
				if c.BookingID != nil && *c.BookingID == *bookingID {
					complaintNote = fmt.Sprintf(" — يوجد شكوى زبون مرتبطة بهذا الحجز (%s)", model.ComplaintTypeLabels[c.Type])
					break
				}
			}
		}
	}

	message := fmt.Sprintf(
		"تجاوز بالوقت: %s (%s) استغرق %d دقيقة بمهمة %s، بينما المتوقع (متعلَّم من %d عيّنة سابقة) كان %.0f دقيقة (تجاوز %.0f%%)%s",
		employeeName, systemName, actualDurationMinutes, jobLabel, sampleCount, *expected, overrunPercent, complaintNote,
	)
	return s.notifications.CreateForRole("ADMIN", "job_duration_overrun", message)
}

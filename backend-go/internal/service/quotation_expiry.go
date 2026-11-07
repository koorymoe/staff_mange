package service

import (
	"fmt"
	"log"
	"time"

	"staffmange-api/internal/repository"
)

// StartQuotationExpiryReminder (ع) 11-07: عرض سعر جديد أو مرسل قرّب ينتهي (يومين
// أو أقل) ينبّه منشئه مرة وحدة: يتصل بالزبون، أو يمدّد العرض إذا السعر بعده صالح.
// الأسعار تتغير، فعرض قديم ينقبل بسعره القديم يخسّر الشركة.
func StartQuotationExpiryReminder(quotes *repository.QuotationRepository, notif *repository.NotificationRepository) {
	go func() {
		time.Sleep(3 * time.Minute)
		for {
			RemindExpiringQuotations(quotes, notif)
			time.Sleep(3 * time.Hour)
		}
	}()
}

// RemindExpiringQuotations دورة وحدة — يرجّع عدد التنبيهات.
func RemindExpiringQuotations(quotes *repository.QuotationRepository, notif *repository.NotificationRepository) int {
	rows, err := quotes.ClaimExpiringQuotations()
	if err != nil {
		log.Printf("[quotation-expiry] %v", err)
		return 0
	}
	today := time.Now().In(time.FixedZone("Baghdad", 3*3600)).Format("2006-01-02")
	for _, q := range rows {
		day := q.ValidUntil.Format("2006-01-02")
		when := "ينتهي " + day
		if day < today {
			when = "انتهى " + day
		} else if day == today {
			when = "ينتهي اليوم"
		}
		msg := fmt.Sprintf("⏳ عرض السعر %s (%s) %s وبعده ما انقبل — اتصل بالزبون، أو مدّد الصلاحية إذا السعر بعده نفسه.", q.QuotationNumber, q.CustomerName, when)
		if notif != nil && q.CreatorID != "" {
			_ = notif.Create(q.CreatorID, "QUOTATION_EXPIRY", msg)
		}
	}
	return len(rows)
}

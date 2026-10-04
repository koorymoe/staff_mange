package database

// صندوق المراقب أبسط (قرار (ع) 10-04): البنود المعلّقة من المحطات المشالة
// تنسكّر بملاحظة — السجل يبقى، والعدّادات تنزل.
func monitorTrimMigrations() []Migration {
	return []Migration{{
		Version: "0310_monitor_stages_trim",
		SQL: `UPDATE "MonitorReview" SET status = 'OK', note = 'انشال من صندوق المراقب — ترتيب 10-04', "reviewedAt" = now()
		      WHERE status = 'PENDING'
		        AND stage IN ('SOLAR_QUOTED', 'GPS_DEVICE_DONE', 'PROCUREMENT_FULFILLED', 'INVOICE_ADJUSTED', 'QUALITY_VERDICT');`,
	}}
}

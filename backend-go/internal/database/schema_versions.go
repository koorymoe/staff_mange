package database

import "strings"

// Migration هو وحدة ترحيل واحدة قابلة للتتبع: نسخة (version) فريدة + SQL تُنفَّذ
// مرة واحدة فقط (تُسجَّل بجدول "SchemaMigration" بعد نجاحها، انظر migrate.go).
type Migration struct {
	Version string
	SQL     string
}

// migrationNames أسماء وصفية قصيرة لكل عنصر بالترتيب الموجود فعلاً بـ migrations
// (schema_migrations.go) — الترتيب هنا لازم يطابق ترتيب migrations تماماً، هذا الترتيب
// هو تاريخ المشروع ولا يجوز تغييره. الأسماء فقط للتوثيق، ما تأثر بالتنفيذ.
var migrationNames = []string{
	"add_customer_map_latitude",
	"add_customer_map_longitude",
	"add_employee_is_trainee",
	"add_employee_shift_start",
	"add_employee_shift_end",
	"add_booking_materials_ready_at",
	"add_booking_materials_ready_by_id",
	"add_booking_response_minutes",
	"create_vehicle_log",
	"index_vehicle_log_vehicle_id",
	"create_vehicle_incident",
	"index_vehicle_incident_vehicle_id",
	"create_vehicle_monthly_status",
	"create_inventory_check",
	"index_inventory_check_employee_id",
	"index_inventory_check_checked_at",
	"create_quality_issue",
	"index_quality_issue_status",
	"add_gps_device_request_scheduled_at",
	"add_gps_device_request_assigned_technician_id",
	"create_staff_request",
	"create_staff_request_employee",
	"index_staff_request_status",
	"index_staff_request_requester_id",
	"add_project_booking_id",
	"add_employee_role_engineer",
	"add_employee_role_owner",
	"add_employee_role_procurement_admin",
	"add_employee_role_designer",
	"add_employee_role_service_manager",
	"add_employee_status_archived",
	"add_employee_status_deleted",
	"add_employee_status_suspended",
	"add_employee_authz_violations",
	"create_login_audit",
	"index_login_audit_employee_id",
	"index_login_audit_created_at",
	"add_complaint_type",
	"add_complaint_related_employee_id",
	"add_complaint_related_employee_name",
	"index_booking_created_at",
	"index_booking_customer_id",
	"index_booking_status",
	"create_customer_service_tag",
	"index_customer_service_tag_customer_id",
	"index_customer_service_tag_service",
	"create_customer_gps_info",
	"add_kpi_evaluation_cancelled",
	"add_kpi_evaluation_cancelled_at",
	"add_kpi_evaluation_cancelled_by_employee_id",
	"create_kpi_criterion",
	"create_service_manager",
	"index_service_manager_employee_id",
	"index_service_manager_service_id",
	"create_location_ping",
	"index_location_ping_employee_id",
	"create_performance_review",
	"index_performance_review_employee_id",
	"create_quality_follow_up",
	"index_quality_follow_up_booking_id",
	"index_quality_follow_up_status",
	"add_inventory_check_resolved",
	"add_inventory_check_resolved_by_id",
	"add_inventory_check_resolved_at",
	"create_notification",
	"index_notification_employee_id",
	"create_vehicle_daily_rating",
	"index_vehicle_daily_rating_vehicle_id",
	"create_vehicle_wash_rating",
	"index_vehicle_wash_rating_daily_rating_id",
	"index_vehicle_wash_rating_employee_id",
	"add_procurement_request_type",
	"add_vehicle_model",
	"add_vehicle_year",
	"add_vehicle_chassis_number",
	"add_vehicle_engine_number",
	"add_vehicle_fuel_type",
	"add_vehicle_current_odometer",
	"add_vehicle_condition",
	"create_vehicle_document",
	"index_vehicle_document_vehicle_id",
	"create_vehicle_photo",
	"index_vehicle_photo_vehicle_id",
	"create_vehicle_mission",
	"index_vehicle_mission_vehicle_id",
	"index_vehicle_mission_driver_id",
	"index_vehicle_mission_status",
	"create_vehicle_mission_passenger",
	"add_vehicle_log_next_due_odometer",
	"create_vehicle_incident_attachment",
	"index_vehicle_incident_attachment_incident_id",
	"create_vehicle_part",
	"index_vehicle_part_vehicle_id",
	"add_vehicle_incident_location",
	"add_vehicle_incident_driver_id",
	"add_vehicle_incident_people_present",
	"add_vehicle_incident_police_report_number",
	"add_vehicle_incident_repair_cost",
	"add_vehicle_part_cost",
	"create_vehicle_booking",
	"index_vehicle_booking_vehicle_id",
	"index_vehicle_booking_requested_by_id",
	"index_vehicle_booking_status",
	"create_vehicle_mission_rating",
	"index_vehicle_mission_rating_mission_id",
	"create_assistant_conversation",
	"index_assistant_conversation_employee_id_created_at",
	"create_assistant_knowledge",
	"index_assistant_knowledge_topic",
	"index_assistant_knowledge_created_at",
}

// versionedMigrations يبني قائمة الترحيلات المرقّمة والمتتبَّعة من البيانات الخام
// الموجودة أصلاً بـ schema_base.go (baseSchema) وschema_migrations.go (migrations):
//   - 0001_initial_schema: كل عبارات baseSchema مجمّعة كوحدة واحدة (نفس المعاملة
//     التاريخية لها كـ"بنية أساسية" واحدة غير قابلة للتجزئة).
//   - 0002 وما بعدها: كل عبارة من migrations تصير ترحيل مستقل برقمه الخاص، بنفس
//     الترتيب الموجود بالملف تماماً — هذا الترتيب هو تاريخ المشروع ولا يجوز تغييره.
func versionedMigrations() []Migration {
	result := make([]Migration, 0, 1+len(migrations))
	result = append(result, Migration{
		Version: "0001_initial_schema",
		SQL:     strings.Join(baseSchema, ";\n"),
	})
	for i, stmt := range migrations {
		name := "migration"
		if i < len(migrationNames) {
			name = migrationNames[i]
		}
		result = append(result, Migration{
			Version: fmt4Digits(i+2) + "_" + name,
			SQL:     stmt,
		})
	}
	return result
}

// fmt4Digits يرجع الرقم كنص 4 خانات (0002, 0047, ...) بدون استيراد fmt لعملية بسيطة كهذه.
func fmt4Digits(n int) string {
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	for len(s) < 4 {
		s = "0" + s
	}
	return s
}

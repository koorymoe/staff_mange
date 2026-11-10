// الأنواع الخاصة (ENUM) — نفس قاعدة الموديلات: القيمة الجديدة تنضاف هنا.

package models

func init() {
	registerEnum(Enum{Name: "BookingStatus", Values: []string{"IN_PROGRESS", "PENDING", "CONFIRMED", "COMPLETED", "CANCELLED", "WAITING", "PARTIAL"}})
	registerEnum(Enum{Name: "BookingType", Values: []string{"REGULAR", "MAINTENANCE", "INTERNAL", "SOLAR", "SURVEY", "COMPLAINT", "DECOR"}})
	registerEnum(Enum{Name: "BookingUrgency", Values: []string{"ASAP", "BY_PRIORITY", "SPECIFIC_DATE"}})
	registerEnum(Enum{Name: "ComplaintStatus", Values: []string{"NEW", "IN_PROGRESS", "RESOLVED", "CLOSED"}})
	registerEnum(Enum{Name: "EmployeeRole", Values: []string{"ADMIN", "SALES", "HR_COORDINATOR", "TECHNICIAN", "PROJECT_MANAGER", "MONITOR", "FINANCE", "GPS_ADMIN", "GPS_ENGINEER", "QUALITY_ENGINEER", "ENGINEER", "OWNER", "PROCUREMENT_ADMIN", "DESIGNER", "SERVICE_MANAGER", "TECHNICAL", "IT_SUPPORT", "MEDIA", "PUBLIC_RELATIONS", "CAR_WASHER", "SERVICE_DEVELOPER"}})
	registerEnum(Enum{Name: "EmployeeStatus", Values: []string{"ACTIVE", "INACTIVE", "ARCHIVED", "DELETED", "SUSPENDED"}})
	registerEnum(Enum{Name: "EquipmentStatus", Values: []string{"NOT_READY", "READY"}})
	registerEnum(Enum{Name: "ExpenseStatus", Values: []string{"PENDING", "APPROVED", "REJECTED"}})
	registerEnum(Enum{Name: "GpsMaintenanceStatus", Values: []string{"PENDING", "IN_PROGRESS", "COMPLETED"}})
	registerEnum(Enum{Name: "GpsPurchaseType", Values: []string{"DEVICE_SIM", "DEVICE_ONLY"}})
	registerEnum(Enum{Name: "GpsRequestStatus", Values: []string{"PENDING", "APPROVED", "REJECTED", "DELIVERED"}})
	registerEnum(Enum{Name: "GpsSubscriptionStatus", Values: []string{"ACTIVE", "EXPIRED", "CANCELLED"}})
	registerEnum(Enum{Name: "GpsSubscriptionType", Values: []string{"THREE_MONTHS", "SIX_MONTHS", "YEARLY"}})
	registerEnum(Enum{Name: "MaintenanceType", Values: []string{"EXECUTION_ERROR", "DEVICE_ISSUE"}})
	registerEnum(Enum{Name: "MissionStage", Values: []string{"ASSIGNED", "MATERIALS_PREP", "MATERIALS_READY", "EN_ROUTE", "ARRIVED", "WORK_STARTED", "COMPLETED", "STOPPED"}})
	registerEnum(Enum{Name: "ProcurementRequestStatus", Values: []string{"PENDING", "IN_PROGRESS", "FULFILLED", "REJECTED"}})
	registerEnum(Enum{Name: "ProjectPriority", Values: []string{"NORMAL", "URGENT"}})
	registerEnum(Enum{Name: "ProjectSpeed", Values: []string{"URGENT", "INTERNAL", "REGULAR", "DIRECTOR_REQ"}})
	registerEnum(Enum{Name: "QuotationStatus", Values: []string{"NEW", "SENT", "ACCEPTED", "REJECTED"}})
	registerEnum(Enum{Name: "Shift", Values: []string{"MORNING", "EVENING"}})
	registerEnum(Enum{Name: "SimOperator", Values: []string{"ZAIN", "ASIACELL", "KOREK", "OTHER"}})
	registerEnum(Enum{Name: "SimStatus", Values: []string{"AVAILABLE", "IN_USE", "NEEDS_BURN", "BURNED"}})
	registerEnum(Enum{Name: "TechnicianRole", Values: []string{"TECH_1", "TECH_2", "TECH_3"}})
	registerEnum(Enum{Name: "ToolRequestStatus", Values: []string{"PENDING", "APPROVED", "REJECTED", "RETURNED"}})
	registerEnum(Enum{Name: "ToolStatus", Values: []string{"AVAILABLE", "CHECKED_OUT", "DAMAGED", "LOST", "REPAIRING", "RETIRED"}})
	registerEnum(Enum{Name: "WorkStatus", Values: []string{"COMPLETED", "STOPPED"}})
	registerEnum(Enum{Name: "WorkType", Values: []string{"FREE_MAINTENANCE", "PAID"}})
}

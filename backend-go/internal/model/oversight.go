package model

// (دازاي 11-28) «رقابة وتدقيق، تدقيق الحجوزات، رقابة على الموظفين — كلشي اجمعهن بصلاحية وحده
// اريد انطيها لموظف حته يسويلي تقارير»: صلاحية وحدة تفتح المشاهدة بكل شاشات الرقابة والتدقيق.
// **مشاهدة بس**: تنفتح لطلبات القراءة (GET) — أي تعديل أو موافقة يرجع 404 مثل ما ماعنده صلاحية.
// الرواتب والتكاليف مو منها (إلها صلاحيتها).

const OversightView = "oversight_view"

var OversightBundle = []string{"monitoring", "mon_inbox", "mon_reports", "mon_justifications", "mon_attendance", "mon_staff", "mon_matrix", "mon_field", "unit_monitoring", "unit_matrix", "auditing", "audit_issues", "duplicate_review", "leader_invoices_view", "crew_management", "finance", "finance_audit", "quality_control", "complaints", "employee_stats", "staff_stats", "view_bookings", "bookings_archive", "mission_tracking", "work_reports", "performance_review", "kpi_management"}

// (دازاي 11-28) صلاحية «رقابة وتدقيق — مشاهدة»: تفتح بالقائمة كل شاشات الرقابة والتدقيق.
// الخادم يفتحها للقراءة بس (backend-go/internal/model/oversight.go) — أي تعديل يرجع مرفوض.
export const OVERSIGHT_VIEW = 'oversight_view'
export const OVERSIGHT_BUNDLE = ['monitoring', 'mon_inbox', 'mon_reports', 'mon_justifications', 'mon_attendance', 'mon_staff', 'mon_matrix', 'mon_field', 'unit_monitoring', 'unit_matrix', 'auditing', 'audit_issues', 'duplicate_review', 'leader_invoices_view', 'crew_management', 'finance', 'finance_audit', 'quality_control', 'complaints', 'employee_stats', 'staff_stats', 'view_bookings', 'bookings_archive', 'mission_tracking', 'work_reports', 'performance_review', 'kpi_management']

export const expandOversight = (perms: string[]): string[] =>
  perms.includes(OVERSIGHT_VIEW) ? [...new Set([...perms, ...OVERSIGHT_BUNDLE])] : perms

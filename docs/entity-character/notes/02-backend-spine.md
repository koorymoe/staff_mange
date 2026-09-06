# المرحلة 02 — العمود الفقري للخادم والصلاحيات

**الفرع المقروء:** `prod`  
**النطاق:** `cmd/api/main.go` (1589 سطراً)، middleware، config/timeutil/storage/safeguard، و`model/permission.go`.

## ما تحققت منه

- كل مسارات التطبيق تُركّب في `backend-go/cmd/api/main.go`. القائمة أدناه مستخرجة من تعريفات `mux.Handle` الحالية وعددها 510.
- `RequireAuth` يعيد قراءة حالة الحساب ودوره من قاعدة البيانات في كل طلب، ويفحص إبطال الجلسة؛ لا يثق بدور JWT القديم.
- ADMIN وOWNER يتجاوزان حراس الصلاحيات؛ حراس الملكية/الليدر يرجعون لقاعدة البيانات عند الحاجة.
- رفض عملية كتابة متكرر قد يحظر الحساب تلقائياً بعد 5 محاولات خلال 10 دقائق؛ القراءة غير المصرح بها تُسجل ولا تدخل عداد الحظر.
- الملفات تستخدم وسم وصول HMAC عمره 15 دقيقة لأن الصور لا تحمل Authorization؛ التخزين R2 أو قرص محلي بحد رفع 10 MiB وقائمة أنواع بيضاء.
- الوقت المدخل يفسر كتوقيت بغداد ثم يحول UTC.

## تناقض موثق

دليل `docs/الصلاحيات.md` يقول إنه محدث إلى 33 صلاحية، بينما `DefaultPermissions` في `backend-go/internal/model/permission.go` يحتوي 46 صلاحية. هذا **انحراف توثيق مؤكد**، وليس دليلاً بعد على خلل تنفيذ. التأثير: من يعتمد الدليل قد لا يرى صلاحيات الوحدات والصلاحيات الأحدث. الإثبات: عدّ عناصر `{Name: ...}` في المصدر مقابل رأس الدليل.

## فحص الحراس

لا أسجل مساراً محمياً بـ`RequireAuth` فقط كخلل تلقائياً: بعض هذه المسارات يطبق ملكية المورد داخل الـhandler/service. إثبات نقص الحارس يحتاج قراءة منطق المورد في مرحلة 04.

## جدول المسارات → سلسلة الحراسة

| المسار | الحارس كما هو مركب في main |
|---|---|
| `POST /api/auth/login` | `public / handler-specific` |
| `POST /api/files` | `requireAuth` |
| `GET /api/files/token` | `requireAuth` |
| `GET /api/files/` | `public / handler-specific` |
| `GET /api/auth/me` | `requireAuth` |
| `PUT /api/auth/change-password` | `requireAuth, middleware.RequireRole(employeeRepo, notificationRepo, "ADMIN", "OWNER")` |
| `GET /api/employees` | `requireAuth` |
| `GET /api/employees/supervisors` | `requireAuth` |
| `GET /api/letters/addressees` | `requireAuth` |
| `POST /api/letters` | `requireAuth` |
| `GET /api/letters/mine` | `requireAuth` |
| `GET /api/letters` | `requireAuth, requireAdmin` |
| `GET /api/letters/pending-count` | `requireAuth, requireAdmin` |
| `PUT /api/letters/{id}/decide` | `requireAuth, requireAdmin` |
| `GET /api/discipline` | `requireAuth` |
| `GET /api/discipline/events` | `requireAuth` |
| `POST /api/discipline/adjust` | `requireAuth, requireAdmin` |
| `POST /api/discipline/run` | `requireAuth, requireAdmin` |
| `GET /api/employees/archived` | `requireAuth, requireAdmin` |
| `GET /api/security/dashboard` | `requireAuth, requireOwner` |
| `POST /api/security/unlock/{id}` | `requireAuth, requireOwner` |
| `POST /api/security/reset-attempts/{id}` | `requireAuth, requireOwner` |
| `POST /api/security/free-memory` | `requireAuth, requireOwner` |
| `GET /api/employees/match` | `requireAuth` |
| `GET /api/employees/{id}` | `requireAuth` |
| `POST /api/employees` | `requireAuth, requireOwnerAccounts` |
| `PUT /api/employees/{id}` | `requireAuth, requireAdmin` |
| `POST /api/employees/{id}/link-historical` | `requireAuth, requireAdmin` |
| `PUT /api/employees/{id}/skills` | `requireAuth, requireStaffManagement` |
| `GET /api/permissions` | `requireAuth, requireAdmin` |
| `GET /api/employees/{id}/leader-skills` | `requireAuth` |
| `PUT /api/employees/{id}/leader-skills` | `requireAuth, middleware.RequireRole(employeeRepo, notificationRepo, "ADMIN", "OWNER", "HR_COORDINATOR")` |
| `GET /api/permissions/audit-defaults` | `requireAuth, requireAdmin` |
| `GET /api/permissions/role-defaults` | `requireAuth` |
| `GET /api/permissions/employees` | `requireAuth` |
| `GET /api/permissions/employee/{id}` | `requireAuth` |
| `PUT /api/permissions/employee/{id}` | `requireAuth, requireAdmin` |
| `POST /api/permissions/employee/{id}/apply-defaults` | `requireAuth, requireAdmin` |
| `GET /api/services` | `requireAuth` |
| `GET /api/skills` | `requireAuth` |
| `POST /api/services` | `requireAuth, requireContentTech` |
| `POST /api/services/{id}/skills` | `requireAuth, requireContentTech` |
| `DELETE /api/services/{id}` | `requireAuth, requireAdmin` |
| `GET /api/locate` | `requireAuth` |
| `GET /api/bookings/manager-paperwork` | `requireAuth` |
| `PUT /api/services/{id}/manager-paperwork` | `requireAuth, requireAdmin` |
| `GET /api/customers` | `requireAuth` |
| `GET /api/customers/gps` | `requireAuth` |
| `GET /api/customers/lookup` | `requireAuth` |
| `POST /api/customers` | `requireAuth, requireCustomerMgmt` |
| `PUT /api/customers/{id}` | `requireAuth, requireCustomerMgmt` |
| `GET /api/bookings` | `requireAuth` |
| `POST /api/bookings` | `requireAuth` |
| `PUT /api/bookings/{id}/confirm` | `requireAuth, requireBookingCoord` |
| `PUT /api/bookings/{id}/details` | `requireAuth, requireBookingEdit` |
| `PUT /api/bookings/{id}/schedule` | `requireAuth, requireBookingParty` |
| `GET /api/bookings/{id}/schedule-log` | `requireAuth` |
| `GET /api/bookings/archived` | `requireAuth, requireBookingsArchive` |
| `DELETE /api/bookings/{id}` | `requireAuth, requireAdmin` |
| `PUT /api/bookings/{id}/restore` | `requireAuth, requireAdmin` |
| `POST /api/bookings/waiting-reminder-sweep` | `requireAuth, requireAdmin` |
| `PUT /api/auth/command-password` | `requireAuth, middleware.RequireOwner()` |
| `GET /api/bookings/postponed` | `requireAuth, requireCoordinator` |
| `PUT /api/bookings/{id}/postpone` | `requireAuth, requireCoordinator` |
| `PUT /api/bookings/{id}/waiting` | `requireAuth, requireCoordinator` |
| `GET /api/dashboard/today` | `requireAuth` |
| `GET /api/bookings/locate` | `requireAuth` |
| `GET /api/bookings/station-counts` | `requireAuth` |
| `GET /api/bookings/paged` | `requireAuth` |
| `POST /api/bookings/{id}/settle-legacy` | `requireAuth, requireOwner` |
| `PUT /api/bookings/{id}/resume` | `requireAuth, requireCoordinator` |
| `PUT /api/bookings/{id}/assign` | `requireAuth, requireBookingCoord` |
| `DELETE /api/bookings/{id}/assign` | `requireAuth, requireBookingCoord` |
| `POST /api/extra-tasks` | `requireAuth, requireExtraTaskAssign` |
| `GET /api/extra-tasks` | `requireAuth, requireExtraTaskAssign` |
| `PUT /api/extra-tasks/{id}/cancel` | `requireAuth, requireExtraTaskAssign` |
| `GET /api/extra-tasks/mine` | `requireAuth` |
| `GET /api/extra-tasks/mine/count` | `requireAuth` |
| `PUT /api/extra-tasks/{id}/seen` | `requireAuth` |
| `PUT /api/extra-tasks/{id}/start` | `requireAuth` |
| `PUT /api/extra-tasks/{id}/complete` | `requireAuth` |
| `GET /api/ai/signals` | `requireAuth, requireAdmin` |
| `POST /api/ai/process` | `requireAuth, requireAdmin` |
| `POST /api/ai/metrics/recompute` | `requireAuth, requireAdmin` |
| `GET /api/ai/metrics` | `requireAuth, requireAdmin` |
| `GET /api/ai/catalog` | `requireAuth, requireAdmin` |
| `GET /api/ai/work-window` | `requireAuth, requireAdmin` |
| `PUT /api/ai/work-window` | `requireAuth, requireAdmin` |
| `GET /api/bookings/{id}/timeline` | `requireAuth` |
| `PUT /api/bookings/{id}/crew-notes` | `requireAuth, requireBookingCoord` |
| `PUT /api/bookings/{id}/project-notes` | `requireAuth, requireBookingCoord` |
| `PUT /api/bookings/{id}/cancel` | `requireAuth, requireBookingCoord` |
| `GET /api/bookings/stage-bucket` | `requireAuth, requireCoordinator` |
| `GET /api/bookings/stage-bucket-counts` | `requireAuth, requireCoordinator` |
| `PUT /api/bookings/{id}/supervisor` | `requireAuth, requireBookingCoord` |
| `PUT /api/bookings/{id}/start` | `requireAuth, requireBookingParty` |
| `PUT /api/bookings/{id}/arrived` | `requireAuth, requireBookingParty` |
| `PUT /api/bookings/{id}/materials-ready` | `requireAuth, requireBookingParty` |
| `PUT /api/bookings/{id}/stop-work` | `requireAuth, requireBookingParty` |
| `PUT /api/bookings/{id}/resume-work` | `requireAuth, requireBookingParty` |
| `POST /api/bookings/{id}/partial-complete` | `requireAuth, requireBookingParty` |
| `POST /api/bookings/{id}/schedule-continuation` | `requireAuth, requireCoordinator` |
| `GET /api/bookings/{id}/progress` | `requireAuth` |
| `GET /api/bookings/{id}/visits` | `requireAuth` |
| `GET /api/bookings/{id}/suggested-crew` | `requireAuth` |
| `PUT /api/bookings/{id}/type` | `requireAuth, requireAdmin` |
| `PUT /api/bookings/{id}/complete` | `requireAuth, requireBookingParty` |
| `POST /api/bookings/{id}/delete-request` | `requireAuth, requireDeleteRequest` |
| `GET /api/booking-delete-requests` | `requireAuth, requireDeleteApprove` |
| `PUT /api/booking-delete-requests/{id}/decide` | `requireAuth, requireDeleteApprove` |
| `PUT /api/booking-delete-requests/{id}/needs-info` | `requireAuth, requireDeleteApprove` |
| `GET /api/booking-delete-requests/counts` | `requireAuth, requireDeleteApprove` |
| `POST /api/bookings/{id}/coordination-alerts` | `requireAuth, requireCrewManagement` |
| `PUT /api/bookings/{id}/coordination-alerts/resolve` | `requireAuth, requireCrewManagement` |
| `GET /api/bookings/{id}/coordination-alerts` | `requireAuth, requireCrewManagement` |
| `GET /api/coordination-alerts/summaries` | `requireAuth, requireCrewManagement` |
| `GET /api/finance/daily-audit` | `requireAuth, requireVerifyBooking` |
| `PUT /api/bookings/{id}/audit` | `requireAuth, requireVerifyBooking` |
| `GET /api/announcements` | `requireAuth` |
| `POST /api/announcements` | `requireAuth, requireAdmin` |
| `PUT /api/announcements/{id}/active` | `requireAuth, requireAdmin` |
| `DELETE /api/announcements/{id}` | `requireAuth, requireAdmin` |
| `GET /api/audit-issues` | `requireAuth, middleware.RequireRoleOrAnyPermission(permissionRepo, employeeRepo, notificationRepo, []string{"ADMIN", "OWNER", "MONITOR", "QUALITY_ENGINEER", "HR_COORDINATOR", "FINANCE"}, "monitoring", "auditing")` |
| `PUT /api/audit-issues/{id}/resolve` | `requireAuth` |
| `PUT /api/bookings/{id}/verify` | `requireAuth, requireVerifyBooking` |
| `PUT /api/bookings/{id}/unverify` | `requireAuth, requireAdmin` |
| `PUT /api/bookings/{id}/return-to-crew` | `requireAuth, requireProjectMgmt` |
| `PUT /api/bookings/{id}/confirmation-contacted` | `requireAuth, requireCoordinator` |
| `GET /api/bookings/pending-audit` | `requireAuth, requireCrewManagement` |
| `GET /api/bookings/{id}/tool-checks` | `requireAuth, requireCoordinator` |
| `GET /api/cart/booking/{bookingId}` | `requireAuth, requireCartBookingParty` |
| `POST /api/cart/booking/{bookingId}` | `requireAuth, requireCartBookingParty` |
| `PUT /api/cart/{id}` | `requireAuth, requireCartItemParty` |
| `DELETE /api/cart/{id}` | `requireAuth, requireCartItemParty` |
| `GET /api/expenses` | `requireAuth` |
| `POST /api/expenses` | `requireAuth` |
| `PUT /api/expenses/{id}/status` | `requireAuth, requireFinance` |
| `GET /api/inventory/personal` | `requireAuth` |
| `POST /api/inventory/personal` | `requireAuth, requireHROrInventory` |
| `PUT /api/inventory/personal/{id}` | `requireAuth, requireHROrInventory` |
| `DELETE /api/inventory/personal/{id}` | `requireAuth, requireHROrInventory` |
| `GET /api/inventory/tool-events` | `requireAuth` |
| `GET /api/inventory/personal-template` | `requireAuth` |
| `POST /api/inventory/personal-template` | `requireAuth, requireHROrInventory` |
| `DELETE /api/inventory/personal-template/{id}` | `requireAuth, requireHROrInventory` |
| `POST /api/inventory/checks` | `requireAuth` |
| `GET /api/inventory/checks/today` | `requireAuth, requireInventoryView` |
| `GET /api/inventory/checks/mine` | `requireAuth` |
| `GET /api/inventory/checks/booking/{id}` | `requireAuth` |
| `POST /api/inventory/checks/{id}/resolve` | `requireAuth, requireHR` |
| `GET /api/inventory/vehicle` | `requireAuth` |
| `POST /api/inventory/vehicle` | `requireAuth, requireHROrInventory` |
| `PUT /api/inventory/vehicle/{id}` | `requireAuth, requireHROrInventory` |
| `DELETE /api/inventory/vehicle/{id}` | `requireAuth, requireHROrInventory` |
| `GET /api/inventory/vehicle-tool-checks` | `requireAuth` |
| `GET /api/inventory/booking-tool-checks` | `requireAuth` |
| `GET /api/inventory/ondemand` | `requireAuth` |
| `POST /api/inventory/ondemand` | `requireAuth, requireProcurementAdmin` |
| `PUT /api/inventory/ondemand/{id}` | `requireAuth, requireProcurementAdmin` |
| `POST /api/inventory/stock-intake` | `requireAuth, requireProcurementAdmin` |
| `GET /api/inventory/stock-intake` | `requireAuth, requireProcurementAdmin` |
| `GET /api/inventory/requests` | `requireAuth` |
| `POST /api/inventory/requests` | `requireAuth, requireToolRequest` |
| `PUT /api/inventory/requests/{id}/approve` | `requireAuth, requireInventoryApprove` |
| `PUT /api/inventory/requests/{id}/reject` | `requireAuth, requireInventoryApprove` |
| `PUT /api/inventory/requests/{id}/return` | `requireAuth, requireInventoryApprove` |
| `DELETE /api/inventory/requests/{id}` | `requireAuth, requireAdmin` |
| `POST /api/attendance/checkin` | `requireAuth` |
| `POST /api/attendance/checkout` | `requireAuth` |
| `GET /api/attendance/mine` | `requireAuth` |
| `GET /api/attendance/open` | `requireAuth` |
| `GET /api/attendance/today` | `requireAuth, requireMonitor` |
| `GET /api/attendance/today-summary` | `requireAuth, requireMonitor` |
| `GET /api/attendance/employee/{id}` | `requireAuth` |
| `GET /api/attendance/export/employee/{id}` | `requireAuth` |
| `GET /api/attendance/export/today` | `requireAuth, requireMonitor` |
| `PUT /api/attendance/{id}` | `requireAuth, requireMonitor` |
| `GET /api/kpi` | `requireAuth` |
| `GET /api/kpi/employee/{employeeId}` | `requireAuth` |
| `GET /api/kpi/leaderboard/{role}` | `requireAuth` |
| `GET /api/kpi/leaderboard-by-permission/{permission}` | `requireAuth` |
| `POST /api/kpi` | `requireAuth, requireKpi` |
| `DELETE /api/kpi/{id}` | `requireAuth, requireAdmin` |
| `PUT /api/kpi/{id}/cancel` | `requireAuth, requireKpi` |
| `POST /api/employees/{id}/complete-training` | `requireAuth, requireContentTech` |
| `GET /api/notifications` | `requireAuth` |
| `POST /api/notifications/{id}/read` | `requireAuth` |
| `POST /api/notifications/read-all` | `requireAuth` |
| `GET /api/kpi-criteria` | `requireAuth` |
| `POST /api/assistant/ask` | `requireAuth` |
| `POST /api/assistant/manager-chat` | `requireAuth, requireMonitor` |
| `GET /api/assistant/conversations` | `requireAuth, requireOwner` |
| `GET /api/assistant/conversations/employees` | `requireAuth, requireOwner` |
| `GET /api/entity/briefing` | `requireAuth` |
| `GET /api/entity/character/me` | `requireAuth` |
| `POST /api/entity/character/{employeeId}/generate` | `requireAuth, requireAdmin` |
| `POST /api/kpi-criteria` | `requireAuth, requireKpiCriteria` |
| `DELETE /api/kpi-criteria/{id}` | `requireAuth, requireKpiCriteria` |
| `GET /api/smart-kpi/technician/{employeeId}` | `requireAuth` |
| `GET /api/smart-kpi/leaderboard` | `requireAuth` |
| `GET /api/complaints` | `requireAuth` |
| `POST /api/complaints` | `requireAuth` |
| `PUT /api/complaints/{id}` | `requireAuth, requireQuality` |
| `PUT /api/complaints/{id}/contact` | `requireAuth, requireComplaintContact` |
| `PUT /api/complaints/{id}/notes` | `requireAuth, requireComplaintContact` |
| `PUT /api/complaints/{id}/resolve` | `requireAuth, requireQuality` |
| `GET /api/complaints/stats` | `requireAuth` |
| `PUT /api/complaints/{id}/audit` | `requireAuth, middleware.RequireRoleOrAnyPermission(permissionRepo, employeeRepo, notificationRepo, []string{"ADMIN", "OWNER", "MONITOR"}, "monitoring", "auditing")` |
| `GET /api/complaints/{id}/events` | `requireAuth` |
| `GET /api/quality-follow-ups` | `requireAuth, middleware.RequireRoleOrPermission(permissionRepo, employeeRepo, notificationRepo, []string{"ADMIN", "OWNER", "MONITOR", "QUALITY_ENGINEER"}, "quality_control")` |
| `PUT /api/quality-follow-ups/{id}` | `requireAuth, requireQuality` |
| `POST /api/quality-follow-ups/{id}/verdict` | `requireAuth, requireQuality` |
| `POST /api/quality-follow-ups/{id}/inspect` | `requireAuth, requireQuality` |
| `GET /api/training/materials/mine` | `requireAuth` |
| `GET /api/training/assignments/{employeeId}` | `requireAuth` |
| `PUT /api/training/assignments/{employeeId}` | `requireAuth, requireContentTech` |
| `GET /api/training/materials` | `requireAuth` |
| `POST /api/training/materials` | `requireAuth, requireContentTech` |
| `PUT /api/training/materials/{id}` | `requireAuth, requireContentTech` |
| `DELETE /api/training/materials/{id}` | `requireAuth, requireContentTech` |
| `GET /api/missions` | `requireAuth` |
| `GET /api/missions/monitor/live` | `requireAuth` |
| `GET /api/missions/reports/performance` | `requireAuth` |
| `GET /api/missions/my/{employeeId}` | `requireAuth` |
| `GET /api/missions/{id}` | `requireAuth` |
| `POST /api/missions` | `requireAuth, requireBookingCoord` |
| `PUT /api/missions/{id}/stage` | `requireAuth, requireFieldMonitor` |
| `GET /api/projects` | `requireAuth` |
| `GET /api/projects/{id}` | `requireAuth` |
| `POST /api/projects` | `requireAuth, middleware.RequireRoleOrAnyPermission(permissionRepo, employeeRepo, notificationRepo, []string{"PROJECT_MANAGER"}, "project_management", "project_create_only")` |
| `GET /api/projects/delegated-to-me` | `requireAuth` |
| `GET /api/projects/statistics` | `requireAuth, requireProjectMgmt` |
| `GET /api/projects/{id}/delegation-log` | `requireAuth, requireProjectManager` |
| `PUT /api/projects/{id}/delegate` | `requireAuth, requireProjectManager` |
| `PUT /api/projects/{id}` | `requireAuth, allowManagerOrDelegate` |
| `DELETE /api/projects/{id}` | `requireAuth, requireProjectManager` |
| `DELETE /api/projects/{id}/contract` | `requireAuth, requireAdmin` |
| `GET /api/project-work-types` | `requireAuth` |
| `GET /api/project-candidates` | `requireAuth` |
| `GET /api/vip-customers` | `requireAuth, requireAdmin` |
| `GET /api/vip-customers/ids` | `requireAuth` |
| `POST /api/vip-customers` | `requireAuth, requireVipManualAdd` |
| `DELETE /api/vip-customers/{customerId}` | `requireAuth, requireAdmin` |
| `POST /api/project-work-types` | `requireAuth, requireProjectManager` |
| `DELETE /api/project-work-types/{id}` | `requireAuth, requireProjectManager` |
| `GET /api/checklists` | `requireAuth` |
| `POST /api/checklists` | `requireAuth, requireProjectMgmt` |
| `PUT /api/checklists/{id}/photos` | `requireAuth, requireProjectMgmt` |
| `GET /api/tech-showcase` | `requireAuth` |
| `POST /api/tech-showcase` | `requireAuth, requireContentTech` |
| `PUT /api/tech-showcase/{id}/media` | `requireAuth, requireContentTech` |
| `GET /api/exhibitions` | `requireAuth, requireUnitTechnicians` |
| `POST /api/exhibitions` | `requireAuth, requireUnitTechnicians` |
| `PUT /api/exhibitions/{id}/nominate` | `requireAuth, requireAdmin` |
| `PUT /api/exhibitions/{id}/photos` | `requireAuth, requireUnitTechnicians` |
| `PUT /api/exhibitions/{id}/findings` | `requireAuth, requireUnitTechnicians` |
| `POST /api/exhibitions/{id}/report` | `requireAuth, requireUnitTechnicians` |
| `PUT /api/exhibitions/{id}/archive` | `requireAuth, requireAdmin` |
| `GET /api/product-requests` | `requireAuth, requireTechUnitOrProcurement` |
| `POST /api/product-requests` | `requireAuth, requireTechUnitOrProcurement` |
| `PUT /api/product-requests/{id}/approve` | `requireAuth, requireAdmin` |
| `PUT /api/product-requests/{id}/reject` | `requireAuth, requireAdmin` |
| `GET /api/product-procurements` | `requireAuth, requireFund` |
| `POST /api/product-requests/{id}/fulfill` | `requireAuth, requireFund` |
| `PUT /api/product-procurements/{id}/settle` | `requireAuth, requireFund` |
| `GET /api/service-studies` | `requireAuth, requireUnitTechnicians` |
| `POST /api/service-studies` | `requireAuth, requireUnitTechnicians` |
| `PUT /api/service-studies/{id}/assign` | `requireAuth, requireAdmin` |
| `POST /api/service-studies/{id}/reports` | `requireAuth, requireTechUnitOrProcurement` |
| `PUT /api/service-studies/{id}/archive` | `requireAuth, requireAdmin` |
| `GET /api/design-forms` | `requireAuth, requireAdmin` |
| `POST /api/design-forms` | `requireAuth, requireAdmin` |
| `DELETE /api/design-forms/{id}` | `requireAuth, requireAdmin` |
| `GET /api/design-forms/{formId}/submissions` | `requireAuth, requireAdmin` |
| `GET /api/design-forms/{formId}/questions` | `requireAuth, requireAdmin` |
| `POST /api/design-forms/{formId}/questions` | `requireAuth, requireAdmin` |
| `PUT /api/design-form/questions/{id}` | `requireAuth, requireAdmin` |
| `DELETE /api/design-form/questions/{id}` | `requireAuth, requireAdmin` |
| `PUT /api/design-form/questions/reorder` | `requireAuth, requireAdmin` |
| `GET /api/public/design-forms/{token}` | `public / handler-specific` |
| `POST /api/public/design-forms/{token}/submit` | `public / handler-specific` |
| `POST /api/attendance-icon-requests` | `requireAuth` |
| `GET /api/attendance-icon-requests` | `requireAuth, requireAdmin` |
| `PUT /api/attendance-icon-requests/{id}/approve` | `requireAuth, requireAdmin` |
| `PUT /api/attendance-icon-requests/{id}/reject` | `requireAuth, requireAdmin` |
| `POST /api/staff-requests` | `requireAuth, requireProjectMgmtPerm` |
| `GET /api/staff-requests` | `requireAuth` |
| `PUT /api/staff-requests/{id}/status` | `requireAuth, requireHR` |
| `GET /api/service-managers` | `requireAuth` |
| `PUT /api/service-managers` | `requireAuth, requireAdmin` |
| `POST /api/location-pings` | `requireAuth` |
| `GET /api/location-pings/latest` | `requireAuth, requireFieldMonitor` |
| `GET /api/location-pings/path` | `requireAuth, requireFieldMonitor` |
| `POST /api/performance-reviews` | `requireAuth, middleware.RequireAnyPermission(permissionRepo, employeeRepo, notificationRepo, "kpi_management", "performance_review")` |
| `GET /api/performance-reviews` | `requireAuth` |
| `GET /api/performance-reviews/ratable` | `requireAuth` |
| `GET /api/performance-reviews/employee/{employeeId}` | `requireAuth` |
| `GET /api/performance-reviews/my-bookings` | `requireAuth` |
| `GET /api/performance-reviews/evaluator-leaderboard` | `requireAuth, middleware.RequireRoleOrAnyPermission(permissionRepo, employeeRepo, notificationRepo, []string{"ADMIN", "OWNER", "MONITOR"}, "monitoring", "auditing")` |
| `GET /api/procurement` | `requireAuth, requireProcurement` |
| `GET /api/procurement/stats` | `requireAuth, requireProcurement` |
| `POST /api/procurement` | `requireAuth` |
| `PUT /api/procurement/{id}/status` | `requireAuth, requireProcurementAdmin` |
| `PUT /api/procurement/{id}/fulfill` | `requireAuth, requireProcurementAdmin` |
| `GET /api/suppliers/specialties` | `requireAuth` |
| `POST /api/suppliers/specialties` | `requireAuth, requireSuppliersMgmt` |
| `DELETE /api/suppliers/specialties/{id}` | `requireAuth, requireAdmin` |
| `GET /api/geo/resolve-map-link` | `requireAuth` |
| `GET /api/privacy-policy` | `requireAuth` |
| `GET /api/privacy-policy/status` | `requireAuth` |
| `POST /api/privacy-policy/accept` | `requireAuth` |
| `POST /api/privacy-policy` | `requireAuth, requirePrivacyMgmt` |
| `PUT /api/privacy-policy/{id}` | `requireAuth, requirePrivacyMgmt` |
| `DELETE /api/privacy-policy/{id}` | `requireAuth, requirePrivacyMgmt` |
| `GET /api/geo/search` | `requireAuth` |
| `GET /api/suppliers` | `requireAuth` |
| `POST /api/suppliers` | `requireAuth, requireSuppliersMgmt` |
| `PUT /api/suppliers/{id}` | `requireAuth, requireSuppliersMgmt` |
| `DELETE /api/suppliers/{id}` | `requireAuth, requireSuppliersMgmt` |
| `POST /api/suppliers/{id}/rate` | `requireAuth, requireSuppliersMgmt` |
| `GET /api/quotations` | `requireAuth` |
| `GET /api/quotations/{id}` | `requireAuth` |
| `POST /api/quotations` | `requireAuth, allowQuotationOrDelegate` |
| `PUT /api/quotations/{id}` | `requireAuth, allowQuotationOrDelegate` |
| `DELETE /api/quotations/{id}` | `requireAuth, requireAdmin` |
| `GET /api/products` | `requireAuth` |
| `POST /api/products` | `requireAuth, requireContentTech` |
| `PUT /api/products/{id}` | `requireAuth, requireContentTech` |
| `DELETE /api/products/{id}` | `requireAuth, requireAdmin` |
| `GET /api/gps/customers` | `requireAuth, requireGpsData` |
| `POST /api/gps/customers` | `requireAuth, requireGpsSystem` |
| `PUT /api/gps/customers/{id}` | `requireAuth, requireGpsSystem` |
| `GET /api/finance/gps-install-costs` | `requireAuth, requireFinance` |
| `POST /api/leaves` | `requireAuth` |
| `GET /api/leaves/mine` | `requireAuth` |
| `DELETE /api/leaves/{id}` | `requireAuth` |
| `GET /api/leaves/inbox` | `requireAuth` |
| `GET /api/leaves/pending-count` | `requireAuth` |
| `PUT /api/leaves/{id}/preliminary` | `requireAuth` |
| `PUT /api/leaves/{id}/decide` | `requireAuth` |
| `GET /api/dashboard/summary` | `requireAuth` |
| `GET /api/dashboard/today-pulse` | `requireAuth` |
| `GET /api/dashboard/finance-summary` | `requireAuth, middleware.RequireRoleOrAnyPermission(permissionRepo, employeeRepo, notificationRepo, []string{"ADMIN", "OWNER", "FINANCE", "MONITOR"}, "finance", "monitoring"), ` |
| `GET /api/training-programs` | `requireAuth` |
| `POST /api/training-programs` | `requireAuth, requireTrainingManage` |
| `PUT /api/training-programs/{id}` | `requireAuth, requireTrainingManage` |
| `PUT /api/training-programs/{id}/complete` | `requireAuth, requireTrainingManage` |
| `DELETE /api/training-programs/{id}` | `requireAuth, requireTrainingManage` |
| `GET /api/solar/stats` | `requireAuth` |
| `GET /api/solar/low-stock` | `requireAuth` |
| `GET /api/solar/components` | `requireAuth` |
| `POST /api/solar/components` | `requireAuth, requireSolar` |
| `PUT /api/solar/components/{id}` | `requireAuth, requireSolar` |
| `DELETE /api/solar/components/{id}` | `requireAuth, requireSolar` |
| `GET /api/solar/systems` | `requireAuth` |
| `GET /api/solar/systems/{id}` | `requireAuth` |
| `POST /api/solar/systems` | `requireAuth, requireSolar` |
| `PUT /api/solar/systems/{id}` | `requireAuth, requireSolar` |
| `DELETE /api/solar/systems/{id}` | `requireAuth, requireSolar` |
| `POST /api/solar/systems/{id}/process` | `requireAuth, requireSolar` |
| `GET /api/solar/installations` | `requireAuth` |
| `PUT /api/solar/installations/{id}/contacted` | `requireAuth, requireSolar` |
| `GET /api/owner/backups` | `requireAuth, middleware.RequireOwner()` |
| `GET /api/sim/categories` | `requireAuth, simOwner` |
| `GET /api/sim/categories/{id}/exercises` | `requireAuth, simOwner` |
| `GET /api/sim/categories/{id}/lessons` | `requireAuth, simOwner` |
| `GET /api/sim/exercises/{id}` | `requireAuth, simOwner` |
| `POST /api/sim/exercises/{id}/attempts` | `requireAuth, simOwner` |
| `PUT /api/sim/attempts/{id}/progress` | `requireAuth, simOwner` |
| `PUT /api/sim/attempts/{id}/finish` | `requireAuth, simOwner` |
| `GET /api/sim/attempts/mine` | `requireAuth, simOwner` |
| `GET /api/sim/projects` | `requireAuth, simOwner` |
| `GET /api/sim/projects/{id}` | `requireAuth, simOwner` |
| `POST /api/sim/projects` | `requireAuth, simOwner` |
| `DELETE /api/sim/projects/{id}` | `requireAuth, simOwner` |
| `GET /api/sim/review` | `requireAuth, simOwner` |
| `PATCH /api/sim/{kind}/{id}/verify` | `requireAuth, simOwner` |
| `PATCH /api/sim/{kind}/{id}/publish` | `requireAuth, simOwner` |
| `GET /api/funds` | `requireAuth, requireFund` |
| `PUT /api/funds/{id}` | `requireAuth, requireFundAmount` |
| `POST /api/funds/{id}/topup` | `requireAuth, requireFundAmount` |
| `POST /api/funds/disburse` | `requireAuth, requireFund` |
| `GET /api/funds/balances` | `requireAuth, requireFund` |
| `GET /api/funds/transactions` | `requireAuth, requireFund` |
| `PUT /api/funds/settlements/{id}/review` | `requireAuth, requireFund` |
| `PUT /api/funds/settlements/{id}/discharge` | `requireAuth, requireDischarge` |
| `GET /api/funds/discharge-accounts` | `requireAuth, requireFund` |
| `GET /api/funds/my-balance` | `requireAuth` |
| `GET /api/funds/my-transactions` | `requireAuth` |
| `POST /api/funds/settlements` | `requireAuth` |
| `GET /api/gps/sims` | `requireAuth, requireGpsData` |
| `POST /api/gps/sims` | `requireAuth, requireGpsSystem` |
| `PUT /api/gps/sims/{id}` | `requireAuth, requireGpsSystem` |
| `GET /api/gps/sims/available` | `requireAuth, requireGpsSystem` |
| `POST /api/gps/sims/{id}/assign` | `requireAuth, requireGpsSystem` |
| `POST /api/gps/sims/{id}/release` | `requireAuth, requireGpsSystem` |
| `POST /api/gps/sims/{id}/burn` | `requireAuth, requireGpsSystem` |
| `GET /api/gps/subscriptions/follow-up` | `requireAuth, requireGpsOrQuality` |
| `GET /api/gps/devices/{id}/follow-up` | `requireAuth, requireGpsOrQuality` |
| `POST /api/gps/devices/{id}/follow-up` | `requireAuth, requireGpsOrQuality` |
| `GET /api/gps/devices` | `requireAuth, requireGpsData` |
| `POST /api/gps/devices` | `requireAuth, requireGpsSystem` |
| `PUT /api/gps/devices/{id}` | `requireAuth, requireGpsSystem` |
| `GET /api/gps/renewals` | `requireAuth` |
| `POST /api/gps/renewals` | `requireAuth, requireGpsSystem` |
| `PUT /api/gps/renewals/{id}` | `requireAuth, requireGpsSystem` |
| `GET /api/gps/maintenance` | `requireAuth` |
| `POST /api/gps/maintenance` | `requireAuth, requireGpsSystem` |
| `PUT /api/gps/maintenance/{id}` | `requireAuth, requireGpsSystem` |
| `GET /api/gps/settings` | `requireAuth` |
| `PUT /api/gps/settings` | `requireAuth, requireGpsAdmin` |
| `GET /api/gps/stats` | `requireAuth` |
| `GET /api/stats` | `requireAuth, requireAdmin` |
| `GET /api/vehicles/options` | `requireAuth` |
| `GET /api/vehicles` | `requireAuth, requireVehicleMgmt` |
| `POST /api/vehicles` | `requireAuth, requireVehicleMgmt` |
| `GET /api/vehicles/{id}/logs` | `requireAuth, requireVehicleMgmt` |
| `POST /api/vehicles/{id}/logs` | `requireAuth, requireVehicleMgmt` |
| `PUT /api/vehicles/{id}/logs/{logId}` | `requireAuth, requireVehicleMgmt` |
| `DELETE /api/vehicles/{id}/logs/{logId}` | `requireAuth, requireVehicleMgmt` |
| `GET /api/vehicles/{id}/logs/{logId}/receipt-photo` | `requireAuth, requireVehicleMgmt` |
| `GET /api/vehicles/fuel-stats/by-employee` | `requireAuth, requireVehicleMgmt` |
| `GET /api/vehicles/{id}/incidents` | `requireAuth, requireVehicleMgmt` |
| `POST /api/vehicles/{id}/incidents` | `requireAuth, requireVehicleMgmt` |
| `PUT /api/vehicle-incidents/{id}` | `requireAuth, requireVehicleMgmt` |
| `GET /api/vehicles/{id}/monthly-status` | `requireAuth, requireVehicleMgmt` |
| `POST /api/vehicles/{id}/monthly-status` | `requireAuth, requireVehicleMgmt` |
| `POST /api/vehicles/{id}/ratings` | `requireAuth, requireVehicleMgmt` |
| `GET /api/vehicles/{id}/ratings` | `requireAuth, requireVehicleMgmt` |
| `PUT /api/vehicles/{id}` | `requireAuth, requireAdmin` |
| `DELETE /api/vehicles/{id}` | `requireAuth, requireAdmin` |
| `GET /api/vehicles/{id}/documents` | `requireAuth` |
| `POST /api/vehicles/{id}/documents` | `requireAuth, requireVehicleMgmt` |
| `DELETE /api/vehicles/{id}/documents/{docId}` | `requireAuth, requireVehicleMgmt` |
| `GET /api/vehicles/{id}/photos` | `requireAuth` |
| `POST /api/vehicles/{id}/photos` | `requireAuth, requireVehicleMgmt` |
| `DELETE /api/vehicles/{id}/photos/{photoId}` | `requireAuth, requireVehicleMgmt` |
| `POST /api/vehicle-missions` | `requireAuth, requireVehicleMgmt` |
| `PUT /api/vehicle-missions/{id}/end` | `requireAuth, requireVehicleMgmt` |
| `GET /api/vehicle-missions` | `requireAuth, requireVehicleMgmt` |
| `GET /api/vehicle-missions/{id}` | `requireAuth` |
| `POST /api/vehicle-missions/{id}/rating` | `requireAuth, requireVehicleMgmt` |
| `POST /api/vehicle-missions/{id}/tool-check` | `requireAuth, requireLeader` |
| `GET /api/employees/{id}/driver-rating-summary` | `requireAuth` |
| `POST /api/vehicle-bookings` | `requireAuth, requireVehicleMgmt` |
| `PUT /api/vehicle-bookings/{id}/decide` | `requireAuth, requireVehicleMgmt` |
| `PUT /api/vehicle-bookings/{id}/cancel` | `requireAuth, requireVehicleMgmt` |
| `GET /api/vehicle-bookings` | `requireAuth` |
| `GET /api/vehicles/ratings/vehicle-summary` | `requireAuth, requireMonitor` |
| `GET /api/vehicles/ratings/technician-summary` | `requireAuth, requireMonitor` |
| `GET /api/vehicle-incidents/{id}/attachments` | `requireAuth, requireVehicleMgmt` |
| `POST /api/vehicle-incidents/{id}/attachments` | `requireAuth, requireVehicleMgmt` |
| `DELETE /api/vehicle-incidents/{id}/attachments/{attachmentId}` | `requireAuth, requireVehicleMgmt` |
| `GET /api/vehicles/{id}/parts` | `requireAuth, requireVehicleMgmt` |
| `POST /api/vehicles/{id}/parts` | `requireAuth, requireVehicleMgmt` |
| `PUT /api/vehicle-parts/{id}/replace` | `requireAuth, requireVehicleMgmt` |
| `GET /api/vehicles/alerts` | `requireAuth, requireVehicleMgmt` |
| `GET /api/vehicles/dashboard` | `requireAuth, requireVehicleMgmt` |
| `GET /api/vehicles/{id}/expense-summary` | `requireAuth, requireVehicleMgmt` |
| `POST /api/work-reports` | `requireAuth` |
| `GET /api/work-reports` | `requireAuth` |
| `GET /api/quality/issues` | `requireAuth, requireQuality` |
| `POST /api/quality/issues` | `requireAuth, requireQuality` |
| `PUT /api/quality/issues/{id}` | `requireAuth, requireQuality` |
| `GET /api/device-maintenance` | `requireAuth, requireLeader` |
| `POST /api/device-maintenance` | `requireAuth, requireLeader` |
| `PUT /api/device-maintenance/{id}` | `requireAuth, requireLeader` |
| `GET /api/team-inventory/tools` | `requireAuth, requireLeader` |
| `POST /api/team-inventory/tools` | `requireAuth, requireLeader` |
| `GET /api/team-inventory/checks` | `requireAuth, requireLeader` |
| `POST /api/team-inventory/checks` | `requireAuth, requireLeader` |
| `GET /api/free-work-reasons` | `requireAuth` |
| `GET /api/system-price-catalog` | `requireAuth` |
| `GET /api/materials` | `requireAuth` |
| `GET /api/leader-invoices` | `requireAuth, requireLeaderBasket` |
| `GET /api/leader-invoices/{id}` | `requireAuth, requireLeaderBasket` |
| `POST /api/leader-invoices` | `requireAuth, requireLeaderOrServiceManager` |
| `POST /api/leader-invoices/estimate` | `requireAuth, requireExecutionCost` |
| `POST /api/leader-invoices/camera-cost` | `requireAuth, requireExecutionCost` |
| `GET /api/monitor-reviews` | `requireAuth, requireMonitor` |
| `GET /api/monitor-reviews/counts` | `requireAuth, requireMonitor` |
| `POST /api/monitor-reviews/{id}/decide` | `requireAuth, requireMonitor` |
| `GET /api/monitor-desk/counts` | `requireAuth, requireMonitor` |
| `GET /api/network-cost/items` | `requireAuth, requireExecutionCost` |
| `POST /api/network-cost/calculate` | `requireAuth, requireExecutionCost` |
| `GET /api/network-cost/prices` | `requireAuth, requireAdmin` |
| `POST /api/network-cost/prices` | `requireAuth, requireAdmin` |
| `PUT /api/network-cost/prices/{id}` | `requireAuth, requireAdmin` |
| `DELETE /api/network-cost/prices/{id}` | `requireAuth, requireAdmin` |
| `GET /api/leader-invoices/camera-cost/options` | `requireAuth, requireExecutionCost` |
| `GET /api/leader-invoices/by-number` | `requireAuth, requireLeaderBasket` |
| `PUT /api/leader-invoices/{id}/approve` | `requireAuth, requireFinance` |
| `PUT /api/leader-invoices/{id}/audit` | `requireAuth, requireFinance` |
| `PUT /api/leader-invoices/{id}/monitor-request` | `requireAuth, requireFinance` |
| `PUT /api/leader-invoices/{id}/monitor-decide` | `requireAuth, requireMonitor` |
| `PUT /api/leader-invoices/{id}/return` | `requireAuth, middleware.RequireOwnerOnly("إرجاع الفواتير للمحاسب بيد المالك وحده")` |
| `PUT /api/leader-invoices/{id}/revoke` | `requireAuth, requireFinance` |
| `GET /api/leader-invoices/approved-without-number` | `requireAuth, requireFinance` |
| `PUT /api/leader-invoices/{id}/external-number` | `requireAuth, requireFinance` |
| `PUT /api/leader-invoices/{id}/adjust` | `requireAuth, requireFinance` |
| `GET /api/leader-invoices/{id}/adjustments` | `requireAuth` |
| `GET /api/employee-stats/monthly` | `requireAuth, requireAdmin` |
| `GET /api/employee-stats/monthly/export` | `requireAuth, requireAdmin` |
| `GET /api/employee-stats/range` | `requireAuth, requireAdmin` |
| `GET /api/employee-stats/curve/{employeeId}` | `requireAuth, requireAdmin` |
| `GET /api/stats-management/daily` | `requireAuth, requireAdmin` |
| `GET /api/stats-management/weekly` | `requireAuth, requireAdmin` |
| `GET /api/stats-management/projects` | `requireAuth, requireAdmin` |
| `GET /api/stats-management/internal-works` | `requireAuth, requireAdmin` |
| `GET /api/job-duration-estimate` | `requireAuth` |

## متابعة المرحلة

سأستخدم هذا الجدول مع قراءة handlers/services في المرحلة 04 لفصل: مسار عام مقصود، مسار ملكية داخل المنطق، ومسار مكشوف فعلاً. لا تعديل مقترح في هذه الجولة.
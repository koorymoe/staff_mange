package main

import (
	"log"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"

	"staffmange-api/internal/config"
	"staffmange-api/internal/database"
	"staffmange-api/internal/handler"
	"staffmange-api/internal/middleware"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/service"
)

func main() {
	startedAt := time.Now()
	// تحميل ملف .env إذا موجود (بالإنتاج المتغيرات تنجي مباشرة من النظام، فما مشكلة لو الملف مو موجود)
	_ = godotenv.Load()

	cfg := config.Load()
	if len(cfg.JWTSecret) < 16 {
		log.Fatal("JWT_SECRET غير معرّف أو قصير جداً — السيرفر يرفض يشتغل بدون سر تواقيع قوي (16 حرف على الأقل)")
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := database.Migrate(db, cfg.OwnerUsername, cfg.OwnerPassword); err != nil {
		log.Fatalf("failed to run database migrations: %v", err)
	}

	handlerChain := NewHandler(cfg, db, startedAt)

	// http.Server بإعدادات مهلة صريحة بدل http.ListenAndServe الخام — بدونها
	// السيرفر يبقى بلا حماية من هجمات slow-loris (عميل يفتح اتصال ويسحب الرد
	// أو الطلب ببطء متعمد ليشغل الاتصال للأبد). WriteTimeout=30s أكبر بمسافة
	// أمان من أطول عملية شرعية بالسيرفر: نداء Gemini (مساعد ذكي) عنده مهلة
	// عميل HTTP صريحة 20 ثانية بدون أي إعادة محاولة (راجع callGeminiWithTools
	// بـassistant_service.go) — 30 ثانية تكفي 20 ثانية Gemini + وقت معالجة
	// إضافي بدون قطع رد شرعي قيد التقدم.
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handlerChain,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("staffmange-api listening on :%s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// NewHandler يبني كامل شجرة التوجيه (mux) والوسائط (middleware) للـ API — مستخرجة
// من main() حتى يقدر كود الاختبار (httptest) يبنيها ضد قاعدة بيانات حية بدون تشغيل
// السيرفر فعلياً على منفذ شبكة، ويفحص سلوك التوجيه/الصلاحيات الحقيقي عبر HTTP.
func NewHandler(cfg *config.Config, db *sqlx.DB, startedAt time.Time) http.Handler {
	// Repositories
	employeeRepo := repository.NewEmployeeRepository(db)
	loginAuditRepo := repository.NewLoginAuditRepository(db)
	permissionRepo := repository.NewPermissionRepository(db)
	serviceRepo := repository.NewServiceRepository(db)
	customerRepo := repository.NewCustomerRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	qualityFollowUpRepo := repository.NewQualityFollowUpRepository(db)
	cartRepo := repository.NewCartRepository(db)
	expenseRepo := repository.NewExpenseRepository(db)
	inventoryRepo := repository.NewInventoryRepository(db)
	attendanceRepo := repository.NewAttendanceRepository(db)
	kpiRepo := repository.NewKpiRepository(db)
	kpiCriterionRepo := repository.NewKpiCriterionRepository(db)
	smartKpiRepo := repository.NewSmartKpiRepository(db)
	complaintRepo := repository.NewComplaintRepository(db)
	assistantConversationRepo := repository.NewAssistantConversationRepository(db)
	assistantKnowledgeRepo := repository.NewAssistantKnowledgeRepository(db)
	trainingRepo := repository.NewTrainingRepository(db)
	missionRepo := repository.NewMissionRepository(db, bookingRepo)
	projectRepo := repository.NewProjectRepository(db)
	checklistRepo := repository.NewChecklistRepository(db)
	procurementRepo := repository.NewProcurementRepository(db)
	supplierRepo := repository.NewSupplierRepository(db)
	quotationRepo := repository.NewQuotationRepository(db)
	productRepo := repository.NewProductRepository(db)
	systemPriceCatalogRepo := repository.NewSystemPriceCatalogRepository(db)
	materialRepo := repository.NewMaterialRepository(db)
	leaderInvoiceRepo := repository.NewLeaderInvoiceRepository(db)
	employeeCommissionRepo := repository.NewEmployeeCommissionRepository(db)
	gpsRepo := repository.NewGpsRepository(db)
	workReportRepo := repository.NewWorkReportRepository(db)
	statsRepo := repository.NewStatsRepository(db)
	vehicleRepo := repository.NewVehicleRepository(db)
	vehicleMissionRepo := repository.NewVehicleMissionRepository(db)
	vehicleBookingRepo := repository.NewVehicleBookingRepository(db)
	vehicleMissionRatingRepo := repository.NewVehicleMissionRatingRepository(db)
	qualityRepo := repository.NewQualityRepository(db)
	staffRequestRepo := repository.NewStaffRequestRepository(db)
	serviceManagerRepo := repository.NewServiceManagerRepository(db)
	locationPingRepo := repository.NewLocationPingRepository(db)
	performanceReviewRepo := repository.NewPerformanceReviewRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	deviceMaintenanceRepo := repository.NewDeviceMaintenanceRepository(db)
	teamInventoryCheckRepo := repository.NewTeamInventoryCheckRepository(db)
	jobDurationSampleRepo := repository.NewJobDurationSampleRepository(db)

	// Services
	authService := service.NewAuthService(employeeRepo, loginAuditRepo, cfg.JWTSecret)
	employeeService := service.NewEmployeeService(employeeRepo)
	employeeService.SetInventoryRepository(inventoryRepo)
	permissionService := service.NewPermissionService(permissionRepo, employeeRepo)
	serviceCatalogService := service.NewServiceCatalogService(serviceRepo)
	customerService := service.NewCustomerService(customerRepo)
	bookingService := service.NewBookingService(bookingRepo, employeeRepo, customerRepo, qualityFollowUpRepo, notificationRepo, inventoryRepo)
	qualityFollowUpService := service.NewQualityFollowUpService(qualityFollowUpRepo)
	cartService := service.NewCartService(cartRepo)
	expenseService := service.NewExpenseService(expenseRepo)
	inventoryService := service.NewInventoryService(inventoryRepo)
	attendanceService := service.NewAttendanceService(attendanceRepo)
	notificationService := service.NewNotificationService(notificationRepo)
	kpiService := service.NewKpiService(kpiRepo, employeeRepo, notificationRepo)
	kpiCriterionService := service.NewKpiCriterionService(kpiCriterionRepo)
	smartKpiService := service.NewSmartKpiService(smartKpiRepo)
	complaintService := service.NewComplaintService(complaintRepo)
	trainingService := service.NewTrainingService(trainingRepo)
	missionService := service.NewMissionService(missionRepo)
	projectService := service.NewProjectService(projectRepo)
	checklistService := service.NewChecklistService(checklistRepo)
	procurementService := service.NewProcurementService(procurementRepo, permissionRepo)
	supplierService := service.NewSupplierService(supplierRepo)
	quotationService := service.NewQuotationService(quotationRepo)
	productService := service.NewProductService(productRepo)
	gpsService := service.NewGpsService(gpsRepo)
	workReportService := service.NewWorkReportService(workReportRepo)
	statsService := service.NewStatsService(statsRepo)
	vehicleService := service.NewVehicleService(vehicleRepo)
	vehicleMissionService := service.NewVehicleMissionService(vehicleMissionRepo, vehicleRepo)
	vehicleBookingService := service.NewVehicleBookingService(vehicleBookingRepo)
	vehicleMissionRatingService := service.NewVehicleMissionRatingService(vehicleMissionRatingRepo, vehicleMissionRepo)
	qualityService := service.NewQualityService(qualityRepo)
	jobDurationEstimatorService := service.NewJobDurationEstimatorService(jobDurationSampleRepo, notificationRepo, complaintRepo)
	deviceMaintenanceService := service.NewDeviceMaintenanceService(deviceMaintenanceRepo, customerRepo, jobDurationEstimatorService)
	teamInventoryCheckService := service.NewTeamInventoryCheckService(teamInventoryCheckRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	employeeHandler := handler.NewEmployeeHandler(employeeService)
	permissionHandler := handler.NewPermissionHandler(permissionService)
	serviceHandler := handler.NewServiceHandler(serviceCatalogService)
	customerHandler := handler.NewCustomerHandler(customerService)
	bookingHandler := handler.NewBookingHandler(bookingService)
	qualityFollowUpHandler := handler.NewQualityFollowUpHandler(qualityFollowUpService)
	securityHandler := handler.NewSecurityHandler(db, loginAuditRepo, startedAt)
	cartHandler := handler.NewCartHandler(cartService)
	expenseHandler := handler.NewExpenseHandler(expenseService)
	inventoryHandler := handler.NewInventoryHandler(inventoryService)
	attendanceHandler := handler.NewAttendanceHandler(attendanceService, permissionRepo)
	kpiHandler := handler.NewKpiHandler(kpiService)
	notificationHandler := handler.NewNotificationHandler(notificationService)
	assistantService := service.NewAssistantService(cfg.GeminiAPIKey, cfg.GeminiDailyCap, employeeRepo, kpiRepo, performanceReviewRepo, bookingRepo, missionRepo, expenseRepo, gpsRepo, qualityFollowUpRepo, complaintRepo, assistantConversationRepo, assistantKnowledgeRepo, cfg.TutorialsDir)
	assistantHandler := handler.NewAssistantHandler(assistantService, assistantConversationRepo)
	kpiCriterionHandler := handler.NewKpiCriterionHandler(kpiCriterionService)
	smartKpiHandler := handler.NewSmartKpiHandler(smartKpiService)
	complaintHandler := handler.NewComplaintHandler(complaintService)
	trainingHandler := handler.NewTrainingHandler(trainingService)
	missionHandler := handler.NewMissionHandler(missionService)
	projectHandler := handler.NewProjectHandler(projectService)
	checklistHandler := handler.NewChecklistHandler(checklistService)
	procurementHandler := handler.NewProcurementHandler(procurementService)
	supplierHandler := handler.NewSupplierHandler(supplierService)
	quotationHandler := handler.NewQuotationHandler(quotationService)
	productHandler := handler.NewProductHandler(productService)
	leaderInvoiceService := service.NewLeaderInvoiceService(leaderInvoiceRepo, systemPriceCatalogRepo, materialRepo, employeeCommissionRepo, bookingRepo, employeeRepo, jobDurationEstimatorService)
	leaderInvoiceHandler := handler.NewLeaderInvoiceHandler(leaderInvoiceService, systemPriceCatalogRepo, materialRepo)
	jobDurationHandler := handler.NewJobDurationHandler(jobDurationEstimatorService)
	employeeMonthlyStatsService := service.NewEmployeeMonthlyStatsService(employeeRepo, kpiRepo, complaintRepo, leaderInvoiceRepo, bookingRepo, vehicleMissionRatingRepo, employeeCommissionRepo)
	employeeStatsHandler := handler.NewEmployeeStatsHandler(employeeMonthlyStatsService)
	gpsHandler := handler.NewGpsHandler(gpsService)
	workReportHandler := handler.NewWorkReportHandler(workReportService)
	statsHandler := handler.NewStatsHandler(statsService)
	vehicleHandler := handler.NewVehicleHandler(vehicleService)
	vehicleMissionHandler := handler.NewVehicleMissionHandler(vehicleMissionService, vehicleMissionRatingService, vehicleBookingService, inventoryService, employeeRepo)
	vehicleBookingHandler := handler.NewVehicleBookingHandler(vehicleBookingService)
	qualityHandler := handler.NewQualityHandler(qualityService)
	staffRequestHandler := handler.NewStaffRequestHandler(staffRequestRepo)
	serviceManagerHandler := handler.NewServiceManagerHandler(serviceManagerRepo)
	locationPingHandler := handler.NewLocationPingHandler(locationPingRepo)
	performanceReviewService := service.NewPerformanceReviewService(performanceReviewRepo, employeeRepo)
	performanceReviewHandler := handler.NewPerformanceReviewHandler(performanceReviewService)
	deviceMaintenanceHandler := handler.NewDeviceMaintenanceHandler(deviceMaintenanceService)
	teamInventoryCheckHandler := handler.NewTeamInventoryCheckHandler(teamInventoryCheckService)

	requireAuth := middleware.RequireAuth(authService, employeeRepo)
	requireAdmin := middleware.RequireRole(employeeRepo, notificationRepo, "ADMIN")
	// حصراً لحساب المالك (OWNER) — أقوى من الأدمن العادي، ما يشوفها إلا هو
	requireOwner := middleware.RequireRole(employeeRepo, notificationRepo, "OWNER")
	requireFinance := middleware.RequireRole(employeeRepo, notificationRepo, "ADMIN", "FINANCE")
	// تدقيق مبلغ الحجز يعتمد على صلاحية "finance" الممنوحة فعلياً للموظف (مو بس دوره
	// الوظيفي) — المراقب مثلاً عنده هذي الصلاحية افتراضياً ويشوف زر "تدقيق" بالواجهة،
	// فلازم الباك إند يتحقق من نفس الصلاحية بدل دور صارم، وإلا يترفض الطلب ويتسبب
	// بإيقاف حساب الموظف تلقائياً بعد 3 محاولات (حماية أمنية ضد التلاعب بالجلسة).
	requireVerifyBooking := middleware.RequirePermission(permissionRepo, employeeRepo, notificationRepo, "finance")
	requireCoordinator := middleware.RequirePermission(permissionRepo, employeeRepo, notificationRepo, "coordinator")
	requireCrewManagement := middleware.RequirePermission(permissionRepo, employeeRepo, notificationRepo, "crew_management")
	requireHR := middleware.RequireRole(employeeRepo, notificationRepo, "ADMIN", "HR_COORDINATOR")
	// إدارة المخزون (عدة الموظفين الشخصية/أدوات المركبات/العدة القياسية) — مسموحة
	// لأدوار HR/ADMIN كالمعتاد، أو لأي موظف عنده صلاحية "inventory" المخصصة
	// (ممنوحة من صفحة الصلاحيات، مثلاً PROCUREMENT_ADMIN) — توسيع وصول، مو تضييق.
	requireHROrInventory := middleware.RequireRoleOrPermission(permissionRepo, employeeRepo, notificationRepo, []string{"ADMIN", "HR_COORDINATOR"}, "inventory")
	requireLeader := middleware.RequireLeader(employeeRepo, notificationRepo)
	requireInventoryApprove := middleware.RequireRole(employeeRepo, notificationRepo, "ADMIN", "HR_COORDINATOR", "MONITOR")
	// تعديل مهارات موظف يعتمد على صلاحية "staff_management" الممنوحة فعلياً (نفس
	// الصلاحية الي تفتح صفحة "إدارة الكوادر" بالواجهة للمراقب أيضاً) — مو دور
	// وظيفي صارم، وإلا نفس بگ "تدقيق الحسابات" يتكرر: زر يطلع بالواجهة، السيرفر
	// يرفضه، وبعد 3 محاولات ينوقف حساب الموظف تلقائياً.
	requireStaffManagement := middleware.RequirePermission(permissionRepo, employeeRepo, notificationRepo, "staff_management")
	requireMonitor := middleware.RequireRole(employeeRepo, notificationRepo, "ADMIN", "MONITOR")
	requireProjectManager := middleware.RequireRole(employeeRepo, notificationRepo, "ADMIN", "PROJECT_MANAGER")
	requireFieldMonitor := middleware.RequireRole(employeeRepo, notificationRepo, "ADMIN", "HR_COORDINATOR", "MONITOR", "PROJECT_MANAGER")
	requireGpsAdmin := middleware.RequireRole(employeeRepo, notificationRepo, "ADMIN", "GPS_ADMIN")
	requireContentTech := middleware.RequirePermission(permissionRepo, employeeRepo, notificationRepo, "content_technician")
	requireVehicleMgmt := middleware.RequirePermission(permissionRepo, employeeRepo, notificationRepo, "vehicle_management")
	requireProcurement := middleware.RequirePermission(permissionRepo, employeeRepo, notificationRepo, "procurement")
	// توفير المواد وتحديد حالتها يقتصر على إداري الكميات فعلياً (أو الأدمن) — مو أي
	// موظف عنده صلاحية "procurement" العامة (زي الفني/مدير المشاريع الي بس يطلبون مواد).
	requireProcurementAdmin := middleware.RequireRole(employeeRepo, notificationRepo, "ADMIN", "PROCUREMENT_ADMIN")
	requireQuality := middleware.RequirePermission(permissionRepo, employeeRepo, notificationRepo, "quality_control")
	requireProjectMgmtPerm := middleware.RequirePermission(permissionRepo, employeeRepo, notificationRepo, "project_management")
	requireKpi := middleware.RequirePermission(permissionRepo, employeeRepo, notificationRepo, "kpi_management")
	requireKpiCriteria := middleware.RequirePermission(permissionRepo, employeeRepo, notificationRepo, "kpi_criteria_management")
	// إنشاء/تعديل عروض الأسعار يقتصر على من يملك صلاحية "quotation_system" فعلياً
	// (نفس الصلاحية الي تفتح صفحة عروض الأسعار بالواجهة) — مو أي موظف مسجل دخول.
	requireQuotationSystem := middleware.RequirePermission(permissionRepo, employeeRepo, notificationRepo, "quotation_system")
	// إنشاء/تعديل بيانات نظام GPS (عملاء/شرائح/أجهزة/تجديد/صيانة) يقتصر على من
	// يملك صلاحية "gps_system" فعلياً (نفس الصلاحية الي تفتح كل صفحات نظام GPS
	// بالواجهة) — مو أي موظف مسجل دخول.
	requireGpsSystem := middleware.RequirePermission(permissionRepo, employeeRepo, notificationRepo, "gps_system")

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// حماية ضد محاولات تخمين كلمة السر المتكررة: أقصى 8 محاولات دخول بالدقيقة
	// من نفس عنوان IP — هذا بالضبط النوع اللي سبب حظر IP السيرفر من Hetzner
	// كإجراء حماية تلقائي ضدهم لما شافوا محاولات دخول متكررة سريعة.
	requireLoginRateLimit := middleware.RateLimit(8, time.Minute)
	mux.Handle("POST /api/auth/login", requireLoginRateLimit(http.HandlerFunc(authHandler.Login)))
	mux.Handle("GET /api/auth/me", middleware.Chain(http.HandlerFunc(authHandler.Me), requireAuth))

	// موظفين — القراءة تحتاج تسجيل دخول فقط، الإنشاء/التعديل الحساس محمي بدور ADMIN
	mux.Handle("GET /api/employees", middleware.Chain(http.HandlerFunc(employeeHandler.List), requireAuth))
	mux.Handle("GET /api/employees/supervisors", middleware.Chain(http.HandlerFunc(employeeHandler.Supervisors), requireAuth))
	mux.Handle("GET /api/employees/archived", middleware.Chain(http.HandlerFunc(employeeHandler.ListArchived), requireAuth, requireAdmin))
	mux.Handle("GET /api/security/dashboard", middleware.Chain(http.HandlerFunc(securityHandler.Dashboard), requireAuth, requireOwner))
	mux.Handle("POST /api/security/free-memory", middleware.Chain(http.HandlerFunc(securityHandler.FreeMemory), requireAuth, requireOwner))
	mux.Handle("GET /api/employees/match", middleware.Chain(http.HandlerFunc(employeeHandler.Match), requireAuth))
	mux.Handle("GET /api/employees/{id}", middleware.Chain(http.HandlerFunc(employeeHandler.Get), requireAuth))
	mux.Handle("POST /api/employees", middleware.Chain(http.HandlerFunc(employeeHandler.Create), requireAuth, requireAdmin))
	mux.Handle("PUT /api/employees/{id}", middleware.Chain(http.HandlerFunc(employeeHandler.Update), requireAuth, requireAdmin))
	mux.Handle("POST /api/employees/{id}/link-historical", middleware.Chain(http.HandlerFunc(employeeHandler.LinkHistoricalRecords), requireAuth, requireAdmin))
	mux.Handle("PUT /api/employees/{id}/skills", middleware.Chain(http.HandlerFunc(employeeHandler.SetSkills), requireAuth, requireStaffManagement))

	// الصلاحيات — العرض متاح لأي مسجل دخول، التعديل والتطبيق التلقائي محصور بمدير النظام فقط
	mux.Handle("GET /api/permissions", middleware.Chain(http.HandlerFunc(permissionHandler.ListAll), requireAuth))
	mux.Handle("GET /api/permissions/role-defaults", middleware.Chain(http.HandlerFunc(permissionHandler.RoleDefaults), requireAuth))
	mux.Handle("GET /api/permissions/employee/{id}", middleware.Chain(http.HandlerFunc(permissionHandler.ListForEmployee), requireAuth))
	mux.Handle("PUT /api/permissions/employee/{id}", middleware.Chain(http.HandlerFunc(permissionHandler.SetForEmployee), requireAuth, requireAdmin))
	mux.Handle("POST /api/permissions/employee/{id}/apply-defaults", middleware.Chain(http.HandlerFunc(permissionHandler.ApplyDefaults), requireAuth, requireAdmin))

	// الخدمات والمهارات — القراءة لأي مسجل دخول، الإضافة لمدير النظام فقط
	mux.Handle("GET /api/services", middleware.Chain(http.HandlerFunc(serviceHandler.List), requireAuth))
	mux.Handle("POST /api/services", middleware.Chain(http.HandlerFunc(serviceHandler.Create), requireAuth, requireContentTech))
	mux.Handle("POST /api/services/{id}/skills", middleware.Chain(http.HandlerFunc(serviceHandler.CreateSkill), requireAuth, requireContentTech))
	mux.Handle("DELETE /api/services/{id}", middleware.Chain(http.HandlerFunc(serviceHandler.Delete), requireAuth, requireAdmin))

	// العملاء — أي مسجل دخول يقدر يبحث وينشئ عميل (يطابق سلوك المبيعات بالباك إند القديم)
	mux.Handle("GET /api/customers", middleware.Chain(http.HandlerFunc(customerHandler.List), requireAuth))
	mux.Handle("GET /api/customers/gps", middleware.Chain(http.HandlerFunc(customerHandler.ListGps), requireAuth))
	mux.Handle("GET /api/customers/lookup", middleware.Chain(http.HandlerFunc(customerHandler.Lookup), requireAuth))
	mux.Handle("POST /api/customers", middleware.Chain(http.HandlerFunc(customerHandler.FindOrCreate), requireAuth))
	mux.Handle("PUT /api/customers/{id}", middleware.Chain(http.HandlerFunc(customerHandler.Update), requireAuth))

	// الحجوزات — دورة حياة الحجز الكاملة، كل خطوة تتطلب تسجيل دخول فقط (الصلاحية الدقيقة تُفرض بالواجهة حالياً)
	mux.Handle("GET /api/bookings", middleware.Chain(http.HandlerFunc(bookingHandler.List), requireAuth))
	mux.Handle("POST /api/bookings", middleware.Chain(http.HandlerFunc(bookingHandler.Create), requireAuth))
	mux.Handle("PUT /api/bookings/{id}/confirm", middleware.Chain(http.HandlerFunc(bookingHandler.Confirm), requireAuth))
	mux.Handle("PUT /api/bookings/{id}/details", middleware.Chain(http.HandlerFunc(bookingHandler.UpdateDetails), requireAuth))
	mux.Handle("PUT /api/bookings/{id}/schedule", middleware.Chain(http.HandlerFunc(bookingHandler.Schedule), requireAuth))
	mux.Handle("GET /api/bookings/{id}/schedule-log", middleware.Chain(http.HandlerFunc(bookingHandler.ScheduleLog), requireAuth))
	mux.Handle("PUT /api/bookings/{id}/assign", middleware.Chain(http.HandlerFunc(bookingHandler.Assign), requireAuth))
	mux.Handle("PUT /api/bookings/{id}/supervisor", middleware.Chain(http.HandlerFunc(bookingHandler.Supervisor), requireAuth))
	mux.Handle("PUT /api/bookings/{id}/start", middleware.Chain(http.HandlerFunc(bookingHandler.Start), requireAuth))
	mux.Handle("PUT /api/bookings/{id}/arrived", middleware.Chain(http.HandlerFunc(bookingHandler.MarkArrived), requireAuth))
	mux.Handle("PUT /api/bookings/{id}/materials-ready", middleware.Chain(http.HandlerFunc(bookingHandler.SetMaterialsReady), requireAuth))
	mux.Handle("PUT /api/bookings/{id}/complete", middleware.Chain(http.HandlerFunc(bookingHandler.Complete), requireAuth))
	mux.Handle("PUT /api/bookings/{id}/verify", middleware.Chain(http.HandlerFunc(bookingHandler.Verify), requireAuth, requireVerifyBooking))
	// "تم" الإداري بعد تواصله فعلياً مع الزبون — خطوة سابقة ومنفصلة عن التثبيت
	// نفسه (نفس صلاحية تنسيق الحجوزات coordinator المستخدمة أصلاً بـCoordinator.tsx).
	mux.Handle("PUT /api/bookings/{id}/confirmation-contacted", middleware.Chain(http.HandlerFunc(bookingHandler.MarkConfirmationContacted), requireAuth, requireCoordinator))
	// تدقيق المراقب على الحجوزات الموجّهة قبل التثبيت (crew_management، صلاحية جديدة
	// يقدر الأدمن يمنحها لأي موظف مراقب من صفحة الصلاحيات).
	mux.Handle("GET /api/bookings/pending-audit", middleware.Chain(http.HandlerFunc(bookingHandler.PendingAudit), requireAuth, requireCrewManagement))
	mux.Handle("GET /api/bookings/{id}/tool-checks", middleware.Chain(http.HandlerFunc(bookingHandler.ToolChecks), requireAuth, requireCoordinator))

	// سلة الحجز
	mux.Handle("GET /api/cart/booking/{bookingId}", middleware.Chain(http.HandlerFunc(cartHandler.ListForBooking), requireAuth))
	mux.Handle("POST /api/cart/booking/{bookingId}", middleware.Chain(http.HandlerFunc(cartHandler.Create), requireAuth))
	mux.Handle("PUT /api/cart/{id}", middleware.Chain(http.HandlerFunc(cartHandler.Update), requireAuth))
	mux.Handle("DELETE /api/cart/{id}", middleware.Chain(http.HandlerFunc(cartHandler.Delete), requireAuth))

	// المصاريف — أي موظف يقدر يرسل مصروف، الموافقة/الرفض للمحاسب ومدير النظام فقط
	mux.Handle("GET /api/expenses", middleware.Chain(http.HandlerFunc(expenseHandler.List), requireAuth))
	mux.Handle("POST /api/expenses", middleware.Chain(http.HandlerFunc(expenseHandler.Create), requireAuth))
	mux.Handle("PUT /api/expenses/{id}/status", middleware.Chain(http.HandlerFunc(expenseHandler.UpdateStatus), requireAuth, requireFinance))

	// المخزون — أدوات شخصية / مركبات / أدوات مشتركة / طلبات الأدوات
	mux.Handle("GET /api/inventory/personal", middleware.Chain(http.HandlerFunc(inventoryHandler.ListPersonalTools), requireAuth))
	mux.Handle("POST /api/inventory/personal", middleware.Chain(http.HandlerFunc(inventoryHandler.CreatePersonalTool), requireAuth, requireHROrInventory))
	mux.Handle("PUT /api/inventory/personal/{id}", middleware.Chain(http.HandlerFunc(inventoryHandler.UpdatePersonalTool), requireAuth, requireHROrInventory))
	mux.Handle("DELETE /api/inventory/personal/{id}", middleware.Chain(http.HandlerFunc(inventoryHandler.DeletePersonalTool), requireAuth, requireHROrInventory))

	// العدة القياسية (PersonalToolTemplateItem) — القراءة مفتوحة لأي موظف، الإضافة/الحذف
	// لمن عنده صلاحية "inventory" أو HR/ADMIN (نفس requireHROrInventory).
	mux.Handle("GET /api/inventory/personal-template", middleware.Chain(http.HandlerFunc(inventoryHandler.ListPersonalToolTemplateItems), requireAuth))
	mux.Handle("POST /api/inventory/personal-template", middleware.Chain(http.HandlerFunc(inventoryHandler.CreatePersonalToolTemplateItem), requireAuth, requireHROrInventory))
	mux.Handle("DELETE /api/inventory/personal-template/{id}", middleware.Chain(http.HandlerFunc(inventoryHandler.DeletePersonalToolTemplateItem), requireAuth, requireHROrInventory))

	// جرد يومي: الموظف يؤكد جرد عدته الخاصة، الإداري يشوف نتائج اليوم لكل الموظفين
	mux.Handle("POST /api/inventory/checks", middleware.Chain(http.HandlerFunc(inventoryHandler.CreateInventoryCheck), requireAuth))
	mux.Handle("GET /api/inventory/checks/today", middleware.Chain(http.HandlerFunc(inventoryHandler.TodaysInventoryChecks), requireAuth))
	mux.Handle("POST /api/inventory/checks/{id}/resolve", middleware.Chain(http.HandlerFunc(inventoryHandler.ResolveInventoryCheck), requireAuth, requireHR))

	mux.Handle("GET /api/inventory/vehicle", middleware.Chain(http.HandlerFunc(inventoryHandler.ListVehicleTools), requireAuth))
	mux.Handle("POST /api/inventory/vehicle", middleware.Chain(http.HandlerFunc(inventoryHandler.CreateVehicleTool), requireAuth))
	// تعديل/حذف أداة مركبة موجودة يقتصر على إدارة الكوادر (نفس صلاحية أدوات
	// الأدوات الشخصية requireHR) — الإنشاء يبقى مفتوح لأي موظف (نفس السلوك القديم).
	mux.Handle("PUT /api/inventory/vehicle/{id}", middleware.Chain(http.HandlerFunc(inventoryHandler.UpdateVehicleTool), requireAuth, requireHROrInventory))
	mux.Handle("DELETE /api/inventory/vehicle/{id}", middleware.Chain(http.HandlerFunc(inventoryHandler.DeleteVehicleTool), requireAuth, requireHROrInventory))
	mux.Handle("GET /api/inventory/vehicle-tool-checks", middleware.Chain(http.HandlerFunc(inventoryHandler.ListVehicleToolChecks), requireAuth))
	mux.Handle("GET /api/inventory/booking-tool-checks", middleware.Chain(http.HandlerFunc(inventoryHandler.ListAllBookingToolChecks), requireAuth))

	mux.Handle("GET /api/inventory/ondemand", middleware.Chain(http.HandlerFunc(inventoryHandler.ListOnDemandTools), requireAuth))
	mux.Handle("POST /api/inventory/ondemand", middleware.Chain(http.HandlerFunc(inventoryHandler.CreateOnDemandTool), requireAuth))
	// تعديل أداة "حسب الحاجة" يقتصر على الأدمن أو إداري الكميات (نفس canManageOnDemand
	// بالواجهة: isAdmin || PROCUREMENT_ADMIN) — يطابق requireProcurementAdmin الموجود.
	mux.Handle("PUT /api/inventory/ondemand/{id}", middleware.Chain(http.HandlerFunc(inventoryHandler.UpdateOnDemandTool), requireAuth, requireProcurementAdmin))

	mux.Handle("GET /api/inventory/requests", middleware.Chain(http.HandlerFunc(inventoryHandler.ListToolRequests), requireAuth))
	// طلب أداة "حسب الحاجة" مقصور على الليدر فقط (isLeader فريش من قاعدة البيانات) —
	// الموظف العادي يبقى يشوف حالة طلباته (GET) بس ما يقدر ينشئ طلب جديد.
	mux.Handle("POST /api/inventory/requests", middleware.Chain(http.HandlerFunc(inventoryHandler.CreateToolRequest), requireAuth, requireLeader))
	mux.Handle("PUT /api/inventory/requests/{id}/approve", middleware.Chain(http.HandlerFunc(inventoryHandler.ApproveToolRequest), requireAuth, requireInventoryApprove))
	mux.Handle("PUT /api/inventory/requests/{id}/reject", middleware.Chain(http.HandlerFunc(inventoryHandler.RejectToolRequest), requireAuth, requireInventoryApprove))
	mux.Handle("PUT /api/inventory/requests/{id}/return", middleware.Chain(http.HandlerFunc(inventoryHandler.ReturnToolRequest), requireAuth))
	mux.Handle("DELETE /api/inventory/requests/{id}", middleware.Chain(http.HandlerFunc(inventoryHandler.DeleteToolRequest), requireAuth, requireInventoryApprove))

	// تقييم الأداء اليدوي (KPI)
	mux.Handle("POST /api/attendance/checkin", middleware.Chain(http.HandlerFunc(attendanceHandler.CheckIn), requireAuth))
	mux.Handle("POST /api/attendance/checkout", middleware.Chain(http.HandlerFunc(attendanceHandler.CheckOut), requireAuth))
	mux.Handle("GET /api/attendance/mine", middleware.Chain(http.HandlerFunc(attendanceHandler.Mine), requireAuth))
	mux.Handle("GET /api/attendance/open", middleware.Chain(http.HandlerFunc(attendanceHandler.OpenSession), requireAuth))
	mux.Handle("GET /api/attendance/today", middleware.Chain(http.HandlerFunc(attendanceHandler.Today), requireAuth, requireMonitor))
	mux.Handle("GET /api/attendance/today-summary", middleware.Chain(http.HandlerFunc(attendanceHandler.TodaySummary), requireAuth, requireMonitor))
	mux.Handle("GET /api/attendance/employee/{id}", middleware.Chain(http.HandlerFunc(attendanceHandler.MonthlyReport), requireAuth))
	mux.Handle("GET /api/attendance/export/employee/{id}", middleware.Chain(http.HandlerFunc(attendanceHandler.ExportEmployeeMonth), requireAuth))
	mux.Handle("GET /api/attendance/export/today", middleware.Chain(http.HandlerFunc(attendanceHandler.ExportToday), requireAuth, requireMonitor))
	mux.Handle("PUT /api/attendance/{id}", middleware.Chain(http.HandlerFunc(attendanceHandler.Correct), requireAuth, requireMonitor))

	mux.Handle("GET /api/kpi", middleware.Chain(http.HandlerFunc(kpiHandler.List), requireAuth))
	mux.Handle("GET /api/kpi/employee/{employeeId}", middleware.Chain(http.HandlerFunc(kpiHandler.ListForEmployee), requireAuth))
	mux.Handle("GET /api/kpi/leaderboard/{role}", middleware.Chain(http.HandlerFunc(kpiHandler.RoleLeaderboard), requireAuth))
	mux.Handle("POST /api/kpi", middleware.Chain(http.HandlerFunc(kpiHandler.Create), requireAuth, requireKpi))
	mux.Handle("DELETE /api/kpi/{id}", middleware.Chain(http.HandlerFunc(kpiHandler.Delete), requireAuth, requireAdmin))
	mux.Handle("PUT /api/kpi/{id}/cancel", middleware.Chain(http.HandlerFunc(kpiHandler.Cancel), requireAuth, requireKpi))
	mux.Handle("POST /api/employees/{id}/complete-training", middleware.Chain(http.HandlerFunc(kpiHandler.CompleteTraining), requireAuth, requireContentTech))

	mux.Handle("GET /api/notifications", middleware.Chain(http.HandlerFunc(notificationHandler.List), requireAuth))
	mux.Handle("POST /api/notifications/{id}/read", middleware.Chain(http.HandlerFunc(notificationHandler.MarkRead), requireAuth))
	mux.Handle("POST /api/notifications/read-all", middleware.Chain(http.HandlerFunc(notificationHandler.MarkAllRead), requireAuth))
	mux.Handle("GET /api/kpi-criteria", middleware.Chain(http.HandlerFunc(kpiCriterionHandler.List), requireAuth))
	mux.Handle("POST /api/assistant/ask", middleware.Chain(http.HandlerFunc(assistantHandler.Ask), requireAuth))
	mux.Handle("POST /api/assistant/manager-chat", middleware.Chain(http.HandlerFunc(assistantHandler.ManagerChat), requireAuth, requireMonitor))
	mux.Handle("GET /api/assistant/conversations", middleware.Chain(http.HandlerFunc(assistantHandler.ListConversations), requireAuth, requireOwner))
	mux.Handle("GET /api/assistant/conversations/employees", middleware.Chain(http.HandlerFunc(assistantHandler.ListConversationEmployees), requireAuth, requireOwner))
	mux.Handle("POST /api/kpi-criteria", middleware.Chain(http.HandlerFunc(kpiCriterionHandler.Create), requireAuth, requireKpiCriteria))
	mux.Handle("DELETE /api/kpi-criteria/{id}", middleware.Chain(http.HandlerFunc(kpiCriterionHandler.Delete), requireAuth, requireKpiCriteria))

	// تقييم الأداء التلقائي (Smart KPI) — الرانك الأسبوعي/الشهري للفنيين
	mux.Handle("GET /api/smart-kpi/technician/{employeeId}", middleware.Chain(http.HandlerFunc(smartKpiHandler.Technician), requireAuth))
	mux.Handle("GET /api/smart-kpi/leaderboard", middleware.Chain(http.HandlerFunc(smartKpiHandler.Leaderboard), requireAuth))

	// الشكاوى
	mux.Handle("GET /api/complaints", middleware.Chain(http.HandlerFunc(complaintHandler.List), requireAuth))
	mux.Handle("POST /api/complaints", middleware.Chain(http.HandlerFunc(complaintHandler.Create), requireAuth))
	// تحديث حالة الشكوى وحلّها يقتصران على من يملك صلاحية "quality_control" (نفس
	// أدوار المتابعة بالواجهة: مهندس الجودة/المراقب/الأدمن) — لا أي موظف مسجل دخول
	// (مثلاً موظف مبيعات عنده صلاحية "complaints" بس لتسجيل شكوى جديدة فقط).
	mux.Handle("PUT /api/complaints/{id}", middleware.Chain(http.HandlerFunc(complaintHandler.Update), requireAuth, requireQuality))
	mux.Handle("PUT /api/complaints/{id}/resolve", middleware.Chain(http.HandlerFunc(complaintHandler.Resolve), requireAuth, requireQuality))
	mux.Handle("GET /api/complaints/stats", middleware.Chain(http.HandlerFunc(complaintHandler.Stats), requireAuth))
	mux.Handle("GET /api/quality-follow-ups", middleware.Chain(http.HandlerFunc(qualityFollowUpHandler.List), requireAuth, requireQuality))
	mux.Handle("PUT /api/quality-follow-ups/{id}", middleware.Chain(http.HandlerFunc(qualityFollowUpHandler.Update), requireAuth, requireQuality))

	// التدريب — عرض متاح لأي مسجل دخول، التعيين وإدارة المواد لمدير النظام فقط
	mux.Handle("GET /api/training/materials/mine", middleware.Chain(http.HandlerFunc(trainingHandler.MaterialsMine), requireAuth))
	mux.Handle("GET /api/training/assignments/{employeeId}", middleware.Chain(http.HandlerFunc(trainingHandler.Assignments), requireAuth))
	mux.Handle("PUT /api/training/assignments/{employeeId}", middleware.Chain(http.HandlerFunc(trainingHandler.SetAssignments), requireAuth, requireContentTech))
	mux.Handle("GET /api/training/materials", middleware.Chain(http.HandlerFunc(trainingHandler.ListMaterials), requireAuth))
	mux.Handle("POST /api/training/materials", middleware.Chain(http.HandlerFunc(trainingHandler.CreateMaterial), requireAuth, requireContentTech))
	mux.Handle("PUT /api/training/materials/{id}", middleware.Chain(http.HandlerFunc(trainingHandler.UpdateMaterial), requireAuth, requireContentTech))
	mux.Handle("DELETE /api/training/materials/{id}", middleware.Chain(http.HandlerFunc(trainingHandler.DeleteMaterial), requireAuth, requireContentTech))

	// تتبع المهام (missions)
	mux.Handle("GET /api/missions", middleware.Chain(http.HandlerFunc(missionHandler.List), requireAuth))
	mux.Handle("GET /api/missions/monitor/live", middleware.Chain(http.HandlerFunc(missionHandler.MonitorLive), requireAuth))
	mux.Handle("GET /api/missions/reports/performance", middleware.Chain(http.HandlerFunc(missionHandler.PerformanceReport), requireAuth))
	mux.Handle("GET /api/missions/my/{employeeId}", middleware.Chain(http.HandlerFunc(missionHandler.ListForEmployee), requireAuth))
	mux.Handle("GET /api/missions/{id}", middleware.Chain(http.HandlerFunc(missionHandler.Get), requireAuth))
	mux.Handle("POST /api/missions", middleware.Chain(http.HandlerFunc(missionHandler.Create), requireAuth))
	mux.Handle("PUT /api/missions/{id}/stage", middleware.Chain(http.HandlerFunc(missionHandler.UpdateStage), requireAuth))

	// إدارة المشاريع (projects)
	mux.Handle("GET /api/projects", middleware.Chain(http.HandlerFunc(projectHandler.List), requireAuth))
	// إنشاء مشروع جديد يتطلب نفس دور مدير المشاريع/الأدمن المطلوب للتعديل والحذف
	// تحته مباشرة — كان مفتوح غلط لأي موظف مسجل دخول فقط (requireAuth بدون requireProjectManager)،
	// عدم اتساق مع PUT/DELETE على نفس المورد.
	mux.Handle("POST /api/projects", middleware.Chain(http.HandlerFunc(projectHandler.Create), requireAuth, requireProjectManager))
	mux.Handle("PUT /api/projects/{id}", middleware.Chain(http.HandlerFunc(projectHandler.Update), requireAuth, requireProjectManager))
	mux.Handle("DELETE /api/projects/{id}", middleware.Chain(http.HandlerFunc(projectHandler.Delete), requireAuth, requireProjectManager))

	// الكشوفات: فورمات فارغة يطبعها المهندس، يمليها بالموقع، وبعدين يرفع صور
	// الفورمة المالية — أي موظف مسجل دخول يقدر ينشئ/يرفع (مو حصراً مدير مشاريع).
	mux.Handle("GET /api/checklists", middleware.Chain(http.HandlerFunc(checklistHandler.List), requireAuth))
	mux.Handle("POST /api/checklists", middleware.Chain(http.HandlerFunc(checklistHandler.Create), requireAuth))
	mux.Handle("PUT /api/checklists/{id}/photos", middleware.Chain(http.HandlerFunc(checklistHandler.AddPhotos), requireAuth))

	// طلبات الكادر — مدير المشاريع (أو صاحب صلاحية إدارة المشاريع) يطلب، وإدارة الكوادر تلبي
	mux.Handle("POST /api/staff-requests", middleware.Chain(http.HandlerFunc(staffRequestHandler.Create), requireAuth, requireProjectMgmtPerm))
	mux.Handle("GET /api/staff-requests", middleware.Chain(http.HandlerFunc(staffRequestHandler.List), requireAuth))
	mux.Handle("PUT /api/staff-requests/{id}/status", middleware.Chain(http.HandlerFunc(staffRequestHandler.UpdateStatus), requireAuth, requireHR))

	// مسؤول خدمة عام (تعميم فكرة أبو الجي بي اس لأي مجموعة خدمات) — الأدمن فقط يحدد المسؤوليات
	mux.Handle("GET /api/service-managers", middleware.Chain(http.HandlerFunc(serviceManagerHandler.List), requireAuth))
	mux.Handle("PUT /api/service-managers", middleware.Chain(http.HandlerFunc(serviceManagerHandler.Set), requireAuth, requireAdmin))

	// تتبع الموقع الحي للفرق الميدانية
	// إرسال نقطة موقع مفتوح لأي موظف مسجل دخول (يرسل موقعه هو بس، محمي داخل الهاندلر
	// عبر EmployeeIDFromContext). القراءة (مين وين الحين، مسار أي موظف) محصورة بمن
	// يدير الفرق الميدانية فعلاً — حتى ما يقدر أي فني يتتبع مواقع زملائه.
	mux.Handle("POST /api/location-pings", middleware.Chain(http.HandlerFunc(locationPingHandler.Create), requireAuth))
	mux.Handle("GET /api/location-pings/latest", middleware.Chain(http.HandlerFunc(locationPingHandler.Latest), requireAuth, requireFieldMonitor))
	mux.Handle("GET /api/location-pings/path", middleware.Chain(http.HandlerFunc(locationPingHandler.Path), requireAuth, requireFieldMonitor))

	// تقييم الأداء (منفصل عن KPI مال الغرامات) — الليدر يقيّم فنييه، الإداري يقيّم الليدرات
	mux.Handle("POST /api/performance-reviews", middleware.Chain(http.HandlerFunc(performanceReviewHandler.Create), requireAuth))
	mux.Handle("GET /api/performance-reviews", middleware.Chain(http.HandlerFunc(performanceReviewHandler.List), requireAuth))
	mux.Handle("GET /api/performance-reviews/employee/{employeeId}", middleware.Chain(http.HandlerFunc(performanceReviewHandler.ListForEmployee), requireAuth))

	// المشتريات (procurement)
	mux.Handle("GET /api/procurement", middleware.Chain(http.HandlerFunc(procurementHandler.List), requireAuth, requireProcurement))
	mux.Handle("GET /api/procurement/stats", middleware.Chain(http.HandlerFunc(procurementHandler.Stats), requireAuth, requireProcurement))
	// إنشاء الطلب مو مربوط بصلاحية "procurement" العامة (الي تفتح الاطّلاع بس) —
	// كل نوع طلب (شخصي/زبون) له صلاحية مستقلة تُفحص داخل الهاندلر نفسه.
	mux.Handle("POST /api/procurement", middleware.Chain(http.HandlerFunc(procurementHandler.Create), requireAuth))
	mux.Handle("PUT /api/procurement/{id}/status", middleware.Chain(http.HandlerFunc(procurementHandler.UpdateStatus), requireAuth, requireProcurementAdmin))
	mux.Handle("PUT /api/procurement/{id}/fulfill", middleware.Chain(http.HandlerFunc(procurementHandler.Fulfill), requireAuth, requireProcurementAdmin))

	// الموردون (suppliers)
	mux.Handle("GET /api/suppliers/specialties", middleware.Chain(http.HandlerFunc(supplierHandler.ListSpecialties), requireAuth))
	mux.Handle("POST /api/suppliers/specialties", middleware.Chain(http.HandlerFunc(supplierHandler.CreateSpecialty), requireAuth, requireContentTech))
	mux.Handle("DELETE /api/suppliers/specialties/{id}", middleware.Chain(http.HandlerFunc(supplierHandler.DeleteSpecialty), requireAuth, requireAdmin))
	mux.Handle("GET /api/suppliers", middleware.Chain(http.HandlerFunc(supplierHandler.List), requireAuth))
	mux.Handle("POST /api/suppliers", middleware.Chain(http.HandlerFunc(supplierHandler.Create), requireAuth, requireContentTech))
	mux.Handle("PUT /api/suppliers/{id}", middleware.Chain(http.HandlerFunc(supplierHandler.Update), requireAuth, requireContentTech))
	mux.Handle("DELETE /api/suppliers/{id}", middleware.Chain(http.HandlerFunc(supplierHandler.Delete), requireAuth, requireAdmin))
	mux.Handle("POST /api/suppliers/{id}/rate", middleware.Chain(http.HandlerFunc(supplierHandler.Rate), requireAuth))

	// عروض الأسعار (quotations)
	mux.Handle("GET /api/quotations", middleware.Chain(http.HandlerFunc(quotationHandler.List), requireAuth))
	mux.Handle("GET /api/quotations/{id}", middleware.Chain(http.HandlerFunc(quotationHandler.Get), requireAuth))
	mux.Handle("POST /api/quotations", middleware.Chain(http.HandlerFunc(quotationHandler.Create), requireAuth, requireQuotationSystem))
	mux.Handle("PUT /api/quotations/{id}", middleware.Chain(http.HandlerFunc(quotationHandler.Update), requireAuth, requireQuotationSystem))
	mux.Handle("DELETE /api/quotations/{id}", middleware.Chain(http.HandlerFunc(quotationHandler.Delete), requireAuth, requireAdmin))

	// المنتجات (products) — التعديل/الحذف محصوران بصلاحية content_technician، والإنشاء
	// يطابق نفس التعديل (نفس المورد) بدل ما يبقى مفتوح لأي موظف مسجل دخول.
	mux.Handle("GET /api/products", middleware.Chain(http.HandlerFunc(productHandler.List), requireAuth))
	mux.Handle("POST /api/products", middleware.Chain(http.HandlerFunc(productHandler.Create), requireAuth, requireContentTech))
	mux.Handle("PUT /api/products/{id}", middleware.Chain(http.HandlerFunc(productHandler.Update), requireAuth, requireContentTech))
	mux.Handle("DELETE /api/products/{id}", middleware.Chain(http.HandlerFunc(productHandler.Delete), requireAuth, requireAdmin))

	// نظام GPS — عملاء / شرائح SIM / طلبات الأجهزة / التجديد / الصيانة / الأسعار / الإحصائيات
	// الإنشاء/التعديل بكل موارد GPS يتطلب صلاحية "gps_system" فعلياً (نفس الصلاحية
	// الي تفتح كل صفحات نظام GPS بالواجهة) — القراءة تبقى مفتوحة لأي مسجل دخول.
	// POST هنا مفتوحة لأي مسجل دخول (requireAuth بس) — هذا التسجيل الأساسي مو عملية
	// إدارية، وموظف المبيعات (GpsPurchase.tsx) لازم يقدر يسجل بيانات الزبون وهو
	// يرسل طلب شراء جهاز GPS جديد، قبل حتى ما يوصل الطلب لإداري GPS للموافقة.
	mux.Handle("GET /api/gps/customers", middleware.Chain(http.HandlerFunc(gpsHandler.ListCustomers), requireAuth))
	mux.Handle("POST /api/gps/customers", middleware.Chain(http.HandlerFunc(gpsHandler.CreateCustomer), requireAuth))
	mux.Handle("PUT /api/gps/customers/{id}", middleware.Chain(http.HandlerFunc(gpsHandler.UpdateCustomer), requireAuth, requireGpsSystem))

	mux.Handle("GET /api/gps/sims", middleware.Chain(http.HandlerFunc(gpsHandler.ListSims), requireAuth))
	mux.Handle("POST /api/gps/sims", middleware.Chain(http.HandlerFunc(gpsHandler.CreateSim), requireAuth, requireGpsSystem))
	mux.Handle("PUT /api/gps/sims/{id}", middleware.Chain(http.HandlerFunc(gpsHandler.UpdateSim), requireAuth, requireGpsSystem))

	// نفس الشي هنا: صف GpsDeviceRequest الجديد يبدأ status='PENDING' افتراضياً
	// بالمخطط نفسه (schema_base.go) — هذا "طلب" بانتظار مراجعة إداري GPS، مو جهاز
	// مفعّل فعلياً، فتقييد الإنشاء بصلاحية gps_system كان يمنع بالضبط سيناريو تقديم
	// الطلب من موظف مبيعات ما عنده هذي الصلاحية. الموافقة (PUT) تبقى محمية.
	mux.Handle("GET /api/gps/devices", middleware.Chain(http.HandlerFunc(gpsHandler.ListDevices), requireAuth))
	mux.Handle("POST /api/gps/devices", middleware.Chain(http.HandlerFunc(gpsHandler.CreateDevice), requireAuth))
	mux.Handle("PUT /api/gps/devices/{id}", middleware.Chain(http.HandlerFunc(gpsHandler.UpdateDevice), requireAuth, requireGpsSystem))

	mux.Handle("GET /api/gps/renewals", middleware.Chain(http.HandlerFunc(gpsHandler.ListRenewals), requireAuth))
	mux.Handle("POST /api/gps/renewals", middleware.Chain(http.HandlerFunc(gpsHandler.CreateRenewal), requireAuth, requireGpsSystem))
	mux.Handle("PUT /api/gps/renewals/{id}", middleware.Chain(http.HandlerFunc(gpsHandler.UpdateRenewal), requireAuth, requireGpsSystem))

	mux.Handle("GET /api/gps/maintenance", middleware.Chain(http.HandlerFunc(gpsHandler.ListMaintenance), requireAuth))
	mux.Handle("POST /api/gps/maintenance", middleware.Chain(http.HandlerFunc(gpsHandler.CreateMaintenance), requireAuth, requireGpsSystem))
	mux.Handle("PUT /api/gps/maintenance/{id}", middleware.Chain(http.HandlerFunc(gpsHandler.UpdateMaintenance), requireAuth, requireGpsSystem))

	mux.Handle("GET /api/gps/settings", middleware.Chain(http.HandlerFunc(gpsHandler.ListSettings), requireAuth))
	mux.Handle("PUT /api/gps/settings", middleware.Chain(http.HandlerFunc(gpsHandler.UpsertSettings), requireAuth, requireGpsAdmin))

	mux.Handle("GET /api/gps/stats", middleware.Chain(http.HandlerFunc(gpsHandler.Stats), requireAuth))

	// الإحصائيات العامة (stats) — لوحة معلومات المدير/المشرف
	mux.Handle("GET /api/stats", middleware.Chain(http.HandlerFunc(statsHandler.Overview), requireAuth))

	// إدارة المركبات — وقود/تنظيف/تبديل زيت، أعطال وأضرار، حالة شهرية
	mux.Handle("GET /api/vehicles", middleware.Chain(http.HandlerFunc(vehicleHandler.List), requireAuth, requireVehicleMgmt))
	mux.Handle("POST /api/vehicles", middleware.Chain(http.HandlerFunc(vehicleHandler.Create), requireAuth, requireVehicleMgmt))
	mux.Handle("GET /api/vehicles/{id}/logs", middleware.Chain(http.HandlerFunc(vehicleHandler.ListLogs), requireAuth, requireVehicleMgmt))
	mux.Handle("POST /api/vehicles/{id}/logs", middleware.Chain(http.HandlerFunc(vehicleHandler.CreateLog), requireAuth, requireVehicleMgmt))
	mux.Handle("GET /api/vehicles/{id}/incidents", middleware.Chain(http.HandlerFunc(vehicleHandler.ListIncidents), requireAuth, requireVehicleMgmt))
	mux.Handle("POST /api/vehicles/{id}/incidents", middleware.Chain(http.HandlerFunc(vehicleHandler.CreateIncident), requireAuth, requireVehicleMgmt))
	mux.Handle("PUT /api/vehicle-incidents/{id}", middleware.Chain(http.HandlerFunc(vehicleHandler.UpdateIncident), requireAuth, requireVehicleMgmt))
	mux.Handle("GET /api/vehicles/{id}/monthly-status", middleware.Chain(http.HandlerFunc(vehicleHandler.ListMonthlyStatus), requireAuth, requireVehicleMgmt))
	mux.Handle("POST /api/vehicles/{id}/monthly-status", middleware.Chain(http.HandlerFunc(vehicleHandler.SetMonthlyStatus), requireAuth, requireVehicleMgmt))
	mux.Handle("POST /api/vehicles/{id}/ratings", middleware.Chain(http.HandlerFunc(vehicleHandler.CreateDailyRating), requireAuth, requireVehicleMgmt))
	mux.Handle("GET /api/vehicles/{id}/ratings", middleware.Chain(http.HandlerFunc(vehicleHandler.ListDailyRatings), requireAuth, requireVehicleMgmt))
	// تعديل/حذف بيانات السيارة الأساسية محصور بالمالك/الأدمن (requireAdmin) فقط —
	// أضيق من صلاحية "إدارة المركبات" العامة (vehicle_management) المستخدمة لبقية
	// عمليات السيارة (سجلات، وثائق، صور...)، بناءً على طلب صريح من مالك النظام.
	mux.Handle("PUT /api/vehicles/{id}", middleware.Chain(http.HandlerFunc(vehicleHandler.Update), requireAuth, requireAdmin))
	mux.Handle("DELETE /api/vehicles/{id}", middleware.Chain(http.HandlerFunc(vehicleHandler.Delete), requireAuth, requireAdmin))

	// ملف السيارة الكامل: وثائق وصور
	mux.Handle("GET /api/vehicles/{id}/documents", middleware.Chain(http.HandlerFunc(vehicleHandler.ListDocuments), requireAuth))
	mux.Handle("POST /api/vehicles/{id}/documents", middleware.Chain(http.HandlerFunc(vehicleHandler.CreateDocument), requireAuth, requireVehicleMgmt))
	mux.Handle("DELETE /api/vehicles/{id}/documents/{docId}", middleware.Chain(http.HandlerFunc(vehicleHandler.DeleteDocument), requireAuth, requireVehicleMgmt))
	mux.Handle("GET /api/vehicles/{id}/photos", middleware.Chain(http.HandlerFunc(vehicleHandler.ListPhotos), requireAuth))
	mux.Handle("POST /api/vehicles/{id}/photos", middleware.Chain(http.HandlerFunc(vehicleHandler.CreatePhoto), requireAuth, requireVehicleMgmt))
	mux.Handle("DELETE /api/vehicles/{id}/photos/{photoId}", middleware.Chain(http.HandlerFunc(vehicleHandler.DeletePhoto), requireAuth, requireVehicleMgmt))

	// نظام المهمة: كل خروج سيارة يصير سجل مهمة متابَع (سبب، وجهة، عداد، ركاب)
	mux.Handle("POST /api/vehicle-missions", middleware.Chain(http.HandlerFunc(vehicleMissionHandler.Start), requireAuth))
	mux.Handle("PUT /api/vehicle-missions/{id}/end", middleware.Chain(http.HandlerFunc(vehicleMissionHandler.End), requireAuth))
	mux.Handle("GET /api/vehicle-missions", middleware.Chain(http.HandlerFunc(vehicleMissionHandler.List), requireAuth, requireVehicleMgmt))
	mux.Handle("GET /api/vehicle-missions/{id}", middleware.Chain(http.HandlerFunc(vehicleMissionHandler.Get), requireAuth))
	mux.Handle("POST /api/vehicle-missions/{id}/rating", middleware.Chain(http.HandlerFunc(vehicleMissionHandler.CreateRating), requireAuth, requireVehicleMgmt))
	// فحص أدوات المركبة العامة عند بدء مهمة — حصراً لليدر (نفس requireLeader
	// المستخدم بصيانة الأجهزة)، الموظف العادي ما يشوف/يستخدم هالراوت إطلاقاً.
	mux.Handle("POST /api/vehicle-missions/{id}/tool-check", middleware.Chain(http.HandlerFunc(vehicleMissionHandler.CreateToolCheck), requireAuth, requireLeader))
	mux.Handle("GET /api/employees/{id}/driver-rating-summary", middleware.Chain(http.HandlerFunc(vehicleMissionHandler.DriverRatingSummary), requireAuth))

	// نظام حجز المركبات (مسبق) — منفصل عن بدء المهمة الفعلي أعلاه.
	mux.Handle("POST /api/vehicle-bookings", middleware.Chain(http.HandlerFunc(vehicleBookingHandler.Create), requireAuth))
	mux.Handle("PUT /api/vehicle-bookings/{id}/decide", middleware.Chain(http.HandlerFunc(vehicleBookingHandler.Decide), requireAuth, requireVehicleMgmt))
	mux.Handle("PUT /api/vehicle-bookings/{id}/cancel", middleware.Chain(http.HandlerFunc(vehicleBookingHandler.Cancel), requireAuth))
	mux.Handle("GET /api/vehicle-bookings", middleware.Chain(http.HandlerFunc(vehicleBookingHandler.List), requireAuth))
	// ملخصات التذكير — للمراقب بس (نظرة شاملة على كل السيارات والفنيين)
	// ملاحظة: هذولا فقط لوحة المراقبة (MonitorDashboard) تستدعيهم — عمداً مقيدين
	// بدور المراقب/الأدمن حصراً، لأنهم يعرضون راتب مقترح للفنيين (بيانات حساسة
	// ماريد الفني نفسه يشوفها). لا تحولها لصلاحية "vehicle_management" العامة —
	// هذي الصلاحية ممكن تنمنح لفنيين لتسجيل صيانة مركبات عادية بدون قصد كشف رواتب.
	mux.Handle("GET /api/vehicles/ratings/vehicle-summary", middleware.Chain(http.HandlerFunc(vehicleHandler.VehicleScoreSummaries), requireAuth, requireMonitor))
	mux.Handle("GET /api/vehicles/ratings/technician-summary", middleware.Chain(http.HandlerFunc(vehicleHandler.TechnicianWashSummaries), requireAuth, requireMonitor))

	// المرحلة 2: مرفقات الأعطال/الأضرار، متابعة الإطارات والبطاريات، تنبيهات الاستحقاق
	mux.Handle("GET /api/vehicle-incidents/{id}/attachments", middleware.Chain(http.HandlerFunc(vehicleHandler.ListIncidentAttachments), requireAuth, requireVehicleMgmt))
	mux.Handle("POST /api/vehicle-incidents/{id}/attachments", middleware.Chain(http.HandlerFunc(vehicleHandler.CreateIncidentAttachment), requireAuth, requireVehicleMgmt))
	mux.Handle("DELETE /api/vehicle-incidents/{id}/attachments/{attachmentId}", middleware.Chain(http.HandlerFunc(vehicleHandler.DeleteIncidentAttachment), requireAuth, requireVehicleMgmt))
	mux.Handle("GET /api/vehicles/{id}/parts", middleware.Chain(http.HandlerFunc(vehicleHandler.ListParts), requireAuth, requireVehicleMgmt))
	mux.Handle("POST /api/vehicles/{id}/parts", middleware.Chain(http.HandlerFunc(vehicleHandler.CreatePart), requireAuth, requireVehicleMgmt))
	mux.Handle("PUT /api/vehicle-parts/{id}/replace", middleware.Chain(http.HandlerFunc(vehicleHandler.ReplacePart), requireAuth, requireVehicleMgmt))
	mux.Handle("GET /api/vehicles/alerts", middleware.Chain(http.HandlerFunc(vehicleHandler.Alerts), requireAuth, requireVehicleMgmt))
	mux.Handle("GET /api/vehicles/dashboard", middleware.Chain(http.HandlerFunc(vehicleHandler.FleetDashboard), requireAuth, requireVehicleMgmt))
	mux.Handle("GET /api/vehicles/{id}/expense-summary", middleware.Chain(http.HandlerFunc(vehicleHandler.ExpenseSummary), requireAuth, requireVehicleMgmt))

	// الجودة — مشاكل تنفيذية ميدانية + مشاكل رقابية/إدارية
	// تقارير العمل — الفني يرسل تقرير عن حجزه، المراقب/الجودة يشوفون كل التقارير
	mux.Handle("POST /api/work-reports", middleware.Chain(http.HandlerFunc(workReportHandler.Create), requireAuth))
	mux.Handle("GET /api/work-reports", middleware.Chain(http.HandlerFunc(workReportHandler.List), requireAuth))

	mux.Handle("GET /api/quality/issues", middleware.Chain(http.HandlerFunc(qualityHandler.List), requireAuth, requireQuality))
	mux.Handle("POST /api/quality/issues", middleware.Chain(http.HandlerFunc(qualityHandler.Create), requireAuth, requireQuality))
	mux.Handle("PUT /api/quality/issues/{id}", middleware.Chain(http.HandlerFunc(qualityHandler.Update), requireAuth, requireQuality))

	// صيانة الأجهزة العامة (شيت "صيانة الاجهزة") — حصراً للـليدر (isLeader فريش من
	// قاعدة البيانات بكل طلب، مو من التوكن)
	mux.Handle("GET /api/device-maintenance", middleware.Chain(http.HandlerFunc(deviceMaintenanceHandler.List), requireAuth, requireLeader))
	mux.Handle("POST /api/device-maintenance", middleware.Chain(http.HandlerFunc(deviceMaintenanceHandler.Create), requireAuth, requireLeader))
	mux.Handle("PUT /api/device-maintenance/{id}", middleware.Chain(http.HandlerFunc(deviceMaintenanceHandler.Update), requireAuth, requireLeader))

	// جرد الفريق ("جرد العدد") — حصراً للـليدر أيضاً
	mux.Handle("GET /api/team-inventory/tools", middleware.Chain(http.HandlerFunc(teamInventoryCheckHandler.ListTools), requireAuth, requireLeader))
	mux.Handle("POST /api/team-inventory/tools", middleware.Chain(http.HandlerFunc(teamInventoryCheckHandler.CreateTool), requireAuth, requireLeader))
	mux.Handle("GET /api/team-inventory/checks", middleware.Chain(http.HandlerFunc(teamInventoryCheckHandler.List), requireAuth, requireLeader))
	mux.Handle("POST /api/team-inventory/checks", middleware.Chain(http.HandlerFunc(teamInventoryCheckHandler.Create), requireAuth, requireLeader))

	// فوترة الليدر (تحل محل شيت جوجل) — القراءة (الكتالوج/المواد/عرض الفواتير)
	// متاحة لأي مستخدم مسجّل دخول (يحتاجها أي مشرف/إداري يراجع الفواتير)، لكن
	// الإنشاء حصراً لليدر (isLeader فريش من قاعدة البيانات، نفس نمط requireLeader
	// أعلاه) لأن هذي الميزة مخصصة لعمل الليدر فقط.
	// عرض/إدارة "سلة" الليدر (فاتورة الليدر) صارت بصلاحية مستقلة leader_basket —
	// افتراضياً تلقائياً لكل ليدر (isLeader، شوف grantLeaderBasketToLeaders)، لكن
	// الأدمن يقدر كمان يمنحها لموظف MONITOR من صفحة الصلاحيات بدون ما يصير ليدر
	// فعلياً (requireLeaderOrPermission يسمح بالاثنين).
	requireLeaderBasket := middleware.RequireLeaderOrPermission(permissionRepo, employeeRepo, notificationRepo, "leader_basket")
	mux.Handle("GET /api/system-price-catalog", middleware.Chain(http.HandlerFunc(leaderInvoiceHandler.ListCatalog), requireAuth))
	mux.Handle("GET /api/materials", middleware.Chain(http.HandlerFunc(leaderInvoiceHandler.ListMaterials), requireAuth))
	mux.Handle("GET /api/leader-invoices", middleware.Chain(http.HandlerFunc(leaderInvoiceHandler.List), requireAuth, requireLeaderBasket))
	mux.Handle("GET /api/leader-invoices/{id}", middleware.Chain(http.HandlerFunc(leaderInvoiceHandler.Get), requireAuth, requireLeaderBasket))
	mux.Handle("POST /api/leader-invoices", middleware.Chain(http.HandlerFunc(leaderInvoiceHandler.Create), requireAuth, requireLeader))

	// إحصائيات الموظفين الشهرية — حصراً للمالك/الأدمن (requireAdmin يسمح OWNER
	// تلقائياً لأنه يتخطى أي قيد أدوار بـRequireRole).
	mux.Handle("GET /api/employee-stats/monthly", middleware.Chain(http.HandlerFunc(employeeStatsHandler.Monthly), requireAuth, requireAdmin))
	mux.Handle("GET /api/employee-stats/monthly/export", middleware.Chain(http.HandlerFunc(employeeStatsHandler.MonthlyExport), requireAuth, requireAdmin))

	// تقدير مدة العمل المتعلَّم (learned baseline) — قراءة فقط، متاح لأي مستخدم
	// مسجّل دخول (يحتاجها المنسق قبل تثبيت موعد/فريق).
	mux.Handle("GET /api/job-duration-estimate", middleware.Chain(http.HandlerFunc(jobDurationHandler.Estimate), requireAuth))

	return middleware.Chain(mux, middleware.Recovery, middleware.Logging, middleware.Metrics, middleware.CORS(cfg.CORSOrigins), middleware.BodyLimit(middleware.MaxBodyBytes))
}

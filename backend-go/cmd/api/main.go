package main

import (
	"log"
	"net/http"
	"time"

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

	if err := database.Migrate(db); err != nil {
		log.Fatalf("failed to run database migrations: %v", err)
	}

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
	trainingRepo := repository.NewTrainingRepository(db)
	missionRepo := repository.NewMissionRepository(db, bookingRepo)
	projectRepo := repository.NewProjectRepository(db)
	procurementRepo := repository.NewProcurementRepository(db)
	supplierRepo := repository.NewSupplierRepository(db)
	quotationRepo := repository.NewQuotationRepository(db)
	productRepo := repository.NewProductRepository(db)
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

	// Services
	authService := service.NewAuthService(employeeRepo, loginAuditRepo, cfg.JWTSecret)
	employeeService := service.NewEmployeeService(employeeRepo)
	permissionService := service.NewPermissionService(permissionRepo, employeeRepo)
	serviceCatalogService := service.NewServiceCatalogService(serviceRepo)
	customerService := service.NewCustomerService(customerRepo)
	bookingService := service.NewBookingService(bookingRepo, employeeRepo, customerRepo, qualityFollowUpRepo, notificationRepo)
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
	assistantService := service.NewAssistantService(cfg.GeminiAPIKey, cfg.GeminiDailyCap, employeeRepo, kpiRepo, performanceReviewRepo, bookingRepo, missionRepo, expenseRepo, gpsRepo, qualityFollowUpRepo, complaintRepo)
	assistantHandler := handler.NewAssistantHandler(assistantService)
	kpiCriterionHandler := handler.NewKpiCriterionHandler(kpiCriterionService)
	smartKpiHandler := handler.NewSmartKpiHandler(smartKpiService)
	complaintHandler := handler.NewComplaintHandler(complaintService)
	trainingHandler := handler.NewTrainingHandler(trainingService)
	missionHandler := handler.NewMissionHandler(missionService)
	projectHandler := handler.NewProjectHandler(projectService)
	procurementHandler := handler.NewProcurementHandler(procurementService)
	supplierHandler := handler.NewSupplierHandler(supplierService)
	quotationHandler := handler.NewQuotationHandler(quotationService)
	productHandler := handler.NewProductHandler(productService)
	gpsHandler := handler.NewGpsHandler(gpsService)
	workReportHandler := handler.NewWorkReportHandler(workReportService)
	statsHandler := handler.NewStatsHandler(statsService)
	vehicleHandler := handler.NewVehicleHandler(vehicleService)
	vehicleMissionHandler := handler.NewVehicleMissionHandler(vehicleMissionService, vehicleMissionRatingService, vehicleBookingService)
	vehicleBookingHandler := handler.NewVehicleBookingHandler(vehicleBookingService)
	qualityHandler := handler.NewQualityHandler(qualityService)
	staffRequestHandler := handler.NewStaffRequestHandler(staffRequestRepo)
	serviceManagerHandler := handler.NewServiceManagerHandler(serviceManagerRepo)
	locationPingHandler := handler.NewLocationPingHandler(locationPingRepo)
	performanceReviewService := service.NewPerformanceReviewService(performanceReviewRepo, employeeRepo)
	performanceReviewHandler := handler.NewPerformanceReviewHandler(performanceReviewService)

	requireAuth := middleware.RequireAuth(authService, employeeRepo)
	requireAdmin := middleware.RequireRole(employeeRepo, "ADMIN")
	// حصراً لحساب المالك (OWNER) — أقوى من الأدمن العادي، ما يشوفها إلا هو
	requireOwner := middleware.RequireRole(employeeRepo, "OWNER")
	requireFinance := middleware.RequireRole(employeeRepo, "ADMIN", "FINANCE")
	// تدقيق مبلغ الحجز يعتمد على صلاحية "finance" الممنوحة فعلياً للموظف (مو بس دوره
	// الوظيفي) — المراقب مثلاً عنده هذي الصلاحية افتراضياً ويشوف زر "تدقيق" بالواجهة،
	// فلازم الباك إند يتحقق من نفس الصلاحية بدل دور صارم، وإلا يترفض الطلب ويتسبب
	// بإيقاف حساب الموظف تلقائياً بعد 3 محاولات (حماية أمنية ضد التلاعب بالجلسة).
	requireVerifyBooking := middleware.RequirePermission(permissionRepo, employeeRepo, "finance")
	requireHR := middleware.RequireRole(employeeRepo, "ADMIN", "HR_COORDINATOR")
	requireInventoryApprove := middleware.RequireRole(employeeRepo, "ADMIN", "HR_COORDINATOR", "MONITOR")
	// تعديل مهارات موظف يعتمد على صلاحية "staff_management" الممنوحة فعلياً (نفس
	// الصلاحية الي تفتح صفحة "إدارة الكوادر" بالواجهة للمراقب أيضاً) — مو دور
	// وظيفي صارم، وإلا نفس بگ "تدقيق الحسابات" يتكرر: زر يطلع بالواجهة، السيرفر
	// يرفضه، وبعد 3 محاولات ينوقف حساب الموظف تلقائياً.
	requireStaffManagement := middleware.RequirePermission(permissionRepo, employeeRepo, "staff_management")
	requireMonitor := middleware.RequireRole(employeeRepo, "ADMIN", "MONITOR")
	requireProjectManager := middleware.RequireRole(employeeRepo, "ADMIN", "PROJECT_MANAGER")
	requireFieldMonitor := middleware.RequireRole(employeeRepo, "ADMIN", "HR_COORDINATOR", "MONITOR", "PROJECT_MANAGER")
	requireGpsAdmin := middleware.RequireRole(employeeRepo, "ADMIN", "GPS_ADMIN")
	requireContentTech := middleware.RequirePermission(permissionRepo, employeeRepo, "content_technician")
	requireVehicleMgmt := middleware.RequirePermission(permissionRepo, employeeRepo, "vehicle_management")
	requireProcurement := middleware.RequirePermission(permissionRepo, employeeRepo, "procurement")
	// توفير المواد وتحديد حالتها يقتصر على إداري الكميات فعلياً (أو الأدمن) — مو أي
	// موظف عنده صلاحية "procurement" العامة (زي الفني/مدير المشاريع الي بس يطلبون مواد).
	requireProcurementAdmin := middleware.RequireRole(employeeRepo, "ADMIN", "PROCUREMENT_ADMIN")
	requireQuality := middleware.RequirePermission(permissionRepo, employeeRepo, "quality_control")
	requireProjectMgmtPerm := middleware.RequirePermission(permissionRepo, employeeRepo, "project_management")
	requireKpi := middleware.RequirePermission(permissionRepo, employeeRepo, "kpi_management")
	requireKpiCriteria := middleware.RequirePermission(permissionRepo, employeeRepo, "kpi_criteria_management")

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
	mux.Handle("PUT /api/bookings/{id}/materials-ready", middleware.Chain(http.HandlerFunc(bookingHandler.SetMaterialsReady), requireAuth))
	mux.Handle("PUT /api/bookings/{id}/complete", middleware.Chain(http.HandlerFunc(bookingHandler.Complete), requireAuth))
	mux.Handle("PUT /api/bookings/{id}/verify", middleware.Chain(http.HandlerFunc(bookingHandler.Verify), requireAuth, requireVerifyBooking))

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
	mux.Handle("POST /api/inventory/personal", middleware.Chain(http.HandlerFunc(inventoryHandler.CreatePersonalTool), requireAuth, requireHR))
	mux.Handle("PUT /api/inventory/personal/{id}", middleware.Chain(http.HandlerFunc(inventoryHandler.UpdatePersonalTool), requireAuth, requireHR))
	mux.Handle("DELETE /api/inventory/personal/{id}", middleware.Chain(http.HandlerFunc(inventoryHandler.DeletePersonalTool), requireAuth, requireHR))

	// جرد يومي: الموظف يؤكد جرد عدته الخاصة، الإداري يشوف نتائج اليوم لكل الموظفين
	mux.Handle("POST /api/inventory/checks", middleware.Chain(http.HandlerFunc(inventoryHandler.CreateInventoryCheck), requireAuth))
	mux.Handle("GET /api/inventory/checks/today", middleware.Chain(http.HandlerFunc(inventoryHandler.TodaysInventoryChecks), requireAuth))
	mux.Handle("POST /api/inventory/checks/{id}/resolve", middleware.Chain(http.HandlerFunc(inventoryHandler.ResolveInventoryCheck), requireAuth, requireHR))

	mux.Handle("GET /api/inventory/vehicle", middleware.Chain(http.HandlerFunc(inventoryHandler.ListVehicleTools), requireAuth))
	mux.Handle("POST /api/inventory/vehicle", middleware.Chain(http.HandlerFunc(inventoryHandler.CreateVehicleTool), requireAuth))
	mux.Handle("PUT /api/inventory/vehicle/{id}", middleware.Chain(http.HandlerFunc(inventoryHandler.UpdateVehicleTool), requireAuth))
	mux.Handle("DELETE /api/inventory/vehicle/{id}", middleware.Chain(http.HandlerFunc(inventoryHandler.DeleteVehicleTool), requireAuth))

	mux.Handle("GET /api/inventory/ondemand", middleware.Chain(http.HandlerFunc(inventoryHandler.ListOnDemandTools), requireAuth))
	mux.Handle("POST /api/inventory/ondemand", middleware.Chain(http.HandlerFunc(inventoryHandler.CreateOnDemandTool), requireAuth))
	mux.Handle("PUT /api/inventory/ondemand/{id}", middleware.Chain(http.HandlerFunc(inventoryHandler.UpdateOnDemandTool), requireAuth))

	mux.Handle("GET /api/inventory/requests", middleware.Chain(http.HandlerFunc(inventoryHandler.ListToolRequests), requireAuth))
	mux.Handle("POST /api/inventory/requests", middleware.Chain(http.HandlerFunc(inventoryHandler.CreateToolRequest), requireAuth))
	mux.Handle("PUT /api/inventory/requests/{id}/approve", middleware.Chain(http.HandlerFunc(inventoryHandler.ApproveToolRequest), requireAuth, requireInventoryApprove))
	mux.Handle("PUT /api/inventory/requests/{id}/reject", middleware.Chain(http.HandlerFunc(inventoryHandler.RejectToolRequest), requireAuth, requireInventoryApprove))
	mux.Handle("PUT /api/inventory/requests/{id}/return", middleware.Chain(http.HandlerFunc(inventoryHandler.ReturnToolRequest), requireAuth))

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
	mux.Handle("POST /api/kpi-criteria", middleware.Chain(http.HandlerFunc(kpiCriterionHandler.Create), requireAuth, requireKpiCriteria))
	mux.Handle("DELETE /api/kpi-criteria/{id}", middleware.Chain(http.HandlerFunc(kpiCriterionHandler.Delete), requireAuth, requireKpiCriteria))

	// تقييم الأداء التلقائي (Smart KPI) — الرانك الأسبوعي/الشهري للفنيين
	mux.Handle("GET /api/smart-kpi/technician/{employeeId}", middleware.Chain(http.HandlerFunc(smartKpiHandler.Technician), requireAuth))
	mux.Handle("GET /api/smart-kpi/leaderboard", middleware.Chain(http.HandlerFunc(smartKpiHandler.Leaderboard), requireAuth))

	// الشكاوى
	mux.Handle("GET /api/complaints", middleware.Chain(http.HandlerFunc(complaintHandler.List), requireAuth))
	mux.Handle("POST /api/complaints", middleware.Chain(http.HandlerFunc(complaintHandler.Create), requireAuth))
	mux.Handle("PUT /api/complaints/{id}", middleware.Chain(http.HandlerFunc(complaintHandler.Update), requireAuth))
	mux.Handle("PUT /api/complaints/{id}/resolve", middleware.Chain(http.HandlerFunc(complaintHandler.Resolve), requireAuth))
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
	mux.Handle("POST /api/projects", middleware.Chain(http.HandlerFunc(projectHandler.Create), requireAuth))
	mux.Handle("PUT /api/projects/{id}", middleware.Chain(http.HandlerFunc(projectHandler.Update), requireAuth))
	mux.Handle("DELETE /api/projects/{id}", middleware.Chain(http.HandlerFunc(projectHandler.Delete), requireAuth, requireProjectManager))

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
	mux.Handle("POST /api/quotations", middleware.Chain(http.HandlerFunc(quotationHandler.Create), requireAuth))
	mux.Handle("PUT /api/quotations/{id}", middleware.Chain(http.HandlerFunc(quotationHandler.Update), requireAuth))
	mux.Handle("DELETE /api/quotations/{id}", middleware.Chain(http.HandlerFunc(quotationHandler.Delete), requireAuth, requireAdmin))

	// المنتجات (products)
	mux.Handle("GET /api/products", middleware.Chain(http.HandlerFunc(productHandler.List), requireAuth))
	mux.Handle("POST /api/products", middleware.Chain(http.HandlerFunc(productHandler.Create), requireAuth))
	mux.Handle("PUT /api/products/{id}", middleware.Chain(http.HandlerFunc(productHandler.Update), requireAuth, requireContentTech))
	mux.Handle("DELETE /api/products/{id}", middleware.Chain(http.HandlerFunc(productHandler.Delete), requireAuth, requireAdmin))

	// نظام GPS — عملاء / شرائح SIM / طلبات الأجهزة / التجديد / الصيانة / الأسعار / الإحصائيات
	mux.Handle("GET /api/gps/customers", middleware.Chain(http.HandlerFunc(gpsHandler.ListCustomers), requireAuth))
	mux.Handle("POST /api/gps/customers", middleware.Chain(http.HandlerFunc(gpsHandler.CreateCustomer), requireAuth))
	mux.Handle("PUT /api/gps/customers/{id}", middleware.Chain(http.HandlerFunc(gpsHandler.UpdateCustomer), requireAuth))

	mux.Handle("GET /api/gps/sims", middleware.Chain(http.HandlerFunc(gpsHandler.ListSims), requireAuth))
	mux.Handle("POST /api/gps/sims", middleware.Chain(http.HandlerFunc(gpsHandler.CreateSim), requireAuth))
	mux.Handle("PUT /api/gps/sims/{id}", middleware.Chain(http.HandlerFunc(gpsHandler.UpdateSim), requireAuth))

	mux.Handle("GET /api/gps/devices", middleware.Chain(http.HandlerFunc(gpsHandler.ListDevices), requireAuth))
	mux.Handle("POST /api/gps/devices", middleware.Chain(http.HandlerFunc(gpsHandler.CreateDevice), requireAuth))
	mux.Handle("PUT /api/gps/devices/{id}", middleware.Chain(http.HandlerFunc(gpsHandler.UpdateDevice), requireAuth))

	mux.Handle("GET /api/gps/renewals", middleware.Chain(http.HandlerFunc(gpsHandler.ListRenewals), requireAuth))
	mux.Handle("POST /api/gps/renewals", middleware.Chain(http.HandlerFunc(gpsHandler.CreateRenewal), requireAuth))
	mux.Handle("PUT /api/gps/renewals/{id}", middleware.Chain(http.HandlerFunc(gpsHandler.UpdateRenewal), requireAuth))

	mux.Handle("GET /api/gps/maintenance", middleware.Chain(http.HandlerFunc(gpsHandler.ListMaintenance), requireAuth))
	mux.Handle("POST /api/gps/maintenance", middleware.Chain(http.HandlerFunc(gpsHandler.CreateMaintenance), requireAuth))
	mux.Handle("PUT /api/gps/maintenance/{id}", middleware.Chain(http.HandlerFunc(gpsHandler.UpdateMaintenance), requireAuth))

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
	mux.Handle("PUT /api/vehicles/{id}", middleware.Chain(http.HandlerFunc(vehicleHandler.Update), requireAuth, requireVehicleMgmt))

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

	handlerChain := middleware.Chain(mux, middleware.Recovery, middleware.Logging, middleware.Metrics, middleware.CORS(cfg.CORSOrigins))

	log.Printf("staffmange-api listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, handlerChain); err != nil {
		log.Fatal(err)
	}
}

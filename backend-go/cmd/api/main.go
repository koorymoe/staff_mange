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
	"staffmange-api/internal/storage"
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
	revolvingFundRepo := repository.NewRevolvingFundRepository(db)
	attendanceRepo := repository.NewAttendanceRepository(db)
	kpiRepo := repository.NewKpiRepository(db)
	announcementRepo := repository.NewAnnouncementRepository(db)
	kpiCriterionRepo := repository.NewKpiCriterionRepository(db)
	smartKpiRepo := repository.NewSmartKpiRepository(db)
	complaintRepo := repository.NewComplaintRepository(db)
	assistantConversationRepo := repository.NewAssistantConversationRepository(db)
	assistantKnowledgeRepo := repository.NewAssistantKnowledgeRepository(db)
	trainingRepo := repository.NewTrainingRepository(db)
	missionRepo := repository.NewMissionRepository(db, bookingRepo)
	projectRepo := repository.NewProjectRepository(db)
	projectWorkTypeRepo := repository.NewProjectWorkTypeRepository(db)
	vipCustomerRepo := repository.NewVipCustomerRepository(db)
	checklistRepo := repository.NewChecklistRepository(db)
	techShowcaseRepo := repository.NewTechShowcaseRepository(db)
	attendanceIconRequestRepo := repository.NewAttendanceIconRequestRepository(db)
	procurementRepo := repository.NewProcurementRepository(db)
	supplierRepo := repository.NewSupplierRepository(db)
	quotationRepo := repository.NewQuotationRepository(db)
	solarRepo := repository.NewSolarRepository(db)
	backupRunRepo := repository.NewBackupRunRepository(db)
	trainingProgramRepo := repository.NewTrainingProgramRepository(db)
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

	// ═══ نظام الغرامات التلقائي ═══
	// النظام يغرّم لحاله ويعلن، والنقاط ترجع بالشغل النظيف مو بالطلب.
	disciplineRepo := repository.NewDisciplineRepository(db)
	employeeLetterRepo := repository.NewEmployeeLetterRepository(db)
	bookingProgressRepo := repository.NewBookingProgressRepository(db)
	deviceMaintenanceRepo := repository.NewDeviceMaintenanceRepository(db)
	teamInventoryCheckRepo := repository.NewTeamInventoryCheckRepository(db)
	jobDurationSampleRepo := repository.NewJobDurationSampleRepository(db)

	// Services
	lockoutRepo := repository.NewSecurityLockoutRepository(db)
	middleware.SetLockoutRepository(lockoutRepo)
	authService := service.NewAuthService(employeeRepo, loginAuditRepo, lockoutRepo, cfg.JWTSecret)
	employeeService := service.NewEmployeeService(employeeRepo)
	employeeService.SetInventoryRepository(inventoryRepo)
	permissionService := service.NewPermissionService(permissionRepo, employeeRepo)
	serviceCatalogService := service.NewServiceCatalogService(serviceRepo)
	customerService := service.NewCustomerService(customerRepo)
	bookingService := service.NewBookingService(bookingRepo, employeeRepo, customerRepo, qualityFollowUpRepo, notificationRepo, inventoryRepo)

	disciplineService := service.NewDisciplineService(disciplineRepo, announcementRepo, notificationRepo, employeeRepo)
	// نربط فحص عدالة التوزيع بخدمة الحجوزات بعد بناء الاثنين (تفادي
	// اعتماد دائري بينهن)
	bookingService.SetDisciplineChecker(disciplineService)
	disciplineService.StartBackgroundSweeps()
	qualityFollowUpService := service.NewQualityFollowUpService(qualityFollowUpRepo, notificationRepo)
	cartService := service.NewCartService(cartRepo)
	expenseService := service.NewExpenseService(expenseRepo)
	inventoryService := service.NewInventoryService(inventoryRepo)
	// موافقة على طلب أداة غير متوفرة تولّد طلب مشتريات يوصل للمحاسب.
	inventoryService.SetProcurementRepository(procurementRepo)
	attendanceService := service.NewAttendanceService(attendanceRepo)
	notificationService := service.NewNotificationService(notificationRepo)
	kpiService := service.NewKpiService(kpiRepo, employeeRepo, notificationRepo, announcementRepo)
	kpiCriterionService := service.NewKpiCriterionService(kpiCriterionRepo)
	smartKpiService := service.NewSmartKpiService(smartKpiRepo)
	complaintService := service.NewComplaintService(complaintRepo)
	trainingService := service.NewTrainingService(trainingRepo)
	missionService := service.NewMissionService(missionRepo)
	projectService := service.NewProjectService(projectRepo)
	// أي مشروع ينضاف يرحّل صاحبه للشخصيات المهمة تلقائياً
	projectService.SetVipRepositories(vipCustomerRepo, customerRepo)
	// وصول المشروع لمرحلة «٥. البدء بالتنفيذ» يفتح حجزه عند إداري
	// الحجوزات — الربط بعد البناء حتى ما يصير اعتماد دائري.
	projectService.SetBookingUnlocker(bookingService)

	// تذكير معاودة الاتصال بالزبون الي ما رد — بدونه الحجز يقعد
	// بالطابور بلا حركة، لا ملغي ولا شغّال.
	bookingReminderService := service.NewBookingReminderService(bookingRepo, notificationRepo)
	bookingReminderService.StartBackgroundSweeps()
	projectWorkTypeService := service.NewProjectWorkTypeService(projectWorkTypeRepo)
	checklistService := service.NewChecklistService(checklistRepo)
	techShowcaseService := service.NewTechShowcaseService(techShowcaseRepo)
	attendanceIconRequestService := service.NewAttendanceIconRequestService(attendanceIconRequestRepo, employeeRepo)
	procurementService := service.NewProcurementService(procurementRepo, permissionRepo)
	supplierService := service.NewSupplierService(supplierRepo)
	quotationService := service.NewQuotationService(quotationRepo, permissionRepo)
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
	permissionHandler := handler.NewPermissionHandler(permissionService, lockoutRepo)
	serviceHandler := handler.NewServiceHandler(serviceCatalogService)
	customerHandler := handler.NewCustomerHandler(customerService)
	bookingHandler := handler.NewBookingHandler(bookingService, permissionRepo)
	bookingHandler.SetReminderService(bookingReminderService)
	qualityFollowUpHandler := handler.NewQualityFollowUpHandler(qualityFollowUpService)
	securityHandler := handler.NewSecurityHandler(db, loginAuditRepo, lockoutRepo, startedAt)

	// تخزين الملفات: R2 إذا انضبطت إعداداته، وإلا القرص المحلي. الاثنين
	// نفس الواجهة فباقي النظام ما يعرف أيهم شغّال.
	fileStore := buildFileStore(cfg)
	log.Printf("[storage] تخزين الملفات: %s", fileStore.Kind())
	fileHandler := handler.NewFileHandler(fileStore, []byte(cfg.JWTSecret))
	cartHandler := handler.NewCartHandler(cartService)
	expenseHandler := handler.NewExpenseHandler(expenseService)
	inventoryHandler := handler.NewInventoryHandler(inventoryService)
	revolvingFundHandler := handler.NewRevolvingFundHandler(revolvingFundRepo)
	dashboardHandler := handler.NewDashboardHandler(db)
	leaveHandler := handler.NewLeaveRequestHandler(repository.NewLeaveRequestRepository(db), permissionRepo, notificationRepo)
	gpsInstallCostHandler := handler.NewGpsInstallCostHandler(repository.NewGpsInstallCostRepository(db))
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
	// منو يشوف كل المشاريع: مدير النظام/المالك/مدير المشاريع، أو صاحب صلاحية
	// إدارة المشاريع (أو إضافة مشروع فقط). غيرهم يشوف بس المشاريع الموجّهة له.
	canSeeAllProjects := func(r *http.Request) bool {
		role, _ := r.Context().Value(middleware.ContextRole).(string)
		if role == "ADMIN" || role == "OWNER" || role == "PROJECT_MANAGER" || role == "MONITOR" {
			return true
		}
		id := middleware.EmployeeIDFromContext(r)
		if id == "" {
			return false
		}
		perms, err := permissionRepo.ListForEmployee(id)
		if err != nil {
			return false
		}
		for _, p := range perms {
			switch p.Name {
			case "project_management", "project_create_only", "monitoring", "auditing":
				return true
			}
		}
		return false
	}
	projectHandler := handler.NewProjectHandler(projectService, canSeeAllProjects)
	projectWorkTypeHandler := handler.NewProjectWorkTypeHandler(projectWorkTypeService)
	vipCustomerHandler := handler.NewVipCustomerHandler(vipCustomerRepo, customerService)
	checklistHandler := handler.NewChecklistHandler(checklistService)
	techShowcaseHandler := handler.NewTechShowcaseHandler(techShowcaseService)
	attendanceIconRequestHandler := handler.NewAttendanceIconRequestHandler(attendanceIconRequestService)
	procurementHandler := handler.NewProcurementHandler(procurementService)
	geoHandler := handler.NewGeoHandler()
	supplierHandler := handler.NewSupplierHandler(supplierService)
	privacyPolicyHandler := handler.NewPrivacyPolicyHandler(repository.NewPrivacyPolicyRepository(db))
	mapLinkHandler := handler.NewMapLinkHandler()
	quotationHandler := handler.NewQuotationHandler(quotationService)
	solarHandler := handler.NewSolarHandler(solarRepo)
	backupHandler := handler.NewBackupHandler(backupRunRepo)
	// سعر المنظومة لحجز الطاقة الشمسية — ينحسب من الكتالوك مو ينكتب بالإيد
	bookingService.SetSolarPricer(solarRepo)
	trainingProgramHandler := handler.NewTrainingProgramHandler(trainingProgramRepo)
	productHandler := handler.NewProductHandler(productService)
	leaderInvoiceService := service.NewLeaderInvoiceService(leaderInvoiceRepo, systemPriceCatalogRepo, materialRepo, employeeCommissionRepo, bookingRepo, employeeRepo, jobDurationEstimatorService)
	leaderInvoiceHandler := handler.NewLeaderInvoiceHandler(leaderInvoiceService, systemPriceCatalogRepo, materialRepo)
	networkPriceRepo := repository.NewNetworkPriceRepository(db)
	// ── صندوق المراقب ──
	// الربط بـSetMonitorFeed بعد البناء (مو بالمنشئ) حتى ما يصير اعتماد
	// دائري بين خدمة الحجوزات وخدمة المراقبة.
	monitorReviewRepo := repository.NewMonitorReviewRepository(db)
	monitorReviewService := service.NewMonitorReviewService(monitorReviewRepo, employeeRepo, notificationRepo)
	monitorReviewHandler := handler.NewMonitorReviewHandler(monitorReviewService)
	bookingService.SetMonitorFeed(monitorReviewService)
	leaderInvoiceService.SetMonitorFeed(monitorReviewService)
	// بقية الأقسام: كل واحد بلحظة قراره الي ما ينراجع —
	// المشتريات وقت صرف الفلوس، الجودة وقت الحكم السلبي،
	// الجي بي اس وقت تسليم الجهاز.
	procurementService.SetMonitorFeed(monitorReviewService)
	qualityFollowUpService.SetMonitorFeed(monitorReviewService)
	gpsService.SetMonitorFeed(monitorReviewService)
	networkCostHandler := handler.NewNetworkCostHandler(networkPriceRepo)
	jobDurationHandler := handler.NewJobDurationHandler(jobDurationEstimatorService)
	employeeMonthlyStatsService := service.NewEmployeeMonthlyStatsService(employeeRepo, kpiRepo, complaintRepo, leaderInvoiceRepo, bookingRepo, vehicleMissionRatingRepo, employeeCommissionRepo, jobDurationSampleRepo)
	// النقاط الكاملة بجدول الإحصاءات تجي من نفس حاسبة المخطط، حتى
	// الرقم بالجدول والرقم بالمخطط ما يختلفون.
	employeeMonthlyStatsService.SetSmartKpi(smartKpiService)
	employeeStatsHandler := handler.NewEmployeeStatsHandler(employeeMonthlyStatsService)
	internalWorksRepo := repository.NewInternalWorksRepository(db)
	statsManagementService := service.NewStatsManagementService(employeeRepo, bookingRepo, employeeCommissionRepo, projectRepo, leaderInvoiceRepo, attendanceRepo, employeeMonthlyStatsService, internalWorksRepo)
	statsManagementHandler := handler.NewStatsManagementHandler(statsManagementService)
	exhibitionRepo := repository.NewExhibitionRepository(db)
	exhibitionService := service.NewExhibitionService(exhibitionRepo, assistantService)
	exhibitionHandler := handler.NewExhibitionHandler(exhibitionService)
	productRequestRepo := repository.NewProductRequestRepository(db)
	productRequestService := service.NewProductRequestService(productRequestRepo, productRepo)
	productRequestHandler := handler.NewProductRequestHandler(productRequestService)
	productProcurementService := service.NewProductProcurementService(repository.NewProductProcurementRepository(db), productRequestService)
	productProcurementHandler := handler.NewProductProcurementHandler(productProcurementService)
	serviceStudyRepo := repository.NewServiceStudyRepository(db)
	serviceStudyService := service.NewServiceStudyService(serviceStudyRepo)
	serviceStudyHandler := handler.NewServiceStudyHandler(serviceStudyService)
	designFormRepo := repository.NewDesignFormRepository(db)
	designFormService := service.NewDesignFormService(designFormRepo)
	designFormHandler := handler.NewDesignFormHandler(designFormService)
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
	performanceReviewService := service.NewPerformanceReviewService(performanceReviewRepo, employeeRepo, bookingRepo)
	performanceReviewHandler := handler.NewPerformanceReviewHandler(performanceReviewService)
	deviceMaintenanceHandler := handler.NewDeviceMaintenanceHandler(deviceMaintenanceService)
	teamInventoryCheckHandler := handler.NewTeamInventoryCheckHandler(teamInventoryCheckService)

	requireAuth := middleware.RequireAuth(authService, employeeRepo)
	requireAdmin := middleware.RequireRole(employeeRepo, notificationRepo, "ADMIN")
	// حصراً لحساب المالك (OWNER) — أقوى من الأدمن العادي، ما يشوفها إلا هو
	requireOwner := middleware.RequireRole(employeeRepo, notificationRepo, "OWNER")
	// فتح الحسابات: المالك وحده، بلا تسجيل مخالفة على مدير النظام
	requireOwnerAccounts := middleware.RequireOwnerOnly("فتح الحسابات للمالك وحده")
	requireFinance := middleware.RequireRole(employeeRepo, notificationRepo, "ADMIN", "FINANCE")
	// تدقيق مبلغ الحجز يعتمد على صلاحية "finance" الممنوحة فعلياً للموظف (مو بس دوره
	// الوظيفي) — المراقب مثلاً عنده هذي الصلاحية افتراضياً ويشوف زر "تدقيق" بالواجهة،
	// فلازم الباك إند يتحقق من نفس الصلاحية بدل دور صارم، وإلا يترفض الطلب ويتسبب
	// بإيقاف حساب الموظف تلقائياً بعد 3 محاولات (حماية أمنية ضد التلاعب بالجلسة).
	requireVerifyBooking := middleware.RequirePermission(permissionRepo, employeeRepo, notificationRepo, "finance")
	requireCoordinator := middleware.RequirePermission(permissionRepo, employeeRepo, notificationRepo, "coordinator")
	requireCrewManagement := middleware.RequirePermission(permissionRepo, employeeRepo, notificationRepo, "crew_management")
	requireHR := middleware.RequireRole(employeeRepo, notificationRepo, "ADMIN", "HR_COORDINATOR")
	// تدقيق أمني: مسارات إدارية كانت مفتوحة لأي موظف مسجّل دخول (requireAuth
	// فقط) — أي فني يقدر يثبّت حجز أو يعدّل زبون أو يضيف أدوات. صارت مربوطة
	// بصلاحياتها الطبيعية، مع إبقاء المسارات الذاتية (حضور/إشعارات) مفتوحة.
	requireCustomerMgmt := middleware.RequireAnyPermission(permissionRepo, employeeRepo, notificationRepo, "manage_customers", "sales_booking", "coordinator")
	requireBookingCoord := middleware.RequireAnyPermission(permissionRepo, employeeRepo, notificationRepo, "coordinator", "crew_management")
	requireKpiMgmt := middleware.RequirePermission(permissionRepo, employeeRepo, notificationRepo, "kpi_management")
	// تعديل تفاصيل الحجز تستخدمه صفحات ثانية غير التنسيق (قائمة الحجوزات
	// لتخصيص مركبة، وخريطة المهام) — فنوسّعها لصلاحياتهن حتى ما ننكسر شغل شغّال.
	requireBookingEdit := middleware.RequireAnyPermission(permissionRepo, employeeRepo, notificationRepo,
		"coordinator", "crew_management", "view_bookings", "mission_tracking", "sales_booking")
	// إدارة المخزون (عدة الموظفين الشخصية/أدوات المركبات/العدة القياسية) — مسموحة
	// لأدوار HR/ADMIN كالمعتاد، أو لأي موظف عنده صلاحية "inventory" المخصصة
	// (ممنوحة من صفحة الصلاحيات، مثلاً PROCUREMENT_ADMIN) — توسيع وصول، مو تضييق.
	requireHROrInventory := middleware.RequireRoleOrPermission(permissionRepo, employeeRepo, notificationRepo, []string{"ADMIN", "HR_COORDINATOR"}, "inventory")
	requireLeader := middleware.RequireLeader(employeeRepo, notificationRepo)
	requireInventoryView := middleware.RequirePermission(permissionRepo, employeeRepo, notificationRepo, "inventory")
	// حساب كلفة التنفيذ: صلاحية تنعطى وتنسحب، مو دور ثابت
	requireExecutionCost := middleware.RequirePermission(permissionRepo, employeeRepo, notificationRepo, "execution_cost")
	// موافقة/رفض طلبات الأدوات: كانت دور صارم بدون منفذ صلاحية، فإداري الكميات
	// — وهو صاحب الشغلة أصلاً — ما كان يقدر يوافق أبداً مهما انمنحت له
	// صلاحيات. صارت صلاحية مستقلة تُمنح لأي موظف، مع إبقاء الأدوار القديمة.
	requireInventoryApprove := middleware.RequireRoleOrPermission(permissionRepo, employeeRepo, notificationRepo, []string{"ADMIN", "HR_COORDINATOR", "MONITOR"}, "tool_requests_approve")
	// تعديل مهارات موظف يعتمد على صلاحية "staff_management" الممنوحة فعلياً (نفس
	// الصلاحية الي تفتح صفحة "إدارة الكوادر" بالواجهة للمراقب أيضاً) — مو دور
	// وظيفي صارم، وإلا نفس بگ "تدقيق الحسابات" يتكرر: زر يطلع بالواجهة، السيرفر
	// يرفضه، وبعد 3 محاولات ينوقف حساب الموظف تلقائياً.
	requireStaffManagement := middleware.RequirePermission(permissionRepo, employeeRepo, notificationRepo, "staff_management")
	requireMonitor := middleware.RequireRole(employeeRepo, notificationRepo, "ADMIN", "MONITOR")
	requireProjectManager := middleware.RequireRole(employeeRepo, notificationRepo, "ADMIN", "PROJECT_MANAGER")
	// إدارة المشاريع: بالدور أو بصلاحية project_management الممنوحة يدوياً
	requireProjectMgmt := middleware.RequireRoleOrPermission(permissionRepo, employeeRepo, notificationRepo, []string{"ADMIN", "PROJECT_MANAGER"}, "project_management")
	requireFieldMonitor := middleware.RequireRole(employeeRepo, notificationRepo, "ADMIN", "HR_COORDINATOR", "MONITOR", "PROJECT_MANAGER")
	requireGpsAdmin := middleware.RequireRole(employeeRepo, notificationRepo, "ADMIN", "GPS_ADMIN")
	requireContentTech := middleware.RequirePermission(permissionRepo, employeeRepo, notificationRepo, "content_technician")
	// الموردون بصلاحية مستقلة (انفصلت عن content_technician الواسعة)
	requireSuppliersMgmt := middleware.RequirePermission(permissionRepo, employeeRepo, notificationRepo, "suppliers_management")
	requireVehicleMgmt := middleware.RequirePermission(permissionRepo, employeeRepo, notificationRepo, "vehicle_management")
	requireProcurement := middleware.RequirePermission(permissionRepo, employeeRepo, notificationRepo, "procurement")
	// توفير المواد وتحديد حالتها يقتصر على إداري الكميات فعلياً (أو الأدمن) — مو أي
	// موظف عنده صلاحية "procurement" العامة (زي الفني/مدير المشاريع الي بس يطلبون مواد).
	requireProcurementAdmin := middleware.RequireRole(employeeRepo, notificationRepo, "ADMIN", "PROCUREMENT_ADMIN")
	// unit_technicians مو content_technician عمداً — صلاحية مستقلة لوحدة
	// التقنيين تحديداً (المعارض/المنتجات/الخدمات) حتى ما تنفتح بالغلط لأي حد
	// عنده content_technician من مكان ثاني (مواد التدريب/الموردين/المنتجات).
	requireUnitTechnicians := middleware.RequirePermission(permissionRepo, employeeRepo, notificationRepo, "unit_technicians")
	requireTechUnitOrProcurement := middleware.RequireRoleOrPermission(permissionRepo, employeeRepo, notificationRepo, []string{"PROCUREMENT_ADMIN"}, "unit_technicians")
	requireQuality := middleware.RequirePermission(permissionRepo, employeeRepo, notificationRepo, "quality_control")
	requireProjectMgmtPerm := middleware.RequirePermission(permissionRepo, employeeRepo, notificationRepo, "project_management")
	requireKpi := middleware.RequirePermission(permissionRepo, employeeRepo, notificationRepo, "kpi_management")
	requireKpiCriteria := middleware.RequirePermission(permissionRepo, employeeRepo, notificationRepo, "kpi_criteria_management")
	// إنشاء/تعديل عروض الأسعار: أي مستوى من الثلاث (إضافة فقط / إضافة وتعديل
	// عروضي / إضافة وتعديل واطلاع الكل) يقدر يوصل للـPOST/PUT — الفرز الدقيق
	// (مين يشوف شنو، ومين يقدر يعدّل شنو) يصير داخل QuotationService نفسه.
	// quotation_system القديمة تبقى مرادف quotation_manage_all لأي موظف
	// كانت ممنوحة له من قبل.
	requireQuotationAccess := middleware.RequireAnyPermission(permissionRepo, employeeRepo, notificationRepo,
		service.QuotationPermCreate, service.QuotationPermEditOwn, service.QuotationPermManageAll, "quotation_system")
	// إنشاء/تعديل بيانات نظام GPS (عملاء/شرائح/أجهزة/تجديد/صيانة) يقتصر على من
	// يملك صلاحية "gps_system" فعلياً (نفس الصلاحية الي تفتح كل صفحات نظام GPS
	// بالواجهة) — مو أي موظف مسجل دخول.
	requireFund := middleware.RequirePermission(permissionRepo, employeeRepo, notificationRepo, "revolving_fund")
	// قيمة مبالغ الدوار نفسها (تعديل الرصيد والتغذية) — مدير النظام والمالك
	// فقط، أو أي موظف ينطونه صلاحية revolving_fund_amount صراحةً.
	requireFundAmount := middleware.RequireRoleOrPermission(permissionRepo, employeeRepo, notificationRepo, []string{"ADMIN", "OWNER"}, "revolving_fund_amount")
	// التخريج يرجّع فلوس للدوار — صلاحية مستقلة عن شغل الدوار اليومي.
	requireDischarge := middleware.RequireRoleOrPermission(permissionRepo, employeeRepo, notificationRepo, []string{"ADMIN", "OWNER", "FINANCE"}, "fund_discharge")
	// الاتصال بالزبون مو حكر على مهندس الجودة — أي موظف ينطيه المدير
	// صلاحية complaint_contact يكدر يتصل ويأشر النتيجة باسمه.
	requireComplaintContact := middleware.RequireRoleOrPermission(permissionRepo, employeeRepo, notificationRepo, []string{"ADMIN", "OWNER", "QUALITY_ENGINEER", "MONITOR"}, "complaint_contact")
	requireGpsSystem := middleware.RequirePermission(permissionRepo, employeeRepo, notificationRepo, "gps_system")
	// متابعة تجديد اشتراكات الجي بي اس تخص الاثنين: مهندس الجودة يتصل
	// بالزبائن، ومسؤول الجي بي اس يشوف منو خلصت مهلته وشريحته تحتاج حرق.
	requireGpsOrQuality := middleware.RequireAnyPermission(permissionRepo, employeeRepo, notificationRepo, "gps_system", "quality_control")
	// قوائم زبائن الجي بي اس فيها أسماء وأرقام ٤٩٣ زبون — كانت مفتوحة لأي
	// موظف مسجّل دخول، يعني الفني يقدر يسحبها كلها من F12 وهو ما إله علاقة
	// بالجي بي اس أصلاً. صارت محصورة بالي شغلهم فعلاً بالجي بي اس.
	requireVipManualAdd := middleware.RequireRoleOrPermission(permissionRepo, employeeRepo, notificationRepo,
		[]string{"ADMIN", "OWNER"}, "vip_manual_add")
	requireGpsData := middleware.RequireRoleOrPermission(permissionRepo, employeeRepo, notificationRepo,
		[]string{"ADMIN", "OWNER", "SALES"}, "gps_system")

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// حماية ضد محاولات تخمين كلمة السر المتكررة: أقصى 8 محاولات دخول بالدقيقة
	// من نفس عنوان IP — هذا بالضبط النوع اللي سبب حظر IP السيرفر من Hetzner
	// كإجراء حماية تلقائي ضدهم لما شافوا محاولات دخول متكررة سريعة.
	requireLoginRateLimit := middleware.RateLimit(8, time.Minute)
	mux.Handle("POST /api/auth/login", requireLoginRateLimit(http.HandlerFunc(authHandler.Login)))
	// رفع الملفات يحتاج تسجيل دخول. العرض بعد — الملفات تحمل صور وصولات
	// ووثائق زبائن، وما تنعرض لأي أحد يخمّن الرابط.
	mux.Handle("POST /api/files", middleware.Chain(http.HandlerFunc(fileHandler.Upload), requireAuth))
	mux.Handle("GET /api/files/token", middleware.Chain(http.HandlerFunc(fileHandler.Token), requireAuth))
	// العرض ما يمر بـrequireAuth لأن <img> ما يرسل ترويسة Authorization —
	// التحقق يصير بالوسم الموقّع بالمعالج نفسه.
	mux.Handle("GET /api/files/", http.HandlerFunc(fileHandler.Serve))

	mux.Handle("GET /api/auth/me", middleware.Chain(http.HandlerFunc(authHandler.Me), requireAuth))
	// تغيير كلمة المرور لمدير النظام والمالك بس. الموظف ما يغيّر سره
	// بنفسه — كلمات السر تنتحدد من «إدارة الموظفين» عند الأدمن، حتى
	// تبقى معروفة عند الإدارة وما يقفل موظف حسابه على نفسه.
	mux.Handle("PUT /api/auth/change-password", middleware.Chain(http.HandlerFunc(authHandler.ChangePassword), requireAuth,
		middleware.RequireRole(employeeRepo, notificationRepo, "ADMIN", "OWNER")))

	// موظفين — القراءة تحتاج تسجيل دخول فقط، الإنشاء/التعديل الحساس محمي بدور ADMIN
	mux.Handle("GET /api/employees", middleware.Chain(http.HandlerFunc(employeeHandler.List), requireAuth))
	mux.Handle("GET /api/employees/supervisors", middleware.Chain(http.HandlerFunc(employeeHandler.Supervisors), requireAuth))
	disciplineHandler := handler.NewDisciplineHandler(disciplineService)
	employeeLetterHandler := handler.NewEmployeeLetterHandler(employeeLetterRepo, notificationRepo)
	bookingProgressHandler := handler.NewBookingProgressHandler(bookingProgressRepo, bookingRepo, notificationRepo)
	// نقاط الانضباط: كل موظف يشوف الأرصدة (الشفافية جزء من العقوبة)،
	// وتشغيل الفحص يدوياً للمدير حصراً.
	// ── الطلبات: كتاب رسمي من الموظف للإدارة ──
	// التقديم وقراءة طلباتك: لكل موظف. البت: للمالك ومدير النظام.
	mux.Handle("GET /api/letters/addressees", middleware.Chain(http.HandlerFunc(employeeLetterHandler.Addressees), requireAuth))
	mux.Handle("POST /api/letters", middleware.Chain(http.HandlerFunc(employeeLetterHandler.Create), requireAuth))
	mux.Handle("GET /api/letters/mine", middleware.Chain(http.HandlerFunc(employeeLetterHandler.Mine), requireAuth))
	mux.Handle("GET /api/letters", middleware.Chain(http.HandlerFunc(employeeLetterHandler.Inbox), requireAuth, requireAdmin))
	mux.Handle("GET /api/letters/pending-count", middleware.Chain(http.HandlerFunc(employeeLetterHandler.PendingCount), requireAuth, requireAdmin))
	mux.Handle("PUT /api/letters/{id}/decide", middleware.Chain(http.HandlerFunc(employeeLetterHandler.Decide), requireAuth, requireAdmin))

	mux.Handle("GET /api/discipline", middleware.Chain(http.HandlerFunc(disciplineHandler.List), requireAuth))
	mux.Handle("GET /api/discipline/events", middleware.Chain(http.HandlerFunc(disciplineHandler.Events), requireAuth))
	// التعديل اليدوي: المالك ومدير النظام بس. requireAdmin يمرّر OWNER أصلاً.
	mux.Handle("POST /api/discipline/adjust", middleware.Chain(http.HandlerFunc(disciplineHandler.Adjust), requireAuth, requireAdmin))
	mux.Handle("POST /api/discipline/run", middleware.Chain(http.HandlerFunc(disciplineHandler.Run), requireAuth, requireAdmin))
	mux.Handle("GET /api/employees/archived", middleware.Chain(http.HandlerFunc(employeeHandler.ListArchived), requireAuth, requireAdmin))
	mux.Handle("GET /api/security/dashboard", middleware.Chain(http.HandlerFunc(securityHandler.Dashboard), requireAuth, requireOwner))
	mux.Handle("POST /api/security/unlock/{id}", middleware.Chain(http.HandlerFunc(securityHandler.Unlock), requireAuth, requireOwner))
	mux.Handle("POST /api/security/reset-attempts/{id}", middleware.Chain(http.HandlerFunc(securityHandler.ResetAttempts), requireAuth, requireOwner))
	mux.Handle("POST /api/security/free-memory", middleware.Chain(http.HandlerFunc(securityHandler.FreeMemory), requireAuth, requireOwner))
	mux.Handle("GET /api/employees/match", middleware.Chain(http.HandlerFunc(employeeHandler.Match), requireAuth))
	mux.Handle("GET /api/employees/{id}", middleware.Chain(http.HandlerFunc(employeeHandler.Get), requireAuth))
	// فتح حساب جديد = المالك وحده. مدير النظام يدير كلشي بالنظام لكن
	// **منو يدخل النظام** قرار المالك — وهذا يصير أهم بعد ما يجي
	// النظام الأكبر ويصير الحساب الواحد يفتح عالمين.
	mux.Handle("POST /api/employees", middleware.Chain(http.HandlerFunc(employeeHandler.Create), requireAuth, requireOwnerAccounts))
	mux.Handle("PUT /api/employees/{id}", middleware.Chain(http.HandlerFunc(employeeHandler.Update), requireAuth, requireAdmin))
	mux.Handle("POST /api/employees/{id}/link-historical", middleware.Chain(http.HandlerFunc(employeeHandler.LinkHistoricalRecords), requireAuth, requireAdmin))
	mux.Handle("PUT /api/employees/{id}/skills", middleware.Chain(http.HandlerFunc(employeeHandler.SetSkills), requireAuth, requireStaffManagement))

	// الصلاحيات — العرض متاح لأي مسجل دخول، التعديل والتطبيق التلقائي محصور بمدير النظام فقط
	// قائمة كل صلاحيات النظام = خارطة للمهاجم يعرف بيها شنو موجود وشنو
	// يستهدف. الموظف ما يحتاجها — صلاحياته هو تجي بـ/permissions/employee/{id}
	mux.Handle("GET /api/permissions", middleware.Chain(http.HandlerFunc(permissionHandler.ListAll), requireAuth, requireAdmin))
	mux.Handle("GET /api/permissions/role-defaults", middleware.Chain(http.HandlerFunc(permissionHandler.RoleDefaults), requireAuth))
	// قائمة الموظفين الي يوصلون لصلاحية معيّنة — لتعبئة القوائم المنسدلة
	// (مثلاً: مين المسؤول عن المشروع، ومين يسوي الكشف)
	mux.Handle("GET /api/permissions/employees", middleware.Chain(http.HandlerFunc(permissionHandler.EmployeesWithPermission), requireAuth))
	mux.Handle("GET /api/permissions/employee/{id}", middleware.Chain(http.HandlerFunc(permissionHandler.ListForEmployee), requireAuth))
	mux.Handle("PUT /api/permissions/employee/{id}", middleware.Chain(http.HandlerFunc(permissionHandler.SetForEmployee), requireAuth, requireAdmin))
	mux.Handle("POST /api/permissions/employee/{id}/apply-defaults", middleware.Chain(http.HandlerFunc(permissionHandler.ApplyDefaults), requireAuth, requireAdmin))

	// الخدمات والمهارات — القراءة لأي مسجل دخول، الإضافة لمدير النظام فقط
	mux.Handle("GET /api/services", middleware.Chain(http.HandlerFunc(serviceHandler.List), requireAuth))
	// المهارات بقائمة مسطّحة — لاختيار مهارات البرنامج التدريبي
	mux.Handle("GET /api/skills", middleware.Chain(http.HandlerFunc(serviceHandler.ListSkills), requireAuth))
	mux.Handle("POST /api/services", middleware.Chain(http.HandlerFunc(serviceHandler.Create), requireAuth, requireContentTech))
	mux.Handle("POST /api/services/{id}/skills", middleware.Chain(http.HandlerFunc(serviceHandler.CreateSkill), requireAuth, requireContentTech))
	mux.Handle("DELETE /api/services/{id}", middleware.Chain(http.HandlerFunc(serviceHandler.Delete), requireAuth, requireAdmin))

	// العملاء — أي مسجل دخول يقدر يبحث وينشئ عميل (يطابق سلوك المبيعات بالباك إند القديم)
	mux.Handle("GET /api/customers", middleware.Chain(http.HandlerFunc(customerHandler.List), requireAuth))
	mux.Handle("GET /api/customers/gps", middleware.Chain(http.HandlerFunc(customerHandler.ListGps), requireAuth))
	mux.Handle("GET /api/customers/lookup", middleware.Chain(http.HandlerFunc(customerHandler.Lookup), requireAuth))
	mux.Handle("POST /api/customers", middleware.Chain(http.HandlerFunc(customerHandler.FindOrCreate), requireAuth, requireCustomerMgmt))
	mux.Handle("PUT /api/customers/{id}", middleware.Chain(http.HandlerFunc(customerHandler.Update), requireAuth, requireCustomerMgmt))

	// الحجوزات — دورة حياة الحجز الكاملة، كل خطوة تتطلب تسجيل دخول فقط (الصلاحية الدقيقة تُفرض بالواجهة حالياً)
	mux.Handle("GET /api/bookings", middleware.Chain(http.HandlerFunc(bookingHandler.List), requireAuth))
	mux.Handle("POST /api/bookings", middleware.Chain(http.HandlerFunc(bookingHandler.Create), requireAuth))
	mux.Handle("PUT /api/bookings/{id}/confirm", middleware.Chain(http.HandlerFunc(bookingHandler.Confirm), requireAuth, requireBookingCoord))
	mux.Handle("PUT /api/bookings/{id}/details", middleware.Chain(http.HandlerFunc(bookingHandler.UpdateDetails), requireAuth, requireBookingEdit))
	// حارس ملكية الحجز: أي إجراء على مسار عمل الحجز (موعد/بدء/وصول/تجهيز
	// مواد/إنهاء) ينسمح بس لـ:
	//   - الموظف المكلّف بالحجز فعلاً (أو مشرفه/مسؤول مصاريفه/الي رحّله)
	//   - أو صاحب صلاحية تنسيق/إدارة الكوادر (يتصرف بكل الحجوزات بحكم دوره)
	// قبل هذا الحارس، أي موظف مسجّل دخول كان يقدر ينهي حجز موظف ثاني أو
	// يغيّر موعده (ثغرة IDOR).
	requireBookingParty := func(next http.Handler) http.Handler {
		coordGuard := requireBookingCoord(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if id := middleware.EmployeeIDFromContext(r); id != "" {
				if ok, err := bookingService.IsAssignedTo(r.PathValue("id"), id); err == nil && ok {
					next.ServeHTTP(w, r)
					return
				}
			}
			coordGuard.ServeHTTP(w, r)
		})
	}

	mux.Handle("PUT /api/bookings/{id}/schedule", middleware.Chain(http.HandlerFunc(bookingHandler.Schedule), requireAuth, requireBookingParty))
	mux.Handle("GET /api/bookings/{id}/schedule-log", middleware.Chain(http.HandlerFunc(bookingHandler.ScheduleLog), requireAuth))

	// ═══ أرشيف الحجوزات وتأجيلها وانتظار رد الزبون ═══
	// الأرشيف للي عنده صلاحية تنسيق أو إدارة — هو سجل قرارات الإلغاء.
	mux.Handle("GET /api/bookings/archived", middleware.Chain(http.HandlerFunc(bookingHandler.ListArchived), requireAuth, requireCoordinator))
	// الحذف صار أرشفة: الحجز يختفي من الشاشات ويضل بالأرشيف بسببه
	mux.Handle("DELETE /api/bookings/{id}", middleware.Chain(http.HandlerFunc(bookingHandler.ArchiveBooking), requireAuth, requireAdmin))
	mux.Handle("PUT /api/bookings/{id}/restore", middleware.Chain(http.HandlerFunc(bookingHandler.RestoreBooking), requireAuth, requireAdmin))
	// الكنسة اليدوية للإدارة والفحص — التلقائية تشتغل بالخلفية
	mux.Handle("POST /api/bookings/waiting-reminder-sweep", middleware.Chain(http.HandlerFunc(bookingHandler.RunWaitingReminderSweep), requireAuth, requireAdmin))
	// باسورد مركز القيادة — المالك بس بهاي المرحلة
	mux.Handle("PUT /api/auth/command-password", middleware.Chain(http.HandlerFunc(authHandler.SetCommandPassword), requireAuth, middleware.RequireOwner()))
	// المؤجلة بلا موعد: قائمة مستقلة لأنها منزاحة عن جدول اليوم قصداً
	mux.Handle("GET /api/bookings/postponed", middleware.Chain(http.HandlerFunc(bookingHandler.ListPostponed), requireAuth, requireCoordinator))
	// التأجيل والانتظار شغل المنسّق — هو الي يتصل بالزبون
	mux.Handle("PUT /api/bookings/{id}/postpone", middleware.Chain(http.HandlerFunc(bookingHandler.Postpone), requireAuth, requireCoordinator))
	mux.Handle("PUT /api/bookings/{id}/waiting", middleware.Chain(http.HandlerFunc(bookingHandler.MarkWaiting), requireAuth, requireCoordinator))
	mux.Handle("PUT /api/bookings/{id}/resume", middleware.Chain(http.HandlerFunc(bookingHandler.ResumeFromWaiting), requireAuth, requireCoordinator))
	mux.Handle("PUT /api/bookings/{id}/assign", middleware.Chain(http.HandlerFunc(bookingHandler.Assign), requireAuth, requireBookingCoord))
	mux.Handle("PUT /api/bookings/{id}/supervisor", middleware.Chain(http.HandlerFunc(bookingHandler.Supervisor), requireAuth, requireBookingCoord))
	mux.Handle("PUT /api/bookings/{id}/start", middleware.Chain(http.HandlerFunc(bookingHandler.Start), requireAuth, requireBookingParty))
	mux.Handle("PUT /api/bookings/{id}/arrived", middleware.Chain(http.HandlerFunc(bookingHandler.MarkArrived), requireAuth, requireBookingParty))
	mux.Handle("PUT /api/bookings/{id}/materials-ready", middleware.Chain(http.HandlerFunc(bookingHandler.SetMaterialsReady), requireAuth, requireBookingParty))
	// توقف العمل ورجوعه — بيد طرف الحجز نفسه (الليدر/الكادر المكلّف)
	mux.Handle("PUT /api/bookings/{id}/stop-work", middleware.Chain(http.HandlerFunc(bookingHandler.StopWork), requireAuth, requireBookingParty))
	mux.Handle("PUT /api/bookings/{id}/resume-work", middleware.Chain(http.HandlerFunc(bookingHandler.ResumeWork), requireAuth, requireBookingParty))
	// ── الإنجاز الجزئي ──
	// نفس صلاحية «تم الإنجاز» بالضبط: الي يقدر يقفل الحجز يقدر يأشّره
	// جزئي. لو ضيّقناها أكثر، الليدر يوكف بلا خيار آخر اليوم ويأشّر
	// «تم الإنجاز» على شغل ناقص — وهاي المشكلة الي نحلها أصلاً.
	mux.Handle("POST /api/bookings/{id}/partial-complete", middleware.Chain(http.HandlerFunc(bookingProgressHandler.PartialComplete), requireAuth, requireBookingParty))
	mux.Handle("POST /api/bookings/{id}/schedule-continuation", middleware.Chain(http.HandlerFunc(bookingProgressHandler.ScheduleContinuation), requireAuth, requireCoordinator))
	mux.Handle("GET /api/bookings/{id}/progress", middleware.Chain(http.HandlerFunc(bookingProgressHandler.Reports), requireAuth))
	mux.Handle("GET /api/bookings/{id}/suggested-crew", middleware.Chain(http.HandlerFunc(bookingProgressHandler.SuggestedCrew), requireAuth))

	// تغيير نوع الحجز: المالك ومدير النظام بس. النوع يأثر على الإحصاءات
	// والعمولات وحساب الصيانة — فمو قرار إداري يومي.
	mux.Handle("PUT /api/bookings/{id}/type", middleware.Chain(http.HandlerFunc(bookingHandler.ChangeType), requireAuth, requireAdmin))
	mux.Handle("PUT /api/bookings/{id}/complete", middleware.Chain(http.HandlerFunc(bookingHandler.Complete), requireAuth, requireBookingParty))

	// حذف الحجوزات التجريبية والملغاة: الإداري يطلب، والمراقب أو مدير
	// النظام يبت. الحذف الفوري ما ينفع لأنه ما يترد.
	bookingDeleteRepo := repository.NewBookingDeleteRequestRepository(db)
	bookingDeleteHandler := handler.NewBookingDeleteRequestHandler(bookingDeleteRepo, notificationRepo)
	requireDeleteRequest := middleware.RequireRoleOrPermission(permissionRepo, employeeRepo, notificationRepo, []string{"ADMIN", "OWNER", "HR_COORDINATOR", "MONITOR"}, "booking_delete_request")
	requireDeleteApprove := middleware.RequireRoleOrPermission(permissionRepo, employeeRepo, notificationRepo, []string{"ADMIN", "OWNER", "MONITOR"}, "booking_delete_approve")
	mux.Handle("POST /api/bookings/{id}/delete-request", middleware.Chain(http.HandlerFunc(bookingDeleteHandler.Create), requireAuth, requireDeleteRequest))
	mux.Handle("GET /api/booking-delete-requests", middleware.Chain(http.HandlerFunc(bookingDeleteHandler.List), requireAuth, requireDeleteApprove))
	mux.Handle("PUT /api/booking-delete-requests/{id}/decide", middleware.Chain(http.HandlerFunc(bookingDeleteHandler.Decide), requireAuth, requireDeleteApprove))
	// التدقيق: المحاسب ما يقدر يمرر حجز بلا مبلغ — إما يكتب المبلغ من
	// الفاتورة، أو يأشر خطأ والنظام يوجّهه للمعني (غير مطابق → رقابة
	// وجودة، خطأ سعر → رقابة وإداري).
	bookingAuditRepo := repository.NewBookingAuditRepository(db)
	bookingAuditHandler := handler.NewBookingAuditHandler(bookingAuditRepo, bookingRepo, notificationRepo)
	// التدقيق اليومي: نفس واجهة التدقيق بس بيوم واحد، مع مجاميع اليوم
	dailyAuditHandler := handler.NewDailyAuditHandler(repository.NewDailyAuditRepository(db))
	mux.Handle("GET /api/finance/daily-audit", middleware.Chain(http.HandlerFunc(dailyAuditHandler.Day), requireAuth, requireVerifyBooking))
	mux.Handle("PUT /api/bookings/{id}/audit", middleware.Chain(http.HandlerFunc(bookingAuditHandler.Audit), requireAuth, requireVerifyBooking))
	// شريط الإعلانات: يقراه كل موظف، وينزّله المالك ومدير النظام بس
	announcementHandler := handler.NewAnnouncementHandler(announcementRepo)
	mux.Handle("GET /api/announcements", middleware.Chain(http.HandlerFunc(announcementHandler.List), requireAuth))
	mux.Handle("POST /api/announcements", middleware.Chain(http.HandlerFunc(announcementHandler.Create), requireAuth, requireAdmin))
	mux.Handle("PUT /api/announcements/{id}/active", middleware.Chain(http.HandlerFunc(announcementHandler.SetActive), requireAuth, requireAdmin))
	mux.Handle("DELETE /api/announcements/{id}", middleware.Chain(http.HandlerFunc(announcementHandler.Delete), requireAuth, requireAdmin))
	mux.Handle("GET /api/audit-issues", middleware.Chain(http.HandlerFunc(bookingAuditHandler.ListIssues), requireAuth))
	mux.Handle("PUT /api/audit-issues/{id}/resolve", middleware.Chain(http.HandlerFunc(bookingAuditHandler.ResolveIssue), requireAuth))
	mux.Handle("PUT /api/bookings/{id}/verify", middleware.Chain(http.HandlerFunc(bookingHandler.Verify), requireAuth, requireVerifyBooking))
	// إرجاع الحجز للتدقيق: التدقيق جان قرار نهائي ما إله رجعة. مدير
	// النظام حصراً يكدر يفتحه من جديد حتى ينصلّح أي غلط بالمبلغ.
	mux.Handle("PUT /api/bookings/{id}/unverify", middleware.Chain(http.HandlerFunc(bookingHandler.Unverify), requireAuth, requireAdmin))
	// "تم" الإداري بعد تواصله فعلياً مع الزبون — خطوة سابقة ومنفصلة عن التثبيت
	// نفسه (نفس صلاحية تنسيق الحجوزات coordinator المستخدمة أصلاً بـCoordinator.tsx).
	// إرجاع حجز محوّل لإدارة المشاريع رجعة لكادر الشد — لمدير المشاريع لما
	// يفتح التفاصيل ويلكاه مو مال مشروع.
	mux.Handle("PUT /api/bookings/{id}/return-to-crew", middleware.Chain(http.HandlerFunc(bookingHandler.ReturnToCrew), requireAuth, requireProjectMgmt))
	mux.Handle("PUT /api/bookings/{id}/confirmation-contacted", middleware.Chain(http.HandlerFunc(bookingHandler.MarkConfirmationContacted), requireAuth, requireCoordinator))
	// تدقيق المراقب على الحجوزات الموجّهة قبل التثبيت (crew_management، صلاحية جديدة
	// يقدر الأدمن يمنحها لأي موظف مراقب من صفحة الصلاحيات).
	mux.Handle("GET /api/bookings/pending-audit", middleware.Chain(http.HandlerFunc(bookingHandler.PendingAudit), requireAuth, requireCrewManagement))
	mux.Handle("GET /api/bookings/{id}/tool-checks", middleware.Chain(http.HandlerFunc(bookingHandler.ToolChecks), requireAuth, requireCoordinator))

	// سلة الحجز
	// سلة الحجز: نفس قاعدة مسار عمل الحجز — الموظف طرف بالحجز أو منسّق.
	// {bookingId} نستخدم حارس الحجز نفسه، و{id} (عنصر سلة) نتحقق من حجزه.
	requireCartBookingParty := func(next http.Handler) http.Handler {
		coordGuard := requireBookingCoord(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if eid := middleware.EmployeeIDFromContext(r); eid != "" {
				if ok, err := bookingService.IsAssignedTo(r.PathValue("bookingId"), eid); err == nil && ok {
					next.ServeHTTP(w, r)
					return
				}
			}
			coordGuard.ServeHTTP(w, r)
		})
	}
	requireCartItemParty := func(next http.Handler) http.Handler {
		coordGuard := requireBookingCoord(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if eid := middleware.EmployeeIDFromContext(r); eid != "" {
				if ok, err := bookingService.IsCartItemOfAssignedBooking(r.PathValue("id"), eid); err == nil && ok {
					next.ServeHTTP(w, r)
					return
				}
			}
			coordGuard.ServeHTTP(w, r)
		})
	}
	mux.Handle("GET /api/cart/booking/{bookingId}", middleware.Chain(http.HandlerFunc(cartHandler.ListForBooking), requireAuth, requireCartBookingParty))
	mux.Handle("POST /api/cart/booking/{bookingId}", middleware.Chain(http.HandlerFunc(cartHandler.Create), requireAuth, requireCartBookingParty))
	mux.Handle("PUT /api/cart/{id}", middleware.Chain(http.HandlerFunc(cartHandler.Update), requireAuth, requireCartItemParty))
	mux.Handle("DELETE /api/cart/{id}", middleware.Chain(http.HandlerFunc(cartHandler.Delete), requireAuth, requireCartItemParty))

	// المصاريف — أي موظف يقدر يرسل مصروف، الموافقة/الرفض للمحاسب ومدير النظام فقط
	mux.Handle("GET /api/expenses", middleware.Chain(http.HandlerFunc(expenseHandler.List), requireAuth))
	mux.Handle("POST /api/expenses", middleware.Chain(http.HandlerFunc(expenseHandler.Create), requireAuth))
	mux.Handle("PUT /api/expenses/{id}/status", middleware.Chain(http.HandlerFunc(expenseHandler.UpdateStatus), requireAuth, requireFinance))

	// المخزون — أدوات شخصية / مركبات / أدوات مشتركة / طلبات الأدوات
	mux.Handle("GET /api/inventory/personal", middleware.Chain(http.HandlerFunc(inventoryHandler.ListPersonalTools), requireAuth))
	mux.Handle("POST /api/inventory/personal", middleware.Chain(http.HandlerFunc(inventoryHandler.CreatePersonalTool), requireAuth, requireHROrInventory))
	mux.Handle("PUT /api/inventory/personal/{id}", middleware.Chain(http.HandlerFunc(inventoryHandler.UpdatePersonalTool), requireAuth, requireHROrInventory))
	mux.Handle("DELETE /api/inventory/personal/{id}", middleware.Chain(http.HandlerFunc(inventoryHandler.DeletePersonalTool), requireAuth, requireHROrInventory))
	// سجل حركة الأدوات — يبيّن متى انفقدت كل أداة ومنو سجّل الفقدان
	mux.Handle("GET /api/inventory/tool-events", middleware.Chain(http.HandlerFunc(inventoryHandler.ToolEvents), requireAuth))

	// العدة القياسية (PersonalToolTemplateItem) — القراءة مفتوحة لأي موظف، الإضافة/الحذف
	// لمن عنده صلاحية "inventory" أو HR/ADMIN (نفس requireHROrInventory).
	mux.Handle("GET /api/inventory/personal-template", middleware.Chain(http.HandlerFunc(inventoryHandler.ListPersonalToolTemplateItems), requireAuth))
	mux.Handle("POST /api/inventory/personal-template", middleware.Chain(http.HandlerFunc(inventoryHandler.CreatePersonalToolTemplateItem), requireAuth, requireHROrInventory))
	mux.Handle("DELETE /api/inventory/personal-template/{id}", middleware.Chain(http.HandlerFunc(inventoryHandler.DeletePersonalToolTemplateItem), requireAuth, requireHROrInventory))

	// جرد يومي: الموظف يؤكد جرد عدته الخاصة، الإداري يشوف نتائج اليوم لكل الموظفين
	mux.Handle("POST /api/inventory/checks", middleware.Chain(http.HandlerFunc(inventoryHandler.CreateInventoryCheck), requireAuth))
	// نتائج جرد كل الفنيين = شاشة متابعة، مو شغل الفني. جانت مفتوحة لأي
	// موظف مسجّل دخول، فالفني يشوف نواقص زملاءه. صارت بصلاحية «جرد
	// الأدوات» — الفني يسوّي جرده هو ويشوف حالته من المسار الي بعده.
	mux.Handle("GET /api/inventory/checks/today", middleware.Chain(http.HandlerFunc(inventoryHandler.TodaysInventoryChecks), requireAuth, requireInventoryView))
	mux.Handle("GET /api/inventory/checks/mine", middleware.Chain(http.HandlerFunc(inventoryHandler.MyLastInventoryCheck), requireAuth))
	mux.Handle("POST /api/inventory/checks/{id}/resolve", middleware.Chain(http.HandlerFunc(inventoryHandler.ResolveInventoryCheck), requireAuth, requireHR))

	mux.Handle("GET /api/inventory/vehicle", middleware.Chain(http.HandlerFunc(inventoryHandler.ListVehicleTools), requireAuth))
	mux.Handle("POST /api/inventory/vehicle", middleware.Chain(http.HandlerFunc(inventoryHandler.CreateVehicleTool), requireAuth, requireHROrInventory))
	// تعديل/حذف أداة مركبة موجودة يقتصر على إدارة الكوادر (نفس صلاحية أدوات
	// الأدوات الشخصية requireHR) — الإنشاء يبقى مفتوح لأي موظف (نفس السلوك القديم).
	mux.Handle("PUT /api/inventory/vehicle/{id}", middleware.Chain(http.HandlerFunc(inventoryHandler.UpdateVehicleTool), requireAuth, requireHROrInventory))
	mux.Handle("DELETE /api/inventory/vehicle/{id}", middleware.Chain(http.HandlerFunc(inventoryHandler.DeleteVehicleTool), requireAuth, requireHROrInventory))
	mux.Handle("GET /api/inventory/vehicle-tool-checks", middleware.Chain(http.HandlerFunc(inventoryHandler.ListVehicleToolChecks), requireAuth))
	mux.Handle("GET /api/inventory/booking-tool-checks", middleware.Chain(http.HandlerFunc(inventoryHandler.ListAllBookingToolChecks), requireAuth))

	mux.Handle("GET /api/inventory/ondemand", middleware.Chain(http.HandlerFunc(inventoryHandler.ListOnDemandTools), requireAuth))
	mux.Handle("POST /api/inventory/ondemand", middleware.Chain(http.HandlerFunc(inventoryHandler.CreateOnDemandTool), requireAuth, requireProcurementAdmin))
	// تعديل أداة "حسب الحاجة" يقتصر على الأدمن أو إداري الكميات (نفس canManageOnDemand
	// بالواجهة: isAdmin || PROCUREMENT_ADMIN) — يطابق requireProcurementAdmin الموجود.
	mux.Handle("PUT /api/inventory/ondemand/{id}", middleware.Chain(http.HandlerFunc(inventoryHandler.UpdateOnDemandTool), requireAuth, requireProcurementAdmin))

	// إضافة الكميات للمخزون — إداري الكميات حصراً
	mux.Handle("POST /api/inventory/stock-intake", middleware.Chain(http.HandlerFunc(inventoryHandler.AddStock), requireAuth, requireProcurementAdmin))
	mux.Handle("GET /api/inventory/stock-intake", middleware.Chain(http.HandlerFunc(inventoryHandler.ListStockIntakes), requireAuth, requireProcurementAdmin))

	mux.Handle("GET /api/inventory/requests", middleware.Chain(http.HandlerFunc(inventoryHandler.ListToolRequests), requireAuth))
	// طلب أداة "حسب الحاجة" مقصور على الليدر فقط (isLeader فريش من قاعدة البيانات) —
	// الموظف العادي يبقى يشوف حالة طلباته (GET) بس ما يقدر ينشئ طلب جديد.
	mux.Handle("POST /api/inventory/requests", middleware.Chain(http.HandlerFunc(inventoryHandler.CreateToolRequest), requireAuth, requireLeader))
	mux.Handle("PUT /api/inventory/requests/{id}/approve", middleware.Chain(http.HandlerFunc(inventoryHandler.ApproveToolRequest), requireAuth, requireInventoryApprove))
	mux.Handle("PUT /api/inventory/requests/{id}/reject", middleware.Chain(http.HandlerFunc(inventoryHandler.RejectToolRequest), requireAuth, requireInventoryApprove))
	// إرجاع الأداة يشيل مسؤوليتها عن الموظف — نفس مستوى الموافقة، لأن
	// إذا أي موظف يقدر يأشر «رجعتها» تروح كل محاسبة العدة المفقودة.
	mux.Handle("PUT /api/inventory/requests/{id}/return", middleware.Chain(http.HandlerFunc(inventoryHandler.ReturnToolRequest), requireAuth, requireInventoryApprove))
	// حذف طلب أداة يمحي أثر الطلب نهائياً — محصور بمدير النظام والمالك فقط،
	// مو بكل من يقدر يوافق أو يرفض.
	mux.Handle("DELETE /api/inventory/requests/{id}", middleware.Chain(http.HandlerFunc(inventoryHandler.DeleteToolRequest), requireAuth, requireAdmin))

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
	// تأشير الاتصال والملاحظات: أي موظف يقدر — المهم النظام يخزن منو
	// اتصل. حصرها بالجودة يخلي الشكوى تضل بلا متابعة لو المهندس مشغول.
	mux.Handle("PUT /api/complaints/{id}/contact", middleware.Chain(http.HandlerFunc(complaintHandler.SetContacted), requireAuth, requireComplaintContact))
	mux.Handle("PUT /api/complaints/{id}/notes", middleware.Chain(http.HandlerFunc(complaintHandler.SetNotes), requireAuth, requireComplaintContact))
	mux.Handle("PUT /api/complaints/{id}/resolve", middleware.Chain(http.HandlerFunc(complaintHandler.Resolve), requireAuth, requireQuality))
	mux.Handle("GET /api/complaints/stats", middleware.Chain(http.HandlerFunc(complaintHandler.Stats), requireAuth))
	mux.Handle("GET /api/quality-follow-ups", middleware.Chain(http.HandlerFunc(qualityFollowUpHandler.List), requireAuth, requireQuality))
	mux.Handle("PUT /api/quality-follow-ups/{id}", middleware.Chain(http.HandlerFunc(qualityFollowUpHandler.Update), requireAuth, requireQuality))
	// حكم الجودة: تقرير إيجابي/سلبي، والكشف الميداني. نفس صلاحية الشاشة
	// نفسها — الي يشوف المتابعة هو الي يحكم بيها.
	mux.Handle("POST /api/quality-follow-ups/{id}/verdict", middleware.Chain(http.HandlerFunc(qualityFollowUpHandler.Verdict), requireAuth, requireQuality))
	mux.Handle("POST /api/quality-follow-ups/{id}/inspect", middleware.Chain(http.HandlerFunc(qualityFollowUpHandler.Inspect), requireAuth, requireQuality))

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
	mux.Handle("POST /api/missions", middleware.Chain(http.HandlerFunc(missionHandler.Create), requireAuth, requireBookingCoord))
	// تقدّم مراحل المهمة: إدارة المشاريع أو المراقب — مو أي موظف يحرّك
	// مهمة غيره.
	mux.Handle("PUT /api/missions/{id}/stage", middleware.Chain(http.HandlerFunc(missionHandler.UpdateStage), requireAuth, requireFieldMonitor))

	// إدارة المشاريع (projects)
	mux.Handle("GET /api/projects", middleware.Chain(http.HandlerFunc(projectHandler.List), requireAuth))
	mux.Handle("GET /api/projects/{id}", middleware.Chain(http.HandlerFunc(projectHandler.Get), requireAuth))
	// إنشاء مشروع جديد يتطلب نفس دور مدير المشاريع/الأدمن المطلوب للتعديل والحذف
	// تحته مباشرة — كان مفتوح غلط لأي موظف مسجل دخول فقط (requireAuth بدون requireProjectManager)،
	// عدم اتساق مع PUT/DELETE على نفس المورد.
	// الإنشاء يقبل كذلك صلاحية "إضافة مشروع فقط" المبسّطة — بينما التعديل/الحذف
	// والترحيل بين المراحل يبقى محصوراً بمدير المشاريع/الأدمن.
	mux.Handle("POST /api/projects", middleware.Chain(http.HandlerFunc(projectHandler.Create), requireAuth,
		middleware.RequireRoleOrAnyPermission(permissionRepo, employeeRepo, notificationRepo,
			[]string{"PROJECT_MANAGER"}, "project_management", "project_create_only")))
	// المشاريع المُسلَّمة للموظف الحالي — ما تحتاج صلاحية إدارة مشاريع عامة،
	// التسليم نفسه هو الصلاحية وعلى هالمشاريع بس.
	mux.Handle("GET /api/projects/delegated-to-me", middleware.Chain(http.HandlerFunc(projectHandler.ListDelegatedToMe), requireAuth))
	// requireProjectMgmt مو requireProjectManager: بقية مسارات المشاريع تقبل
	// صلاحية project_management الممنوحة يدوياً، وهذا المسار كان الوحيد المقيّد
	// بالدور — فمهندس عنده الصلاحية يشوف رابط "إحصائيات المشاريع" بقائمته ويوكع بـ403.
	mux.Handle("GET /api/projects/statistics", middleware.Chain(http.HandlerFunc(projectHandler.Statistics), requireAuth, requireProjectMgmt))
	mux.Handle("GET /api/projects/{id}/delegation-log", middleware.Chain(http.HandlerFunc(projectHandler.DelegationLog), requireAuth, requireProjectManager))
	// التسليم/السحب بيد مدير المشاريع بس — الموظف المُسلَّم إله ما يقدر يسلّمه لغيره
	mux.Handle("PUT /api/projects/{id}/delegate", middleware.Chain(http.HandlerFunc(projectHandler.Delegate), requireAuth, requireProjectManager))

	// التعديل: مدير المشاريع على أي مشروع، أو الموظف المُسلَّم إله على مشروعه هو
	// فقط. أي موظف ثاني ينرفض حتى لو حاول يعدّل مشروع مو إله.
	allowManagerOrDelegate := func(next http.Handler) http.Handler {
		managerGuard := requireProjectManager(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			employeeID := middleware.EmployeeIDFromContext(r)
			if employeeID != "" {
				if ok, err := projectService.IsDelegatedTo(r.PathValue("id"), employeeID); err == nil && ok {
					next.ServeHTTP(w, r)
					return
				}
			}
			managerGuard.ServeHTTP(w, r)
		})
	}
	mux.Handle("PUT /api/projects/{id}", middleware.Chain(http.HandlerFunc(projectHandler.Update), requireAuth, allowManagerOrDelegate))
	mux.Handle("DELETE /api/projects/{id}", middleware.Chain(http.HandlerFunc(projectHandler.Delete), requireAuth, requireProjectManager))
	// حذف ملف العقد المرفوع — لمدير النظام حصراً
	mux.Handle("DELETE /api/projects/{id}/contract", middleware.Chain(http.HandlerFunc(projectHandler.DeleteContract), requireAuth, requireAdmin))

	// أنواع الأعمال ("نوع العمل" بحقل المشروع) — إعدادات وحدة إدارة المشاريع:
	// أي موظف مسجل دخول يشوف القائمة (يحتاجها بفورمة المشروع)، بس الإضافة/الحذف
	// محصورة بنفس صلاحية مدير المشاريع المستخدمة لبقية إعدادات هذي الوحدة.
	mux.Handle("GET /api/project-work-types", middleware.Chain(http.HandlerFunc(projectWorkTypeHandler.List), requireAuth))
	// مرشحو المشروع (المسؤول / منفّذ الكشف) مصنّفين ومرتّبين
	mux.Handle("GET /api/project-candidates", middleware.Chain(http.HandlerFunc(projectWorkTypeHandler.ListCandidates), requireAuth))

	// الشخصيات المهمة (VIP): أي موظف يعلّم بضغطة زر، بس التفاصيل الكاملة
	// (رقم الزبون وشنو طلب ومنو علّمه) تُعرض لمدير النظام حصراً.
	mux.Handle("GET /api/vip-customers", middleware.Chain(http.HandlerFunc(vipCustomerHandler.List), requireAuth, requireAdmin))
	mux.Handle("GET /api/vip-customers/ids", middleware.Chain(http.HandlerFunc(vipCustomerHandler.ListIDs), requireAuth))
	// الإضافة اليدوية للشخصية المهمة: الإداري ومدير النظام حصراً — مو أي
	// موظف عنده صلاحية العملاء. (الترحيل التلقائي من المشاريع يمر بالخدمة
	// مو بهذا المسار، فما يتأثر.)
	mux.Handle("POST /api/vip-customers", middleware.Chain(http.HandlerFunc(vipCustomerHandler.Mark), requireAuth, requireVipManualAdd))
	mux.Handle("DELETE /api/vip-customers/{customerId}", middleware.Chain(http.HandlerFunc(vipCustomerHandler.Unmark), requireAuth, requireAdmin))
	mux.Handle("POST /api/project-work-types", middleware.Chain(http.HandlerFunc(projectWorkTypeHandler.Create), requireAuth, requireProjectManager))
	mux.Handle("DELETE /api/project-work-types/{id}", middleware.Chain(http.HandlerFunc(projectWorkTypeHandler.Delete), requireAuth, requireProjectManager))

	// الكشوفات: فورمات فارغة يطبعها المهندس، يمليها بالموقع، وبعدين يرفع صور
	// الفورمة المالية — أي موظف مسجل دخول يقدر ينشئ/يرفع (مو حصراً مدير مشاريع).
	mux.Handle("GET /api/checklists", middleware.Chain(http.HandlerFunc(checklistHandler.List), requireAuth))
	mux.Handle("POST /api/checklists", middleware.Chain(http.HandlerFunc(checklistHandler.Create), requireAuth, requireProjectMgmt))
	mux.Handle("PUT /api/checklists/{id}/photos", middleware.Chain(http.HandlerFunc(checklistHandler.AddPhotos), requireAuth, requireProjectMgmt))

	// معرض أعمال التقنيين — أي موظف مسجل دخول يتصفحه (إلهام)، بس صاحب صلاحية
	// "التقني" (content_technician) بس يقدر يضيف عمل/يرفع وسائط.
	mux.Handle("GET /api/tech-showcase", middleware.Chain(http.HandlerFunc(techShowcaseHandler.List), requireAuth))
	mux.Handle("POST /api/tech-showcase", middleware.Chain(http.HandlerFunc(techShowcaseHandler.Create), requireAuth, requireContentTech))
	mux.Handle("PUT /api/tech-showcase/{id}/media", middleware.Chain(http.HandlerFunc(techShowcaseHandler.AddMedia), requireAuth, requireContentTech))

	// وحدة التقنيين — إدارة المعارض
	mux.Handle("GET /api/exhibitions", middleware.Chain(http.HandlerFunc(exhibitionHandler.List), requireAuth, requireUnitTechnicians))
	mux.Handle("POST /api/exhibitions", middleware.Chain(http.HandlerFunc(exhibitionHandler.Create), requireAuth, requireUnitTechnicians))
	mux.Handle("PUT /api/exhibitions/{id}/nominate", middleware.Chain(http.HandlerFunc(exhibitionHandler.Nominate), requireAuth, requireAdmin))
	mux.Handle("PUT /api/exhibitions/{id}/photos", middleware.Chain(http.HandlerFunc(exhibitionHandler.AddPhotos), requireAuth, requireUnitTechnicians))
	mux.Handle("PUT /api/exhibitions/{id}/findings", middleware.Chain(http.HandlerFunc(exhibitionHandler.SetFindings), requireAuth, requireUnitTechnicians))
	mux.Handle("POST /api/exhibitions/{id}/report", middleware.Chain(http.HandlerFunc(exhibitionHandler.GenerateReport), requireAuth, requireUnitTechnicians))
	mux.Handle("PUT /api/exhibitions/{id}/archive", middleware.Chain(http.HandlerFunc(exhibitionHandler.Archive), requireAuth, requireAdmin))

	// وحدة التقنيين — إدارة المنتجات
	mux.Handle("GET /api/product-requests", middleware.Chain(http.HandlerFunc(productRequestHandler.List), requireAuth, requireTechUnitOrProcurement))
	mux.Handle("POST /api/product-requests", middleware.Chain(http.HandlerFunc(productRequestHandler.Create), requireAuth, requireTechUnitOrProcurement))
	mux.Handle("PUT /api/product-requests/{id}/approve", middleware.Chain(http.HandlerFunc(productRequestHandler.Approve), requireAuth, requireAdmin))
	mux.Handle("PUT /api/product-requests/{id}/reject", middleware.Chain(http.HandlerFunc(productRequestHandler.Reject), requireAuth, requireAdmin))

	// تجهيز طلب المنتج من الدوار — أبو الحسابات (نفسه أبو الكميات)
	mux.Handle("GET /api/product-procurements", middleware.Chain(http.HandlerFunc(productProcurementHandler.List), requireAuth, requireFund))
	mux.Handle("POST /api/product-requests/{id}/fulfill", middleware.Chain(http.HandlerFunc(productProcurementHandler.Fulfill), requireAuth, requireFund))
	mux.Handle("PUT /api/product-procurements/{id}/settle", middleware.Chain(http.HandlerFunc(productProcurementHandler.Settle), requireAuth, requireFund))

	// وحدة التقنيين — إدارة الخدمات
	mux.Handle("GET /api/service-studies", middleware.Chain(http.HandlerFunc(serviceStudyHandler.List), requireAuth, requireUnitTechnicians))
	mux.Handle("POST /api/service-studies", middleware.Chain(http.HandlerFunc(serviceStudyHandler.Create), requireAuth, requireUnitTechnicians))
	mux.Handle("PUT /api/service-studies/{id}/assign", middleware.Chain(http.HandlerFunc(serviceStudyHandler.Assign), requireAuth, requireAdmin))
	mux.Handle("POST /api/service-studies/{id}/reports", middleware.Chain(http.HandlerFunc(serviceStudyHandler.AddReport), requireAuth, requireTechUnitOrProcurement))
	mux.Handle("PUT /api/service-studies/{id}/archive", middleware.Chain(http.HandlerFunc(serviceStudyHandler.Archive), requireAuth, requireAdmin))

	// وحدة التصميم — عدة استمارات مستقلة، كل وحدة بأسئلتها وبرابطها العام الخاص
	mux.Handle("GET /api/design-forms", middleware.Chain(http.HandlerFunc(designFormHandler.ListForms), requireAuth, requireAdmin))
	mux.Handle("POST /api/design-forms", middleware.Chain(http.HandlerFunc(designFormHandler.CreateForm), requireAuth, requireAdmin))
	mux.Handle("DELETE /api/design-forms/{id}", middleware.Chain(http.HandlerFunc(designFormHandler.DeleteForm), requireAuth, requireAdmin))
	mux.Handle("GET /api/design-forms/{formId}/submissions", middleware.Chain(http.HandlerFunc(designFormHandler.ListSubmissions), requireAuth, requireAdmin))
	mux.Handle("GET /api/design-forms/{formId}/questions", middleware.Chain(http.HandlerFunc(designFormHandler.List), requireAuth, requireAdmin))
	mux.Handle("POST /api/design-forms/{formId}/questions", middleware.Chain(http.HandlerFunc(designFormHandler.Create), requireAuth, requireAdmin))
	mux.Handle("PUT /api/design-form/questions/{id}", middleware.Chain(http.HandlerFunc(designFormHandler.Update), requireAuth, requireAdmin))
	mux.Handle("DELETE /api/design-form/questions/{id}", middleware.Chain(http.HandlerFunc(designFormHandler.Delete), requireAuth, requireAdmin))
	mux.Handle("PUT /api/design-form/questions/reorder", middleware.Chain(http.HandlerFunc(designFormHandler.Reorder), requireAuth, requireAdmin))

	// رابط عام للزبون (بدون تسجيل دخول) — يشوف الاستمارة ويرسل جوابه فقط
	// النماذج العامة: بلا تسجيل دخول، فلازم حد صارم — بدونه أي واحد عنده
	// الرابط يقدر يغرق قاعدة البيانات بآلاف التقديمات الضخمة.
	publicFormReadLimit := middleware.RateLimit(30, time.Minute)
	publicFormSubmitLimit := middleware.RateLimit(5, time.Minute)
	mux.Handle("GET /api/public/design-forms/{token}", publicFormReadLimit(http.HandlerFunc(designFormHandler.PublicGet)))
	mux.Handle("POST /api/public/design-forms/{token}/submit", publicFormSubmitLimit(http.HandlerFunc(designFormHandler.PublicSubmit)))

	// طلبات تغيير أيقونة الحضور — أي موظف يطلب، ومدير النظام بس يوافق/يرفض
	mux.Handle("POST /api/attendance-icon-requests", middleware.Chain(http.HandlerFunc(attendanceIconRequestHandler.Create), requireAuth))
	mux.Handle("GET /api/attendance-icon-requests", middleware.Chain(http.HandlerFunc(attendanceIconRequestHandler.ListPending), requireAuth, requireAdmin))
	mux.Handle("PUT /api/attendance-icon-requests/{id}/approve", middleware.Chain(http.HandlerFunc(attendanceIconRequestHandler.Approve), requireAuth, requireAdmin))
	mux.Handle("PUT /api/attendance-icon-requests/{id}/reject", middleware.Chain(http.HandlerFunc(attendanceIconRequestHandler.Reject), requireAuth, requireAdmin))

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
	mux.Handle("POST /api/performance-reviews", middleware.Chain(http.HandlerFunc(performanceReviewHandler.Create), requireAuth, requireKpiMgmt))
	mux.Handle("GET /api/performance-reviews", middleware.Chain(http.HandlerFunc(performanceReviewHandler.List), requireAuth))
	mux.Handle("GET /api/performance-reviews/ratable", middleware.Chain(http.HandlerFunc(performanceReviewHandler.Ratable), requireAuth))
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
	mux.Handle("POST /api/suppliers/specialties", middleware.Chain(http.HandlerFunc(supplierHandler.CreateSpecialty), requireAuth, requireSuppliersMgmt))
	mux.Handle("DELETE /api/suppliers/specialties/{id}", middleware.Chain(http.HandlerFunc(supplierHandler.DeleteSpecialty), requireAuth, requireAdmin))
	// حل روابط الخرائط (خصوصاً المختصرة مثل maps.app.goo.gl الي ما تحتوي
	// إحداثيات) — لازم يمر بالسيرفر لأن المتصفح ما يقدر يتبع التحويل بسبب CORS.
	mux.Handle("GET /api/geo/resolve-map-link", middleware.Chain(http.HandlerFunc(mapLinkHandler.Resolve), requireAuth))

	// سياسة الخصوصية — القراءة والموافقة لأي موظف، والإدارة بصلاحية مستقلة
	requirePrivacyMgmt := middleware.RequirePermission(permissionRepo, employeeRepo, notificationRepo, "privacy_policy_manage")
	mux.Handle("GET /api/privacy-policy", middleware.Chain(http.HandlerFunc(privacyPolicyHandler.List), requireAuth))
	mux.Handle("GET /api/privacy-policy/status", middleware.Chain(http.HandlerFunc(privacyPolicyHandler.Status), requireAuth))
	mux.Handle("POST /api/privacy-policy/accept", middleware.Chain(http.HandlerFunc(privacyPolicyHandler.Accept), requireAuth))
	mux.Handle("POST /api/privacy-policy", middleware.Chain(http.HandlerFunc(privacyPolicyHandler.Create), requireAuth, requirePrivacyMgmt))
	mux.Handle("PUT /api/privacy-policy/{id}", middleware.Chain(http.HandlerFunc(privacyPolicyHandler.Update), requireAuth, requirePrivacyMgmt))
	mux.Handle("DELETE /api/privacy-policy/{id}", middleware.Chain(http.HandlerFunc(privacyPolicyHandler.Delete), requireAuth, requirePrivacyMgmt))

	// بحث المناطق يمر من سيرفرنا — خدمة الخرائط تحجب المتصفحات
	mux.Handle("GET /api/geo/search", middleware.Chain(http.HandlerFunc(geoHandler.Search), requireAuth))

	mux.Handle("GET /api/suppliers", middleware.Chain(http.HandlerFunc(supplierHandler.List), requireAuth))
	mux.Handle("POST /api/suppliers", middleware.Chain(http.HandlerFunc(supplierHandler.Create), requireAuth, requireSuppliersMgmt))
	mux.Handle("PUT /api/suppliers/{id}", middleware.Chain(http.HandlerFunc(supplierHandler.Update), requireAuth, requireSuppliersMgmt))
	// الحذف داخل نفس صلاحية إدارة الموردين — منو يقدر يضيف ويعدل يقدر يحذف،
	// حتى ما يضطر يرجع للأدمن على كل مورد غلط.
	mux.Handle("DELETE /api/suppliers/{id}", middleware.Chain(http.HandlerFunc(supplierHandler.Delete), requireAuth, requireSuppliersMgmt))
	mux.Handle("POST /api/suppliers/{id}/rate", middleware.Chain(http.HandlerFunc(supplierHandler.Rate), requireAuth, requireSuppliersMgmt))

	// عروض الأسعار (quotations)
	mux.Handle("GET /api/quotations", middleware.Chain(http.HandlerFunc(quotationHandler.List), requireAuth))
	mux.Handle("GET /api/quotations/{id}", middleware.Chain(http.HandlerFunc(quotationHandler.Get), requireAuth))
	// الموظف الموجّه له مشروع لازم يقدر يسوي عرض سعر لمشروعه — التوجيه نفسه
	// هو الصلاحية. غيره يضل محتاج صلاحية عروض الأسعار.
	allowQuotationOrDelegate := func(next http.Handler) http.Handler {
		guard := requireQuotationAccess(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if id := middleware.EmployeeIDFromContext(r); id != "" {
				if ok, err := projectService.HasAnyDelegation(id); err == nil && ok {
					next.ServeHTTP(w, r)
					return
				}
			}
			guard.ServeHTTP(w, r)
		})
	}
	mux.Handle("POST /api/quotations", middleware.Chain(http.HandlerFunc(quotationHandler.Create), requireAuth, allowQuotationOrDelegate))
	mux.Handle("PUT /api/quotations/{id}", middleware.Chain(http.HandlerFunc(quotationHandler.Update), requireAuth, allowQuotationOrDelegate))
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
	mux.Handle("GET /api/gps/customers", middleware.Chain(http.HandlerFunc(gpsHandler.ListCustomers), requireAuth, requireGpsData))
	mux.Handle("POST /api/gps/customers", middleware.Chain(http.HandlerFunc(gpsHandler.CreateCustomer), requireAuth, requireGpsSystem))
	mux.Handle("PUT /api/gps/customers/{id}", middleware.Chain(http.HandlerFunc(gpsHandler.UpdateCustomer), requireAuth, requireGpsSystem))

	// ── الدوار ────────────────────────────────────────────────────────────────
	// الإدارة والصرف والتدقيق: صلاحية "الدوار" (المحاسب). أما رصيد الموظف
	// نفسه ورفع تسويته فمفتوحين لأي موظف مسجّل دخول — كل واحد يشوف حركاته هو بس.
	// حساب تكاليف الشد — تفصيلي لكل الكوادر، ضمن خانة الحسابات
	mux.Handle("GET /api/finance/gps-install-costs", middleware.Chain(http.HandlerFunc(gpsInstallCostHandler.Summary), requireAuth, requireFinance))

	// ── الإجازات ──────────────────────────────────────────────────────────────
	// أي موظف يقدّم ويشوف طلباته. البت بالطلبات محصور بالمخوّل حسب مسار
	// الكادر (يُتحقق داخل الهاندلر لأنه يعتمد على مسار الطلب نفسه).
	mux.Handle("POST /api/leaves", middleware.Chain(http.HandlerFunc(leaveHandler.Create), requireAuth))
	mux.Handle("GET /api/leaves/mine", middleware.Chain(http.HandlerFunc(leaveHandler.Mine), requireAuth))
	mux.Handle("DELETE /api/leaves/{id}", middleware.Chain(http.HandlerFunc(leaveHandler.Cancel), requireAuth))
	mux.Handle("GET /api/leaves/inbox", middleware.Chain(http.HandlerFunc(leaveHandler.Inbox), requireAuth))
	mux.Handle("GET /api/leaves/pending-count", middleware.Chain(http.HandlerFunc(leaveHandler.PendingCount), requireAuth))
	mux.Handle("PUT /api/leaves/{id}/preliminary", middleware.Chain(http.HandlerFunc(leaveHandler.Preliminary), requireAuth))
	mux.Handle("PUT /api/leaves/{id}/decide", middleware.Chain(http.HandlerFunc(leaveHandler.Decide), requireAuth))

	// أرقام اللوحة الرئيسية بدون سحب أرشيف الشركة كامل
	mux.Handle("GET /api/dashboard/summary", middleware.Chain(http.HandlerFunc(dashboardHandler.Summary), requireAuth))

	// مجاميع المحاسب والمراقب — تنحسب بقاعدة البيانات بدل تنزيل الأرشيف
	// اثنين يشوفون هذي الأرقام: المحاسب بدوره، والمراقب بصلاحيته. لازم
	// الاثنين سوه — حارس بالدور بس يرفض المراقب، وحارس بالصلاحية بس
	// يرفض المحاسب الجديد الي ما انمنحت له صلاحية صريحة بعد. وأي رفض
	// هنا مو مجرد ٤٠٣: النظام يعدّه محاولة تلاعب ويقفل الحساب بعد ٣ مرات،
	// والصفحة تعاود الطلب لحالها — يعني القفل يجي بثواني.
	mux.Handle("GET /api/dashboard/finance-summary", middleware.Chain(
		http.HandlerFunc(dashboardHandler.FinanceSummary),
		requireAuth,
		middleware.RequireRoleOrAnyPermission(permissionRepo, employeeRepo, notificationRepo,
			[]string{"ADMIN", "OWNER", "FINANCE", "MONITOR"}, "finance", "monitoring"),
	))

	// ═══ نظام الطاقة الشمسية (منقول من Google Sheets) ═══
	// القراءة لأي موظف مسجّل: الفني بالموقع يحتاج يشوف مكوّنات المنظومة
	// ومواصفاتها وهو شغال. التعديل والتجهيز بصلاحية "solar_system".
	requireSolar := middleware.RequireRoleOrPermission(permissionRepo, employeeRepo, notificationRepo,
		[]string{"ADMIN", "OWNER"}, "solar_system")

	// ═══ برامج التدريب ═══
	// القراءة لأي موظف (يشوف تدريباته)، والإدارة لإداري الكوادر والتقني.
	requireTraining := middleware.RequireRoleOrAnyPermission(permissionRepo, employeeRepo, notificationRepo,
		[]string{"ADMIN", "OWNER", "HR_COORDINATOR"}, "staff_management", "content_technician")
	mux.Handle("GET /api/training-programs", middleware.Chain(http.HandlerFunc(trainingProgramHandler.List), requireAuth))
	mux.Handle("POST /api/training-programs", middleware.Chain(http.HandlerFunc(trainingProgramHandler.Create), requireAuth, requireTraining))
	mux.Handle("PUT /api/training-programs/{id}", middleware.Chain(http.HandlerFunc(trainingProgramHandler.Update), requireAuth, requireTraining))
	mux.Handle("PUT /api/training-programs/{id}/complete", middleware.Chain(http.HandlerFunc(trainingProgramHandler.Complete), requireAuth, requireTraining))
	mux.Handle("DELETE /api/training-programs/{id}", middleware.Chain(http.HandlerFunc(trainingProgramHandler.Delete), requireAuth, requireTraining))

	mux.Handle("GET /api/solar/stats", middleware.Chain(http.HandlerFunc(solarHandler.Stats), requireAuth))
	mux.Handle("GET /api/solar/low-stock", middleware.Chain(http.HandlerFunc(solarHandler.LowStock), requireAuth))

	mux.Handle("GET /api/solar/components", middleware.Chain(http.HandlerFunc(solarHandler.ListComponents), requireAuth))
	mux.Handle("POST /api/solar/components", middleware.Chain(http.HandlerFunc(solarHandler.CreateComponent), requireAuth, requireSolar))
	mux.Handle("PUT /api/solar/components/{id}", middleware.Chain(http.HandlerFunc(solarHandler.UpdateComponent), requireAuth, requireSolar))
	mux.Handle("DELETE /api/solar/components/{id}", middleware.Chain(http.HandlerFunc(solarHandler.DeleteComponent), requireAuth, requireSolar))

	mux.Handle("GET /api/solar/systems", middleware.Chain(http.HandlerFunc(solarHandler.ListSystems), requireAuth))
	mux.Handle("GET /api/solar/systems/{id}", middleware.Chain(http.HandlerFunc(solarHandler.GetSystem), requireAuth))
	mux.Handle("POST /api/solar/systems", middleware.Chain(http.HandlerFunc(solarHandler.CreateSystem), requireAuth, requireSolar))
	mux.Handle("PUT /api/solar/systems/{id}", middleware.Chain(http.HandlerFunc(solarHandler.UpdateSystem), requireAuth, requireSolar))
	mux.Handle("DELETE /api/solar/systems/{id}", middleware.Chain(http.HandlerFunc(solarHandler.DeleteSystem), requireAuth, requireSolar))

	// التجهيز يخصم من المخزن فعلياً — صلاحية إجبارية
	mux.Handle("POST /api/solar/systems/{id}/process", middleware.Chain(http.HandlerFunc(solarHandler.ProcessSystem), requireAuth, requireSolar))
	mux.Handle("GET /api/solar/installations", middleware.Chain(http.HandlerFunc(solarHandler.ListInstallations), requireAuth))
	mux.Handle("PUT /api/solar/installations/{id}/contacted", middleware.Chain(http.HandlerFunc(solarHandler.MarkContacted), requireAuth, requireSolar))

	// ── مراقبة النسخ الاحتياطية — للمالك وحده ──
	// requireOwner يرجّع 404 مو 403: أي حساب ثاني (حتى ADMIN) ما يعرف
	// إن هذا المسار موجود أصلاً. لا تضيف هنا أي middleware ثاني.
	mux.Handle("GET /api/owner/backups", middleware.Chain(http.HandlerFunc(backupHandler.Overview), requireAuth, middleware.RequireOwner()))

	mux.Handle("GET /api/funds", middleware.Chain(http.HandlerFunc(revolvingFundHandler.ListFunds), requireAuth, requireFund))
	mux.Handle("PUT /api/funds/{id}", middleware.Chain(http.HandlerFunc(revolvingFundHandler.UpdateFund), requireAuth, requireFundAmount))
	mux.Handle("POST /api/funds/{id}/topup", middleware.Chain(http.HandlerFunc(revolvingFundHandler.Topup), requireAuth, requireFundAmount))
	mux.Handle("POST /api/funds/disburse", middleware.Chain(http.HandlerFunc(revolvingFundHandler.Disburse), requireAuth, requireFund))
	mux.Handle("GET /api/funds/balances", middleware.Chain(http.HandlerFunc(revolvingFundHandler.Balances), requireAuth, requireFund))
	mux.Handle("GET /api/funds/transactions", middleware.Chain(http.HandlerFunc(revolvingFundHandler.Txns), requireAuth, requireFund))
	mux.Handle("PUT /api/funds/settlements/{id}/review", middleware.Chain(http.HandlerFunc(revolvingFundHandler.ReviewSettlement), requireAuth, requireFund))
	mux.Handle("PUT /api/funds/settlements/{id}/discharge", middleware.Chain(http.HandlerFunc(revolvingFundHandler.Discharge), requireAuth, requireDischarge))
	mux.Handle("GET /api/funds/discharge-accounts", middleware.Chain(http.HandlerFunc(revolvingFundHandler.DischargeAccounts), requireAuth, requireFund))

	mux.Handle("GET /api/funds/my-balance", middleware.Chain(http.HandlerFunc(revolvingFundHandler.MyBalance), requireAuth))
	mux.Handle("GET /api/funds/my-transactions", middleware.Chain(http.HandlerFunc(revolvingFundHandler.MyTxns), requireAuth))
	mux.Handle("POST /api/funds/settlements", middleware.Chain(http.HandlerFunc(revolvingFundHandler.SubmitSettlement), requireAuth))

	mux.Handle("GET /api/gps/sims", middleware.Chain(http.HandlerFunc(gpsHandler.ListSims), requireAuth, requireGpsData))
	mux.Handle("POST /api/gps/sims", middleware.Chain(http.HandlerFunc(gpsHandler.CreateSim), requireAuth, requireGpsSystem))
	mux.Handle("PUT /api/gps/sims/{id}", middleware.Chain(http.HandlerFunc(gpsHandler.UpdateSim), requireAuth, requireGpsSystem))

	// دورة حياة الشريحة: الربط والتحرير والحرق كلها شغل مسؤول الجي بي اس.
	mux.Handle("GET /api/gps/sims/available", middleware.Chain(http.HandlerFunc(gpsHandler.ListAvailableSims), requireAuth, requireGpsSystem))
	mux.Handle("POST /api/gps/sims/{id}/assign", middleware.Chain(http.HandlerFunc(gpsHandler.AssignSim), requireAuth, requireGpsSystem))
	mux.Handle("POST /api/gps/sims/{id}/release", middleware.Chain(http.HandlerFunc(gpsHandler.ReleaseSim), requireAuth, requireGpsSystem))
	mux.Handle("POST /api/gps/sims/{id}/burn", middleware.Chain(http.HandlerFunc(gpsHandler.BurnSim), requireAuth, requireGpsSystem))

	// متابعة التجديد: القائمة يشوفها الاثنين (مهندس الجودة يتصل، ومسؤول
	// الجي بي اس يشوف منو يحتاج حرق)، وتسجيل نتيجة الاتصال لمهندس الجودة.
	mux.Handle("GET /api/gps/subscriptions/follow-up", middleware.Chain(http.HandlerFunc(gpsHandler.SubscriptionFollowUps), requireAuth, requireGpsOrQuality))
	mux.Handle("GET /api/gps/devices/{id}/follow-up", middleware.Chain(http.HandlerFunc(gpsHandler.ListFollowUps), requireAuth, requireGpsOrQuality))
	mux.Handle("POST /api/gps/devices/{id}/follow-up", middleware.Chain(http.HandlerFunc(gpsHandler.CreateFollowUp), requireAuth, requireGpsOrQuality))

	// نفس الشي هنا: صف GpsDeviceRequest الجديد يبدأ status='PENDING' افتراضياً
	// بالمخطط نفسه (schema_base.go) — هذا "طلب" بانتظار مراجعة إداري GPS، مو جهاز
	// مفعّل فعلياً، فتقييد الإنشاء بصلاحية gps_system كان يمنع بالضبط سيناريو تقديم
	// الطلب من موظف مبيعات ما عنده هذي الصلاحية. الموافقة (PUT) تبقى محمية.
	mux.Handle("GET /api/gps/devices", middleware.Chain(http.HandlerFunc(gpsHandler.ListDevices), requireAuth, requireGpsData))
	mux.Handle("POST /api/gps/devices", middleware.Chain(http.HandlerFunc(gpsHandler.CreateDevice), requireAuth, requireGpsSystem))
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
	// قائمة مبسطة لاختيار سيارة لحجز — بلا صلاحية إدارة الأسطول
	mux.Handle("GET /api/vehicles/options", middleware.Chain(http.HandlerFunc(vehicleHandler.Options), requireAuth))
	mux.Handle("GET /api/vehicles", middleware.Chain(http.HandlerFunc(vehicleHandler.List), requireAuth, requireVehicleMgmt))
	mux.Handle("POST /api/vehicles", middleware.Chain(http.HandlerFunc(vehicleHandler.Create), requireAuth, requireVehicleMgmt))
	mux.Handle("GET /api/vehicles/{id}/logs", middleware.Chain(http.HandlerFunc(vehicleHandler.ListLogs), requireAuth, requireVehicleMgmt))
	mux.Handle("POST /api/vehicles/{id}/logs", middleware.Chain(http.HandlerFunc(vehicleHandler.CreateLog), requireAuth, requireVehicleMgmt))
	mux.Handle("PUT /api/vehicles/{id}/logs/{logId}", middleware.Chain(http.HandlerFunc(vehicleHandler.UpdateLog), requireAuth, requireVehicleMgmt))
	mux.Handle("DELETE /api/vehicles/{id}/logs/{logId}", middleware.Chain(http.HandlerFunc(vehicleHandler.DeleteLog), requireAuth, requireVehicleMgmt))
	// صورة الوصل بمسار مستقل — القوائم ترجع علم وجودها فقط حتى ما تنبلع
	// ميغابايتات base64 بكل جلب لسجلات السيارة.
	mux.Handle("GET /api/vehicles/{id}/logs/{logId}/receipt-photo", middleware.Chain(http.HandlerFunc(vehicleHandler.LogReceiptPhoto), requireAuth, requireVehicleMgmt))
	// كم مرة عبّأ كل موظف بالشهر
	mux.Handle("GET /api/vehicles/fuel-stats/by-employee", middleware.Chain(http.HandlerFunc(vehicleHandler.EmployeeFuelStats), requireAuth, requireVehicleMgmt))
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
	mux.Handle("POST /api/vehicle-missions", middleware.Chain(http.HandlerFunc(vehicleMissionHandler.Start), requireAuth, requireVehicleMgmt))
	mux.Handle("PUT /api/vehicle-missions/{id}/end", middleware.Chain(http.HandlerFunc(vehicleMissionHandler.End), requireAuth, requireVehicleMgmt))
	mux.Handle("GET /api/vehicle-missions", middleware.Chain(http.HandlerFunc(vehicleMissionHandler.List), requireAuth, requireVehicleMgmt))
	mux.Handle("GET /api/vehicle-missions/{id}", middleware.Chain(http.HandlerFunc(vehicleMissionHandler.Get), requireAuth))
	mux.Handle("POST /api/vehicle-missions/{id}/rating", middleware.Chain(http.HandlerFunc(vehicleMissionHandler.CreateRating), requireAuth, requireVehicleMgmt))
	// فحص أدوات المركبة العامة عند بدء مهمة — حصراً لليدر (نفس requireLeader
	// المستخدم بصيانة الأجهزة)، الموظف العادي ما يشوف/يستخدم هالراوت إطلاقاً.
	mux.Handle("POST /api/vehicle-missions/{id}/tool-check", middleware.Chain(http.HandlerFunc(vehicleMissionHandler.CreateToolCheck), requireAuth, requireLeader))
	mux.Handle("GET /api/employees/{id}/driver-rating-summary", middleware.Chain(http.HandlerFunc(vehicleMissionHandler.DriverRatingSummary), requireAuth))

	// نظام حجز المركبات (مسبق) — منفصل عن بدء المهمة الفعلي أعلاه.
	mux.Handle("POST /api/vehicle-bookings", middleware.Chain(http.HandlerFunc(vehicleBookingHandler.Create), requireAuth, requireVehicleMgmt))
	mux.Handle("PUT /api/vehicle-bookings/{id}/decide", middleware.Chain(http.HandlerFunc(vehicleBookingHandler.Decide), requireAuth, requireVehicleMgmt))
	mux.Handle("PUT /api/vehicle-bookings/{id}/cancel", middleware.Chain(http.HandlerFunc(vehicleBookingHandler.Cancel), requireAuth, requireVehicleMgmt))
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
	// فواتير الليدر: يشوفها الليدر نفسه، وصاحب صلاحية سلة الليدر، *والمحاسب* —
	// لأنه الفاتورة لازم ترحّل له بتفاصيلها حتى يدققها ويعتمدها.
	requireLeaderBasket := middleware.RequireLeaderOrAnyPermission(permissionRepo, employeeRepo, notificationRepo, "leader_basket", "finance")
	// أسباب الشغل المجاني — يقراها أي موظف يسوي فاتورة
	mux.Handle("GET /api/free-work-reasons", middleware.Chain(http.HandlerFunc(leaderInvoiceHandler.FreeReasons), requireAuth))
	mux.Handle("GET /api/system-price-catalog", middleware.Chain(http.HandlerFunc(leaderInvoiceHandler.ListCatalog), requireAuth))
	mux.Handle("GET /api/materials", middleware.Chain(http.HandlerFunc(leaderInvoiceHandler.ListMaterials), requireAuth))
	mux.Handle("GET /api/leader-invoices", middleware.Chain(http.HandlerFunc(leaderInvoiceHandler.List), requireAuth, requireLeaderBasket))
	mux.Handle("GET /api/leader-invoices/{id}", middleware.Chain(http.HandlerFunc(leaderInvoiceHandler.Get), requireAuth, requireLeaderBasket))
	mux.Handle("POST /api/leader-invoices", middleware.Chain(http.HandlerFunc(leaderInvoiceHandler.Create), requireAuth, requireLeader))
	// حساب تقريبي بدون حفظ لما زبون يستفسر — نفس صلاحية إنشاء الفاتورة (الليدر)
	// حساب تكلفة التنصيب للتنفيذ: فقرة رئيسية بكل الحسابات وكل الأدوار،
	// فما بيها قيد غير تسجيل الدخول — هي حاسبة ما تكشف بيانات أحد
	mux.Handle("POST /api/leader-invoices/estimate", middleware.Chain(http.HandlerFunc(leaderInvoiceHandler.Estimate), requireAuth, requireExecutionCost))
	mux.Handle("POST /api/leader-invoices/camera-cost", middleware.Chain(http.HandlerFunc(leaderInvoiceHandler.CameraCost), requireAuth, requireExecutionCost))
	// ── صندوق المراقب ──
	// شغل المراقب: يشوف ويأشّر. requireMonitor = ADMIN أو MONITOR،
	// والمالك داخل بالأدمن.
	mux.Handle("GET /api/monitor-reviews", middleware.Chain(http.HandlerFunc(monitorReviewHandler.List), requireAuth, requireMonitor))
	mux.Handle("GET /api/monitor-reviews/counts", middleware.Chain(http.HandlerFunc(monitorReviewHandler.Counts), requireAuth, requireMonitor))
	mux.Handle("POST /api/monitor-reviews/{id}/decide", middleware.Chain(http.HandlerFunc(monitorReviewHandler.Decide), requireAuth, requireMonitor))

	// ── تكلفة الشبكات ──
	// الاستمارة والحساب: نفس قيد حاسبة الكاميرات (صلاحية حساب التنفيذ).
	// أما تعديل الأسعار فمحصور بالمالك ومدير النظام — سعر يتغيّر يعني
	// كل فاتورة جاية تتغير معاه.
	mux.Handle("GET /api/network-cost/items", middleware.Chain(http.HandlerFunc(networkCostHandler.ListActive), requireAuth, requireExecutionCost))
	mux.Handle("POST /api/network-cost/calculate", middleware.Chain(http.HandlerFunc(networkCostHandler.Calculate), requireAuth, requireExecutionCost))
	mux.Handle("GET /api/network-cost/prices", middleware.Chain(http.HandlerFunc(networkCostHandler.ListAll), requireAuth, requireAdmin))
	mux.Handle("POST /api/network-cost/prices", middleware.Chain(http.HandlerFunc(networkCostHandler.CreatePrice), requireAuth, requireAdmin))
	mux.Handle("PUT /api/network-cost/prices/{id}", middleware.Chain(http.HandlerFunc(networkCostHandler.UpdatePrice), requireAuth, requireAdmin))
	mux.Handle("DELETE /api/network-cost/prices/{id}", middleware.Chain(http.HandlerFunc(networkCostHandler.DeactivatePrice), requireAuth, requireAdmin))
	mux.Handle("GET /api/leader-invoices/camera-cost/options", middleware.Chain(http.HandlerFunc(leaderInvoiceHandler.CameraCostOptions), requireAuth, requireExecutionCost))
	// الاعتماد محصور بمدير/محاسب فقط — الليدر ما يقدر يعتمد فاتورته بنفسه
	// البحث بالفاتورة المحاسبية — لازم يجي قبل مسار {id} حتى ما ينحسب
	// "by-number" معرّف فاتورة
	mux.Handle("GET /api/leader-invoices/by-number", middleware.Chain(http.HandlerFunc(leaderInvoiceHandler.FindByExternalNumber), requireAuth, requireLeaderBasket))
	mux.Handle("PUT /api/leader-invoices/{id}/approve", middleware.Chain(http.HandlerFunc(leaderInvoiceHandler.Approve), requireAuth, requireFinance))
	// ربط رقم فاتورة محاسبية بفاتورة معتمدة أصلاً، وتعديل المبالغ —
	// الاثنين للمحاسب/المدير حصراً
	mux.Handle("PUT /api/leader-invoices/{id}/external-number", middleware.Chain(http.HandlerFunc(leaderInvoiceHandler.SetExternalNumber), requireAuth, requireFinance))
	mux.Handle("PUT /api/leader-invoices/{id}/adjust", middleware.Chain(http.HandlerFunc(leaderInvoiceHandler.Adjust), requireAuth, requireFinance))
	mux.Handle("GET /api/leader-invoices/{id}/adjustments", middleware.Chain(http.HandlerFunc(leaderInvoiceHandler.Adjustments), requireAuth))

	// إحصائيات الموظفين الشهرية — حصراً للمالك/الأدمن (requireAdmin يسمح OWNER
	// تلقائياً لأنه يتخطى أي قيد أدوار بـRequireRole).
	mux.Handle("GET /api/employee-stats/monthly", middleware.Chain(http.HandlerFunc(employeeStatsHandler.Monthly), requireAuth, requireAdmin))
	mux.Handle("GET /api/employee-stats/monthly/export", middleware.Chain(http.HandlerFunc(employeeStatsHandler.MonthlyExport), requireAuth, requireAdmin))
	mux.Handle("GET /api/employee-stats/range", middleware.Chain(http.HandlerFunc(employeeStatsHandler.Range), requireAuth, requireAdmin))
	mux.Handle("GET /api/employee-stats/curve/{employeeId}", middleware.Chain(http.HandlerFunc(employeeStatsHandler.Curve), requireAuth, requireAdmin))

	// إدارة الإحصائيات: يومية/أسبوعية/مشاريع — حصراً لمدير النظام.
	mux.Handle("GET /api/stats-management/daily", middleware.Chain(http.HandlerFunc(statsManagementHandler.Daily), requireAuth, requireAdmin))
	mux.Handle("GET /api/stats-management/weekly", middleware.Chain(http.HandlerFunc(statsManagementHandler.Weekly), requireAuth, requireAdmin))
	mux.Handle("GET /api/stats-management/projects", middleware.Chain(http.HandlerFunc(statsManagementHandler.ProjectStages), requireAuth, requireAdmin))
	mux.Handle("GET /api/stats-management/internal-works", middleware.Chain(http.HandlerFunc(statsManagementHandler.InternalWorks), requireAuth, requireAdmin))

	// تقدير مدة العمل المتعلَّم (learned baseline) — قراءة فقط، متاح لأي مستخدم
	// مسجّل دخول (يحتاجها المنسق قبل تثبيت موعد/فريق).
	mux.Handle("GET /api/job-duration-estimate", middleware.Chain(http.HandlerFunc(jobDurationHandler.Estimate), requireAuth))

	return middleware.Chain(mux, middleware.Recovery, middleware.SecurityHeaders, middleware.Logging, middleware.Metrics, middleware.CORS(cfg.CORSOrigins), middleware.BodyLimit(middleware.MaxBodyBytes))
}

// buildFileStore يختار باكند التخزين. R2 لو إعداداته كاملة وينجح
// إنشاؤه، وإلا القرص المحلي — النظام ما يوقف لأن التخزين البعيد
// مو مضبوط.
func buildFileStore(cfg *config.Config) storage.Store {
	r2cfg := storage.R2Config{
		Bucket:    cfg.R2Bucket,
		AccessKey: cfg.R2AccessKey,
		SecretKey: cfg.R2SecretKey,
		Endpoint:  cfg.R2Endpoint,
	}
	if r2cfg.Configured() {
		if s, err := storage.NewR2Store(r2cfg); err == nil {
			return s
		} else {
			log.Printf("[storage] تعذر تهيئة R2 (%v) — نرجع للقرص المحلي", err)
		}
	}
	s, err := storage.NewLocalStore(cfg.UploadsDir)
	if err != nil {
		// ⚠️ ما نطيح السيرفر. فشل تهيئة مجلد الملفات كان log.Fatalf —
		// يعني حاوية بمجلد عمل مو قابل للكتابة توقّف النظام كله وهو
		// شغّال تمام. الرفع والعرض بس ينعطلون، والباقي يكمل.
		log.Printf("[storage] ⚠️ تعذر تهيئة مجلد التخزين %q: %v — الرفع والعرض معطّلين، وباقي النظام شغّال",
			cfg.UploadsDir, err)
		return storage.NewUnavailableStore(err.Error())
	}
	return s
}

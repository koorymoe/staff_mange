package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"staffmange-api/internal/apps/access"
	"staffmange-api/internal/apps/field"
	"staffmange-api/internal/apps/finance"
	"staffmange-api/internal/apps/invoices"
	"staffmange-api/internal/apps/me"
	"staffmange-api/internal/apps/org"
	"staffmange-api/internal/apps/people"
	"staffmange-api/internal/apps/requests"
	"staffmange-api/internal/apps/sections"
	"staffmange-api/internal/apps/system"
	"staffmange-api/internal/authz"
	"staffmange-api/internal/core"
	"staffmange-api/internal/monitor"
	"strconv"

	"staffmange-api/internal/logredact"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"

	"staffmange-api/internal/config"
	"staffmange-api/internal/database"
	"staffmange-api/internal/handler"
	"staffmange-api/internal/integrity"
	"staffmange-api/internal/mail"
	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/safeguard"
	"staffmange-api/internal/service"
	"staffmange-api/internal/storage"
	"staffmange-api/internal/telegram"
	"staffmange-api/internal/whatsapp"
)

func main() {
	// (11-04 — ٦ز) كل سطر بالسجل يمر بفلتر يحجب المفاتيح والتوكنات والباسوردات
	log.SetOutput(logredact.Writer{W: os.Stderr})
	startedAt := time.Now()
	// تحميل ملف .env إذا موجود (بالإنتاج المتغيرات تنجي مباشرة من النظام، فما مشكلة لو الملف مو موجود)
	_ = godotenv.Load()

	cfg := config.Load()
	// Sentry (دازاي 11-05): يلزم أخطاء السيرفر — بلا SENTRY_DSN_BACKEND مطفي. النصوص تنتنظف قبل الإرسال.
	monitor.Init(cfg.SentryDSN, cfg.SentryEnvironment, "")
	defer monitor.Flush()
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
	// 🔎 أدوار قراءة الذكاء — ايريس وماتركس يقرون النظام كله بلا الأسرار (دازاي 10-23)
	database.EnsureAIReaders(db)

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
	servicePriceLearningRepo := repository.NewServicePriceLearningRepository(db)
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
	surveyReportRepo := repository.NewBookingSurveyReportRepository(db)
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
	// (DRY 11-02) الحراس بالمستودعات الثلاثة مرة وحدة — نفس Require* بالضبط
	guards := middleware.NewGuards(permissionRepo, employeeRepo, notificationRepo)

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
	// (ع) 10-11: نظام نقاط الانضباط القديم (١٠٠ نقطة) انشال — ما تشتغل كنساته.
	// البيانات القديمة تبقى بالقاعدة وبسجل الانضباط للتاريخ بس.
	qualityFollowUpService := service.NewQualityFollowUpService(qualityFollowUpRepo, notificationRepo)
	cartService := service.NewCartService(cartRepo)
	expenseService := service.NewExpenseService(expenseRepo)
	inventoryService := service.NewInventoryService(inventoryRepo)
	// موافقة على طلب أداة غير متوفرة تولّد طلب مشتريات يوصل للمحاسب.
	inventoryService.SetProcurementRepository(procurementRepo)
	// قرار طلب الأداة يوصل صاحبه — بلا هذا الطلب يروح بلا رجعة
	inventoryService.SetNotificationRepository(notificationRepo)
	attendanceService := service.NewAttendanceService(attendanceRepo)
	notificationService := service.NewNotificationService(notificationRepo)
	kpiService := service.NewKpiService(kpiRepo, employeeRepo, notificationRepo, announcementRepo)

	// محرّك القصص — الكيان ينفّذ مشهداً بدل تنبيه ينقرا وينتسى.
	// ⚠️ يُركَّب بـsetter: القصة **إضافة على** الإجراء الإداري مو شرط
	// له، فالخصم يشتغل كامل حتى لو انطفى المحرّك.
	storyRepo := repository.NewStoryRepository(db)
	storyService := service.NewStoryService(storyRepo, employeeRepo)
	storyHandler := handler.NewStoryHandler(storyService)
	kpiService.SetStories(storyService)
	// ⚠️ كل ربط بـsetter: القصة **إضافة** على الإجراء مو شرط له —
	// وأي خدمة منهن تشتغل كاملة لو انشال المحرّك.
	disciplineService.SetStories(storyService)
	bookingService.SetStories(storyService)
	kpiCriterionService := service.NewKpiCriterionService(kpiCriterionRepo)
	smartKpiService := service.NewSmartKpiService(smartKpiRepo)
	complaintService := service.NewComplaintService(complaintRepo)
	complaintService.SetNewNotifier(func(assignee *string, msg string) {
		if assignee != nil {
			_ = notificationRepo.Create(*assignee, "complaint_new", msg)
		}
		_ = notificationRepo.CreateForRolesOrPermission([]string{"ADMIN", "OWNER", "MONITOR"}, "complaints_watch", "complaint_new", msg)
	})
	complaintService.SetVipNotifier(func(msg string) {
		_ = notificationRepo.CreateForRolesOrPermission([]string{"ADMIN", "OWNER"}, "complaints_vip", "complaint_vip", msg)
	})
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
	// ⚠️ حارس واحد للتقرير والفاتورة — بخدمات مؤشّرة الورق على مسؤول
	// الخدمة مو على الفني. فحصان منفصلان يفترقان بأول تعديل.
	paperworkGuard := service.NewPaperworkGuard(db)
	workReportService := service.NewWorkReportService(workReportRepo, paperworkGuard)
	surveyReportService := service.NewBookingSurveyReportService(surveyReportRepo)
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
	authHandler.SetPerms(permissionRepo) // (11-04) جدول القدرات can
	employeeHandler := handler.NewEmployeeHandler(employeeService)
	employeeHandler.SetNotifications(notificationRepo)
	leaderSkillHandler := handler.NewLeaderSkillHandler(repository.NewLeaderSkillRepository(db))
	departmentRepo := repository.NewDepartmentRepository(db)
	departmentHandler := handler.NewDepartmentHandler(departmentRepo)
	permissionHandler := handler.NewPermissionHandler(permissionService, lockoutRepo)
	serviceHandler := handler.NewServiceHandler(serviceCatalogService)
	customerHandler := handler.NewCustomerHandler(customerService)
	locateHandler := handler.NewLocateHandler(db)
	bookingHandler := handler.NewBookingHandler(bookingService, permissionRepo)
	bookingHandler.SetReminderService(bookingReminderService)
	qualityFollowUpHandler := handler.NewQualityFollowUpHandler(qualityFollowUpService, permissionRepo)
	qualityFollowUpHandler.SetNotifications(notificationRepo)
	securityHandler := handler.NewSecurityHandler(db, loginAuditRepo, lockoutRepo, startedAt)

	// تخزين الملفات: R2 إذا انضبطت إعداداته، وإلا القرص المحلي. الاثنين
	// نفس الواجهة فباقي النظام ما يعرف أيهم شغّال.
	fileStore := buildFileStore(cfg)
	// (٧ل، 11-05) الملفات القديمة (base64 بالقاعدة: صور المنتجات، الوصولات، العقود…)
	// تنتقل لمخزن الملفات بالخلفية — بأمان، كل ملف يتأكد قبل ما يتبدّل
	go func() {
		for col, n := range repository.MigrateInlineFiles(db, fileStore) {
			log.Printf("[storage] %s: انتقل %d ملف من القاعدة لمخزن الملفات", col, n)
		}
	}()
	log.Printf("[storage] تخزين الملفات: %s", fileStore.Kind())
	{
		onOff := func(b bool) string {
			if b {
				return "مفعّل"
			}
			return "مطفي"
		}
		log.Printf("🤖 ماتركس: %s · السكرتير: %s · الجودة: %s · الحجوزات: %s (%s) · Gemini: %s · Whisper: %s",
			onOff(cfg.AIKeyMatrix != ""), onOff(cfg.AIKeySecretary != ""), onOff(cfg.AIKeyQuality != ""), onOff(cfg.AIKeyBooking != ""),
			cfg.AIModel, onOff(cfg.GeminiAPIKey != ""), onOff(cfg.WhisperURL != ""))
	}
	fileHandler := handler.NewFileHandler(fileStore, []byte(cfg.JWTSecret), employeeRepo)
	surveyPhotoHandler := handler.NewSurveyPhotoHandler(repository.NewSurveyPhotoRepository(db), permissionRepo, fileStore)
	// مفاتيح إطفاء الميزات — جدول عام صغير، القراءة للكل والكتابة للمالك.
	systemSwitchHandler := handler.NewSystemSwitchHandler(repository.NewSystemSwitchRepository(db))
	cartHandler := handler.NewCartHandler(cartService)
	expenseHandler := handler.NewExpenseHandler(expenseService, permissionRepo)
	inventoryHandler := handler.NewInventoryHandler(inventoryService)
	// أبو الكميات ومتابع الجرد يشوفون عدة غيرهم بصلاحيتهم.
	inventoryHandler.SetPermissions(permissionRepo)
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
		if model.IsAdminRole(role) || role == "PROJECT_MANAGER" || role == "MONITOR" {
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
	servicePriceHandler := handler.NewServicePriceLearningHandler(servicePriceLearningRepo)
	solarHandler := handler.NewSolarHandler(solarRepo)
	backupHandler := handler.NewBackupHandler(backupRunRepo)
	// سعر المنظومة لحجز الطاقة الشمسية — ينحسب من الكتالوك مو ينكتب بالإيد
	bookingService.SetSolarPricer(solarRepo)
	// اسم طالب الحجز الداخلي يجي من سجل الأقسام، ما ينكتب بالإيد.
	bookingService.SetDepartments(departmentRepo)
	// تكليف الكادر يولّد مهمة ميدانية. بلا هذا السطر شاشة «تتبع المهام»
	// تضل فارغة للأبد مهما اشتغلت الشركة — وهذا الي كان صاير.
	bookingService.SetMissionStarter(missionService)
	// تعويض الحجوزات الشغّالة الي انكلّفت قبل هذا الربط — وإلا الشاشة
	// تضل فاضية لحد ما ينكلّف حجز جديد. بحماية: فشلها ما يمنع الإقلاع.
	safeguard.Go("تعويض مهام الحجوزات القديمة", missionService.BackfillOnce)
	trainingProgramHandler := handler.NewTrainingProgramHandler(trainingProgramRepo)
	productHandler := handler.NewProductHandler(productService)
	leaderInvoiceService := service.NewLeaderInvoiceService(leaderInvoiceRepo, systemPriceCatalogRepo, materialRepo, employeeCommissionRepo, bookingRepo, employeeRepo, jobDurationEstimatorService)
	leaderInvoiceHandler := handler.NewLeaderInvoiceHandler(leaderInvoiceService, systemPriceCatalogRepo, materialRepo)
	// فحص صلاحية نوع فاتورة الخدمة (جي بي اس/داش كام) — النوع يجي
	// بالطلب، فالفحص بالمعالج مو بالمسار.
	leaderInvoiceHandler.SetPermissions(permissionRepo)
	networkPriceRepo := repository.NewNetworkPriceRepository(db)
	// ── صندوق المراقب ──
	// الربط بـSetMonitorFeed بعد البناء (مو بالمنشئ) حتى ما يصير اعتماد
	// دائري بين خدمة الحجوزات وخدمة المراقبة.
	monitorReviewRepo := repository.NewMonitorReviewRepository(db)
	monitorReviewService := service.NewMonitorReviewService(monitorReviewRepo, employeeRepo, notificationRepo)
	monitorReviewHandler := handler.NewMonitorReviewHandler(monitorReviewService)
	monitorDeskRepo := repository.NewMonitorDeskRepository(db)
	monitorDeskHandler := handler.NewMonitorDeskHandler(monitorDeskRepo, monitorReviewRepo)
	bookingService.SetMonitorFeed(monitorReviewService)
	// ═══ نواة الذكاء الاصطناعي ═══
	// نفس أسلوب صندوق المراقب: الربط بعد البناء حتى ما يصير اعتماد
	// متبادل بين خدمة الحجوزات ونواة التحليل.
	// الخط الزمني للحجز — يجمع قصته من سبعة جداول ويحسب التأخيرات
	bookingTimelineService := service.NewBookingTimelineService(
		bookingRepo, bookingProgressRepo, leaderInvoiceRepo, qualityFollowUpRepo, monitorReviewRepo)
	bookingHandler.SetTimelineService(bookingTimelineService)

	// المهام الإضافية — المدير يوجّه شغل لموظف، مو مربوط بحجز
	extraTaskRepo := repository.NewExtraTaskRepository(db)
	extraTaskHandler := handler.NewExtraTaskHandler(extraTaskRepo, notificationRepo)
	extraTaskHandler.SetStories(storyService)

	aiRepo := repository.NewAiRepository(db)
	// achievementRepo يُبنى هنا (بدل مكانه الطبيعي تحت مع بقية شغل
	// الإنجازات) لأن جامع أدلة SELF_REPORT_MISMATCH يحتاجه.
	achievementRepo := repository.NewAchievementRepository(db)
	aiEvidenceService := service.NewAiEvidenceService(aiRepo, bookingRepo, leaderInvoiceRepo, bookingProgressRepo, achievementRepo, employeeRepo)
	aiBrainService := service.NewAiBrainService(aiRepo, aiEvidenceService)
	aiMetricsService := service.NewAiMetricsService(aiRepo)
	// ═══ توصيل النموذج ═══
	// ⚠️ بلا مفتاح يبقى على محرّك القواعد — الميزة تنطفي مو النظام.
	// وحتى بوجود المفتاح، `ModelJudge` يرجع للقواعد عند أي عثرة
	// (شبكة، حد يومي، رد مشوّه). القواعد شبكة أمان دائمة.
	if cfg.AIKeyMatrix != "" {
		modelJudge := service.NewModelJudge(cfg.AIKeyMatrix, cfg.AIModel, cfg.AIDailyCap, service.RulesJudge{})
		// حلقة التعلّم: أمثلة حقيقية من قرارات المراقب — بلا جدول
		// جديد، لأن القرار مخزون أصلاً بصندوق المراقب.
		modelJudge.SetFeedbackSource(aiRepo)
		aiBrainService.SetJudge(modelJudge)
		log.Printf("[ai] حاكم التحليل: %s (سقف يومي %d)", cfg.AIModel, cfg.AIDailyCap)
	} else {
		log.Printf("[ai] ماكو ANTHROPIC_API_KEY — التحليل بمحرّك القواعد")
	}
	// ⚠️ الحلقة تشتغل بأي حال: القواعد وحدها تحلل توقف العمل، والإشارات
	// چانت تتكدس بلا حكم لأن ماكو أحد يشغّل التحليل.
	aiBrainService.StartBackgroundLoop()
	aiHandler := handler.NewAiHandler(aiRepo, aiBrainService, aiMetricsService)
	aiBrainService.SetMonitorFeed(monitorReviewService)
	bookingService.SetAiRecorder(aiRepo)
	// وصل كاشف شذوذ الوقود الموجود من زمان (VehicleService.CheckFuelAnomaly) بماتركس.
	vehicleService.SetAiRecorder(aiRepo)
	leaderInvoiceService.SetAiRecorder(aiRepo)
	complaintService.SetAiRecorder(aiRepo)

	// ═══ تدقيق التكرار (ماتركس) — حجوزات وزبائن مكررون بالغلط ═══
	// مسار مستقل عن AiSignal/صندوق المراقب عمداً (زوج التكرار علاقة
	// بين صفّين، وذاك التصميم يفترض هوية واحدة لكل معرّف).
	duplicateCandidateRepo := repository.NewDuplicateCandidateRepository(db)
	duplicateCandidateService := service.NewDuplicateCandidateService(duplicateCandidateRepo)
	duplicateCandidateHandler := handler.NewDuplicateCandidateHandler(duplicateCandidateService)
	baghdadLoc, _ := time.LoadLocation("Asia/Baghdad")
	if baghdadLoc == nil {
		baghdadLoc = time.FixedZone("Baghdad", 3*3600)
	}
	// إذا السيرفر اشتغل بعد ٩ الصبح، تقرير اليوم يعتبر انرسل (حتى إعادة التشغيل ما تكرره).
	duplicateDailyDay := ""
	if time.Now().In(baghdadLoc).Hour() >= 9 {
		duplicateDailyDay = time.Now().In(baghdadLoc).Format("2006-01-02")
	}
	safeguard.Loop("كنسة تدقيق التكرار", 5*time.Minute, time.Hour, func() {
		bookings, customers, err := duplicateCandidateService.RunScan()
		if err != nil {
			log.Printf("duplicate candidate scan: %v", err)
			return
		}
		if bookings > 0 || customers > 0 {
			log.Printf("[duplicate] حجوزات جديدة: %d، زبائن جدد: %d", bookings, customers)
			// ماتركس يبلّغ: المالك والمدير وصاحب صلاحية تدقيق التكرار.
			_ = notificationRepo.CreateForRolesOrPermission([]string{"OWNER", "ADMIN"}, "duplicate_review", "DUPLICATE",
				fmt.Sprintf("🤖 ماتركس — لگيت تكرار جديد: %d زوج حجوزات و%d زوج زبائن. شوفهن بـ«ماتركس — التكرار» واتخذ إجراء.", bookings, customers))
		}
		// تقرير ماتركس اليومي (٩ الصبح بغداد): كم تكرار بعده معلّق — حتى ما ينسى.
		now := time.Now().In(baghdadLoc)
		day := now.Format("2006-01-02")
		if now.Hour() >= 9 && duplicateDailyDay != day {
			duplicateDailyDay = day
			if rep, err := duplicateCandidateRepo.Report(); err == nil && rep.PendingBookings+rep.PendingCustomers > 0 {
				_ = notificationRepo.CreateForRolesOrPermission([]string{"OWNER", "ADMIN"}, "duplicate_review", "DUPLICATE",
					fmt.Sprintf("🤖 ماتركس — فحص التكرار اليومي: %d حجز مكرر و%d زبون مكرر بعدهم ينتظرون إجراء.", rep.PendingBookings, rep.PendingCustomers))
			}
		}
	})

	// ═══ الفضفضة اليومية (ماتركس) ═══ — ملخص أحكام اليوم للمالك ومدير
	// النظام، مرة وحدة بعد ٨ مساءً بغداد (العلامة تمنع إرسالها مرتين).
	dailyDebriefService := service.NewDailyDebriefService(aiRepo, notificationRepo)
	safeguard.Loop("الفضفضة اليومية", 10*time.Minute, 30*time.Minute, func() {
		if err := dailyDebriefService.RunIfDue(); err != nil {
			log.Printf("daily debrief: %v", err)
		}
	})
	// ═══ الاستكشاف الأسبوعي (ماتركس) ═══ — شبكة ٥ مقاييس × ٤ محاور
	// تُحسب حتمياً ضد كل البيانات كل اثنين، وتنبّه المالك ومدير النظام
	// حصراً بالأنماط الي عبرت بوابتي العيّنة والانحراف.
	aiDiscoveryRepo := repository.NewAiDiscoveryRepository(db)
	aiDiscoveryService := service.NewAiDiscoveryService(aiDiscoveryRepo, aiRepo, employeeRepo, notificationRepo)
	if cfg.AIKeyMatrix != "" {
		aiDiscoveryService.SetNarrator(service.NewModelDiscoveryNarrator(cfg.AIKeyMatrix, cfg.AIModel, cfg.AIDailyCap))
	}
	safeguard.Loop("الاستكشاف الأسبوعي", 15*time.Minute, 6*time.Hour, func() {
		if err := aiDiscoveryService.RunWeeklyIfDue(); err != nil {
			log.Printf("weekly discovery: %v", err)
		}
	})
	// ═══ ماتركس يمدح مو بس يعاتب ═══ — إشعار أسبوعي شخصي لكل موظف
	// (له هو، مو للمالك) بتحسّنه — أول مسار إيجابي بكل إشارات الذكاء.
	weeklyPraiseService := service.NewWeeklyPraiseService(aiRepo, employeeRepo, notificationRepo)
	safeguard.Loop("مديح ماتركس الأسبوعي", 20*time.Minute, 6*time.Hour, func() {
		if err := weeklyPraiseService.RunWeeklyIfDue(); err != nil {
			log.Printf("weekly praise: %v", err)
		}
	})
	// ═══ حماية الموظف من الإرهاق (ماتركس) ═══ — دوام متواصل بلا راحة
	// + ارتفاع إشاراته بنفس الفترة → تنبيه حمائي للمالك ومدير النظام،
	// مو ملاحظة أداء.
	burnoutProtectionService := service.NewBurnoutProtectionService(aiRepo, employeeRepo, attendanceRepo, notificationRepo)
	safeguard.Loop("حماية الإرهاق الأسبوعية", 22*time.Minute, 6*time.Hour, func() {
		if err := burnoutProtectionService.RunWeeklyIfDue(); err != nil {
			log.Printf("burnout protection: %v", err)
		}
	})
	// ═══ ربط الأخطاء بفجوة تدريب حقيقية (ماتركس) ═══ — تكرار توقف
	// العمل بخدمة معيّنة عند موظف ما تدرّب عليها → اقتراح تدريب.
	trainingGapService := service.NewTrainingGapService(aiRepo, notificationRepo)
	safeguard.Loop("فجوة التدريب الأسبوعية", 24*time.Minute, 6*time.Hour, func() {
		if err := trainingGapService.RunWeeklyIfDue(); err != nil {
			log.Printf("training gap: %v", err)
		}
	})
	// ═══ صيانة المركبات مقابل الاستخدام الفعلي (ماتركس) ═══ — مركبة
	// صيانتها متأخرة وبعدها تُرسل بمهام/حجوزات.
	vehicleMaintenanceAlertService := service.NewVehicleMaintenanceAlertService(aiRepo, notificationRepo)
	safeguard.Loop("تنبيه صيانة المركبات الأسبوعي", 26*time.Minute, 6*time.Hour, func() {
		if err := vehicleMaintenanceAlertService.RunWeeklyIfDue(); err != nil {
			log.Printf("vehicle maintenance alert: %v", err)
		}
	})
	// 📦 ملخص المخازن اليومي (طلبات ماتركس 11-04 — إداري الكميات): طلبات معلّقة، تقييم السيارات، صيانة فايتة
	warehouseDaily := service.NewWarehouseDaily(db, aiRepo.ClaimDailyMarker, func() ([]string, error) {
		rows, err := aiRepo.OverdueMaintenanceVehiclesInUse()
		out := make([]string, 0, len(rows))
		for _, r := range rows {
			out = append(out, r.VehicleName+" ("+r.PlateNumber+")")
		}
		return out, err
	}, func(msg string) { _ = notificationRepo.CreateForPermission("inventory", "WAREHOUSE_DAILY", msg) })
	safeguard.Loop("ملخص المخازن اليومي", 27*time.Minute, 30*time.Minute, func() {
		if err := warehouseDaily.RunIfDue(time.Now()); err != nil {
			log.Printf("warehouse daily: %v", err)
		}
	})
	// ═══ صحة الزبون (ماتركس) ═══ — زبائن مخلصين انقطعوا.
	lapsedCustomerAlertService := service.NewLapsedCustomerAlertService(aiRepo, notificationRepo)
	safeguard.Loop("تنبيه الزبائن المنقطعين الأسبوعي", 28*time.Minute, 6*time.Hour, func() {
		if err := lapsedCustomerAlertService.RunWeeklyIfDue(); err != nil {
			log.Printf("lapsed customer alert: %v", err)
		}
	})
	// ═══ ماتركس الاستباقي ═══ — التنبؤ بالتأخير (صباحاً) وفرص البيع (أسبوعياً).
	delayPredictionService := service.NewDelayPredictionService(aiRepo, notificationRepo)
	safeguard.Loop("تنبيه توقع التأخير الصباحي", 17*time.Minute, time.Hour, func() {
		if err := delayPredictionService.RunMorningIfDue(); err != nil {
			log.Printf("delay prediction alert: %v", err)
		}
	})
	salesOpportunityService := service.NewSalesOpportunityService(aiRepo, notificationRepo)
	safeguard.Loop("ملخص فرص البيع الأسبوعي", 33*time.Minute, 6*time.Hour, func() {
		if err := salesOpportunityService.RunWeeklyIfDue(); err != nil {
			log.Printf("sales opportunities alert: %v", err)
		}
	})
	// ماتركس: فحوصات يومية (زبون قرب يزعل، الورق المتأخر، الحضور مقابل الشغل).
	matrixScanService := service.NewMatrixScanService(aiRepo)
	safeguard.Loop("فحوصات ماتركس اليومية", 41*time.Minute, time.Hour, func() {
		if err := matrixScanService.RunDailyIfDue(); err != nil {
			log.Printf("matrix daily scans: %v", err)
		}
	})
	matrixForesightHandler := handler.NewMatrixForesightHandler(delayPredictionService, salesOpportunityService)
	// ماتركس: الكادر الأنسب، منحنى الجدد، المواد، الاستبدال، والتقرير الأسبوعي (أحد الصبح).
	matrixInsightsService := service.NewMatrixInsightsService(aiRepo)
	weeklyReportService := service.NewWeeklyReportService(aiRepo, notificationRepo)
	safeguard.Loop("تقرير المالك الأسبوعي", 37*time.Minute, time.Hour, func() {
		if err := weeklyReportService.RunSundayIfDue(); err != nil {
			log.Printf("weekly owner report: %v", err)
		}
	})
	matrixInsightsHandler := handler.NewMatrixInsightsHandler(matrixInsightsService, weeklyReportService, permissionRepo)
	// ماتركس «موظف ويانه»: تذكيرات وتنبيهات لحاله بعد ٩ الصبح + صندوق قرارات المدير.
	aiActionRepo := repository.NewAiActionRepository(db)
	service.SetAIUsageRepo(repository.NewAiUsageRepository(db))
	// 💵 حد الصرف اليومي بالدولار لكل وكيل (دازاي 10-23)
	service.SetAIBudget(service.AIAgentMatrix, cfg.AIMatrixDailyUSD)
	service.SetAIBudget(service.AIAgentIris, cfg.AIIrisDailyUSD)
	log.Printf("💵 حد الصرف اليومي: ماتركس $%.2f · ايريس $%.2f", cfg.AIMatrixDailyUSD, cfg.AIIrisDailyUSD)
	service.SetAISwitches(repository.NewSystemSwitchRepository(db))
	weeklyReportService.SetActions(aiActionRepo)
	matrixAutopilotService := service.NewMatrixAutopilotService(aiActionRepo, aiRepo,
		notificationRepo, repository.NewSystemSwitchRepository(db), delayPredictionService, matrixInsightsService)
	safeguard.Loop("ماتركس التنفيذي", 41*time.Minute, 30*time.Minute, func() {
		if err := matrixAutopilotService.RunIfDue(); err != nil {
			log.Printf("matrix autopilot: %v", err)
		}
	})
	matrixDecisionsHandler := handler.NewMatrixDecisionsHandler(matrixAutopilotService)
	// عقل التوجيه: تعليمات المدير + هايكو لو المفتاح موجود.
	matrixGuideService := service.NewMatrixGuideService(repository.NewMatrixGuideRepository(db), matrixAutopilotService)
	matrixGuideService.EnableModel(cfg.AIKeyMatrix, cfg.AIModel)
	matrixGuideHandler := handler.NewMatrixGuideHandler(matrixGuideService)
	// ماتركس يقترح ويتعلّم: توقعات يومية + تعليمات من النتائج أسبوعياً + هايكو.
	matrixLearningService := service.NewMatrixLearningService(repository.NewMatrixProposalRepository(db),
		repository.NewMatrixGuideRepository(db), aiActionRepo, aiRepo, notificationRepo, matrixAutopilotService)
	matrixLearningService.EnableModel(cfg.AIKeyMatrix, cfg.AIModel)
	matrixAutopilotService.SetProposalCounter(matrixLearningService.PendingCount)
	// تذكيرات تجديد الجي بي اس القديمة الي على المالك تروح للجودة فوراً.
	go matrixAutopilotService.ReassignGpsToQuality()
	// التصفير الشهري: يوم ٢٧ الساعة ١١ بالليل — النقاط ترجع وساعات البيت تتصفّر (السجل يبقى).
	monthlyResetService := service.NewMonthlyResetService(repository.NewKpiRepository(db), repository.NewRemoteHoursRepository(db), aiRepo, notificationRepo)
	// قرار (ع) 10-06: الي ما يسجّل انصراف ينسكّر تلقائياً بعد شفته بـ٣ ساعات.
	safeguard.Loop("الانصراف التلقائي", 7*time.Minute, 15*time.Minute, func() {
		n, err := attendanceService.AutoCheckout(time.Now(), func(id, msg string) { _ = notificationRepo.Create(id, "ATTENDANCE", msg) })
		attendanceService.ExtendAutoByEvidence(time.Now())
		if err != nil {
			log.Printf("auto checkout: %v", err)
		} else if n > 0 {
			log.Printf("auto checkout: %d جلسة", n)
		}
	})
	safeguard.Loop("التصفير الشهري", 10*time.Minute, 3*time.Minute, func() {
		if err := monthlyResetService.RunIfDue(); err != nil {
			log.Printf("monthly reset: %v", err)
		}
	})
	safeguard.Loop("ماتركس يتعلّم", 47*time.Minute, 30*time.Minute, func() {
		if err := matrixLearningService.RunIfDue(); err != nil {
			log.Printf("matrix learning: %v", err)
		}
	})
	matrixLearningHandler := handler.NewMatrixLearningHandler(matrixLearningService)
	// تقرير الموظف المفصّل + أرقام الشركة والزبائن.
	matrixEmployeeReportService := service.NewMatrixEmployeeReportService(repository.NewMatrixReportRepository(db), matrixAutopilotService)
	matrixEmployeeReportService.EnableModel(cfg.AIKeyMatrix, cfg.AIModel)
	matrixBusinessService := service.NewMatrixBusinessService(repository.NewMatrixBusinessRepository(db))
	// (ع) 10-15: كل ساعة — أي شهر خلص وبيه توقع محفوظ ينحكم عليه صح/غلط ويوصل للمالك
	go func() {
		for {
			matrixBusinessService.EvaluateForecasts(func(msg string) {
				_ = notificationRepo.CreateForRolesOrPermission([]string{"ADMIN", "OWNER"}, "matrix_forecast", "matrix_forecast", msg)
			})
			time.Sleep(time.Hour)
		}
	}()
	matrixLearningService.SetReportService(matrixEmployeeReportService)
	// سجل النشاط: كل حفظ ناجح ينكتب — «شنو سوّى اليوم» بتقرير الموظف.
	employeeActivityService := service.NewEmployeeActivityService(repository.NewEmployeeActivityRepository(db))
	// كل حفظ ناجح: ينسجل بالنشاط، وماتركس يفحص تذكيرات الموظف لحظياً
	// (سجّل حضور ← التذكير ينسكّر والعين ترجع هادئة).
	// آخر ظهور: كتابة وحدة لكل موظف بالدقيقة كحد أعلى (بالذاكرة)، بـgoroutine.
	presenceRepo := repository.NewPresenceRepository(db)
	var presenceMu sync.Mutex
	presenceLast := map[string]time.Time{}
	monitorPerm := func(employeeID string) bool {
		ok, err := permissionRepo.HasPermission(employeeID, "monitoring")
		return err == nil && ok
	}
	middleware.SetMonitorPermChecker(monitorPerm)
	// (دازاي 11-01) الدور من المنصب: صاحب منصب بالهيكل ياخذ دور وحدته الفعلي
	middleware.SetOrgRolesFn(func(employeeID string) []string { return repository.OrgEffectiveRoles(db, employeeID) })
	service.MonitorPermChecker = monitorPerm
	middleware.SetSeenRecorder(func(employeeID string) {
		presenceMu.Lock()
		if time.Since(presenceLast[employeeID]) < time.Minute {
			presenceMu.Unlock()
			return
		}
		presenceLast[employeeID] = time.Now()
		presenceMu.Unlock()
		go presenceRepo.Touch(employeeID)
	})
	middleware.SetActivityRecorder(func(employeeID, method, pattern, path string, status int) {
		employeeActivityService.Record(employeeID, method, pattern, path, status)
		matrixAutopilotService.ResolveFor(employeeID)
	})
	matrixEmployeeReportService.SetActivity(employeeActivityService)
	matrixCommandService := service.NewMatrixCommandService(repository.NewMatrixCommandRepository(db), matrixAutopilotService,
		matrixEmployeeReportService, matrixBusinessService, repository.NewMatrixProposalRepository(db))
	matrixCommandService.EnableModel(cfg.AIKeyMatrix, cfg.AIModel)
	matrixCommandHandler := handler.NewMatrixCommandHandler(matrixCommandService)
	matrixReportHandler := handler.NewMatrixReportHandler(matrixEmployeeReportService, matrixBusinessService)
	bookingHandler.SetInquiryNotifier(func(customerID, employeeID string) {
		if msg := matrixBusinessService.InquiryWarning(customerID); msg != "" && employeeID != "" {
			_ = notificationRepo.Create(employeeID, "AI_AUTOPILOT", msg)
		}
	})
	leaderInvoiceService.SetNotifications(notificationRepo)
	// (ع) 10-16: جدول أسعار الديكور — فاتورة الديكور تنحسب منه
	decorRateRepo := repository.NewDecorRateRepository(db)
	leaderInvoiceService.SetDecorRates(decorRateRepo)
	decorRateHandler := handler.NewDecorRateHandler(decorRateRepo)
	leaderInvoiceService.SetMonitorFeed(monitorReviewService)
	// بقية الأقسام: كل واحد بلحظة قراره الي ما ينراجع —
	// المشتريات وقت صرف الفلوس، الجودة وقت الحكم السلبي،
	// الجي بي اس وقت تسليم الجهاز.
	procurementService.SetMonitorFeed(monitorReviewService)
	procurementService.SetNotificationRepository(notificationRepo)
	qualityFollowUpService.SetMonitorFeed(monitorReviewService)
	// حجز الصيانة المربوط بمتابعة جودة اكتمل؟ نبّه مهندس الجودة يرجع
	// يتصل بالزبون مرة ثانية يتأكد الحل انسوى.
	safeguard.Loop("كنسة اتصال الجودة الثاني", 5*time.Minute, 15*time.Minute, func() {
		if err := qualityFollowUpService.RunRecontactScan(); err != nil {
			log.Printf("quality recontact scan: %v", err)
		}
	})
	gpsService.SetMonitorFeed(monitorReviewService)
	networkCostHandler := handler.NewNetworkCostHandler(networkPriceRepo)
	jobDurationHandler := handler.NewJobDurationHandler(jobDurationEstimatorService)
	employeeMonthlyStatsService := service.NewEmployeeMonthlyStatsService(employeeRepo, kpiRepo, complaintRepo, leaderInvoiceRepo, bookingRepo, vehicleMissionRatingRepo, employeeCommissionRepo, jobDurationSampleRepo)
	// النقاط الكاملة بجدول الإحصاءات تجي من نفس حاسبة المخطط، حتى
	// الرقم بالجدول والرقم بالمخطط ما يختلفون.
	employeeMonthlyStatsService.SetSmartKpi(smartKpiService)
	// ═══ مؤشر الطاقة المستخدمة الشهري (ماتركس) ═══ — مين استخدم طاقة
	// أعلى من زملائه هذا الشهر (حجوزات + ساعات دوام + صيانات)، للمالك
	// ومدير النظام حصراً.
	energyUsageService := service.NewEnergyUsageService(aiRepo, employeeMonthlyStatsService, attendanceService, notificationRepo)
	safeguard.Loop("طاقة ماتركس الشهرية", 25*time.Minute, 6*time.Hour, func() {
		if err := energyUsageService.RunMonthlyIfDue(); err != nil {
			log.Printf("monthly energy: %v", err)
		}
	})
	// ═══ الإنجازات (ماتركس) ═══ — تقرير يومي حر من أي موظف بأي دور،
	// ربط بحجز اختياري، يوصل لمدير النظام والمالك حصراً (مساره الخاص،
	// مو صندوق المراقب — دور MONITOR ما يوصله).
	achievementService := service.NewAchievementService(achievementRepo, bookingRepo)
	// تناقض التقرير الذاتي (SELF_REPORT_MISMATCH) — ادّعى «وحدي» وكادر
	// الحجز الحقيقي أكثر من واحد.
	achievementService.SetAiRecorder(aiRepo)
	achievementService.SetCrewSizer(aiRepo)
	achievementHandler := handler.NewAchievementHandler(achievementService)
	// ماتركس يسمع الفويس: Whisper داخلي (الصوت ما يطلع) + هايكو على النص بلا أسماء.
	achievementVoiceRepo := repository.NewAchievementVoiceRepository(db)
	voiceAnalysisService := service.NewVoiceAnalysisService(achievementVoiceRepo, fileStore, cfg.WhisperURL)
	voiceAnalysisService.EnableModel(cfg.AIKeyMatrix, cfg.AIModel)
	voiceAnalysisService.Start()
	matrixGuideService.SetDialect(voiceAnalysisService.DialectHint)
	achievementHandler.SetVoice(achievementVoiceRepo, fileStore, voiceAnalysisService)
	achievementHandler.SetNotifier(notificationRepo)
	achievementDigestService := service.NewAchievementDigestService(achievementRepo, aiRepo, notificationRepo)
	safeguard.Loop("ملخص الإنجازات اليومي", 12*time.Minute, 30*time.Minute, func() {
		if err := achievementDigestService.RunIfDue(); err != nil {
			log.Printf("achievement digest: %v", err)
		}
	})
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
	designAssetHandler := handler.NewDesignAssetHandler(repository.NewDesignAssetRepository(db), employeeRepo)
	designFormRepo := repository.NewDesignFormRepository(db)
	designFormService := service.NewDesignFormService(designFormRepo)
	designFormHandler := handler.NewDesignFormHandler(designFormService)
	gpsHandler := handler.NewGpsHandler(gpsService)
	workReportHandler := handler.NewWorkReportHandler(workReportService)
	surveyReportHandler := handler.NewBookingSurveyReportHandler(surveyReportService)
	// مهندس الجودة يراجع تقارير غيره بصلاحيته، مو بدوره.
	workReportHandler.SetPermissions(permissionRepo)
	workReportHandler.SetSignatures(repository.NewCustomerSignatureRepository(db))
	statsHandler := handler.NewStatsHandler(statsService)
	vehicleHandler := handler.NewVehicleHandler(vehicleService)
	vehicleMissionHandler := handler.NewVehicleMissionHandler(vehicleMissionService, vehicleMissionRatingService, vehicleBookingService, inventoryService, employeeRepo)
	vehicleBookingHandler := handler.NewVehicleBookingHandler(vehicleBookingService)
	qualityHandler := handler.NewQualityHandler(qualityService)
	staffRequestHandler := handler.NewStaffRequestHandler(staffRequestRepo)
	// قرار طلب الكادر يوصل مدير المشاريع الي طلبه
	staffRequestHandler.SetNotificationRepository(notificationRepo)
	serviceManagerHandler := handler.NewServiceManagerHandler(serviceManagerRepo)
	locationPingHandler := handler.NewLocationPingHandler(locationPingRepo)
	performanceReviewService := service.NewPerformanceReviewService(performanceReviewRepo, employeeRepo, bookingRepo, notificationRepo, permissionRepo)
	performanceReviewHandler := handler.NewPerformanceReviewHandler(performanceReviewService)
	deviceMaintenanceHandler := handler.NewDeviceMaintenanceHandler(deviceMaintenanceService)
	teamInventoryCheckHandler := handler.NewTeamInventoryCheckHandler(teamInventoryCheckService)

	requireAuth := middleware.RequireAuth(authService, employeeRepo)
	// (ون بيس 11-10) requireAdmin انتقل لقاعدة بالنواة (internal/apps/sections)
	// صلاحية «إحصائيات الموظفين» تفتح الشاشتين — چانت تظهر البند والخادم يرفضه.
	// (ون بيس 11-10) requireEmployeeStats انتقل لقاعدة بالنواة (internal/apps/sections)
	// حصراً لحساب المالك (OWNER) — أقوى من الأدمن العادي، ما يشوفها إلا هو
	// (ون بيس 11-10) requireOwner انتقل لقاعدة بالنواة (internal/apps/sections)
	// فتح الحسابات: المالك وحده، بلا تسجيل مخالفة على مدير النظام
	requireOwnerAccounts := middleware.RequireOwnerOnly("فتح الحسابات للمالك وحده")
	// ⚠️ جان `RequireRole("ADMIN","FINANCE")` بلا أي باب صلاحية — يعني
	// اعتماد فاتورة الليدر وحكم التدقيق وسحب الاعتماد والتعديل وحالة
	// المصروف **كلها بالدور حصراً**. فلمّا المحاسب ياخذ إجازة ما عندنا
	// طريق نخلي غيره يكمل: المالك ينطي صلاحية والاعتماد يبقى مرفوض
	// والفاتورة ما تمشي. الإضافة **زيادة مو تضييق** — ولا حارس ينضيّق،
	// فما تنكسر صلاحية أي موظف شغّال.
	// (ون بيس 11-10) requireFinance انتقل لقاعدة بالنواة (internal/apps/sections)
	// تدقيق مبلغ الحجز يعتمد على صلاحية "finance" الممنوحة فعلياً للموظف (مو بس دوره
	// الوظيفي) — المراقب مثلاً عنده هذي الصلاحية افتراضياً ويشوف زر "تدقيق" بالواجهة،
	// فلازم الباك إند يتحقق من نفس الصلاحية بدل دور صارم، وإلا يترفض الطلب ويتسبب
	// بإيقاف حساب الموظف تلقائياً بعد 3 محاولات (حماية أمنية ضد التلاعب بالجلسة).
	// و`finance_audit` جنبها: المفتاح الي المالك ينطيه لمن يريده حتى
	// يأشّر «مطابق/غير مطابق» بدل المحاسب.
	//
	// 🔴 **والدور يمرّ هنا بقصد**: `RequireAnyPermission` تسجّل **مخالفة**
	// عند الرفض، و**ثلاث محاولات توقّف الحساب تلقائياً** (فوگ). وخريطة
	// افتراضي الأدوار **اقتراح مو إلزام** — فمراقب ما انصبّت له
	// صلاحياته يفتح شاشته ويضغط ثلاث ضغطات و**ينقفل حسابه** على رفض
	// مفروض يكون هادئاً. والقراءة (جلب التدقيق اليومي) على نفس الحارس!
	//
	// فالتوسيع هنا **بالقراءة**، والتضييق يبقى **بالقرار**:
	// `booking_audit_handler.go` يفحص المفتاح للمراقب ويرجّعه ٤٠٣ نظيفاً
	// بلا مخالفة ولا قفل.
	// (11-04 — ٦أ) التكرار الآمن لمسارات الفلوس
	idem := middleware.Idempotent(db)
	// تنظيف مفاتيح التكرار الآمن مرة باليوم (11-04)
	go func() {
		for {
			if n, err := middleware.CleanupIdempotency(db); err != nil {
				log.Printf("[idempotency] تنظيف: %v", err)
			} else if n > 0 {
				log.Printf("[idempotency] انمسح %d مفتاح قديم", n)
			}
			time.Sleep(24 * time.Hour)
		}
	}()
	// (ون بيس 11-10) requireVerifyBooking انتقل لقاعدة بالنواة (internal/apps/sections)
	// ربط حجزين تاريخيين منفصلين كإنجاز جزئي لنفس الشغلة — صلاحية
	// مستقلة (نفس فلسفة finance_audit): المالك ينطيها فرد-فرد، مو دور.
	// (ون بيس 11-10) requirePartialLink انتقل لقاعدة بالنواة (internal/apps/sections)
	// (ون بيس 11-10) requireCoordinator انتقل لقاعدة بالنواة (internal/apps/sections)
	// (ون بيس 11-10) requireCrewManagement انتقل لقاعدة بالنواة (internal/apps/sections)
	// تدقيق أمني: مسارات إدارية كانت مفتوحة لأي موظف مسجّل دخول (requireAuth
	// فقط) — أي فني يقدر يثبّت حجز أو يعدّل زبون أو يضيف أدوات. صارت مربوطة
	// بصلاحياتها الطبيعية، مع إبقاء المسارات الذاتية (حضور/إشعارات) مفتوحة.
	// (ون بيس 11-10) requireCustomerMgmt انتقل لقاعدة بالنواة (internal/apps/sections)
	requireBookingCoord := guards.AnyPermission("coordinator", "crew_management")
	// تعديل تفاصيل الحجز تستخدمه صفحات ثانية غير التنسيق (قائمة الحجوزات
	// لتخصيص مركبة، وخريطة المهام) — فنوسّعها لصلاحياتهن حتى ما ننكسر شغل شغّال.
	requireBookingEdit := guards.AnyPermission(
		"coordinator", "crew_management", "view_bookings", "mission_tracking", "sales_booking")
	// إدارة المخزون (عدة الموظفين الشخصية/أدوات المركبات/العدة القياسية) — مسموحة
	// لأدوار HR/ADMIN كالمعتاد، أو لأي موظف عنده صلاحية "inventory" المخصصة
	// (ممنوحة من صفحة الصلاحيات، مثلاً PROCUREMENT_ADMIN) — توسيع وصول، مو تضييق.
	// (ون بيس 11-10) requireHROrInventory انتقل لقاعدة بالنواة (internal/apps/sections)
	requireLeader := middleware.RequireLeader(employeeRepo, notificationRepo)
	// تقرير ١ (B2): صلاحية «صيانة الأجهزة» تفتح البند لغير الليدر — الخادم يقبلها هم.
	// (ون بيس 11-10) requireDeviceMaint انتقل لقاعدة بالنواة (internal/apps/sections)
	// (ون بيس 11-10) الليدر أو مسؤول خدمة الحجز صار invoices.create بالنواة
	// (ون بيس 11-10) requireInventoryView انتقل لقاعدة بالنواة (internal/apps/sections)
	// (ون بيس 11-10) requireToolRequest انتقل لقاعدة بالنواة (internal/apps/sections)
	// حساب كلفة التنفيذ: صلاحية تنعطى وتنسحب، مو دور ثابت
	// ⚠️ **واعية للّيدر** مو صلاحية صرفة — وهذا إصلاح خلل حظر حسابات:
	//
	// الليدر الي يسوي فاتورة **لازم** يحسب كلفتها؛ الحساب جزء من الفاتورة
	// مو ميزة منفصلة. وصفحة الفاتورة تدزّ طلب الحساب **تلقائياً** وهو يكتب،
	// فالليدر الي ما عنده صلاحية `execution_cost` چان يجمّع مخالفات وصول
	// بلا ما يدري — وبخمس مخالفات بعشر دقائق **ينحظر حسابه تلقائياً**
	// وينرمي لشاشة الدخول، وفكّ الحظر للمالك وحده.
	//
	// السبب إن ترحيل 0189 منح `execution_cost` **مرة وحدة** للّيدرية
	// الموجودين وقتها، وأي واحد ينترقّى ليدراً بعدها ما ياخذها.
	//
	// وباقي مسارات فواتير الليدر أصلاً واعية للّيدر (requireLeaderBasket)،
	// فهذا الحارس چان **متناقض** وياهن.
	requireExecutionCost := guards.LeaderOrPermission("execution_cost")
	// موافقة/رفض طلبات الأدوات: كانت دور صارم بدون منفذ صلاحية، فإداري الكميات
	// — وهو صاحب الشغلة أصلاً — ما كان يقدر يوافق أبداً مهما انمنحت له
	// صلاحيات. صارت صلاحية مستقلة تُمنح لأي موظف، مع إبقاء الأدوار القديمة.
	// (ون بيس 11-10) requireInventoryApprove انتقل لقاعدة بالنواة (internal/apps/sections)
	// تعديل مهارات موظف يعتمد على صلاحية "staff_management" الممنوحة فعلياً (نفس
	// الصلاحية الي تفتح صفحة "إدارة الكوادر" بالواجهة للمراقب أيضاً) — مو دور
	// وظيفي صارم، وإلا نفس بگ "تدقيق الحسابات" يتكرر: زر يطلع بالواجهة، السيرفر
	// يرفضه، وبعد 3 محاولات ينوقف حساب الموظف تلقائياً.
	// requireStaffManagement انتقل للنواة (people.skills) — 11-10
	// ═══ حرّاس الصلاحيات المستقلة ═══
	//
	// ⚠️ الإخفاء بالقائمة الجانبية **مو حماية**: الموظف يكتب المسار
	// بالمتصفح أو يدز طلب مباشر ويوصل. لازم السيرفر يرفض بعد، وإلا
	// المنع تجميلي بس.
	// (ون بيس 11-10) requireBookingsArchive انتقل لقاعدة بالنواة (internal/apps/sections)
	// (ون بيس 11-10) requireExtraTaskAssign انتقل لقاعدة بالنواة (internal/apps/sections)
	// (ون بيس 11-10) requireTrainingManage انتقل لقاعدة بالنواة (internal/apps/sections)
	// ⚠️ RequireRoleOrAnyPermission لا RequireRole: صاحب النظام يريد يمنح
	// "يتابع ويدقّق" (monitoring/auditing) لموظف واحد بلا ما يحوّل دوره
	// لمراقب فعلياً — نفس مبدأ الصلاحيات الفردية المعتمد بكل الشاشة.
	requireMonitor := guards.RoleOrAnyPermission([]string{"ADMIN", "MONITOR"}, "monitoring", "auditing")
	// (دازاي 10-25) صلاحيات المراقب المفصّلة: كل مسار مراقب يقبل الحزمة كاملة
	// («monitoring»/«auditing») أو صلاحيته الصغيرة بس.
	_ = requireMonitor // كل مسارات المراقب صارت على requireMon[sub]
	// التبريرات تنفتح من صندوق المراقب هم — صاحب الصندوق أو صاحب التبريرات
	// (ون بيس 11-10) requireMonJust انتقل لقاعدة بالنواة (internal/apps/sections)
	requireMon := map[string]func(http.Handler) http.Handler{}
	for _, sub := range model.MonitorSubPermissions {
		requireMon[sub] = guards.RoleOrAnyPermission([]string{"ADMIN", "MONITOR"}, "monitoring", "auditing", sub)
	}
	requireProjectManager := middleware.RequireRole(employeeRepo, notificationRepo, "ADMIN", "PROJECT_MANAGER")
	// إدارة المشاريع: بالدور أو بصلاحية project_management الممنوحة يدوياً
	// (ون بيس 11-10) requireProjectMgmt انتقل لقاعدة بالنواة (internal/apps/sections)
	// (ون بيس 11-10) requireFieldMonitor انتقل لقاعدة بالنواة (internal/apps/sections)
	// (ون بيس 11-10) requireGpsAdmin انتقل لقاعدة بالنواة (internal/apps/sections)
	// (ون بيس 11-10) requireContentTech انتقل لقاعدة بالنواة (internal/apps/sections)
	// الموردون بصلاحية مستقلة (انفصلت عن content_technician الواسعة)
	// (ون بيس 11-10) requireSuppliersMgmt انتقل لقاعدة بالنواة (internal/apps/sections)
	// (ون بيس 11-10) requireVehicleMgmt انتقل لقاعدة بالنواة (internal/apps/sections)
	// (ون بيس 11-10) requireProcurement انتقل لقاعدة بالنواة (internal/apps/sections)
	// توفير المواد وتحديد حالتها يقتصر على إداري الكميات فعلياً (أو الأدمن) — مو أي
	// موظف عنده صلاحية "procurement" العامة (زي الفني/مدير المشاريع الي بس يطلبون مواد).
	// (ون بيس 11-10) requireProcurementAdmin انتقل لقاعدة بالنواة (internal/apps/sections)
	// unit_technicians مو content_technician عمداً — صلاحية مستقلة لوحدة
	// التقنيين تحديداً (المعارض/المنتجات/الخدمات) حتى ما تنفتح بالغلط لأي حد
	// عنده content_technician من مكان ثاني (مواد التدريب/الموردين/المنتجات).
	// (ون بيس 11-10) requireUnitTechnicians انتقل لقاعدة بالنواة (internal/apps/sections)
	// (ون بيس 11-10) requireTechUnitOrProcurement انتقل لقاعدة بالنواة (internal/apps/sections)
	// قرار (ع) 10-10: مطوّر الخدمات شغله دراسات الخدمات — يشوفها ويقترح ويرفع تقارير
	// (ون بيس 11-10) requireServiceStudies انتقل لقاعدة بالنواة (internal/apps/sections)
	// 🔴 (10-16) چان صلاحية بس: الواجهة تطلّع أزرار الشكوى لمهندس الجودة بدوره
	// (ComplaintsPage fullDetailRoles)، فمهندس جودة ما انمنحله quality_control
	// چان ياخذ ٤٠٣ = مخالفة وقفل حساب. هسه الدور أو الصلاحية — نفس الواجهة.
	// (ون بيس 11-10) requireQuality انتقل لقاعدة بالنواة (internal/apps/sections)
	// (ون بيس 11-10) requireProjectMgmtPerm انتقل لقاعدة بالنواة (internal/apps/sections)
	// (ون بيس 11-10) requireKpi انتقل لقاعدة بالنواة (internal/apps/sections)
	// (ون بيس 11-10) requireKpiCriteria انتقل لقاعدة بالنواة (internal/apps/sections)
	// إنشاء/تعديل عروض الأسعار: أي مستوى من الثلاث (إضافة فقط / إضافة وتعديل
	// عروضي / إضافة وتعديل واطلاع الكل) يقدر يوصل للـPOST/PUT — الفرز الدقيق
	// (مين يشوف شنو، ومين يقدر يعدّل شنو) يصير داخل QuotationService نفسه.
	// quotation_system القديمة تبقى مرادف quotation_manage_all لأي موظف
	// كانت ممنوحة له من قبل.
	requireQuotationAccess := guards.AnyPermission(
		service.QuotationPermCreate, service.QuotationPermEditOwn, service.QuotationPermManageAll, "quotation_system")
	// إنشاء/تعديل بيانات نظام GPS (عملاء/شرائح/أجهزة/تجديد/صيانة) يقتصر على من
	// يملك صلاحية "gps_system" فعلياً (نفس الصلاحية الي تفتح كل صفحات نظام GPS
	// بالواجهة) — مو أي موظف مسجل دخول.
	// (ون بيس 11-10) requireFund انتقل لقاعدة بالنواة (internal/apps/sections)
	// (ون بيس 11-10) مبالغ الدوار والتخريج صارن قواعد بالنواة: finance.fund_update/
	// fund_topup (authz.FundAmount) وfinance.settle_discharge (authz.FundDischarge).
	// الاتصال بالزبون مو حكر على مهندس الجودة — أي موظف ينطيه المدير
	// صلاحية complaint_contact يكدر يتصل ويأشر النتيجة باسمه.
	// (ون بيس 11-10) requireComplaintContact انتقل لقاعدة بالنواة (internal/apps/sections)
	// (ون بيس 11-10) requireGpsSystem انتقل لقاعدة بالنواة (internal/apps/sections)
	// تقرير ١ (B1): المبيعات يطلعلهم «طلب GPS/تسليم/تجديد/صيانة» — الإرسال يقبلهم
	// (ويا صاحب gps_requests) بدل ما يرجع 403 وينقفل الحساب.
	// (ون بيس 11-10) requireGpsRequest انتقل لقاعدة بالنواة (internal/apps/sections)
	// متابعة تجديد اشتراكات الجي بي اس تخص الاثنين: مهندس الجودة يتصل
	// بالزبائن، ومسؤول الجي بي اس يشوف منو خلصت مهلته وشريحته تحتاج حرق.
	// (ون بيس 11-10) requireGpsOrQuality انتقل لقاعدة بالنواة (internal/apps/sections)
	// قوائم زبائن الجي بي اس فيها أسماء وأرقام ٤٩٣ زبون — كانت مفتوحة لأي
	// موظف مسجّل دخول، يعني الفني يقدر يسحبها كلها من F12 وهو ما إله علاقة
	// بالجي بي اس أصلاً. صارت محصورة بالي شغلهم فعلاً بالجي بي اس.
	// (ون بيس 11-10) requireVipManualAdd انتقل لقاعدة بالنواة (internal/apps/sections)
	// (ون بيس 11-10) requireGpsData انتقل لقاعدة بالنواة (internal/apps/sections)

	mux := http.NewServeMux()
	// 🏴‍☠️ «ون بيس» (دازاي 11-10): التطبيقات الجديدة تتسجّل عبر النواة — كل مسار بفعل
	// وقاعدة، والرفض 404. تعيش جنب المسارات القديمة لحد ما تنتقل وحدة وحدة.
	coreRouter := core.NewRouter(mux, requireAuth, permissionRepo, core.NewPolicies())
	coreRouter.SetLeader(func(id string) bool { ok, err := employeeRepo.IsLeaderFreshByID(id); return err == nil && ok })
	coreRouter.OnDeny(func(r *http.Request, id string) { middleware.RecordViolation(r, employeeRepo, notificationRepo, id) })
	meApp := me.New() // «شغلي أنا»: المعالجات تنحط بمكانها، والتسجيل بالأخير
	requestsApp := requests.New()
	accessApp := access.New()   // الصلاحيات نفسها
	peopleApp := people.New()   // الموظفين
	financeApp := finance.New() // الفلوس
	secs := sections.New()      // باقي الأقسام (scripts/onepiece-bulk.py)
	designsApp := core.NewSetApp("designs", map[core.Action]core.Rule{})
	trainingCenterApp := core.NewSetApp("training_center", map[core.Action]core.Rule{})
	invoicesApp := invoices.New(invoices.Deps{IsBookingManager: paperworkGuard.IsBookingManager})
	fieldApp := field.New(field.Deps{
		BookingParty:        func(b, e string) bool { ok, err := bookingRepo.IsAssignedTo(b, e); return err == nil && ok },
		ProjectParty:        repository.NewProjectRepository(db).IsParty,
		BookingLeader:       repository.NewMatrixChainRepository(db).IsBookingLeader,
		BookingTech:         repository.NewMatrixChainRepository(db).IsBookingTech,
		SurveyPhotoUploader: repository.NewSurveyPhotoRepository(db).UploaderOf,
		CanCreateInternal:   bookingHandler.CanCreateInternal,
		CanLeaveDecide:      leaveHandler.CanDecide,
		SurveyParty:         surveyReportRepo.IsBookingParty,
		CanUploadSurvey:     surveyPhotoHandler.CanUploadReq,
	})

	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// حماية ضد محاولات تخمين كلمة السر المتكررة: أقصى 8 محاولات دخول بالدقيقة
	// من نفس عنوان IP — هذا بالضبط النوع اللي سبب حظر IP السيرفر من Hetzner
	// كإجراء حماية تلقائي ضدهم لما شافوا محاولات دخول متكررة سريعة.
	requireLoginRateLimit := middleware.RateLimit(8, time.Minute)
	mux.Handle("POST /api/auth/login", requireLoginRateLimit(http.HandlerFunc(authHandler.Login)))
	// عدّاد لحاله: المكتب كله يطلع من عنوان واحد، فالخروج ما لازم ياكل من محاولات دخول غيره
	mux.Handle("POST /api/auth/logout", middleware.RateLimit(60, time.Minute)(http.HandlerFunc(authHandler.Logout)))
	// رفع الملفات يحتاج تسجيل دخول. العرض بعد — الملفات تحمل صور وصولات
	// ووثائق زبائن، وما تنعرض لأي أحد يخمّن الرابط.
	requestsApp.Set("POST /api/files", "requests.file_upload", fileHandler.Upload)
	secs["files"].Set("GET /api/files/token", "files.get_token", fileHandler.Token)
	// العرض ما يمر بـrequireAuth لأن <img> ما يرسل ترويسة Authorization —
	// التحقق يصير بالوسم الموقّع بالمعالج نفسه.
	mux.Handle("GET /api/files/", http.HandlerFunc(fileHandler.Serve))

	// ═══ مفاتيح النظام ═══
	//
	// ⚠️ **القراءة لكل موظف**: واجهة كل موظف تحتاج تعرف هل الميزة
	// مطفية، وإلا تبقى تعرضها. **والكتابة تختلف بالمفتاح**: بعض
	// المفاتيح (شخصية الكائن) تبقى للمالك حصراً، وبعضها (شريط
	// الإعلانات) مفتوح لمدير النظام كمان بطلب صريح من (ع) — لهذا
	// القرار داخل SystemSwitchHandler.Set نفسه مو حارساً واحداً هنا.
	secs["system_x"].Set("GET /api/system/switches", "system_x.get_switches", systemSwitchHandler.List)

	secs["auth"].Set("GET /api/auth/me", "auth.get_me", authHandler.Me)
	secs["me_x"].Set("PUT /api/me/language", "me_x.put_language", handler.SetMyLanguage(db).ServeHTTP) // بلاغ #11: الأوردو
	// تغيير كلمة المرور لمدير النظام والمالك بس. الموظف ما يغيّر سره
	// بنفسه — كلمات السر تنتحدد من «إدارة الموظفين» عند الأدمن، حتى
	// تبقى معروفة عند الإدارة وما يقفل موظف حسابه على نفسه.
	secs["auth"].Set("PUT /api/auth/change-password", "auth.put_change_password", authHandler.ChangePassword)

	// موظفين — القراءة تحتاج تسجيل دخول فقط، الإنشاء/التعديل الحساس محمي بدور ADMIN
	peopleApp.Set("GET /api/employees", "people.list", employeeHandler.List)
	peopleApp.Set("GET /api/employees/supervisors", "people.supervisors", employeeHandler.Supervisors)
	disciplineHandler := handler.NewDisciplineHandler(disciplineService)
	employeeLetterHandler := handler.NewEmployeeLetterHandler(employeeLetterRepo, notificationRepo)
	bookingVisitRepo := repository.NewBookingVisitRepository(db)
	bookingProgressHandler := handler.NewBookingProgressHandler(bookingProgressRepo, bookingRepo, notificationRepo, bookingVisitRepo)
	bookingProgressHandler.SetAiRecorder(aiRepo)
	// نقاط الانضباط: كل موظف يشوف الأرصدة (الشفافية جزء من العقوبة)،
	// وتشغيل الفحص يدوياً للمدير حصراً.
	// ── الطلبات: كتاب رسمي من الموظف للإدارة ──
	// التقديم وقراءة طلباتك: لكل موظف. البت: للمالك ومدير النظام.
	secs["letters"].Set("GET /api/letters/addressees", "letters.get_addressees", employeeLetterHandler.Addressees)
	requestsApp.Set("POST /api/letters", "requests.letter", employeeLetterHandler.Create)
	secs["letters"].Set("GET /api/letters/mine", "letters.get_mine", employeeLetterHandler.Mine)
	secs["letters"].Set("GET /api/letters", "letters.get_root", employeeLetterHandler.Inbox)
	secs["letters"].Set("GET /api/letters/pending-count", "letters.get_pending_count", employeeLetterHandler.PendingCount)
	secs["letters"].Set("PUT /api/letters/{id}/decide", "letters.put_by_id_decide", employeeLetterHandler.Decide)

	// (ع) 10-11: سجل الانضباط (أرصدة كل الموظفين ومخالفاتهم) للمالك والمدير بس
	secs["discipline"].Set("GET /api/discipline", "discipline.get_root", disciplineHandler.List)
	secs["discipline"].Set("GET /api/discipline/events", "discipline.get_events", disciplineHandler.Events)
	// أرقام مجرّدة بلا أسماء — تنعرض بشريط الإعلانات لكل الموظفين
	secs["discipline"].Set("GET /api/discipline/today-headline", "discipline.get_today_headline", disciplineHandler.TodayHeadline)
	// التعديل اليدوي: المالك ومدير النظام بس. requireAdmin يمرّر OWNER أصلاً.
	secs["discipline"].Set("POST /api/discipline/adjust", "discipline.post_adjust", disciplineHandler.Adjust)
	peopleApp.Set("GET /api/employees/archived", "people.archived", employeeHandler.ListArchived)
	// 🧠 بوابة الالتزام (11-02 — تعديلات ٩.١): الاستمارة + تقرير الإنجاز اليومي
	workGate := handler.NewWorkGateHandler(db, notificationRepo)
	achievementHandler.OnReport = workGate.MirrorAchievement // دازاي 11-05: الإنجاز = تقرير اليوم
	secs["me_x"].Set("GET /api/me/work-gate", "me_x.get_work_gate", workGate.Gate)
	// ═══ بوت بلاغات الموظفين — تلغرام (دازاي 11-05) ═══
	// الـ webhook بلا تسجيل دخول بس محمي برمز سري يرسله التلغرام (بلاه ٤٠٤).
	tgBot := handler.NewTelegramBotHandler(repository.NewBotReportRepository(db), telegram.New(cfg.TelegramBotToken, cfg.TelegramAPIBase), fileStore, permissionRepo, notificationRepo, cfg.JWTSecret)
	tgBot.Start(cfg.TelegramWebhookURL)
	// رمز القراءة بس (Claude يقرا البلاغات وصورها) — بلا كتابة أبداً. فاضي = ٤٠٤.
	tgBot.SetReaderToken(cfg.BotReaderToken)
	// صور البلاغات: كل يوم تنمسح الأقدم من أسبوع (قرار دازاي 11-05 — حتى ما يثگل السيرفر)
	go func() {
		for {
			if n, err := tgBot.PurgeOldPhotos(context.Background()); err != nil {
				log.Printf("[telegram] تنظيف الصور: %v", err)
			} else if n > 0 {
				log.Printf("[telegram] انمسحت %d صورة بلاغ أقدم من أسبوع", n)
			}
			time.Sleep(24 * time.Hour)
		}
	}()
	mux.Handle("GET /api/bot-reader/reports", middleware.RateLimit(60, time.Minute)(http.HandlerFunc(tgBot.ReaderReports)))
	mux.Handle("GET /api/bot-reader/files/{key...}", middleware.RateLimit(120, time.Minute)(http.HandlerFunc(tgBot.ReaderFile)))
	mux.Handle("GET /api/bot-reader/suggestions", middleware.RateLimit(60, time.Minute)(http.HandlerFunc(tgBot.ReaderSuggestions)))
	mux.Handle("POST /api/telegram/webhook", middleware.RateLimit(600, time.Minute)(http.HandlerFunc(tgBot.Webhook)))
	secs["me_x"].Set("GET /api/me/telegram", "me_x.get_telegram", tgBot.MyLink)
	meApp.Set("POST /api/me/telegram/link", "me.telegram_link", tgBot.CreateLink)
	// الربط مرة وحدة: الموظف ما يفك، الفك بس من المالك
	requireTgOwner := authz.Strict(permissionRepo, authz.TelegramLinks)
	secs["telegram_links"].Set("GET /api/telegram-links", "telegram_links.get_root", tgBot.Links, requireTgOwner)
	secs["telegram_links"].Set("DELETE /api/telegram-links/{employeeId}", "telegram_links.delete_by_employeeid", tgBot.OwnerUnlink, requireTgOwner)
	// البلاغ من داخل النظام (أي موظف مسجّل)، وبلاغاته هو بس
	meApp.Set("POST /api/me/bot-reports", "me.bot_report", tgBot.SubmitMine)
	secs["me_x"].Set("GET /api/me/bot-reports", "me_x.get_bot_reports", tgBot.MyReports)
	requireBotReports := authz.Strict(permissionRepo, authz.BotReports)
	secs["bot_reports"].Set("GET /api/bot-reports", "bot_reports.get_root", tgBot.List, requireBotReports)
	secs["bot_reports"].Set("PATCH /api/bot-reports/{id}", "bot_reports.patch_by_id", tgBot.Update, requireBotReports)
	meApp.Set("POST /api/me/profile-survey", "me.profile_survey", workGate.SubmitSurvey)
	meApp.Set("POST /api/me/daily-report", "me.daily_report", workGate.SubmitDaily)
	secs["me_x"].Set("GET /api/me/daily-reports", "me_x.get_daily_reports", workGate.MyDaily)
	secs["me_x"].Set("GET /api/me/rating-duty", "me_x.get_rating_duty", workGate.MyDuty)
	meApp.Set("POST /api/me/rating-duty", "me.rating_duty", workGate.SubmitDuty)
	safeguard.Loop("واجب التقييم اليومي", 9*time.Minute, time.Hour, func() {
		if n, err := workGate.Sweep(); err != nil {
			log.Printf("rating duty: %v", err)
		} else if n > 0 {
			log.Printf("rating duty: %d سجل إهمال جديد", n)
		}
	})
	// (ون بيس 11-10) requireDailyReports انتقل لقاعدة بالنواة (internal/apps/sections)
	// (ون بيس 11-10) requireSurveys انتقل لقاعدة بالنواة (internal/apps/sections)
	secs["field_timing"].Set("GET /api/field-timing", "field_timing.get_root", workGate.FieldTiming)
	secs["daily_reports"].Set("GET /api/daily-reports/{id}/match", "daily_reports.get_by_id_match", workGate.MatchDetail)
	secs["daily_reports"].Set("POST /api/daily-reports/{id}/match-review", "daily_reports.post_by_id_match_review", workGate.ReviewMatch)
	safeguard.Loop("مطابقة تقارير الإنجاز", 13*time.Minute, time.Hour, func() {
		if n, err := workGate.SweepMatches(); err != nil {
			log.Printf("report match: %v", err)
		} else if n > 0 {
			log.Printf("report match: %d تقرير بيه فرق", n)
		}
	})
	secs["daily_reports"].Set("GET /api/daily-reports", "daily_reports.get_root", workGate.ListDaily)
	peopleApp.Set("GET /api/employees/{id}/dossier", "people.dossier", workGate.Dossier)
	secs["negligence"].Set("GET /api/negligence", "negligence.get_root", workGate.Negligence)
	secs["profile_surveys"].Set("GET /api/profile-surveys", "profile_surveys.get_root", workGate.ListSurveys)
	secs["profile_surveys"].Set("POST /api/profile-surveys/{id}/review", "profile_surveys.post_by_id_review", workGate.ReviewSurvey)
	// 🩺 سلامة البيانات (11-02 — شبكة الأمان ٢): قراءة بس، للمالك
	integrityHandler := handler.NewIntegrityHandler(db, notificationRepo)
	// (11-04 — ٦ج) 🗑️ المحذوفات: المالك يشوف ويرجّع
	deletedRecords := handler.NewDeletedRecordsHandler(db)
	secs["owner"].Set("GET /api/owner/deleted", "owner.get_deleted", deletedRecords.List)
	secs["owner"].Set("POST /api/owner/deleted/{id}/restore", "owner.post_deleted_by_id_restore", deletedRecords.Restore)
	secs["integrity"].Set("GET /api/integrity", "integrity.get_root", integrityHandler.Latest)
	secs["integrity"].Set("POST /api/integrity/run", "integrity.post_run", integrityHandler.RunNow)
	secs["integrity"].Set("GET /api/integrity/{key}/sample", "integrity.get_by_key_sample", integrityHandler.Sample)
	safeguard.Loop("فحص سلامة البيانات", 25*time.Minute, 24*time.Hour, func() {
		if _, _, err := integrity.RunAndStore(db, "daily", notificationRepo); err != nil {
			log.Printf("integrity: %v", err)
		}
	})
	secs["security"].Set("GET /api/security/dashboard", "security.get_dashboard", securityHandler.Dashboard)
	secs["security"].Set("POST /api/security/unlock/{id}", "security.post_unlock_by_id", securityHandler.Unlock)
	secs["security"].Set("POST /api/security/reset-attempts/{id}", "security.post_reset_attempts_by_id", securityHandler.ResetAttempts)
	secs["security"].Set("POST /api/security/free-memory", "security.post_free_memory", securityHandler.FreeMemory)
	peopleApp.Set("GET /api/employees/match", "people.match", employeeHandler.Match)
	peopleApp.Set("GET /api/employees/{id}", "people.view", employeeHandler.Get)
	// فتح حساب جديد = المالك وحده. مدير النظام يدير كلشي بالنظام لكن
	// **منو يدخل النظام** قرار المالك — وهذا يصير أهم بعد ما يجي
	// النظام الأكبر ويصير الحساب الواحد يفتح عالمين.
	// (ع): «إضافة موظف» صارت صلاحية تنطى — المالك أو صاحب employee_create،
	// والفحص بالمعالج (بلا أدوار عليا لغير المالك).
	employeeAdminHandler := handler.NewEmployeeAdminHandler(employeeService, permissionRepo, employeeRepo)
	peopleApp.Set("POST /api/employees", "people.create", employeeAdminHandler.Create)
	_ = requireOwnerAccounts
	peopleApp.Set("GET /api/employees/my-admin-scope", "people.my_scope", employeeAdminHandler.MyScope)
	peopleApp.Set("GET /api/employees/{id}/admin-scope", "people.scope_view", employeeAdminHandler.GetScope)
	peopleApp.Set("PUT /api/employees/{id}/admin-scope", "people.scope_set", employeeAdminHandler.SetScope)
	peopleApp.Set("PUT /api/employees/{id}/all", "people.edit_all", employeeAdminHandler.UpdateAll)
	peopleApp.Set("PUT /api/employees/{id}/suspend", "people.suspend", employeeAdminHandler.Suspend)
	peopleApp.Set("PUT /api/employees/{id}/reactivate", "people.reactivate", employeeAdminHandler.Reactivate)
	peopleApp.Set("GET /api/employees/suspended", "people.suspended", employeeAdminHandler.ListSuspended)
	peopleApp.Set("PUT /api/employees/{id}/schedule", "people.schedule", employeeAdminHandler.Schedule)
	secs["work_schedule"].Set("GET /api/work-schedule", "work_schedule.get_root", employeeAdminHandler.WorkSchedule)
	peopleApp.Set("PUT /api/employees/{id}", "people.update", employeeHandler.Update)
	peopleApp.Set("PUT /api/employees/{id}/identity", "people.identity", employeeHandler.UpdateIdentity)
	peopleApp.Set("PUT /api/employees/{id}/profile", "people.profile", employeeHandler.UpdateProfile)
	peopleApp.Set("POST /api/employees/{id}/link-historical", "people.link_history", employeeHandler.LinkHistoricalRecords)
	peopleApp.Set("PUT /api/employees/{id}/skills", "people.skills", employeeHandler.SetSkills)

	// الصلاحيات — العرض متاح لأي مسجل دخول، التعديل والتطبيق التلقائي محصور بمدير النظام فقط
	// قائمة كل صلاحيات النظام = خارطة للمهاجم يعرف بيها شنو موجود وشنو
	// يستهدف. الموظف ما يحتاجها — صلاحياته هو تجي بـ/permissions/employee/{id}
	accessApp.Set("GET /api/permissions", "access.list", permissionHandler.ListAll)
	// فحص: منو من الموظفين ناقصه صلاحيات دوره. قراءة بس، ما يمنح شي.
	// تقييم مهارات القيادة — چانت أشرطة ما تنحفظ وتنتقل من موظف لموظف.
	// ⚠️ القراءة لأي مسجَّل (الموظف يشوف تقييمه)، والكتابة للمالك
	// والمدير وإداري الكوادر بس.
	peopleApp.Set("GET /api/employees/{id}/leader-skills", "people.leader_skills", leaderSkillHandler.List)
	peopleApp.Set("PUT /api/employees/{id}/leader-skills", "people.leader_skills_set", leaderSkillHandler.Set)
	// ═══ سجل الأقسام ومسؤوليها — للحجز داخل الشركة ═══
	//
	// ⚠️ القراءة لأي مسجَّل: منتقي القسم بشاشة الحجز يحتاجها، وحصرها
	// بصلاحية يخلّي المنتقي فارغاً لمن يفتحه — «ماكو أقسام» بينما
	// السجل مليان.
	// ⚠️ والتعديل «فقط لمالك ومدير النظام حالياً، وخليها صلاحية
	// أيضاً» — نفس نمط الحراس المزدوجة بالنظام.
	requireDepartmentsManage := middleware.RequireRoleOrAnyPermission(
		permissionRepo, employeeRepo, notificationRepo,
		[]string{"ADMIN", "OWNER"}, "departments_manage")
	secs["departments"].Set("GET /api/departments", "departments.get_root", departmentHandler.List)
	secs["departments"].Set("POST /api/departments", "departments.post_root", departmentHandler.Create, requireDepartmentsManage)
	secs["departments"].Set("PUT /api/departments/{id}", "departments.put_by_id", departmentHandler.Update, requireDepartmentsManage)
	secs["department_heads"].Set("POST /api/department-heads", "department_heads.post_root", departmentHandler.CreateHead, requireDepartmentsManage)
	secs["department_heads"].Set("PUT /api/department-heads/{id}", "department_heads.put_by_id", departmentHandler.UpdateHead, requireDepartmentsManage)
	accessApp.Set("GET /api/permissions/audit-defaults", "access.audit_defaults", permissionHandler.AuditDefaults)
	accessApp.Set("GET /api/permissions/role-defaults", "access.role_defaults", permissionHandler.RoleDefaults)
	// قائمة الموظفين الي يوصلون لصلاحية معيّنة — لتعبئة القوائم المنسدلة
	// (مثلاً: مين المسؤول عن المشروع، ومين يسوي الكشف)
	accessApp.Set("GET /api/permissions/employees", "access.holders", permissionHandler.EmployeesWithPermission)
	accessApp.Set("GET /api/permissions/employee/{id}", "access.employee_view", permissionHandler.ListForEmployee)
	accessApp.Set("PUT /api/permissions/employee/{id}", "access.employee_set", permissionHandler.SetForEmployee)
	// (ع) 10-17: تسليم شغل موظف لموظف جديد — المنصب والصلاحيات والشغل المفتوح
	employeeHandoverHandler := handler.NewEmployeeHandoverHandler(repository.NewEmployeeHandoverRepository(db))
	peopleApp.Set("GET /api/employees/{id}/handover", "people.handover_view", employeeHandoverHandler.Preview)
	peopleApp.Set("POST /api/employees/{id}/handover", "people.handover", employeeHandoverHandler.Handover)
	// (دازاي 10-26) «الصلاحيات والوصول»: مجموعات، مصدر كل صلاحية، حفظ بنسخة، مطابقة الصنف،
	// استثناءات (ومؤقتة)، وسجل — المالك ومدير النظام.
	accessHandler := handler.NewPermissionAccessHandler(db)
	accessApp.Set("GET /api/permissions/catalog", "access.catalog", accessHandler.Catalog)
	peopleApp.Set("GET /api/employees/{id}/access", "people.access_view", accessHandler.Access)
	peopleApp.Set("PUT /api/employees/{id}/access", "people.access_set", accessHandler.Save)
	accessApp.Set("GET /api/permissions/role-drift", "access.role_drift", accessHandler.RoleDrift)
	accessApp.Set("POST /api/permissions/role-drift/{id}/align", "access.role_drift_fix", accessHandler.Align)
	accessApp.Set("POST /api/permissions/exceptions", "access.exception_add", accessHandler.AddException)
	accessApp.Set("DELETE /api/permissions/exceptions", "access.exception_del", accessHandler.RemoveException)
	accessApp.Set("GET /api/permissions/audit-log", "access.audit_log", accessHandler.AuditLog)
	safeguard.Loop("الصلاحيات المؤقتة", 6*time.Minute, time.Hour, func() {
		if n := accessHandler.SweepExpired(); n > 0 {
			log.Printf("permissions: انشالت %d صلاحية مؤقتة خلص وقتها", n)
		}
	})
	accessApp.Set("POST /api/permissions/apply-defaults-all", "access.reset_all", permissionHandler.ApplyDefaultsAll)
	accessApp.Set("POST /api/permissions/employee/{id}/apply-defaults", "access.employee_reset", permissionHandler.ApplyDefaults)

	// الخدمات والمهارات — القراءة لأي مسجل دخول، الإضافة لمدير النظام فقط
	secs["services"].Set("GET /api/services", "services.get_root", serviceHandler.List)
	// المهارات بقائمة مسطّحة — لاختيار مهارات البرنامج التدريبي
	secs["skills"].Set("GET /api/skills", "skills.get_root", serviceHandler.ListSkills)
	secs["services"].Set("POST /api/services", "services.post_root", serviceHandler.Create)
	secs["services"].Set("POST /api/services/{id}/skills", "services.post_by_id_skills", serviceHandler.CreateSkill)
	secs["services"].Set("DELETE /api/services/{id}", "services.delete_by_id", serviceHandler.Delete)
	// ⚠️ إداري بس: هذا يحدد **منو مسؤول عن ورق** خدمة كاملة — قرار
	// تنظيمي مو تفضيل شخصي.
	// ⚠️ `requireAuth` بس ومقصود: الاستعلام محصور بالبناء على خدمات
	// المُستدعي — موظف مو مسؤول خدمة يرجع بقائمة فاضية، مو بخطأ.
	// ⚠️ `requireAuth` بس: يرجّع **مكاناً** مو بيانات. والوصول
	// للتفاصيل يبقى محكوماً بصلاحية الشاشة نفسها — البحث ما يصير
	// باباً خلفياً.
	secs["locate"].Set("GET /api/locate", "locate.get_root", locateHandler.Locate)
	secs["bookings"].Set("GET /api/bookings/manager-paperwork", "bookings.get_manager_paperwork", bookingHandler.ManagerPaperwork)
	secs["services"].Set("PUT /api/services/{id}/manager-paperwork", "services.put_by_id_manager_paperwork", serviceHandler.SetManagerPaperwork)
	// نوع الخدمة (جي بي اس / داش كام) — نفس حارس «الورق على المسؤول»:
	// تأشير كتالوك يخص المالك ومدير النظام.
	secs["services"].Set("PUT /api/services/{id}/kind", "services.put_by_id_kind", serviceHandler.SetKind)
	secs["services"].Set("PUT /api/services/{id}/rules", "services.put_by_id_rules", serviceHandler.SetRules)
	// ⏱ مواعيد الالتزام ويا الزبون (دازاي 11-06): لوحة للمنسّق + تنبيه مرة وحدة لكل حجز
	secs["sla"].Set("GET /api/sla/board", "sla.get_board", handler.SlaBoardHandler(db).ServeHTTP)
	slaAlerter := service.NewSlaAlerter(db, func(kind, msg string) {
		if kind == "OVERDUE" {
			_ = notificationRepo.CreateForRole("ADMIN", "SLA", msg)
			return
		}
		_ = notificationRepo.CreateForPermission("coordinator", "SLA", msg)
	})
	safeguard.Loop("تنبيهات الالتزام", 4*time.Minute, 15*time.Minute, func() {
		if err := slaAlerter.Tick(); err != nil {
			log.Printf("sla: %v", err)
		}
	})

	// العملاء — أي مسجل دخول يقدر يبحث وينشئ عميل (يطابق سلوك المبيعات بالباك إند القديم)
	// ═══ بيانات الزبائن — أسماء وأرقام هواتف وعناوين ═══
	//
	// ⚠️⚠️ چانت `requireAuth` بس: أي موظف مسجّل يسحب دليل زبائن
	// الشركة كامل بنداء واحد. وماكو «نطاق شخصي» للزبائن (بعكس
	// الحجوزات)، فالحارس هنا يرفض.
	//
	// ⚠️ `complaints` و`quality_control` و`monitoring` **إجبارية**
	// بالقائمة: شاشة الشكاوى تستدعي الزبائن وصاحبها عنده
	// `complaints` مو `manage_customers` — بدونهن تنكسر شاشة شغّالة.
	requireCustomerRead := middleware.RequireRoleOrAnyPermission(
		permissionRepo, employeeRepo, notificationRepo,
		[]string{"ADMIN", "OWNER", "HR_COORDINATOR", "MONITOR", "FINANCE", "SALES"}, // SALES: تقرير ١ (B13)
		"manage_customers", "sales_booking", "coordinator",
		"complaints", "quality_control", "monitoring")
	secs["customers"].Set("GET /api/customers", "customers.get_root", customerHandler.List, requireCustomerRead)
	secs["customers"].Set("GET /api/customers/page", "customers.get_page", customerHandler.Page, requireCustomerRead)
	secs["customers"].Set("GET /api/customers/gps", "customers.get_gps", customerHandler.ListGps, requireCustomerRead)
	secs["customers"].Set("GET /api/customers/lookup", "customers.get_lookup", customerHandler.Lookup, requireCustomerRead)
	secs["customers"].Set("GET /api/customers/{id}", "customers.get_by_id", customerHandler.Get, requireCustomerRead)
	secs["customers"].Set("POST /api/customers", "customers.post_root", customerHandler.FindOrCreate)
	secs["customers"].Set("PUT /api/customers/{id}", "customers.put_by_id", customerHandler.Update)

	// الحجوزات — دورة حياة الحجز الكاملة، كل خطوة تتطلب تسجيل دخول فقط (الصلاحية الدقيقة تُفرض بالواجهة حالياً)
	secs["bookings"].Set("GET /api/bookings", "bookings.get_root", bookingHandler.List)
	// حجوزات الشغل داخل الشركة — لمن يسوي فاتورتها أو يدقّقها.
	// ⚠️ لازم تجي **قبل** `GET /api/bookings/{id}` وإلا "internal"
	// تنحسب معرّف حجز.
	// حجوزات خدمتي مرشّحة بالنوع — نفس حارس فاتورة الخدمة بالضبط،
	// حتى ما يصير باب يشوف بيه حجوزات وما يكدر يفوترها.
	secs["bookings"].Set("GET /api/bookings/service-paperwork", "bookings.get_service_paperwork", bookingHandler.ServicePaperworkByKind)
	secs["bookings"].Set("GET /api/bookings/internal", "bookings.get_internal", bookingHandler.ListInternal)
	fieldApp.Set("POST /api/bookings", "field.booking_create", bookingHandler.Create)
	// (ع) 10-16: حجز ديكور مربوط — المنسق/المدير/المراقب
	// (ع) 10-17: السواق (إجازة السوق والتعهد والتخويل) وبطاقات البنزين — إدارة المركبات
	vehicleDriversHandler := handler.NewVehicleDriversHandler(db)
	secs["vehicle_drivers"].Set("GET /api/vehicle-drivers", "vehicle_drivers.get_root", vehicleDriversHandler.List)
	secs["vehicle_drivers"].Set("PUT /api/vehicle-drivers/{employeeId}", "vehicle_drivers.put_by_employeeid", vehicleDriversHandler.Save)
	secs["fuel_cards"].Set("GET /api/fuel-cards", "fuel_cards.get_root", vehicleDriversHandler.Cards)
	secs["fuel_cards"].Set("PUT /api/fuel-cards", "fuel_cards.put_root", vehicleDriversHandler.SaveCard)
	secs["decor_rates"].Set("GET /api/decor-rates", "decor_rates.get_root", decorRateHandler.List)
	secs["decor_rates"].Set("PUT /api/decor-rates", "decor_rates.put_root", decorRateHandler.Save)
	secs["bookings"].Set("POST /api/bookings/{id}/decor", "bookings.post_by_id_decor", bookingHandler.CreateDecor)
	secs["bookings"].Set("GET /api/bookings/{id}/decor", "bookings.get_by_id_decor", bookingHandler.DecorChildren)
	secs["bookings"].Set("PUT /api/bookings/{id}/confirm", "bookings.put_by_id_confirm", bookingHandler.Confirm)
	secs["bookings"].Set("PUT /api/bookings/{id}/details", "bookings.put_by_id_details", bookingHandler.UpdateDetails, requireBookingEdit)
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

	secs["bookings"].Set("PUT /api/bookings/{id}/schedule", "bookings.put_by_id_schedule", bookingHandler.Schedule, requireBookingParty)
	secs["bookings"].Set("GET /api/bookings/{id}/schedule-log", "bookings.get_by_id_schedule_log", bookingHandler.ScheduleLog)

	// ═══ أرشيف الحجوزات وتأجيلها وانتظار رد الزبون ═══
	// الأرشيف للي عنده صلاحية تنسيق أو إدارة — هو سجل قرارات الإلغاء.
	secs["bookings"].Set("GET /api/bookings/archived", "bookings.get_archived", bookingHandler.ListArchived)
	// الحذف صار أرشفة: الحجز يختفي من الشاشات ويضل بالأرشيف بسببه
	secs["bookings"].Set("DELETE /api/bookings/{id}", "bookings.delete_by_id", bookingHandler.ArchiveBooking)
	secs["bookings"].Set("PUT /api/bookings/{id}/restore", "bookings.put_by_id_restore", bookingHandler.RestoreBooking)
	secs["bookings"].Set("PUT /api/bookings/{id}/partial-job-link", "bookings.put_by_id_partial_job_link", bookingHandler.SetPartialJobLink)
	secs["bookings"].Set("PUT /api/bookings/{id}/mark-survey", "bookings.put_by_id_mark_survey", bookingHandler.MarkAsSurveyRetroactively)
	// الكنسة اليدوية للإدارة والفحص — التلقائية تشتغل بالخلفية
	secs["bookings"].Set("POST /api/bookings/waiting-reminder-sweep", "bookings.post_waiting_reminder_sweep", bookingHandler.RunWaitingReminderSweep)
	// باسورد مركز القيادة — المالك بس بهاي المرحلة
	// 🕶️ الدخول المتخفي (ع 10-23): المالك وحده يفتح حساب موظف، وكل كتابة تنسجل بـStealthLog
	stealthLog := func(ownerID, employeeID, method, path string, status int) {
		if _, err := db.Exec(`INSERT INTO "StealthLog" ("ownerId", "employeeId", method, path, status) VALUES ($1, $2, $3, $4, $5)`,
			ownerID, employeeID, method, path, status); err != nil {
			log.Printf("stealth log: %v", err)
		}
	}
	middleware.SetStealthRecorder(stealthLog)
	authHandler.SetStealthLog(stealthLog)
	// 🤖 حالة مفاتيح الذكاء (دازاي 10-23) — شغّال/مطفي بس، المفتاح نفسه ما يطلع أبداً.
	secs["ai"].Set("GET /api/ai/status", "ai.get_status", func(w http.ResponseWriter, _ *http.Request) {
		mSpent, mLimit := service.AIBudgetState(service.AIAgentMatrix)
		iSpent, iLimit := service.AIBudgetState(service.AIAgentIris)
		qSpent, qLimit := service.AIBudgetState(service.WaAgentQuality)
		bSpent, bLimit := service.AIBudgetState(service.WaAgentBooking)
		handler.WriteJSON(w, http.StatusOK, map[string]any{
			"claude":    map[string]any{"on": cfg.AIKeyMatrix != "", "model": cfg.AIModel, "dailyCap": cfg.AIDailyCap},
			"matrix":    map[string]any{"on": cfg.AIKeyMatrix != "", "ownKey": cfg.AIKeyMatrix != cfg.AnthropicAPIKey, "model": cfg.AIModel, "dailyCap": cfg.AIDailyCap, "spentTodayUSD": mSpent, "dailyLimitUSD": mLimit},
			"secretary": map[string]any{"on": cfg.AIKeySecretary != "", "ownKey": cfg.AIKeySecretary != cfg.AnthropicAPIKey, "model": cfg.AIModelSecretary, "dailyCap": cfg.AISecretaryDailyCap, "spentTodayUSD": iSpent, "dailyLimitUSD": iLimit},
			"quality":   map[string]any{"on": cfg.AIKeyQuality != "", "ownKey": cfg.AIKeyQuality != cfg.AnthropicAPIKey, "built": true, "model": cfg.AIModel, "spentTodayUSD": qSpent, "dailyLimitUSD": qLimit},
			"booking":   map[string]any{"on": cfg.AIKeyBooking != "", "ownKey": cfg.AIKeyBooking != cfg.AnthropicAPIKey, "built": true, "model": cfg.AIModel, "spentTodayUSD": bSpent, "dailyLimitUSD": bLimit},
			"whatsapp":  map[string]any{"on": cfg.WaToken != "" && cfg.WaPhoneID != "" && cfg.WaAppSecret != ""},
			"brains":    aiBrainsStatus(cfg.AIModel, cfg.AIModelChat),
			"gemini":    map[string]any{"on": cfg.GeminiAPIKey != "", "dailyCap": cfg.GeminiDailyCap},
			"whisper":   map[string]any{"on": cfg.WhisperURL != ""},
		})
	})
	aiAgentNames := handler.NewAiAgentNameHandler(db)
	secs["ai"].Set("GET /api/ai/agents", "ai.get_agents", aiAgentNames.List) // قراءة للأدمن، التعديل للمالك
	// 🧠 الدماغين (دازاي 10-28) — docs/ai-brains-contract.md
	aiBrains := handler.NewAiBrainHandler(db)
	aiBrains.LoadBudgets(map[string]float64{service.AIBrainInternal: cfg.AIInternalDailyUSD, service.AIBrainExternal: cfg.AIExternalDailyUSD})
	secs["ai"].Set("PUT /api/ai/budgets/{brain}", "ai.put_budgets_by_brain", aiBrains.SetBudget)
	secs["ai"].Set("GET /api/ai/suggestions", "ai.get_suggestions", aiBrains.Suggestions)
	secs["ai"].Set("POST /api/ai/suggestions/{id}/decide", "ai.post_suggestions_by_id_decide", aiBrains.Decide)
	secs["ai"].Set("GET /api/ai/agent-names", "ai.get_agent_names", aiAgentNames.Names)
	secs["ai"].Set("PUT /api/ai/agents/{agent}", "ai.put_agents_by_agent", aiAgentNames.Rename)
	secs["auth"].Set("POST /api/auth/stealth", "auth.post_stealth", authHandler.Stealth)
	secs["auth"].Set("PUT /api/auth/command-password", "auth.put_command_password", authHandler.SetCommandPassword)
	secs["auth"].Set("GET /api/auth/command-accounts", "auth.get_command_accounts", authHandler.CommandAccounts)
	secs["auth"].Set("DELETE /api/auth/command-password/{employeeId}", "auth.delete_command_password_by_employeeid", authHandler.ClearCommandPassword)
	// المؤجلة بلا موعد: قائمة مستقلة لأنها منزاحة عن جدول اليوم قصداً
	secs["bookings"].Set("GET /api/bookings/postponed", "bookings.get_postponed", bookingHandler.ListPostponed)
	// التأجيل والانتظار شغل المنسّق — هو الي يتصل بالزبون
	secs["bookings"].Set("PUT /api/bookings/{id}/postpone", "bookings.put_by_id_postpone", bookingHandler.Postpone)
	secs["bookings"].Set("PUT /api/bookings/{id}/waiting", "bookings.put_by_id_waiting", bookingHandler.MarkWaiting)
	// ⚠️ للمالك وحده: هذا قفل حجز بلا تفاصيل — يتفتح لغيره يصير باب
	// خلفي يقفل بيه أي واحد شغله بضغطة بدل ما يوثّقه.
	// نفس حارس `GET /api/bookings` — هاي نفس البيانات بصفحات، فما يجوز
	// تكون أضيق ولا أوسع منه.
	// لوحة اليوم — كل أرقامها بنداء واحد
	todayHandler := handler.NewTodayHandler(repository.NewTodayRepository(db))
	secs["dashboard"].Set("GET /api/dashboard/today", "dashboard.get_today", todayHandler.Board)
	// «الرقم = القائمة» (ع 10-16): قائمة كل بطاقة باللوحة. بيها أسماء زبائن
	// وأرقامهم، فمحصورة بالي تطلعله «لوحة اليوم» أصلاً — ⚠️ نفس شرط الواجهة
	// بـDashboard.tsx (canDrill بـutils/drill.ts) حرفياً، وإلا البطاقة تنضغط
	// وتاخذ ٤٠٣ = مخالفة على موظف بريء.
	secs["drill"].Set("GET /api/drill/{key}", "drill.get_by_key", todayHandler.Drill)
	secs["bookings"].Set("GET /api/bookings/locate", "bookings.get_locate", bookingHandler.Locate)
	secs["bookings"].Set("GET /api/bookings/station-counts", "bookings.get_station_counts", bookingHandler.StationCounts)
	secs["bookings"].Set("GET /api/bookings/paged", "bookings.get_paged", bookingHandler.Paged)
	secs["bookings"].Set("POST /api/bookings/{id}/settle-legacy", "bookings.post_by_id_settle_legacy", bookingHandler.SettleLegacy)
	secs["bookings"].Set("PUT /api/bookings/{id}/resume", "bookings.put_by_id_resume", bookingHandler.ResumeFromWaiting)
	// رسول ماتركس (ع 10-13): المالك/المدير يوجّه شغل → العين تسلّمه للموظف
	courier := handler.NewMatrixCourierHandler(db)
	secs["matrix_courier"].Set("GET /api/matrix-courier/state", "matrix_courier.get_state", courier.State)
	secs["matrix_courier"].Set("GET /api/matrix-courier/inbox", "matrix_courier.get_inbox", courier.Inbox)
	meApp.Set("POST /api/matrix-courier/seen", "me.courier_seen", courier.Seen)
	meApp.Set("POST /api/matrix-courier/{id}/delivered", "me.courier_delivered", courier.Delivered)
	secs["bookings"].Set("PUT /api/bookings/{id}/assign", "bookings.put_by_id_assign", bookingHandler.Assign, courier.Wrap("booking"))
	// «كشف»: الإداري يطلّع ليدر يعاين الموقع قبل ما يصير مشروع.
	secs["bookings"].Set("PUT /api/bookings/{id}/send-to-survey", "bookings.put_by_id_send_to_survey", bookingHandler.SendToSurvey)
	// ترحيل للتقني (قرار (ع) 10-07): الإداري يرحّل، والتقني المرحّل إله بس يكمّل الخطوات.
	secs["bookings"].Set("GET /api/bookings/handover-candidates", "bookings.get_handover_candidates", bookingHandler.HandoverCandidates)
	secs["bookings"].Set("PUT /api/bookings/{id}/handover", "bookings.put_by_id_handover", bookingHandler.Handover)
	fieldApp.Set("PUT /api/bookings/{id}/tech/{step}", "field.booking_tech_step", bookingHandler.TechStep)
	// إلغاء تكليف موظف — نفس حارس التكليف: الي يكدر يكلّف يكدر يشيل.
	// حارس أشد يعني الإداري يكلّف بالغلط وينتظر المدير حتى يصلّحها.
	secs["bookings"].Set("DELETE /api/bookings/{id}/assign", "bookings.delete_by_id_assign", bookingHandler.Unassign)
	// ملاحظات موجّهة: وحدة للكادر ووحدة لمدير المشاريع.
	// ═══ المهام الإضافية ═══
	// التوجيه والمتابعة للمدير؛ التنفيذ للموظف نفسه.
	// ⚠️ مسارات التنفيذ requireAuth بس — الحارس ما يعرف صاحب المهمة،
	// فالمستودع يفحص "assignedToId" بكل عملية. بدونه أي موظف ينهي
	// مهمة موظف ثاني بمعرّفها.
	secs["extra_tasks"].Set("POST /api/extra-tasks", "extra_tasks.post_root", extraTaskHandler.Create, courier.Wrap("extra_task"))
	// مهمة بالفويس: مسودة بس — التوجيه الفعلي يبقى من POST /api/extra-tasks.
	secs["extra_tasks"].Set("POST /api/extra-tasks/voice-draft", "extra_tasks.post_voice_draft", handler.NewVoiceTaskHandler(voiceAnalysisService, employeeRepo).Draft)
	secs["extra_tasks"].Set("GET /api/extra-tasks", "extra_tasks.get_root", extraTaskHandler.List)
	secs["extra_tasks"].Set("PUT /api/extra-tasks/{id}/cancel", "extra_tasks.put_by_id_cancel", extraTaskHandler.Cancel)
	secs["extra_tasks"].Set("GET /api/extra-tasks/mine", "extra_tasks.get_mine", extraTaskHandler.Mine)
	secs["extra_tasks"].Set("GET /api/extra-tasks/mine/count", "extra_tasks.get_mine_count", extraTaskHandler.MyOpenCount)
	meApp.Set("PUT /api/extra-tasks/{id}/seen", "me.extra_task_seen", extraTaskHandler.MarkSeen)
	meApp.Set("PUT /api/extra-tasks/{id}/start", "me.extra_task_start", extraTaskHandler.Start)
	meApp.Set("PUT /api/extra-tasks/{id}/complete", "me.extra_task_complete", extraTaskHandler.Complete)

	// ═══ الذكاء الاصطناعي ═══
	// ⚠️ كلها requireAdmin (يشمل المالك): «التقرير يطلع فقط لمدير
	// النظام والمالك» — طلب صريح. تحليل سلوك موظف بيد زميله يتحول
	// لسلاح داخلي.
	secs["ai"].Set("GET /api/ai/signals", "ai.get_signals", aiHandler.ListSignals)
	secs["ai"].Set("POST /api/ai/process", "ai.post_process", aiHandler.Process)
	secs["ai"].Set("POST /api/ai/metrics/recompute", "ai.post_metrics_recompute", aiHandler.RecomputeMetrics)
	secs["ai"].Set("GET /api/ai/metrics", "ai.get_metrics", aiHandler.Metrics)
	// توقع التأخير: نفس من يشوف «تنسيق الحجوزات» بالقائمة (أدوار ADMIN/HR_COORDINATOR/MONITOR أو صلاحية coordinator).
	secs["ai"].Set("GET /api/ai/delay-risks", "ai.get_delay_risks", matrixForesightHandler.DelayRisks)
	// فرص البيع: المالك/مدير النظام أو صلاحية sales_booking (المبيعات بالافتراض).
	secs["ai"].Set("GET /api/ai/opportunities", "ai.get_opportunities", matrixForesightHandler.Opportunities)
	// ═══ ماتركس — ميزات ٣/٧/١٠/١١/١٢: كلها GET ═══
	// الكادر الأنسب: نفس حارس PUT /api/bookings/{id}/assign بالضبط.
	secs["ai"].Set("GET /api/ai/crew-recommendation", "ai.get_crew_recommendation", matrixInsightsHandler.CrewRecommendation)
	secs["ai"].Set("GET /api/ai/new-employee-curves", "ai.get_new_employee_curves", matrixInsightsHandler.NewEmployeeCurves)
	// المواد: نفس حارس شاشة الجرد (صلاحية inventory؛ ADMIN/OWNER يمرّون).
	secs["materials"].Set("PUT /api/materials/{id}/stock", "materials.put_by_id_stock", matrixInsightsHandler.SetMaterialStock)
	secs["ai"].Set("GET /api/ai/stock-forecast", "ai.get_stock_forecast", matrixInsightsHandler.StockForecast)
	// الاستبدال: ADMIN/OWNER أو it_assets أو vehicle_management — والمعالج يفلتر الأقسام.
	secs["ai"].Set("GET /api/ai/replacement-suggestions", "ai.get_replacement_suggestions", matrixInsightsHandler.ReplacementSuggestions)
	secs["ai"].Set("GET /api/ai/my-watch", "ai.get_my_watch", matrixDecisionsHandler.MyWatch)
	secs["ai"].Set("GET /api/ai/guide", "ai.get_guide", matrixGuideHandler.Guide)
	secs["ai"].Set("GET /api/ai/guide-rules", "ai.get_guide_rules", matrixGuideHandler.List)
	secs["ai"].Set("POST /api/ai/guide-rules", "ai.post_guide_rules", matrixGuideHandler.Create)
	secs["ai"].Set("PUT /api/ai/guide-rules/{id}", "ai.put_guide_rules_by_id", matrixGuideHandler.Update)
	secs["ai"].Set("DELETE /api/ai/guide-rules/{id}", "ai.delete_guide_rules_by_id", matrixGuideHandler.Delete)
	secs["ai"].Set("GET /api/ai/proposals", "ai.get_proposals", matrixLearningHandler.List)
	secs["ai"].Set("POST /api/ai/proposals/{id}/approve", "ai.post_proposals_by_id_approve", matrixLearningHandler.Approve)
	secs["ai"].Set("POST /api/ai/proposals/{id}/reject", "ai.post_proposals_by_id_reject", matrixLearningHandler.Reject)
	// المراقب يشوف تقرير الموظف (قراءة بس) — شغله المتابعة، ومن صندوقه يضغط اسم الموظف.
	secs["ai"].Set("GET /api/ai/employee-report", "ai.get_employee_report", matrixReportHandler.Employee)
	// (ع) 10-16: «شنو سوّى هذا الموظف» — اليوم/الشهر بالتفصيل، من عيون ماتركس
	employeeLedgerHandler := handler.NewEmployeeLedgerHandler(repository.NewEmployeeLedgerRepository(db))
	secs["ai"].Set("GET /api/ai/employee-ledger", "ai.get_employee_ledger", employeeLedgerHandler.Get)
	// 📝 بانتظار قرار الزبون +٥ أيام ← الأرشيف «استفسار فقط» (ع 10-23)
	safeguard.Loop("أرشفة الاستفسارات", 3*time.Minute, time.Hour, func() {
		if n, err := bookingRepo.ArchiveStaleInquiries(); err != nil {
			log.Printf("stale inquiries: %v", err)
		} else if n > 0 {
			log.Printf("stale inquiries: %d حجز للأرشيف", n)
		}
	})
	dutiesRepo := repository.NewDutiesRepository(db)
	dutiesHandler := handler.NewDutiesHandler(dutiesRepo)
	// 📸 لقطة الواجبات اليومية (ع 10-22): كل ليلة ١١:٥٠ بغداد — تاريخ يحكم عليه ماتركس
	safeguard.Loop("لقطة الواجبات اليومية", 2*time.Minute, 5*time.Minute, func() {
		if n, err := dutiesRepo.SnapshotIfDue(time.Now()); err != nil {
			log.Printf("duty snapshot: %v", err)
		} else if n > 0 {
			log.Printf("duty snapshot: %d موظف", n)
		}
	})
	secs["duties"].Set("GET /api/duties/me", "duties.get_me", dutiesHandler.Me)
	secs["duties"].Set("GET /api/duties/{employeeId}", "duties.get_by_employeeid", dutiesHandler.ForEmployee)
	secs["duties"].Set("GET /api/duties/history/{employeeId}", "duties.get_history_by_employeeid", dutiesHandler.History)
	secs["ai"].Set("GET /api/ai/group-performance", "ai.get_group_performance", matrixReportHandler.Group)
	secs["ai"].Set("GET /api/ai/feed", "ai.get_feed", matrixCommandHandler.Feed)
	secs["ai"].Set("GET /api/ai/daily-trend", "ai.get_daily_trend", matrixCommandHandler.Trend)
	secs["ai"].Set("GET /api/ai/late-focus", "ai.get_late_focus", matrixCommandHandler.LateFocus)
	secs["ai"].Set("POST /api/ai/ask", "ai.post_ask", matrixCommandHandler.Ask)
	secs["ai"].Set("GET /api/ai/business", "ai.get_business", matrixReportHandler.Business)
	// (ع) 10-19: «استقرارية النظام… لازم تطلع عد المراقب» — نفس حارس تقرير الموظف.
	secs["ai"].Set("GET /api/ai/role-watch", "ai.get_role_watch", matrixDecisionsHandler.RoleWatch)
	secs["ai"].Set("GET /api/ai/decisions", "ai.get_decisions", matrixDecisionsHandler.Get)
	secs["ai"].Set("POST /api/ai/actions/kinds/{kind}/resume", "ai.post_actions_kinds_by_kind_resume", matrixDecisionsHandler.Resume)
	secs["ai"].Set("POST /api/ai/actions/{id}/undo", "ai.post_actions_by_id_undo", matrixDecisionsHandler.Undo)
	secs["ai"].Set("GET /api/ai/weekly-report", "ai.get_weekly_report", matrixInsightsHandler.WeeklyReport)
	secs["ai"].Set("GET /api/ai/catalog", "ai.get_catalog", aiHandler.Catalog)
	secs["ai"].Set("GET /api/ai/work-window", "ai.get_work_window", aiHandler.GetWorkWindow)
	secs["ai"].Set("PUT /api/ai/work-window", "ai.put_work_window", aiHandler.SetWorkWindow)
	// الإنجازات: أي موظف يرفع تقريره — العرض والمراجعة حصراً لمدير
	// النظام والمالك (requireAdmin)، لا المراقب.
	meApp.Set("POST /api/achievements", "me.achievement", achievementHandler.Create)
	// ساعات العمل من البيت — صلاحية «remote_hours_manage».
	remoteHoursHandler := handler.NewRemoteHoursHandler(repository.NewRemoteHoursRepository(db))
	remoteHoursHandler.SetAgent(cfg.RemoteHoursAgentToken, cfg.RemoteHoursAgentEmployee) // 🤖 ساعات Claude (قرار 0011)
	mux.Handle("POST /api/remote-hours/agent", middleware.RateLimit(30, time.Minute)(http.HandlerFunc(remoteHoursHandler.AgentAdd)))
	// (ون بيس 11-10) requireRemoteHours انتقل لقاعدة بالنواة (internal/apps/sections)
	secs["remote_hours"].Set("POST /api/remote-hours", "remote_hours.post_root", remoteHoursHandler.Create)
	secs["remote_hours"].Set("GET /api/remote-hours", "remote_hours.get_root", remoteHoursHandler.List)
	secs["remote_hours"].Set("GET /api/remote-hours/history", "remote_hours.get_history", remoteHoursHandler.History)
	secs["remote_hours"].Set("GET /api/remote-hours/summary", "remote_hours.get_summary", remoteHoursHandler.Summary)
	secs["remote_hours"].Set("GET /api/remote-hours/export", "remote_hours.get_export", remoteHoursHandler.Export)
	// الموظف لنفسه: صلاحية remote_hours_self (أو المسؤول).
	// (ون بيس 11-10) requireRemoteSelf انتقل لقاعدة بالنواة (internal/apps/sections)
	secs["remote_hours"].Set("GET /api/remote-hours/mine", "remote_hours.get_mine", remoteHoursHandler.Mine)
	secs["remote_hours"].Set("POST /api/remote-hours/mine", "remote_hours.post_mine", remoteHoursHandler.AddMine)
	secs["remote_hours"].Set("POST /api/remote-hours/timer/{action}", "remote_hours.post_timer_by_action", remoteHoursHandler.TimerAction)
	// (دازاي 11-01) المطلوب بالشهر وسعر الساعة — الموظف مرة وحدة، والإدارة تعدّل
	secs["remote_hours"].Set("PUT /api/remote-hours/plan", "remote_hours.put_plan", remoteHoursHandler.SetPlan(func(r *http.Request) bool {
		role := middleware.RoleFromContext(r)
		if model.IsAdminRole(role) {
			return true
		}
		ok, _ := permissionRepo.HasPermission(middleware.EmployeeIDFromContext(r), "remote_hours_manage")
		return ok
	}).ServeHTTP)
	secs["remote_hours"].Set("POST /api/remote-hours/reset", "remote_hours.post_reset", remoteHoursHandler.Reset)
	// الحذف: للي سجّلها أو للمدير (يتفحص بالهاندلر) — للمسؤول والموظف لنفسه.
	secs["remote_hours"].Set("DELETE /api/remote-hours/{id}", "remote_hours.delete_by_id", remoteHoursHandler.Delete)
	secs["remote_hours"].Set("PUT /api/remote-hours/{id}", "remote_hours.put_by_id", remoteHoursHandler.Update)
	secs["achievements"].Set("GET /api/achievements/mine", "achievements.get_mine", achievementHandler.Mine)
	// فويس مسج للإنجاز — يبقى بتخزيننا؛ صاحبه يرفع، وصاحبه أو المدير يسمع.
	meApp.Set("POST /api/achievements/{id}/voice", "me.achievement_voice", achievementHandler.UploadVoice)
	secs["achievements"].Set("GET /api/achievements/{id}/voice", "achievements.get_by_id_voice", achievementHandler.Voice)
	secs["achievements"].Set("GET /api/achievements/{id}/voice/analysis", "achievements.get_by_id_voice_analysis", achievementHandler.VoiceAnalysis)
	secs["ai"].Set("GET /api/ai/dialect", "ai.get_dialect", achievementHandler.Dialect)
	secs["achievements"].Set("GET /api/achievements", "achievements.get_root", achievementHandler.List)
	secs["achievements"].Set("PUT /api/achievements/{id}/review", "achievements.put_by_id_review", achievementHandler.Review)
	secs["bookings"].Set("GET /api/bookings/{id}/timeline", "bookings.get_by_id_timeline", bookingHandler.Timeline)
	secs["bookings"].Set("PUT /api/bookings/{id}/crew-notes", "bookings.put_by_id_crew_notes", bookingHandler.SetCrewNotes)
	secs["bookings"].Set("PUT /api/bookings/{id}/project-notes", "bookings.put_by_id_project_notes", bookingHandler.SetProjectNotes)
	// إلغاء بسبب مكتوب — بوقته ومنو ألغى، حتى نفرّق قبل التثبيت وبعده.
	secs["bookings"].Set("PUT /api/bookings/{id}/cancel", "bookings.put_by_id_cancel", bookingHandler.Cancel)
	// سلال المراحل الست
	secs["bookings"].Set("GET /api/bookings/stage-bucket", "bookings.get_stage_bucket", bookingHandler.ListByStageBucket)
	secs["bookings"].Set("GET /api/bookings/stage-bucket-counts", "bookings.get_stage_bucket_counts", bookingHandler.StageBucketCounts)
	secs["bookings"].Set("PUT /api/bookings/{id}/supervisor", "bookings.put_by_id_supervisor", bookingHandler.Supervisor)
	// (ع) 10-14: وقت المرحلة الحقيقي + موقعها (يشتغل حتى لو انضغطت بلا نت)
	stageProofHandler := handler.NewStageProofHandler(db)
	secs["bookings"].Set("POST /api/bookings/{id}/stage-proof", "bookings.post_by_id_stage_proof", stageProofHandler.Record, requireBookingParty)
	// (دازاي 10-25) صور «قبل/بعد» لكادر الحجز — الديكور يصوّر بدل ما يكتب
	workPhotoHandler := handler.NewWorkPhotoHandler(db, fileStore)
	secs["bookings"].Set("GET /api/bookings/{id}/work-photos", "bookings.get_by_id_work_photos", workPhotoHandler.List, requireBookingParty)
	secs["bookings"].Set("POST /api/bookings/{id}/work-photos", "bookings.post_by_id_work_photos", workPhotoHandler.Upload, requireBookingParty)
	secs["bookings"].Set("GET /api/bookings/{id}/work-photos/{photoId}/file", "bookings.get_by_id_work_photos_by_photoid_file", workPhotoHandler.File, requireBookingParty)
	// (ع) 10-14: الميدان — موقع الشركة · استلام · انطلاق · نقاط المسار · خريطة المراقبة
	fieldTrackHandler := handler.NewFieldTrackHandler(db)
	fieldTrackHandler.SetNotifier(func(emp, msg string) { _ = notificationRepo.Create(emp, "stage_reminder", msg) })
	go fieldTrackHandler.RunLoop()
	secs["company_location"].Set("GET /api/company-location", "company_location.get_root", fieldTrackHandler.GetCompany)
	secs["company_location"].Set("PUT /api/company-location", "company_location.put_root", fieldTrackHandler.SetCompany)
	secs["bookings"].Set("PUT /api/bookings/{id}/accept", "bookings.put_by_id_accept", fieldTrackHandler.Accept, requireBookingParty)
	secs["bookings"].Set("PUT /api/bookings/{id}/depart", "bookings.put_by_id_depart", fieldTrackHandler.Depart, requireBookingParty)
	secs["location_pings"].Set("POST /api/location-pings/batch", "location_pings.post_batch", fieldTrackHandler.PingBatch, middleware.SkipInStealth())
	secs["bookings"].Set("GET /api/bookings/{id}/track", "bookings.get_by_id_track", fieldTrackHandler.Track)
	// (ع) 10-15: فحص المتأخرات + فحص دقة ماتركس
	lateAudit := handler.NewLateAuditHandler(db)
	secs["late_audit"].Set("GET /api/late-audit", "late_audit.get_root", lateAudit.List)
	selfCheck := handler.NewMatrixSelfCheckHandler(db, func(msg string) {
		_ = notificationRepo.CreateForRolesOrPermission([]string{"ADMIN", "OWNER"}, "matrix_selfcheck", "matrix_selfcheck", msg)
	})
	go selfCheck.RunDaily()
	secs["matrix_selfcheck"].Set("GET /api/matrix-selfcheck", "matrix_selfcheck.get_root", selfCheck.Get)
	secs["field_map"].Set("GET /api/field-map", "field_map.get_root", fieldTrackHandler.FieldMap)
	// 🧭 ترتيب مسار اليوم (دازاي 11-06): الموظف لنفسه، والمشرف لأي موظف (الفحص بالـhandler)
	secs["route_plan"].Set("GET /api/route-plan", "route_plan.get_root", handler.NewRoutePlanHandler(db).Plan)
	secs["bookings"].Set("PUT /api/bookings/{id}/start", "bookings.put_by_id_start", bookingHandler.Start, requireBookingParty)
	secs["bookings"].Set("PUT /api/bookings/{id}/arrived", "bookings.put_by_id_arrived", bookingHandler.MarkArrived, requireBookingParty)
	secs["bookings"].Set("PUT /api/bookings/{id}/materials-ready", "bookings.put_by_id_materials_ready", bookingHandler.SetMaterialsReady, requireBookingParty)
	// توقف العمل ورجوعه — بيد طرف الحجز نفسه (الليدر/الكادر المكلّف)
	secs["bookings"].Set("PUT /api/bookings/{id}/stop-work", "bookings.put_by_id_stop_work", bookingHandler.StopWork, requireBookingParty)
	secs["bookings"].Set("PUT /api/bookings/{id}/resume-work", "bookings.put_by_id_resume_work", bookingHandler.ResumeWork, requireBookingParty)
	// ── الإنجاز الجزئي ──
	// نفس صلاحية «تم الإنجاز» بالضبط: الي يقدر يقفل الحجز يقدر يأشّره
	// جزئي. لو ضيّقناها أكثر، الليدر يوكف بلا خيار آخر اليوم ويأشّر
	// «تم الإنجاز» على شغل ناقص — وهاي المشكلة الي نحلها أصلاً.
	secs["bookings"].Set("POST /api/bookings/{id}/partial-complete", "bookings.post_by_id_partial_complete", bookingProgressHandler.PartialComplete, requireBookingParty)
	secs["bookings"].Set("POST /api/bookings/{id}/schedule-continuation", "bookings.post_by_id_schedule_continuation", bookingProgressHandler.ScheduleContinuation)
	bookingProgressHandler.SetViewGate(func(r *http.Request, id string) bool {
		return bookingHandler.CanSeeAllBookings(r) || bookingRepo.IsInvolved(id, middleware.EmployeeIDFromContext(r))
	})
	secs["progress_reports"].Set("GET /api/progress-reports/recent", "progress_reports.get_recent", bookingProgressHandler.Recent)
	secs["bookings"].Set("GET /api/bookings/{id}/progress", "bookings.get_by_id_progress", bookingProgressHandler.Reports)
	secs["bookings"].Set("GET /api/bookings/{id}/work-window", "bookings.get_by_id_work_window", bookingProgressHandler.WorkWindow)
	// ⚠️ الطلعات ما إلها حارس خاص: منو يشوف الحجز يشوف طلعاته. إخفاء
	// «منو طلع ومتى» عن الي يشوف الحجز نفسه ما إله معنى.
	secs["bookings"].Set("GET /api/bookings/{id}/visits", "bookings.get_by_id_visits", bookingProgressHandler.Visits)
	secs["bookings"].Set("GET /api/bookings/{id}/suggested-crew", "bookings.get_by_id_suggested_crew", bookingProgressHandler.SuggestedCrew)

	// تغيير نوع الحجز: المالك ومدير النظام بس. النوع يأثر على الإحصاءات
	// والعمولات وحساب الصيانة — فمو قرار إداري يومي.
	secs["bookings"].Set("PUT /api/bookings/{id}/type", "bookings.put_by_id_type", bookingHandler.ChangeType)
	secs["bookings"].Set("PUT /api/bookings/{id}/complete", "bookings.put_by_id_complete", bookingHandler.Complete, requireBookingParty)

	// حذف الحجوزات التجريبية والملغاة: الإداري يطلب، والمراقب أو مدير
	// النظام يبت. الحذف الفوري ما ينفع لأنه ما يترد.
	bookingDeleteRepo := repository.NewBookingDeleteRequestRepository(db)
	bookingDeleteHandler := handler.NewBookingDeleteRequestHandler(bookingDeleteRepo, notificationRepo)
	// (ون بيس 11-10) requireDeleteRequest انتقل لقاعدة بالنواة (internal/apps/sections)
	// (ون بيس 11-10) requireDeleteApprove انتقل لقاعدة بالنواة (internal/apps/sections)
	secs["bookings"].Set("POST /api/bookings/{id}/delete-request", "bookings.post_by_id_delete_request", bookingDeleteHandler.Create)
	secs["booking_delete_requests"].Set("GET /api/booking-delete-requests", "booking_delete_requests.get_root", bookingDeleteHandler.List)
	secs["booking_delete_requests"].Set("PUT /api/booking-delete-requests/{id}/decide", "booking_delete_requests.put_by_id_decide", bookingDeleteHandler.Decide)
	secs["booking_delete_requests"].Set("PUT /api/booking-delete-requests/{id}/needs-info", "booking_delete_requests.put_by_id_needs_info", bookingDeleteHandler.NeedsInfo)
	secs["booking_delete_requests"].Set("GET /api/booking-delete-requests/counts", "booking_delete_requests.get_counts", bookingDeleteHandler.Counts)

	// تنسيق الحجوزات: المراقب يسجّل تقصير الإداري بتثبيت الحجز، وبالعاشر
	// ينشر إعلان يسمّي الإداري والحجز — نفس حارس تبويب التنسيق (crew_management).
	coordinationAlertRepo := repository.NewCoordinationAlertRepository(db)
	coordinationAlertHandler := handler.NewCoordinationAlertHandler(coordinationAlertRepo, bookingRepo, disciplineRepo, announcementRepo, notificationRepo, employeeRepo)
	coordinationAlertHandler.Kpi = kpiRepo
	secs["coordination_alerts"].Set("GET /api/coordination-alerts/shifts", "coordination_alerts.get_shifts", coordinationAlertHandler.Shifts)
	secs["bookings"].Set("POST /api/bookings/{id}/coordination-alerts", "bookings.post_by_id_coordination_alerts", coordinationAlertHandler.Add)
	secs["bookings"].Set("PUT /api/bookings/{id}/coordination-alerts/resolve", "bookings.put_by_id_coordination_alerts_resolve", coordinationAlertHandler.Resolve)
	secs["bookings"].Set("GET /api/bookings/{id}/coordination-alerts", "bookings.get_by_id_coordination_alerts", coordinationAlertHandler.ListForBooking)
	secs["coordination_alerts"].Set("GET /api/coordination-alerts/summaries", "coordination_alerts.get_summaries", coordinationAlertHandler.Summaries)

	// التدقيق: المحاسب ما يقدر يمرر حجز بلا مبلغ — إما يكتب المبلغ من
	// الفاتورة، أو يأشر خطأ والنظام يوجّهه للمعني (غير مطابق → رقابة
	// وجودة، خطأ سعر → رقابة وإداري).
	bookingAuditRepo := repository.NewBookingAuditRepository(db)
	bookingAuditHandler := handler.NewBookingAuditHandler(bookingAuditRepo, bookingRepo, notificationRepo, leaderInvoiceRepo, disciplineRepo, employeeRepo)
	// مفتاح التدقيق: لمّا المالك ينطي المراقب `finance_audit`، المعالج
	// لازم يعرف — وإلا رفضه للمراقب يبقى بالدور حصراً والمنح ما ينفع.
	bookingAuditHandler.SetPermissions(permissionRepo)
	leaderInvoiceService.SetPriceLearning(servicePriceLearningRepo)
	// التدقيق اليومي: نفس واجهة التدقيق بس بيوم واحد، مع مجاميع اليوم
	dailyAuditHandler := handler.NewDailyAuditHandler(repository.NewDailyAuditRepository(db))
	// (ع) 10-14: سجل الآجل — نفس حارس التدقيق
	creditHandler := handler.NewCreditHandler(db)
	financeApp.Set("GET /api/credits", "finance.credits", creditHandler.List)
	financeApp.Set("GET /api/credits/export", "finance.credits_export", creditHandler.Export)
	financeApp.Set("POST /api/credits", "finance.credit_add", creditHandler.Create, idem)
	financeApp.Set("GET /api/credits/{id}/payments", "finance.credit_payments", creditHandler.Payments)
	financeApp.Set("POST /api/credits/{id}/payments", "finance.credit_pay", creditHandler.AddPayment, idem)
	financeApp.Set("GET /api/finance/daily-audit", "finance.daily_audit", dailyAuditHandler.Day)
	secs["bookings"].Set("PUT /api/bookings/{id}/audit", "bookings.put_by_id_audit", bookingAuditHandler.Audit)
	// شريط الإعلانات: يقراه كل موظف، وينزّله المالك ومدير النظام بس
	announcementHandler := handler.NewAnnouncementHandler(announcementRepo)
	secs["announcements"].Set("GET /api/announcements", "announcements.get_root", announcementHandler.List)
	secs["announcements"].Set("POST /api/announcements", "announcements.post_root", announcementHandler.Create)
	secs["announcements"].Set("PUT /api/announcements/{id}/active", "announcements.put_by_id_active", announcementHandler.SetActive)
	secs["announcements"].Set("DELETE /api/announcements/{id}", "announcements.delete_by_id", announcementHandler.Delete)
	// ⚠️⚠️ قبل هذا السطر ماكو حارس صلاحية إطلاقاً هنا — أي موظف مسجّل
	// دخول (حتى فني بالميدان) يقدر يستدعي المسار ويشوف كل البلاغات،
	// لأن `ListIssues` تفلتر حسب الدور بس (جودة/كوادر/محاسب) وتترك
	// أي دور ثاني يمرّ بلا تصفية = يشوف الكل. نفس الفحص المستخدم
	// لبقية شاشات المراقبة، وبنفس منفذ monitoring/auditing.
	secs["audit_issues"].Set("GET /api/audit-issues", "audit_issues.get_root", bookingAuditHandler.ListIssues)
	// ⚠️⚠️ جان حارسه `requireAuth` **وبس** — أي موظف مسجّل دخول، حتى
	// فني بالميدان، يكدر يحلّ بلاغ خطأ تدقيق ويغلقه. نفس حارس القراءة
	// فوق بالضبط.
	secs["audit_issues"].Set("PUT /api/audit-issues/{id}/resolve", "audit_issues.put_by_id_resolve", bookingAuditHandler.ResolveIssue)
	secs["bookings"].Set("PUT /api/bookings/{id}/verify", "bookings.put_by_id_verify", bookingHandler.Verify)
	// إرجاع الحجز للتدقيق: التدقيق جان قرار نهائي ما إله رجعة. مدير
	// النظام حصراً يكدر يفتحه من جديد حتى ينصلّح أي غلط بالمبلغ.
	secs["bookings"].Set("PUT /api/bookings/{id}/unverify", "bookings.put_by_id_unverify", bookingHandler.Unverify)
	// "تم" الإداري بعد تواصله فعلياً مع الزبون — خطوة سابقة ومنفصلة عن التثبيت
	// نفسه (نفس صلاحية تنسيق الحجوزات coordinator المستخدمة أصلاً بـCoordinator.tsx).
	// إرجاع حجز محوّل لإدارة المشاريع رجعة لكادر الشد — لمدير المشاريع لما
	// يفتح التفاصيل ويلكاه مو مال مشروع.
	// حجز «منجز» بالغلط يرجع للكادر حتى يسجّل إنجاز جزئي — صلاحية booking_reopen.
	bookingReopenHandler := handler.NewBookingReopenHandler(repository.NewBookingReopenRepository(db), employeeRepo, notificationRepo)
	secs["bookings"].Set("PUT /api/bookings/{id}/reopen", "bookings.put_by_id_reopen", bookingReopenHandler.Reopen)
	// (ع) 10-19: «بدء العمل» انضغط بالغلط على حجز باچر — المالك يتراجع عنه.
	secs["bookings"].Set("PUT /api/bookings/{id}/unstart", "bookings.put_by_id_unstart", bookingReopenHandler.Unstart)
	secs["bookings"].Set("PUT /api/bookings/{id}/return-to-crew", "bookings.put_by_id_return_to_crew", bookingHandler.ReturnToCrew)
	// ↩️ «رجعت من المشاريع» (ع 10-23): المنسّق يشوفها ويستلمها — نفس حارس «تنسيق الحجوزات» بالواجهة
	// (ون بيس 11-10) requireReturnedDesk انتقل لقاعدة بالنواة (internal/apps/sections)
	secs["bookings"].Set("GET /api/bookings/returned-from-projects", "bookings.get_returned_from_projects", bookingHandler.ReturnedFromProjects)
	// (ع) 11-10 بلاغ #30/#31: «منو وصّل الحجز لهالخانة؟» — سجل كل حفظ مسّ الحجز.
	secs["bookings"].Set("GET /api/bookings/{id}/history", "bookings.get_by_id_history", func(w http.ResponseWriter, r *http.Request) {
		lines, err := employeeActivityService.BookingHistory(r.PathValue("id"))
		if err != nil {
			handler.WriteError(w, http.StatusInternalServerError, "تعذر جلب سجل الحجز")
			return
		}
		handler.WriteJSON(w, http.StatusOK, lines)
	})
	secs["bookings"].Set("PUT /api/bookings/{id}/return-handled", "bookings.put_by_id_return_handled", bookingHandler.ReturnHandled)
	secs["bookings"].Set("PUT /api/bookings/{id}/confirmation-contacted", "bookings.put_by_id_confirmation_contacted", bookingHandler.MarkConfirmationContacted)
	// تدقيق المراقب على الحجوزات الموجّهة قبل التثبيت (crew_management، صلاحية جديدة
	// يقدر الأدمن يمنحها لأي موظف مراقب من صفحة الصلاحيات).
	secs["bookings"].Set("GET /api/bookings/pending-audit", "bookings.get_pending_audit", bookingHandler.PendingAudit)
	secs["bookings"].Set("GET /api/bookings/{id}/tool-checks", "bookings.get_by_id_tool_checks", bookingHandler.ToolChecks)

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
	secs["cart"].Set("GET /api/cart/booking/{bookingId}", "cart.get_booking_by_bookingid", cartHandler.ListForBooking, requireCartBookingParty)
	secs["cart"].Set("POST /api/cart/booking/{bookingId}", "cart.post_booking_by_bookingid", cartHandler.Create, requireCartBookingParty)
	secs["cart"].Set("PUT /api/cart/{id}", "cart.put_by_id", cartHandler.Update, requireCartItemParty)
	secs["cart"].Set("DELETE /api/cart/{id}", "cart.delete_by_id", cartHandler.Delete, requireCartItemParty)

	// المصاريف — أي موظف يقدر يرسل مصروف، الموافقة/الرفض للمحاسب ومدير النظام فقط
	financeApp.Set("GET /api/expenses", "finance.expenses", expenseHandler.List)
	financeApp.Set("POST /api/expenses", "finance.expense_add", expenseHandler.Create, idem)
	// صلاحية «إدارة المصاريف (موافقة ورفض)» چانت ما تشتغل — الحارس ما يعرفها.
	financeApp.Set("PUT /api/expenses/{id}/status", "finance.expense_status", expenseHandler.UpdateStatus)

	// المخزون — أدوات شخصية / مركبات / أدوات مشتركة / طلبات الأدوات
	secs["inventory"].Set("GET /api/inventory/personal", "inventory.get_personal", inventoryHandler.ListPersonalTools)
	secs["inventory"].Set("POST /api/inventory/personal", "inventory.post_personal", inventoryHandler.CreatePersonalTool)
	secs["inventory"].Set("PUT /api/inventory/personal/{id}", "inventory.put_personal_by_id", inventoryHandler.UpdatePersonalTool)
	secs["inventory"].Set("DELETE /api/inventory/personal/{id}", "inventory.delete_personal_by_id", inventoryHandler.DeletePersonalTool)
	// سجل حركة الأدوات — يبيّن متى انفقدت كل أداة ومنو سجّل الفقدان
	secs["inventory"].Set("GET /api/inventory/tool-events", "inventory.get_tool_events", inventoryHandler.ToolEvents)

	// العدة القياسية (PersonalToolTemplateItem) — القراءة مفتوحة لأي موظف، الإضافة/الحذف
	// لمن عنده صلاحية "inventory" أو HR/ADMIN (نفس requireHROrInventory).
	secs["inventory"].Set("GET /api/inventory/personal-template", "inventory.get_personal_template", inventoryHandler.ListPersonalToolTemplateItems)
	secs["inventory"].Set("POST /api/inventory/personal-template", "inventory.post_personal_template", inventoryHandler.CreatePersonalToolTemplateItem)
	secs["inventory"].Set("DELETE /api/inventory/personal-template/{id}", "inventory.delete_personal_template_by_id", inventoryHandler.DeletePersonalToolTemplateItem)

	// ═══ استثناء أداة من نواقص موظف بعينه ═══
	// الحذف يستثنيها تلقائياً؛ هذولا للعرض وللتراجع. نفس حارس حذف
	// الأداة نفسه — منو يحذف هو منو يرجّع.
	secs["inventory"].Set("GET /api/inventory/tool-exemptions", "inventory.get_tool_exemptions", inventoryHandler.ListToolExemptions)
	secs["inventory"].Set("POST /api/inventory/tool-exemptions", "inventory.post_tool_exemptions", inventoryHandler.CreateToolExemption)
	secs["inventory"].Set("DELETE /api/inventory/tool-exemptions", "inventory.delete_tool_exemptions", inventoryHandler.DeleteToolExemption)

	// جرد يومي: الموظف يؤكد جرد عدته الخاصة، الإداري يشوف نتائج اليوم لكل الموظفين
	requestsApp.Set("POST /api/inventory/checks", "requests.inventory_check", inventoryHandler.CreateInventoryCheck)
	// نتائج جرد كل الفنيين = شاشة متابعة، مو شغل الفني. جانت مفتوحة لأي
	// موظف مسجّل دخول، فالفني يشوف نواقص زملاءه. صارت بصلاحية «جرد
	// الأدوات» — الفني يسوّي جرده هو ويشوف حالته من المسار الي بعده.
	secs["inventory"].Set("GET /api/inventory/checks/today", "inventory.get_checks_today", inventoryHandler.TodaysInventoryChecks)
	secs["inventory"].Set("GET /api/inventory/checks/mine", "inventory.get_checks_mine", inventoryHandler.MyLastInventoryCheck)
	// جرد كادر حجز معيّن — الخدمة تتأكد إن الطالب من نفس الحجز، فما
	// يحتاج حارس دور: الليدر يشوف فريقه، والفني يشوف رفاقه بنفس الشغلة.
	secs["inventory"].Set("GET /api/inventory/checks/booking/{id}", "inventory.get_checks_booking_by_id", inventoryHandler.BookingCrewInventory)
	// (ع) 11-04: چان requireHR بس — إداري الكميات (صلاحية inventory) يشوف الزر وياخذ 403
	// (وتنحسب عليه مخالفة)، فولا نقص انسكّر بتاريخ النظام. نفس منطق canManageInventory بالواجهة.
	// (ون بيس 11-10) requireInventoryManage انتقل لقاعدة بالنواة (internal/apps/sections)
	secs["inventory"].Set("POST /api/inventory/checks/{id}/resolve", "inventory.post_checks_by_id_resolve", inventoryHandler.ResolveInventoryCheck)
	secs["inventory"].Set("GET /api/inventory/checks/open", "inventory.get_checks_open", inventoryHandler.OpenShortages)
	// 📨 رفع طلب للمدير (ع) 11-04
	managerRequestRepo := repository.NewManagerRequestRepository(db, notificationRepo)
	managerRequestRepo.StartReminder() // ⏰ (ع) 11-08: بلا رد ٤٨ ساعة ← يرجع يوصل
	managerRequestHandler := handler.NewManagerRequestHandler(db, managerRequestRepo)
	requestsApp.Set("POST /api/manager-requests", "requests.manager_request", managerRequestHandler.Create)
	secs["manager_requests"].Set("GET /api/manager-requests/mine", "manager_requests.get_mine", managerRequestHandler.Mine)
	secs["manager_requests"].Set("GET /api/manager-requests/inventory-context", "manager_requests.get_inventory_context", managerRequestHandler.InventoryContext)
	// 🔔 إشعارات حقيقية على الموبايل (ع) 11-08 — كل إشعار ينبعث لأجهزة صاحبه المشتركة
	pushSubscriber := "mailto:noreply@localhost"
	if len(cfg.CORSOrigins) > 0 && strings.HasPrefix(cfg.CORSOrigins[0], "https://") {
		pushSubscriber = cfg.CORSOrigins[0]
	}
	if webPush, err := service.NewWebPush(db, pushSubscriber); err != nil {
		log.Printf("[push] مطفي: %v", err)
	} else {
		webPush.Start()
		pushHandler := handler.NewPushHandler(webPush)
		secs["push"].Set("GET /api/push/key", "push.get_key", pushHandler.Key)
		meApp.Set("POST /api/push/subscribe", "me.push_subscribe", pushHandler.Subscribe)
		meApp.Set("POST /api/push/unsubscribe", "me.push_unsubscribe", pushHandler.Unsubscribe)
	}
	// 🔄 عقود الصيانة الدورية (ع) 11-08 — المدير والإداري المنسّق
	contractSvc := service.NewMaintenanceContractService(db, bookingService, notificationRepo)
	contractSvc.Start()
	contractHandler := handler.NewMaintenanceContractHandler(contractSvc)
	// (ون بيس 11-10) requireContracts انتقل لقاعدة بالنواة (internal/apps/sections)
	secs["maintenance_contracts"].Set("GET /api/maintenance-contracts", "maintenance_contracts.get_root", contractHandler.List)
	secs["maintenance_contracts"].Set("POST /api/maintenance-contracts", "maintenance_contracts.post_root", contractHandler.Create)
	secs["maintenance_contracts"].Set("PATCH /api/maintenance-contracts/{id}", "maintenance_contracts.patch_by_id", contractHandler.Update)
	// 🏗️ صيت التصاميم (ع) 11-08 — الصلاحيات بالخدمة نفسها
	designSetService := service.NewDesignSetService(db, repository.NewExtraTaskRepository(db), notificationRepo)
	designSetHandler := handler.NewDesignSetHandler(designSetService, permissionRepo)
	// (ك) 11-10: 📐 سيت المخططات بطلب دازاي — نفس صلاحيات صيت التصاميم (بالخدمة)
	planSetHandler := handler.NewPlanSetHandler(service.NewPlanSetService(db, designSetService), permissionRepo)
	for _, rt := range []struct {
		pat string
		fn  http.HandlerFunc
	}{
		{"GET /api/design-sets/meta", designSetHandler.Meta},
		{"GET /api/design-sets", designSetHandler.List},
		{"POST /api/design-sets", designSetHandler.Create},
		{"GET /api/design-sets/{id}", designSetHandler.Get},
		{"POST /api/design-sets/{id}/archive", designSetHandler.Archive},
		{"POST /api/design-sets/{id}/sections", designSetHandler.AddSection},
		{"POST /api/design-sections/{id}/assign", designSetHandler.Assign},
		{"POST /api/design-sections/{id}/files", designSetHandler.AddFile},
		{"POST /api/design-sections/{id}/submit", designSetHandler.Submit},
		{"POST /api/design-sections/{id}/review", designSetHandler.Review},
		{"DELETE /api/design-files/{id}", designSetHandler.RemoveFile},
		{"POST /api/design-tools", designSetHandler.SaveTool},
		{"PUT /api/design-tools/{id}", designSetHandler.SaveTool},
		{"GET /api/plan-sets/meta", planSetHandler.Meta},
		{"GET /api/plan-sets", planSetHandler.List},
		{"POST /api/plan-sets", planSetHandler.Create},
		{"GET /api/plan-sets/{id}", planSetHandler.Get},
		{"POST /api/plan-sets/{id}/archive", planSetHandler.Archive},
		{"POST /api/plan-sets/{id}/files", planSetHandler.AddFile},
		{"DELETE /api/plan-files/{id}", planSetHandler.RemoveFile},
	} {
		// (ون بيس 11-10) الصلاحية بالخدمة (CanSee/CanManage — ErrDesignHidden ← 404)
		designsApp.Add(rt.pat, core.Rule{Label: rt.pat + " — الفحص بالخدمة (صيت التصاميم/المخططات)", Anyone: true}, rt.fn)
	}
	// 📅 متابعة العقود (ع) 11-08
	service.StartContractWatch(db, notificationRepo)
	// 🔁 الرجعات (ع) 11-07 — نفس حارس شغل المراقب
	revisitHandler := handler.NewRevisitHandler(repository.NewRevisitRepository(db))
	secs["revisits"].Set("GET /api/revisits", "revisits.get_root", revisitHandler.Report)
	secs["manager_requests"].Set("GET /api/manager-requests", "manager_requests.get_root", managerRequestHandler.List)
	secs["manager_requests"].Set("POST /api/manager-requests/{id}/handle", "manager_requests.post_by_id_handle", managerRequestHandler.Handle)

	secs["inventory"].Set("GET /api/inventory/vehicle", "inventory.get_vehicle", inventoryHandler.ListVehicleTools)
	secs["inventory"].Set("POST /api/inventory/vehicle", "inventory.post_vehicle", inventoryHandler.CreateVehicleTool)
	// تعديل/حذف أداة مركبة موجودة يقتصر على إدارة الكوادر (نفس صلاحية أدوات
	// الأدوات الشخصية requireHR) — الإنشاء يبقى مفتوح لأي موظف (نفس السلوك القديم).
	secs["inventory"].Set("PUT /api/inventory/vehicle/{id}", "inventory.put_vehicle_by_id", inventoryHandler.UpdateVehicleTool)
	secs["inventory"].Set("DELETE /api/inventory/vehicle/{id}", "inventory.delete_vehicle_by_id", inventoryHandler.DeleteVehicleTool)
	secs["inventory"].Set("GET /api/inventory/vehicle-tool-checks", "inventory.get_vehicle_tool_checks", inventoryHandler.ListVehicleToolChecks)
	secs["inventory"].Set("GET /api/inventory/booking-tool-checks", "inventory.get_booking_tool_checks", inventoryHandler.ListAllBookingToolChecks)

	secs["inventory"].Set("GET /api/inventory/ondemand", "inventory.get_ondemand", inventoryHandler.ListOnDemandTools)
	secs["inventory"].Set("POST /api/inventory/ondemand", "inventory.post_ondemand", inventoryHandler.CreateOnDemandTool)
	// تعديل أداة "حسب الحاجة" يقتصر على الأدمن أو إداري الكميات (نفس canManageOnDemand
	// بالواجهة: isAdmin || PROCUREMENT_ADMIN) — يطابق requireProcurementAdmin الموجود.
	secs["inventory"].Set("PUT /api/inventory/ondemand/{id}", "inventory.put_ondemand_by_id", inventoryHandler.UpdateOnDemandTool)

	// إضافة الكميات للمخزون — إداري الكميات حصراً
	// (ع) 10-16: «أدوات المخزن التخصصية» — أبو الكميات ينطي ويرجّع من شاشته (نفس حارس إضافة الأدوات)
	secs["store_tools"].Set("POST /api/store-tools/{id}/issue", "store_tools.post_by_id_issue", inventoryHandler.IssueStoreTool)
	secs["store_tools"].Set("POST /api/store-tools/issued/{id}/return", "store_tools.post_issued_by_id_return", inventoryHandler.ReturnStoreTool)
	secs["inventory"].Set("POST /api/inventory/stock-intake", "inventory.post_stock_intake", inventoryHandler.AddStock)
	secs["inventory"].Set("GET /api/inventory/stock-intake", "inventory.get_stock_intake", inventoryHandler.ListStockIntakes)

	secs["inventory"].Set("GET /api/inventory/requests", "inventory.get_requests", inventoryHandler.ListToolRequests)
	// طلب أداة: الليدر بحكم دوره، **أو** أي موظف تنمنح له صلاحية
	// «طلب أداة من المخزن». قبل كان الليدر بس — فالفني الي تنكسر
	// عدته بالموقع لازم يدور على ليدره حتى يطلبله، والشغل يوقف.
	secs["inventory"].Set("POST /api/inventory/requests", "inventory.post_requests", inventoryHandler.CreateToolRequest)
	secs["inventory"].Set("PUT /api/inventory/requests/{id}/approve", "inventory.put_requests_by_id_approve", inventoryHandler.ApproveToolRequest)
	secs["inventory"].Set("PUT /api/inventory/requests/{id}/reject", "inventory.put_requests_by_id_reject", inventoryHandler.RejectToolRequest)
	// إرجاع الأداة يشيل مسؤوليتها عن الموظف — نفس مستوى الموافقة، لأن
	// إذا أي موظف يقدر يأشر «رجعتها» تروح كل محاسبة العدة المفقودة.
	secs["inventory"].Set("PUT /api/inventory/requests/{id}/return", "inventory.put_requests_by_id_return", inventoryHandler.ReturnToolRequest)
	// حذف طلب أداة يمحي أثر الطلب نهائياً — محصور بمدير النظام والمالك فقط،
	// مو بكل من يقدر يوافق أو يرفض.
	secs["inventory"].Set("DELETE /api/inventory/requests/{id}", "inventory.delete_requests_by_id", inventoryHandler.DeleteToolRequest)

	// تقييم الأداء اليدوي (KPI)
	meApp.Set("POST /api/attendance/checkin", "me.checkin", attendanceHandler.CheckIn)
	meApp.Set("POST /api/attendance/checkout", "me.checkout", attendanceHandler.CheckOut)
	meApp.Set("POST /api/attendance/auto/{id}/answer", "me.attendance_answer", attendanceHandler.AnswerAuto)
	secs["attendance"].Set("GET /api/attendance/claims", "attendance.get_claims", attendanceHandler.Claims)
	secs["attendance"].Set("POST /api/attendance/claims/{id}/decide", "attendance.post_claims_by_id_decide", attendanceHandler.DecideClaim)
	secs["attendance"].Set("GET /api/attendance/gate", "attendance.get_gate", attendanceHandler.Gate)
	secs["attendance"].Set("GET /api/attendance/mine", "attendance.get_mine", attendanceHandler.Mine)
	secs["attendance"].Set("GET /api/attendance/open", "attendance.get_open", attendanceHandler.OpenSession)
	secs["attendance"].Set("GET /api/attendance/today", "attendance.get_today", attendanceHandler.Today)
	secs["attendance"].Set("GET /api/attendance/today-summary", "attendance.get_today_summary", attendanceHandler.TodaySummary)
	secs["attendance"].Set("GET /api/attendance/employee/{id}", "attendance.get_employee_by_id", attendanceHandler.MonthlyReport)
	secs["attendance"].Set("GET /api/attendance/export/employee/{id}", "attendance.get_export_employee_by_id", attendanceHandler.ExportEmployeeMonth)
	secs["attendance"].Set("GET /api/attendance/export/today", "attendance.get_export_today", attendanceHandler.ExportToday)
	attendanceBoard := handler.NewAttendanceBoardHandler(db)
	attendanceHandler.Board = attendanceBoard
	secs["attendance"].Set("GET /api/attendance/board", "attendance.get_board", attendanceBoard.Board)
	secs["attendance"].Set("PUT /api/attendance/{id}", "attendance.put_by_id", attendanceHandler.Correct)

	// كل تقييمات الشركة — حارسها يطابق حارس شاشة «نقاط الكي بي اي».
	secs["kpi"].Set("GET /api/kpi", "kpi.get_root", kpiHandler.List)
	secs["kpi"].Set("GET /api/kpi/employee/{employeeId}", "kpi.get_employee_by_employeeid", kpiHandler.ListForEmployee)
	secs["kpi"].Set("GET /api/kpi/leaderboard/{role}", "kpi.get_leaderboard_by_role", kpiHandler.RoleLeaderboard)
	secs["kpi"].Set("GET /api/kpi/leaderboard-by-permission/{permission}", "kpi.get_leaderboard_by_permission_by_permission", kpiHandler.PermissionLeaderboard)
	secs["kpi"].Set("POST /api/kpi", "kpi.post_root", kpiHandler.Create)
	secs["kpi"].Set("DELETE /api/kpi/{id}", "kpi.delete_by_id", kpiHandler.Delete)
	secs["kpi"].Set("PUT /api/kpi/{id}/cancel", "kpi.put_by_id_cancel", kpiHandler.Cancel)
	peopleApp.Set("POST /api/employees/{id}/complete-training", "people.training_done", kpiHandler.CompleteTraining)

	secs["notifications"].Set("GET /api/notifications", "notifications.get_root", notificationHandler.List)
	meApp.Set("POST /api/notifications/{id}/read", "me.notification_read", notificationHandler.MarkRead)
	meApp.Set("POST /api/notifications/read-all", "me.notifications_read_all", notificationHandler.MarkAllRead)
	secs["kpi_criteria"].Set("GET /api/kpi-criteria", "kpi_criteria.get_root", kpiCriterionHandler.List)
	requestsApp.Set("POST /api/assistant/ask", "requests.assistant_ask", assistantHandler.Ask)
	secs["assistant"].Set("POST /api/assistant/manager-chat", "assistant.post_manager_chat", assistantHandler.ManagerChat)
	secs["assistant"].Set("GET /api/assistant/conversations", "assistant.get_conversations", assistantHandler.ListConversations)
	secs["assistant"].Set("GET /api/assistant/conversations/employees", "assistant.get_conversations_employees", assistantHandler.ListConversationEmployees)
	// قصص الكيان — كلها **لصاحب الجلسة حصراً**، ماكو ولا مسار ياخذ
	// رقم موظف من الرابط. الهوية من التوكن وحده.
	secs["stories"].Set("GET /api/stories/next", "stories.get_next", storyHandler.Next)
	secs["stories"].Set("GET /api/stories/mine", "stories.get_mine", storyHandler.Mine)
	meApp.Set("POST /api/stories/{id}/claim", "me.story_claim", storyHandler.Claim)
	meApp.Set("POST /api/stories/{id}/advance", "me.story_advance", storyHandler.Advance)

	secs["kpi_criteria"].Set("POST /api/kpi-criteria", "kpi_criteria.post_root", kpiCriterionHandler.Create)
	secs["kpi_criteria"].Set("DELETE /api/kpi-criteria/{id}", "kpi_criteria.delete_by_id", kpiCriterionHandler.Delete)

	// تقييم الأداء التلقائي (Smart KPI) — الرانك الأسبوعي/الشهري للفنيين
	secs["smart_kpi"].Set("GET /api/smart-kpi/technician/{employeeId}", "smart_kpi.get_technician_by_employeeid", smartKpiHandler.Technician)
	secs["smart_kpi"].Set("GET /api/smart-kpi/leaderboard", "smart_kpi.get_leaderboard", smartKpiHandler.Leaderboard)

	// الشكاوى
	secs["complaints"].Set("GET /api/complaints", "complaints.get_root", complaintHandler.List, requireCustomerRead)
	// 🛠️ صيانة وقائية (دازاي 11-06، فكرة ٧)
	secs["maintenance_due"].Set("GET /api/maintenance-due", "maintenance_due.get_root", handler.MaintenanceDueHandler(db).ServeHTTP, requireCustomerRead)
	secs["complaints"].Set("POST /api/complaints", "complaints.post_root", complaintHandler.Create, requireCustomerRead)
	// تحديث حالة الشكوى وحلّها يقتصران على من يملك صلاحية "quality_control" (نفس
	// أدوار المتابعة بالواجهة: مهندس الجودة/المراقب/الأدمن) — لا أي موظف مسجل دخول
	// (مثلاً موظف مبيعات عنده صلاحية "complaints" بس لتسجيل شكوى جديدة فقط).
	secs["complaints"].Set("PUT /api/complaints/{id}", "complaints.put_by_id", complaintHandler.Update)
	// تأشير الاتصال والملاحظات: أي موظف يقدر — المهم النظام يخزن منو
	// اتصل. حصرها بالجودة يخلي الشكوى تضل بلا متابعة لو المهندس مشغول.
	secs["complaints"].Set("PUT /api/complaints/{id}/contact", "complaints.put_by_id_contact", complaintHandler.SetContacted)
	secs["complaints"].Set("PUT /api/complaints/{id}/no-answer", "complaints.put_by_id_no_answer", complaintHandler.NoAnswer)
	secs["complaints"].Set("PUT /api/complaints/{id}/no-answer-close", "complaints.put_by_id_no_answer_close", complaintHandler.DropNoAnswer)
	secs["complaints"].Set("PUT /api/complaints/{id}/notes", "complaints.put_by_id_notes", complaintHandler.SetNotes)
	secs["complaints"].Set("PUT /api/complaints/{id}/link-booking", "complaints.put_by_id_link_booking", complaintHandler.LinkBooking)
	secs["complaints"].Set("PUT /api/complaints/{id}/resolve", "complaints.put_by_id_resolve", complaintHandler.Resolve)
	secs["complaints"].Set("GET /api/complaints/stats", "complaints.get_stats", complaintHandler.Stats, requireCustomerRead)
	// ⚠️ التدقيق للمالك والمراقب والمدير **بس** — مهندس الجودة ما
	// يدقّق نفسه، هو الي انتقيّم شغله بهذي الشاشة.
	secs["complaints"].Set("PUT /api/complaints/{id}/audit", "complaints.put_by_id_audit", complaintHandler.Audit)
	secs["complaints"].Set("GET /api/complaints/{id}/events", "complaints.get_by_id_events", complaintHandler.Events, requireCustomerRead)
	// ⚠️ القراءة تقبل **الدور** مو الصلاحية بس: صاحب النظام طلب
	// «المراقب لازم يشوف متابعة الجودة كاملة»، وخريطة صلاحيات
	// الأدوار **ما تنطبّق تلقائياً** — فمراقب ما انضغط عليه «طبّق
	// الافتراضي» چان ينرفض بالقراءة أصلاً.
	// والكتابة تبقى على requireQuality، وblockMonitorWrite يمنع
	// المراقب منها حتى لو عنده الصلاحية.
	secs["quality_follow_ups"].Set("PUT /api/quality-follow-ups/{id}/monitor-verdict", "quality_follow_ups.put_by_id_monitor_verdict", qualityFollowUpHandler.MonitorVerdict)
	secs["quality_follow_ups"].Set("GET /api/quality-follow-ups", "quality_follow_ups.get_root", qualityFollowUpHandler.List)
	secs["quality_follow_ups"].Set("PUT /api/quality-follow-ups/{id}", "quality_follow_ups.put_by_id", qualityFollowUpHandler.Update)
	// حكم الجودة: تقرير إيجابي/سلبي، والكشف الميداني. نفس صلاحية الشاشة
	// نفسها — الي يشوف المتابعة هو الي يحكم بيها.
	secs["quality_follow_ups"].Set("POST /api/quality-follow-ups/{id}/verdict", "quality_follow_ups.post_by_id_verdict", qualityFollowUpHandler.Verdict)
	secs["quality_follow_ups"].Set("POST /api/quality-follow-ups/{id}/inspect", "quality_follow_ups.post_by_id_inspect", qualityFollowUpHandler.Inspect)
	// بعد ما مهندس الجودة يفتح حجز صيانة حقيقي من شاشة المتابعة — يربطها
	// بيه بدل ما تروح بالهوا (نفس صلاحية الحكم والكشف بالضبط).
	secs["quality_follow_ups"].Set("PUT /api/quality-follow-ups/{id}/link-booking", "quality_follow_ups.put_by_id_link_booking", qualityFollowUpHandler.LinkBooking)

	// التدريب — عرض متاح لأي مسجل دخول، التعيين وإدارة المواد لمدير النظام فقط
	secs["training"].Set("GET /api/training/materials/mine", "training.get_materials_mine", trainingHandler.MaterialsMine)
	secs["training"].Set("GET /api/training/assignments/{employeeId}", "training.get_assignments_by_employeeid", trainingHandler.Assignments)
	secs["training"].Set("PUT /api/training/assignments/{employeeId}", "training.put_assignments_by_employeeid", trainingHandler.SetAssignments)
	secs["training"].Set("GET /api/training/materials", "training.get_materials", trainingHandler.ListMaterials)
	secs["training"].Set("POST /api/training/materials", "training.post_materials", trainingHandler.CreateMaterial)
	secs["training"].Set("PUT /api/training/materials/{id}", "training.put_materials_by_id", trainingHandler.UpdateMaterial)
	secs["training"].Set("DELETE /api/training/materials/{id}", "training.delete_materials_by_id", trainingHandler.DeleteMaterial)

	// تتبع المهام (missions)
	// (فحص 10-11) كل المهمات (مواقع الزبائن والحركة) لأهل المتابعة بس — الفني يشوف مهامه بـ/my
	// (ون بيس 11-10) requireMissionView انتقل لقاعدة بالنواة (internal/apps/sections)
	secs["missions"].Set("GET /api/missions", "missions.get_root", missionHandler.List)
	secs["missions"].Set("GET /api/missions/monitor/live", "missions.get_monitor_live", missionHandler.MonitorLive)
	secs["missions"].Set("GET /api/missions/reports/performance", "missions.get_reports_performance", missionHandler.PerformanceReport)
	secs["missions"].Set("GET /api/missions/my/{employeeId}", "missions.get_my_by_employeeid", missionHandler.ListForEmployee)
	missionHandler.Supervisor = middleware.RoleOrAnyPermissionCheck(permissionRepo, []string{"ADMIN", "HR_COORDINATOR", "MONITOR", "PROJECT_MANAGER"}, "monitoring", "auditing")
	secs["missions"].Set("GET /api/missions/{id}", "missions.get_by_id", missionHandler.Get)
	secs["missions"].Set("POST /api/missions", "missions.post_root", missionHandler.Create)
	// تقدّم مراحل المهمة: إدارة المشاريع أو المراقب — مو أي موظف يحرّك
	// مهمة غيره.
	secs["missions"].Set("PUT /api/missions/{id}/stage", "missions.put_by_id_stage", missionHandler.UpdateStage)

	// إدارة المشاريع (projects)
	secs["projects"].Set("GET /api/projects", "projects.get_root", projectHandler.List)
	secs["projects"].Set("GET /api/projects/{id}", "projects.get_by_id", projectHandler.Get)
	// إنشاء مشروع جديد يتطلب نفس دور مدير المشاريع/الأدمن المطلوب للتعديل والحذف
	// تحته مباشرة — كان مفتوح غلط لأي موظف مسجل دخول فقط (requireAuth بدون requireProjectManager)،
	// عدم اتساق مع PUT/DELETE على نفس المورد.
	// الإنشاء يقبل كذلك صلاحية "إضافة مشروع فقط" المبسّطة — بينما التعديل/الحذف
	// والترحيل بين المراحل يبقى محصوراً بمدير المشاريع/الأدمن.
	secs["projects"].Set("POST /api/projects", "projects.post_root", projectHandler.Create)
	// المشاريع المُسلَّمة للموظف الحالي — ما تحتاج صلاحية إدارة مشاريع عامة،
	// التسليم نفسه هو الصلاحية وعلى هالمشاريع بس.
	secs["projects"].Set("GET /api/projects/delegated-to-me", "projects.get_delegated_to_me", projectHandler.ListDelegatedToMe)
	// requireProjectMgmt مو requireProjectManager: بقية مسارات المشاريع تقبل
	// صلاحية project_management الممنوحة يدوياً، وهذا المسار كان الوحيد المقيّد
	// بالدور — فمهندس عنده الصلاحية يشوف رابط "إحصائيات المشاريع" بقائمته ويوكع بـ403.
	secs["projects"].Set("GET /api/projects/statistics", "projects.get_statistics", projectHandler.Statistics)
	secs["projects"].Set("PUT /api/projects/{id}/fill-info", "projects.put_by_id_fill_info", projectHandler.FillInfo)
	secs["projects"].Set("GET /api/projects/{id}/delegation-log", "projects.get_by_id_delegation_log", projectHandler.DelegationLog)
	// التسليم/السحب بيد مدير المشاريع بس — الموظف المُسلَّم إله ما يقدر يسلّمه لغيره
	secs["projects"].Set("PUT /api/projects/{id}/delegate", "projects.put_by_id_delegate", projectHandler.Delegate, courier.Wrap("project"))

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
	secs["projects"].Set("PUT /api/projects/{id}", "projects.put_by_id", projectHandler.Update, allowManagerOrDelegate)
	// (ع) 10-19: المشروع المرفوض يرجع لمرحلته قبل الرفض — نفس حارس الترحيل.
	secs["projects"].Set("PUT /api/projects/{id}/restore", "projects.put_by_id_restore", projectHandler.Restore, allowManagerOrDelegate)
	// (ع) 10-16: رقابة المشرف على عرض السعر — «تواصلت ويا الزبون» بنفس حارس تعديل المشروع
	quoteWatch := handler.NewProjectQuoteWatchHandler(db)
	secs["projects"].Set("POST /api/projects/{id}/customer-contacted", "projects.post_by_id_customer_contacted", quoteWatch.MarkContacted, allowManagerOrDelegate)
	secs["project_quote_watch"].Set("GET /api/project-quote-watch", "project_quote_watch.get_root", quoteWatch.List)
	// (دازاي 10-31) إجراءات المراقب على المتأخر: مهلة، تصعيد لمكتب الإدارة، مخالفة (وتراجع)
	quoteWatch.SetActionDeps(extraTaskRepo, disciplineRepo, notificationRepo)
	secs["project_quote_watch"].Set("POST /api/project-quote-watch/{id}/action", "project_quote_watch.post_by_id_action", quoteWatch.Act)
	secs["project_monitor_actions"].Set("POST /api/project-monitor-actions/{id}/revoke", "project_monitor_actions.post_by_id_revoke", quoteWatch.Revoke)
	// علامة فقط (نوع الإجراء ووقته) — بلا حارس صلاحية حتى «إضافة مشروع فقط» ما ياخذ 403 ومخالفة
	secs["projects"].Set("GET /api/projects/monitor-watch", "projects.get_monitor_watch", quoteWatch.WatchList)
	secs["projects"].Set("DELETE /api/projects/{id}", "projects.delete_by_id", projectHandler.Delete)
	// (ع) 10-19: حذف نهائي للمرفوض بسبب — نفس حارس الحذف، ونسخته تنحفظ بـ"DeletedProject".
	secs["projects"].Set("DELETE /api/projects/{id}/purge", "projects.delete_by_id_purge", projectHandler.Purge)
	// حذف ملف العقد المرفوع — لمدير النظام حصراً
	secs["projects"].Set("DELETE /api/projects/{id}/contract", "projects.delete_by_id_contract", projectHandler.DeleteContract)

	// أنواع الأعمال ("نوع العمل" بحقل المشروع) — إعدادات وحدة إدارة المشاريع:
	// أي موظف مسجل دخول يشوف القائمة (يحتاجها بفورمة المشروع)، بس الإضافة/الحذف
	// محصورة بنفس صلاحية مدير المشاريع المستخدمة لبقية إعدادات هذي الوحدة.
	secs["project_work_types"].Set("GET /api/project-work-types", "project_work_types.get_root", projectWorkTypeHandler.List)
	// مرشحو المشروع (المسؤول / منفّذ الكشف) مصنّفين ومرتّبين
	secs["project_candidates"].Set("GET /api/project-candidates", "project_candidates.get_root", projectWorkTypeHandler.ListCandidates)

	// الشخصيات المهمة (VIP): أي موظف يعلّم بضغطة زر، بس التفاصيل الكاملة
	// (رقم الزبون وشنو طلب ومنو علّمه) تُعرض لمدير النظام حصراً.
	secs["vip_customers"].Set("GET /api/vip-customers", "vip_customers.get_root", vipCustomerHandler.List)
	secs["vip_customers"].Set("GET /api/vip-customers/ids", "vip_customers.get_ids", vipCustomerHandler.ListIDs)
	// الإضافة اليدوية للشخصية المهمة: الإداري ومدير النظام حصراً — مو أي
	// موظف عنده صلاحية العملاء. (الترحيل التلقائي من المشاريع يمر بالخدمة
	// مو بهذا المسار، فما يتأثر.)
	secs["vip_customers"].Set("POST /api/vip-customers", "vip_customers.post_root", vipCustomerHandler.Mark)
	secs["vip_customers"].Set("DELETE /api/vip-customers/{customerId}", "vip_customers.delete_by_customerid", vipCustomerHandler.Unmark)
	secs["project_work_types"].Set("POST /api/project-work-types", "project_work_types.post_root", projectWorkTypeHandler.Create)
	secs["project_work_types"].Set("DELETE /api/project-work-types/{id}", "project_work_types.delete_by_id", projectWorkTypeHandler.Delete)

	// الكشوفات: فورمات فارغة يطبعها المهندس، يمليها بالموقع، وبعدين يرفع صور
	// الفورمة المالية — أي موظف مسجل دخول يقدر ينشئ/يرفع (مو حصراً مدير مشاريع).
	secs["checklists"].Set("GET /api/checklists", "checklists.get_root", checklistHandler.List)
	secs["checklists"].Set("POST /api/checklists", "checklists.post_root", checklistHandler.Create)
	secs["checklists"].Set("PUT /api/checklists/{id}/photos", "checklists.put_by_id_photos", checklistHandler.AddPhotos)

	// معرض أعمال التقنيين — أي موظف مسجل دخول يتصفحه (إلهام)، بس صاحب صلاحية
	// "التقني" (content_technician) بس يقدر يضيف عمل/يرفع وسائط.
	secs["tech_showcase"].Set("GET /api/tech-showcase", "tech_showcase.get_root", techShowcaseHandler.List)
	secs["tech_showcase"].Set("POST /api/tech-showcase", "tech_showcase.post_root", techShowcaseHandler.Create)
	secs["tech_showcase"].Set("PUT /api/tech-showcase/{id}/media", "tech_showcase.put_by_id_media", techShowcaseHandler.AddMedia)

	// وحدة التقنيين — إدارة المعارض
	secs["exhibitions"].Set("GET /api/exhibitions", "exhibitions.get_root", exhibitionHandler.List)
	secs["exhibitions"].Set("POST /api/exhibitions", "exhibitions.post_root", exhibitionHandler.Create)
	secs["exhibitions"].Set("PUT /api/exhibitions/{id}/nominate", "exhibitions.put_by_id_nominate", exhibitionHandler.Nominate)
	secs["exhibitions"].Set("PUT /api/exhibitions/{id}/photos", "exhibitions.put_by_id_photos", exhibitionHandler.AddPhotos)
	secs["exhibitions"].Set("PUT /api/exhibitions/{id}/findings", "exhibitions.put_by_id_findings", exhibitionHandler.SetFindings)
	secs["exhibitions"].Set("POST /api/exhibitions/{id}/report", "exhibitions.post_by_id_report", exhibitionHandler.GenerateReport)
	secs["exhibitions"].Set("PUT /api/exhibitions/{id}/archive", "exhibitions.put_by_id_archive", exhibitionHandler.Archive)

	// وحدة التقنيين — إدارة المنتجات
	secs["product_requests"].Set("GET /api/product-requests", "product_requests.get_root", productRequestHandler.List)
	secs["product_requests"].Set("POST /api/product-requests", "product_requests.post_root", productRequestHandler.Create)
	secs["product_requests"].Set("PUT /api/product-requests/{id}/approve", "product_requests.put_by_id_approve", productRequestHandler.Approve)
	secs["product_requests"].Set("PUT /api/product-requests/{id}/reject", "product_requests.put_by_id_reject", productRequestHandler.Reject)

	// تجهيز طلب المنتج من الدوار — أبو الحسابات (نفسه أبو الكميات)
	secs["product_procurements"].Set("GET /api/product-procurements", "product_procurements.get_root", productProcurementHandler.List)
	secs["product_requests"].Set("POST /api/product-requests/{id}/fulfill", "product_requests.post_by_id_fulfill", productProcurementHandler.Fulfill)
	secs["product_procurements"].Set("PUT /api/product-procurements/{id}/settle", "product_procurements.put_by_id_settle", productProcurementHandler.Settle)

	// وحدة التقنيين — إدارة الخدمات
	secs["service_studies"].Set("GET /api/service-studies", "service_studies.get_root", serviceStudyHandler.List)
	secs["service_studies"].Set("POST /api/service-studies", "service_studies.post_root", serviceStudyHandler.Create)
	secs["service_studies"].Set("PUT /api/service-studies/{id}/assign", "service_studies.put_by_id_assign", serviceStudyHandler.Assign)
	secs["service_studies"].Set("POST /api/service-studies/{id}/reports", "service_studies.post_by_id_reports", serviceStudyHandler.AddReport)
	secs["service_studies"].Set("PUT /api/service-studies/{id}/archive", "service_studies.put_by_id_archive", serviceStudyHandler.Archive)

	// وحدة التصميم — عدة استمارات مستقلة، كل وحدة بأسئلتها وبرابطها العام الخاص
	// ═══ معرض التصاميم ═══
	//
	// معرض **مستقل** ما ينربط بحجز — قراره، لأن المصممة ما تشوف
	// الحجوزات. والصلاحية معزولة عن فورمة التصميم: ممكن تريد واحداً
	// يشوف المعرض بلا ما يحرّر الفورمة.
	//
	// ⚠️ و`unit_design` مقبولة بعد: صلاحية الوحدة معناها بالنظام «تفتح
	// الوحدة كاملة بكل صفحاتها»، فهي تفتح البند بالقائمة الجانبية. لو
	// السيرفر ما يقبلها، المصممة تشوف «معرض التصاميم» وتضغطه وترجعها
	// الصفحة «غير مصرح لك» — والحارسان يتناقضان.
	requireDesignGallery := middleware.RequireRoleOrAnyPermission(
		permissionRepo, employeeRepo, notificationRepo,
		[]string{"ADMIN", "OWNER"}, "design_gallery", "unit_design")
	secs["design_assets"].Set("GET /api/design-assets", "design_assets.get_root", designAssetHandler.List, requireDesignGallery)
	secs["design_assets"].Set("GET /api/design-assets/categories", "design_assets.get_categories", designAssetHandler.Categories, requireDesignGallery)
	secs["design_assets"].Set("POST /api/design-assets", "design_assets.post_root", designAssetHandler.Create, requireDesignGallery)
	secs["design_assets"].Set("PUT /api/design-assets/{id}/archive", "design_assets.put_by_id_archive", designAssetHandler.SetArchived, requireDesignGallery)

	// ═══ فورمة التصميم — تنفتح للمصممة ═══
	//
	// ⚠️ التسعة كلهن چانن `requireAdmin` حصراً، **صفر منفذ صلاحية** —
	// يعني حتى لو انمنحت المصممة كل شي بالواجهة، الخادم يرفضها ٤٠٣.
	// ⚠️ **إضافة منفذ، بلا سحب وصول**: مدير النظام يبقى يوصلهن مثل
	// ما هو بالضبط.
	requireDesignForms := middleware.RequireRoleOrAnyPermission(
		permissionRepo, employeeRepo, notificationRepo,
		[]string{"ADMIN", "OWNER"}, "design_forms")
	secs["design_forms"].Set("GET /api/design-forms", "design_forms.get_root", designFormHandler.ListForms, requireDesignForms)
	secs["design_forms"].Set("POST /api/design-forms", "design_forms.post_root", designFormHandler.CreateForm, requireDesignForms)
	secs["design_forms"].Set("DELETE /api/design-forms/{id}", "design_forms.delete_by_id", designFormHandler.DeleteForm, requireDesignForms)
	secs["design_forms"].Set("GET /api/design-forms/{formId}/submissions", "design_forms.get_by_formid_submissions", designFormHandler.ListSubmissions, requireDesignForms)
	secs["design_forms"].Set("GET /api/design-forms/{formId}/questions", "design_forms.get_by_formid_questions", designFormHandler.List, requireDesignForms)
	secs["design_forms"].Set("POST /api/design-forms/{formId}/questions", "design_forms.post_by_formid_questions", designFormHandler.Create, requireDesignForms)
	secs["design_form"].Set("PUT /api/design-form/questions/{id}", "design_form.put_questions_by_id", designFormHandler.Update, requireDesignForms)
	secs["design_form"].Set("DELETE /api/design-form/questions/{id}", "design_form.delete_questions_by_id", designFormHandler.Delete, requireDesignForms)
	secs["design_form"].Set("PUT /api/design-form/questions/reorder", "design_form.put_questions_reorder", designFormHandler.Reorder, requireDesignForms)

	// رابط عام للزبون (بدون تسجيل دخول) — يشوف الاستمارة ويرسل جوابه فقط
	// النماذج العامة: بلا تسجيل دخول، فلازم حد صارم — بدونه أي واحد عنده
	// الرابط يقدر يغرق قاعدة البيانات بآلاف التقديمات الضخمة.
	publicFormReadLimit := middleware.RateLimit(30, time.Minute)
	publicFormSubmitLimit := middleware.RateLimit(5, time.Minute)
	mux.Handle("GET /api/public/design-forms/{token}", publicFormReadLimit(http.HandlerFunc(designFormHandler.PublicGet)))
	mux.Handle("POST /api/public/design-forms/{token}/submit", publicFormSubmitLimit(http.HandlerFunc(designFormHandler.PublicSubmit)))

	// ═══ تطبيق الزبائن — قرار (ع) 10-14 ═══
	// مسارات /api/c/ بحارسها المستقل (توكن الزبون ما يفتح شي من النظام).
	{
		client, model := matrixCommandService.Client()
		custMatrix := service.NewCustomerMatrix(db, client, model)
		cust := handler.NewCustomerAppHandler(db, cfg.JWTSecret, mail.FromEnv(), custMatrix, fileStore, bookingService, customerRepo, notificationRepo)
		authLimit := middleware.RateLimit(10, time.Minute)
		mux.Handle("POST /api/c/signup", authLimit(http.HandlerFunc(cust.Signup)))
		mux.Handle("POST /api/c/verify", authLimit(http.HandlerFunc(cust.Verify)))
		mux.Handle("POST /api/c/resend", authLimit(http.HandlerFunc(cust.Resend)))
		mux.Handle("POST /api/c/login", authLimit(http.HandlerFunc(cust.Login)))
		mux.Handle("POST /api/c/forgot", authLimit(http.HandlerFunc(cust.Forgot)))
		mux.Handle("POST /api/c/reset", authLimit(http.HandlerFunc(cust.Reset)))
		mux.Handle("GET /api/c/img/{id}", middleware.RateLimit(240, time.Minute)(http.HandlerFunc(cust.Image)))
		mux.Handle("GET /api/c/me", cust.RequireCustomer(http.HandlerFunc(cust.Me)))
		mux.Handle("PUT /api/c/me", cust.RequireCustomer(http.HandlerFunc(cust.UpdateMe)))
		mux.Handle("GET /api/c/me/location", cust.RequireCustomer(http.HandlerFunc(cust.MyLocation)))
		mux.Handle("PUT /api/c/me/location", cust.RequireCustomer(http.HandlerFunc(cust.SetMyLocation)))
		mux.Handle("GET /api/c/home", cust.RequireCustomer(http.HandlerFunc(cust.Home)))
		mux.Handle("POST /api/c/home/seen", cust.RequireCustomer(http.HandlerFunc(cust.Seen)))
		mux.Handle("GET /api/c/bookings", cust.RequireCustomer(http.HandlerFunc(cust.Bookings)))
		mux.Handle("GET /api/c/chat", cust.RequireCustomer(http.HandlerFunc(cust.Chat)))
		mux.Handle("POST /api/c/chat", middleware.RateLimit(20, time.Minute)(cust.RequireCustomer(http.HandlerFunc(cust.Send))))
		// إدارة المالك/المدير
		secs["customer_app"].Set("GET /api/customer-app/promos", "customer_app.get_promos", cust.AdminPromos)
		secs["customer_app"].Set("POST /api/customer-app/promos", "customer_app.post_promos", cust.AdminCreatePromo)
		secs["customer_app"].Set("POST /api/customer-app/promos/suggest", "customer_app.post_promos_suggest", cust.AdminSuggest)
		secs["customer_app"].Set("PUT /api/customer-app/promos/{id}", "customer_app.put_promos_by_id", cust.AdminUpdatePromo)
		secs["customer_app"].Set("DELETE /api/customer-app/promos/{id}", "customer_app.delete_promos_by_id", cust.AdminDeletePromo)
		secs["customer_app"].Set("POST /api/customer-app/promos/{id}/image", "customer_app.post_promos_by_id_image", cust.AdminPromoImage)
		secs["customer_app"].Set("GET /api/customer-app/accounts", "customer_app.get_accounts", cust.AdminAccounts)
		secs["customer_app"].Set("GET /api/customer-app/accounts/{id}/chat", "customer_app.get_accounts_by_id_chat", cust.AdminChat)
		// التصميم الجديد (10-14): الكتالوك، المفضلة، طلب بالنموذج، تفاصيل الحجز، التقييم، الفروع
		rc := func(f http.HandlerFunc) http.Handler { return cust.RequireCustomer(f) }
		mux.Handle("GET /api/c/services", rc(cust.Services))
		mux.Handle("GET /api/c/services/{id}", rc(cust.ServiceDetail))
		mux.Handle("GET /api/c/img/s/{id}", middleware.RateLimit(240, time.Minute)(http.HandlerFunc(cust.ServiceImage)))
		mux.Handle("GET /api/c/favorites", rc(cust.Favorites))
		mux.Handle("POST /api/c/favorites/{id}", rc(cust.AddFavorite))
		mux.Handle("DELETE /api/c/favorites/{id}", rc(cust.RemoveFavorite))
		mux.Handle("POST /api/c/bookings", middleware.RateLimit(6, time.Minute)(rc(cust.CreateBooking)))
		mux.Handle("GET /api/c/bookings/{id}", rc(cust.BookingDetail))
		mux.Handle("GET /api/c/bookings/{id}/tech-photo", rc(cust.TechPhoto))
		mux.Handle("POST /api/c/bookings/{id}/rate", rc(cust.Rate))
		mux.Handle("GET /api/c/branches", rc(cust.Branches))
		secs["customer_app"].Set("GET /api/customer-app/services", "customer_app.get_services", cust.AdminServices)
		secs["customer_app"].Set("PUT /api/customer-app/services/{id}", "customer_app.put_services_by_id", cust.AdminSaveService)
		secs["customer_app"].Set("POST /api/customer-app/services/{id}/image", "customer_app.post_services_by_id_image", cust.AdminServiceImage)
		secs["customer_app"].Set("GET /api/customer-app/branches", "customer_app.get_branches", cust.Branches)
		secs["customer_app"].Set("POST /api/customer-app/branches", "customer_app.post_branches", cust.AdminSaveBranch)
		secs["customer_app"].Set("PUT /api/customer-app/branches/{id}", "customer_app.put_branches_by_id", cust.AdminSaveBranch)
		secs["customer_app"].Set("DELETE /api/customer-app/branches/{id}", "customer_app.delete_branches_by_id", cust.AdminDeleteBranch)
		secs["customer_app"].Set("GET /api/customer-app/mail-status", "customer_app.get_mail_status", cust.AdminMailStatus)
		secs["customer_app"].Set("POST /api/customer-app/mail-test", "customer_app.post_mail_test", cust.AdminMailTest)
		secs["customer_app"].Set("GET /api/customer-app/ratings", "customer_app.get_ratings", cust.AdminRatings)
	}

	// طلبات تغيير أيقونة الحضور — أي موظف يطلب، ومدير النظام بس يوافق/يرفض
	meApp.Set("POST /api/attendance-icon-requests", "me.attendance_icon_request", attendanceIconRequestHandler.Create)
	secs["attendance_icon_requests"].Set("GET /api/attendance-icon-requests", "attendance_icon_requests.get_root", attendanceIconRequestHandler.ListPending)
	secs["attendance_icon_requests"].Set("PUT /api/attendance-icon-requests/{id}/approve", "attendance_icon_requests.put_by_id_approve", attendanceIconRequestHandler.Approve)
	secs["attendance_icon_requests"].Set("PUT /api/attendance-icon-requests/{id}/reject", "attendance_icon_requests.put_by_id_reject", attendanceIconRequestHandler.Reject)

	// طلبات الكادر — مدير المشاريع (أو صاحب صلاحية إدارة المشاريع) يطلب، وإدارة الكوادر تلبي
	secs["staff_requests"].Set("POST /api/staff-requests", "staff_requests.post_root", staffRequestHandler.Create)
	secs["staff_requests"].Set("GET /api/staff-requests", "staff_requests.get_root", staffRequestHandler.List)
	secs["staff_requests"].Set("PUT /api/staff-requests/{id}/status", "staff_requests.put_by_id_status", staffRequestHandler.UpdateStatus)

	// مسؤول خدمة عام (تعميم فكرة أبو الجي بي اس لأي مجموعة خدمات) — الأدمن فقط يحدد المسؤوليات
	secs["service_managers"].Set("GET /api/service-managers", "service_managers.get_root", serviceManagerHandler.List)
	secs["service_managers"].Set("PUT /api/service-managers", "service_managers.put_root", serviceManagerHandler.Set)

	// تتبع الموقع الحي للفرق الميدانية
	// إرسال نقطة موقع مفتوح لأي موظف مسجل دخول (يرسل موقعه هو بس، محمي داخل الهاندلر
	// عبر EmployeeIDFromContext). القراءة (مين وين الحين، مسار أي موظف) محصورة بمن
	// يدير الفرق الميدانية فعلاً — حتى ما يقدر أي فني يتتبع مواقع زملائه.
	secs["location_pings"].Set("POST /api/location-pings", "location_pings.post_root", locationPingHandler.Create, middleware.SkipInStealth())
	secs["location_pings"].Set("GET /api/location-pings/latest", "location_pings.get_latest", locationPingHandler.Latest)
	secs["location_pings"].Set("GET /api/location-pings/path", "location_pings.get_path", locationPingHandler.Path)

	// تقييم الأداء (منفصل عن KPI مال الغرامات) — الليدر يقيّم فنييه، الإداري يقيّم الليدرات
	// ⚠️ الحارس يقبل kpi_management **أو** performance_review: الليدر
	// لازم يقيّم كادره، وأكثر الليدرات ما عندهم kpi_management (هاي
	// للغرامات المالية، شي ثاني تماماً). ربطهن بحارس واحد كان يمنع
	// الليدر من تقييم فريقه أصلاً.
	//
	// ومنو يقيّم منو محسوم بالخدمة (authorizeReview): الليدر فنيي
	// فريقه بس، وإداري الكوادر الليدرات بس. فتوسيع الحارس ما يفتح
	// باب — يخلي الي عنده الحق يوصله.
	//
	// (دازاي 10-31) الحارس كان يرفض الليدر والمراقب وإداري الكوادر الي ما عندهم
	// الصلاحيتين — التبويب يطلعلهم والحفظ يرد «ما عندك صلاحية» (وتنسجل مخالفة).
	// هسه يطابق authorizeReview: ليدر، MONITOR/HR_COORDINATOR، monitoring/auditing.
	secs["performance_reviews"].Set("POST /api/performance-reviews", "performance_reviews.post_root", performanceReviewHandler.Create)
	// كل تقييمات الأداء — سجلات شخصية، للإدارة والمراقبة.
	secs["performance_reviews"].Set("GET /api/performance-reviews", "performance_reviews.get_root", performanceReviewHandler.List)
	secs["performance_reviews"].Set("GET /api/performance-reviews/ratable", "performance_reviews.get_ratable", performanceReviewHandler.Ratable)
	secs["performance_reviews"].Set("GET /api/performance-reviews/employee/{employeeId}", "performance_reviews.get_employee_by_employeeid", performanceReviewHandler.ListForEmployee)
	secs["performance_reviews"].Set("GET /api/performance-reviews/my-bookings", "performance_reviews.get_my_bookings", performanceReviewHandler.MyBookings)
	// «تقييم بين الإداريين» — ADMIN/OWNER/MONITOR بس (منو يقارن
	// إداري الكوادر ببعض)، مو إداري الكوادر نفسه ولا التيم ليدر.
	secs["performance_reviews"].Set("GET /api/performance-reviews/evaluator-leaderboard", "performance_reviews.get_evaluator_leaderboard", performanceReviewHandler.EvaluatorLeaderboard)

	// المشتريات (procurement)
	secs["procurement"].Set("GET /api/procurement", "procurement.get_root", procurementHandler.List)
	secs["procurement"].Set("GET /api/procurement/stats", "procurement.get_stats", procurementHandler.Stats)
	// إنشاء الطلب مو مربوط بصلاحية "procurement" العامة (الي تفتح الاطّلاع بس) —
	// كل نوع طلب (شخصي/زبون) له صلاحية مستقلة تُفحص داخل الهاندلر نفسه.
	requestsApp.Set("POST /api/procurement", "requests.procurement", procurementHandler.Create)
	secs["procurement"].Set("PUT /api/procurement/{id}/status", "procurement.put_by_id_status", procurementHandler.UpdateStatus)
	secs["procurement"].Set("PUT /api/procurement/{id}/fulfill", "procurement.put_by_id_fulfill", procurementHandler.Fulfill)
	// (ع) 10-15: رقابة المشتريات
	purchaseControl := handler.NewPurchaseControlHandler(db, func(msg string) {
		_ = notificationRepo.CreateForRolesOrPermission([]string{"ADMIN", "MONITOR"}, "monitoring", "purchase_alert", msg)
	})
	procurementHandler.SetAfterFulfill(purchaseControl.AfterFulfill)
	procurementHandler.SetBeforeFulfill(purchaseControl.BeforeFulfill)
	expenseHandler.SetAfterPurchase(purchaseControl.AfterExpensePurchase)
	// (ون بيس 11-10) requirePurchaseWatch انتقل لقاعدة بالنواة (internal/apps/sections)
	go purchaseControl.RunStandardsLoop()
	secs["procurement_standards"].Set("GET /api/procurement-standards", "procurement_standards.get_root", purchaseControl.Standards)
	secs["procurement_standards"].Set("PUT /api/procurement-standards", "procurement_standards.put_root", purchaseControl.SetStandard)
	secs["procurement_standards"].Set("POST /api/procurement-standards/accept", "procurement_standards.post_accept", purchaseControl.AcceptSuggestion)
	secs["procurement_policy"].Set("PUT /api/procurement-policy/ceiling", "procurement_policy.put_ceiling", purchaseControl.SetCeiling)
	secs["procurement"].Set("GET /api/procurement/control", "procurement.get_control", purchaseControl.Control)
	secs["procurement"].Set("POST /api/procurement/{id}/review", "procurement.post_by_id_review", purchaseControl.Review)
	secs["procurement"].Set("PUT /api/procurement/{id}/verify", "procurement.put_by_id_verify", purchaseControl.Verify)
	secs["procurement"].Set("GET /api/procurement/prices", "procurement.get_prices", purchaseControl.Prices)
	secs["procurement"].Set("GET /api/procurement/alerts", "procurement.get_alerts", purchaseControl.Alerts)
	secs["procurement_alerts"].Set("PUT /api/procurement-alerts/{id}", "procurement_alerts.put_by_id", purchaseControl.ReviewAlert)
	secs["procurement_policy"].Set("GET /api/procurement-policy", "procurement_policy.get_root", purchaseControl.GetPolicy)
	secs["procurement_policy"].Set("PUT /api/procurement-policy", "procurement_policy.put_root", purchaseControl.SetPolicy)
	secs["procurement_quotes"].Set("GET /api/procurement-quotes/{id}", "procurement_quotes.get_by_id", purchaseControl.Quotes)
	secs["procurement_monthly"].Set("GET /api/procurement-monthly", "procurement_monthly.get_root", purchaseControl.Monthly)
	secs["procurement_bias"].Set("GET /api/procurement-bias", "procurement_bias.get_root", purchaseControl.Bias)
	// (ع) 10-15: علاقات الموردين — العلاقات العامة + وحدة التقنيين
	supplierRelations := handler.NewSupplierRelationsHandler(db, func(msg string) {
		_ = notificationRepo.CreateForRolesOrPermission([]string{"ADMIN", "OWNER", "PUBLIC_RELATIONS"}, "supplier_relations", "supplier_relations", msg)
	})
	go supplierRelations.RunReminders()
	// (ون بيس 11-10) requireSupplierRel انتقل لقاعدة بالنواة (internal/apps/sections)
	secs["supplier_relations"].Set("GET /api/supplier-relations", "supplier_relations.get_root", supplierRelations.List)
	secs["supplier_relations"].Set("POST /api/supplier-relations/{kind}/{id}/contact", "supplier_relations.post_by_kind_by_id_contact", supplierRelations.AddContact)
	secs["supplier_relations"].Set("GET /api/supplier-relations/{kind}/{id}/timeline", "supplier_relations.get_by_kind_by_id_timeline", supplierRelations.Timeline)
	secs["supplier_relations"].Set("PUT /api/supplier-relations/policy", "supplier_relations.put_policy", supplierRelations.SetPolicy)
	secs["suppliers"].Set("GET /api/suppliers/scores", "suppliers.get_scores", purchaseControl.SupplierScores)

	// الموردون (suppliers)
	secs["suppliers"].Set("GET /api/suppliers/specialties", "suppliers.get_specialties", supplierHandler.ListSpecialties)
	secs["suppliers"].Set("POST /api/suppliers/specialties", "suppliers.post_specialties", supplierHandler.CreateSpecialty)
	secs["suppliers"].Set("DELETE /api/suppliers/specialties/{id}", "suppliers.delete_specialties_by_id", supplierHandler.DeleteSpecialty)
	// حل روابط الخرائط (خصوصاً المختصرة مثل maps.app.goo.gl الي ما تحتوي
	// إحداثيات) — لازم يمر بالسيرفر لأن المتصفح ما يقدر يتبع التحويل بسبب CORS.
	secs["geo"].Set("GET /api/geo/resolve-map-link", "geo.get_resolve_map_link", mapLinkHandler.Resolve)

	// سياسة الخصوصية — القراءة والموافقة لأي موظف، والإدارة بصلاحية مستقلة
	// (ون بيس 11-10) requirePrivacyMgmt انتقل لقاعدة بالنواة (internal/apps/sections)
	secs["privacy_policy"].Set("GET /api/privacy-policy", "privacy_policy.get_root", privacyPolicyHandler.List)
	secs["privacy_policy"].Set("GET /api/privacy-policy/status", "privacy_policy.get_status", privacyPolicyHandler.Status)
	meApp.Set("POST /api/privacy-policy/accept", "me.privacy_accept", privacyPolicyHandler.Accept)
	// 📜 سجل موافقات المبادئ (دازاي 10-29) — الموظف نفسه أو المالك/مدير النظام
	secs["privacy_policy"].Set("GET /api/privacy-policy/acceptances/{employeeId}", "privacy_policy.get_acceptances_by_employeeid", privacyPolicyHandler.Acceptances)
	secs["privacy_policy"].Set("POST /api/privacy-policy", "privacy_policy.post_root", privacyPolicyHandler.Create)
	secs["privacy_policy"].Set("PUT /api/privacy-policy/{id}", "privacy_policy.put_by_id", privacyPolicyHandler.Update)
	secs["privacy_policy"].Set("DELETE /api/privacy-policy/{id}", "privacy_policy.delete_by_id", privacyPolicyHandler.Delete)

	// بحث المناطق يمر من سيرفرنا — خدمة الخرائط تحجب المتصفحات
	secs["geo"].Set("GET /api/geo/search", "geo.get_search", geoHandler.Search)

	// ── موردين التقنيين (قرار (ع) 10-06): التقني يضيف، والمدير يختار لكل تقني ──
	techSupplierHandler := handler.NewTechSupplierHandler(repository.NewTechSupplierRepository(db))
	secs["tech_suppliers"].Set("GET /api/tech-suppliers/can", "tech_suppliers.get_can", techSupplierHandler.Can)
	secs["tech_suppliers"].Set("GET /api/tech-suppliers/mine", "tech_suppliers.get_mine", techSupplierHandler.Mine)
	requestsApp.Set("POST /api/tech-suppliers", "requests.tech_supplier", techSupplierHandler.Create)
	secs["tech_suppliers"].Set("GET /api/tech-suppliers", "tech_suppliers.get_root", techSupplierHandler.All)
	secs["tech_suppliers"].Set("PUT /api/tech-suppliers/{id}", "tech_suppliers.put_by_id", techSupplierHandler.Update)
	secs["tech_suppliers"].Set("DELETE /api/tech-suppliers/{id}", "tech_suppliers.delete_by_id", techSupplierHandler.Delete)
	secs["tech_suppliers"].Set("PUT /api/tech-suppliers/access/{employeeId}", "tech_suppliers.put_access_by_employeeid", techSupplierHandler.SetAccess)
	secs["suppliers"].Set("GET /api/suppliers", "suppliers.get_root", supplierHandler.List)
	secs["suppliers"].Set("POST /api/suppliers", "suppliers.post_root", supplierHandler.Create)
	secs["suppliers"].Set("PUT /api/suppliers/{id}", "suppliers.put_by_id", supplierHandler.Update)
	// الحذف داخل نفس صلاحية إدارة الموردين — منو يقدر يضيف ويعدل يقدر يحذف،
	// حتى ما يضطر يرجع للأدمن على كل مورد غلط.
	secs["suppliers"].Set("DELETE /api/suppliers/{id}", "suppliers.delete_by_id", supplierHandler.Delete)
	secs["suppliers"].Set("POST /api/suppliers/{id}/rate", "suppliers.post_by_id_rate", supplierHandler.Rate)

	// عروض الأسعار (quotations)
	secs["quotations"].Set("GET /api/quotations", "quotations.get_root", quotationHandler.List)
	secs["quotations"].Set("GET /api/quotations/{id}", "quotations.get_by_id", quotationHandler.Get)
	// النسخ المؤرشفة — نفس حارس قراءة العرض: الي يشوف العرض يشوف تاريخه
	secs["quotation_by_project"].Set("GET /api/quotation-by-project/{projectId}", "quotation_by_project.get_by_projectid", quotationHandler.ByProject)
	secs["quotations"].Set("GET /api/quotations/{id}/versions", "quotations.get_by_id_versions", quotationHandler.Versions)

	// ═══ تعلّم أسعار الخدمات ═══
	// 🔴 **للمحاسب والمراقب والمالك والمدير وبس** — قرار صاحب النظام
	// (١٦ أيلول ٢٠٢٦): «المفروض تطلع فقط للمراقب وللمحاسب… يعني أكيد
	// تطلع للمالك وللمدير».
	//
	// ⚠️ چانت مفتوحة لـ`execution_cost` و`invoice_manual` بحجة إن
	// صاحب الفاتورة اليدوية يشوف عيّناته — والنتيجة إن **فني** عنده
	// حساب الكلفة طلعله تبويب أسعار الشركة كله. وسعر الخدمة قرار
	// إدارة، مو معلومة كل من يسوي فاتورة.
	// (ون بيس 11-10) requireServicePriceRead انتقل لقاعدة بالنواة (internal/apps/sections)
	// القراءة لصاحب حساب الكلفة والمحاسب والمراقب والمالك — الي
	// يشتغل بالأسعار يشوف عيّناتها. المراجعة **للمحاسب** والاعتماد
	// **للمالك ومدير النظام** — ترتيب (ع) نصاً.
	secs["service_prices"].Set("GET /api/service-prices/samples", "service_prices.get_samples", servicePriceHandler.Samples)
	secs["service_prices"].Set("GET /api/service-prices/suggestions", "service_prices.get_suggestions", servicePriceHandler.Suggestions)
	secs["service_prices"].Set("PUT /api/service-prices/suggestions/{id}/review", "service_prices.put_suggestions_by_id_review", servicePriceHandler.Review)
	// ⚠️ الاعتماد للمالك ومدير النظام وبس: سعر الخدمة قرار إدارة،
	// ولو انفتح لغيرهم يصير المعدّل الآلي سعراً رسمياً بلا قرار.
	secs["service_prices"].Set("PUT /api/service-prices/suggestions/{id}/decide", "service_prices.put_suggestions_by_id_decide", servicePriceHandler.Decide)
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
	// شروط وأحكام عرض السعر: يقراها أي واحد يسوي عرض، ويعدّلها صاحب صلاحية عرض السعر بس.
	quotationTermsHandler := handler.NewQuotationTermsHandler(db)
	secs["quotation_terms"].Set("GET /api/quotation-terms", "quotation_terms.get_root", quotationTermsHandler.List, allowQuotationOrDelegate)
	secs["quotation_terms"].Set("PUT /api/quotation-terms", "quotation_terms.put_root", quotationTermsHandler.Save, requireQuotationAccess)
	// (دازاي 11-06) قوالب الملحق الفني لعرض المشاريع
	annexHandler := handler.NewQuotationAnnexHandler(db)
	secs["quotation_annex_templates"].Set("GET /api/quotation-annex-templates", "quotation_annex_templates.get_root", annexHandler.List, allowQuotationOrDelegate)
	secs["quotation_annex_templates"].Set("PUT /api/quotation-annex-templates", "quotation_annex_templates.put_root", annexHandler.Save, requireQuotationAccess)
	secs["quotations"].Set("POST /api/quotations", "quotations.post_root", quotationHandler.Create, allowQuotationOrDelegate)
	secs["quotations"].Set("PUT /api/quotations/{id}", "quotations.put_by_id", quotationHandler.Update, allowQuotationOrDelegate)
	secs["quotations"].Set("DELETE /api/quotations/{id}", "quotations.delete_by_id", quotationHandler.Delete)

	// المنتجات (products) — التعديل/الحذف محصوران بصلاحية content_technician، والإنشاء
	// يطابق نفس التعديل (نفس المورد) بدل ما يبقى مفتوح لأي موظف مسجل دخول.
	secs["products"].Set("GET /api/products", "products.get_root", productHandler.List)
	secs["products"].Set("POST /api/products", "products.post_root", productHandler.Create)
	secs["products"].Set("PUT /api/products/{id}", "products.put_by_id", productHandler.Update)
	specsSuggest := &handler.ProductSpecsSuggest{}
	// (11-10 بلاغ #17) مؤمل يريد المواصفات تطلع من الموديل وهو يكتب عرض المشاريع — صاحب عروض الأسعار هم.
	secs["products"].Set("POST /api/products/specs-suggest", "products.post_specs_suggest", specsSuggest.Suggest, guards.AnyPermission("content_technician", service.QuotationPermCreate, service.QuotationPermEditOwn, service.QuotationPermManageAll, "quotation_system"))
	secs["products"].Set("DELETE /api/products/{id}", "products.delete_by_id", productHandler.Delete)

	// نظام GPS — عملاء / شرائح SIM / طلبات الأجهزة / التجديد / الصيانة / الأسعار / الإحصائيات
	// الإنشاء/التعديل بكل موارد GPS يتطلب صلاحية "gps_system" فعلياً (نفس الصلاحية
	// الي تفتح كل صفحات نظام GPS بالواجهة) — القراءة تبقى مفتوحة لأي مسجل دخول.
	// POST هنا مفتوحة لأي مسجل دخول (requireAuth بس) — هذا التسجيل الأساسي مو عملية
	// إدارية، وموظف المبيعات (GpsPurchase.tsx) لازم يقدر يسجل بيانات الزبون وهو
	// يرسل طلب شراء جهاز GPS جديد، قبل حتى ما يوصل الطلب لإداري GPS للموافقة.
	secs["gps"].Set("GET /api/gps/customers", "gps.get_customers", gpsHandler.ListCustomers)
	secs["gps"].Set("POST /api/gps/customers", "gps.post_customers", gpsHandler.CreateCustomer)
	secs["gps"].Set("PUT /api/gps/customers/{id}", "gps.put_customers_by_id", gpsHandler.UpdateCustomer)

	// ── الدوار ────────────────────────────────────────────────────────────────
	// الإدارة والصرف والتدقيق: صلاحية "الدوار" (المحاسب). أما رصيد الموظف
	// نفسه ورفع تسويته فمفتوحين لأي موظف مسجّل دخول — كل واحد يشوف حركاته هو بس.
	// حساب تكاليف الشد — تفصيلي لكل الكوادر، ضمن خانة الحسابات
	// صلاحية «حساب تكاليف الشد» تظهر البند بالقائمة — فلازم تفتح المسار كمان،
	// وإلا الموظف يضغط وينرفض ويتسجّل عليه تجاوز.
	financeApp.Set("GET /api/finance/gps-install-costs", "finance.gps_costs", gpsInstallCostHandler.Summary)

	// ── الإجازات ──────────────────────────────────────────────────────────────
	// أي موظف يقدّم ويشوف طلباته. البت بالطلبات محصور بالمخوّل حسب مسار
	// الكادر (يُتحقق داخل الهاندلر لأنه يعتمد على مسار الطلب نفسه).
	meApp.Set("POST /api/leaves", "me.leave_request", leaveHandler.Create)
	secs["leaves"].Set("GET /api/leaves/mine", "leaves.get_mine", leaveHandler.Mine)
	meApp.Set("DELETE /api/leaves/{id}", "me.leave_cancel", leaveHandler.Cancel)
	secs["leaves"].Set("GET /api/leaves/inbox", "leaves.get_inbox", leaveHandler.Inbox)
	secs["leaves"].Set("GET /api/leaves/pending-count", "leaves.get_pending_count", leaveHandler.PendingCount)
	// 🗓️ «إجازات الموظفين» (ع 10-23) — نفس حارس البند بالقائمة: المدير/المراقب أو إدارة الكوادر أو الموافقة على الإجازات
	// (ون بيس 11-10) requireLeaveRegister انتقل لقاعدة بالنواة (internal/apps/sections)
	secs["leaves"].Set("GET /api/leaves/register", "leaves.get_register", leaveHandler.Register)
	fieldApp.Set("PUT /api/leaves/{id}/preliminary", "field.leave_preliminary", leaveHandler.Preliminary)
	secs["leaves"].Set("PUT /api/leaves/{id}/decide", "leaves.put_by_id_decide", leaveHandler.Decide, courier.Wrap("leave"))

	// أرقام اللوحة الرئيسية بدون سحب أرشيف الشركة كامل
	secs["dashboard"].Set("GET /api/dashboard/summary", "dashboard.get_summary", dashboardHandler.Summary)
	// نبض اليوم: أرقام الشاشة الرئيسية للإداري. مجاميع بس — ماكو بيانات
	// زبون ولا مبالغ، فما يحتاج صلاحية أضيق من تسجيل الدخول.
	secs["dashboard"].Set("GET /api/dashboard/today-pulse", "dashboard.get_today_pulse", dashboardHandler.TodayPulse)

	// مجاميع المحاسب والمراقب — تنحسب بقاعدة البيانات بدل تنزيل الأرشيف
	// اثنين يشوفون هذي الأرقام: المحاسب بدوره، والمراقب بصلاحيته. لازم
	// الاثنين سوه — حارس بالدور بس يرفض المراقب، وحارس بالصلاحية بس
	// يرفض المحاسب الجديد الي ما انمنحت له صلاحية صريحة بعد. وأي رفض
	// هنا مو مجرد ٤٠٣: النظام يعدّه محاولة تلاعب ويقفل الحساب بعد ٣ مرات،
	// والصفحة تعاود الطلب لحالها — يعني القفل يجي بثواني.
	secs["dashboard"].Set("GET /api/dashboard/finance-summary", "dashboard.get_finance_summary", dashboardHandler.FinanceSummary)

	// ═══ نظام الطاقة الشمسية (منقول من Google Sheets) ═══
	// القراءة لأي موظف مسجّل: الفني بالموقع يحتاج يشوف مكوّنات المنظومة
	// ومواصفاتها وهو شغال. التعديل والتجهيز بصلاحية "solar_system".
	// (ون بيس 11-10) requireSolar انتقل لقاعدة بالنواة (internal/apps/sections)

	// ═══ برامج التدريب ═══
	//
	// القراءة لأي موظف (يشوف تدريباته)، والإدارة بصلاحية
	// training_manage المستقلة.
	//
	// ⚠️ قبل، الإدارة جانت تنفتح بـstaff_management أو بمجرد إن دوره
	// HR_COORDINATOR. يعني تنطي موظف صلاحية إدارة الكوادر فينفتحله
	// معاها تعديل برامج التدريب بلا ما تقصد.
	secs["training_programs"].Set("GET /api/training-programs", "training_programs.get_root", trainingProgramHandler.List)
	// ═══ مركز التدريب — قرار (ع) 10-13 ═══
	trainingCenter := handler.NewTrainingCenterHandler(db)
	tcRoute := func(pattern string, f http.HandlerFunc, manage bool) {
		if manage {
			trainingCenterApp.Add(pattern, core.PermRule(pattern+" — إدارة التدريب", "training_manage"), f)
		} else {
			trainingCenterApp.Add(pattern, core.Rule{Label: pattern + " — أي موظف (تدريبه هو)", Anyone: true}, f)
		}
	}
	tcRoute("GET /api/training-center/insights", trainingCenter.Insights, true)
	tcRoute("GET /api/training-center/tracks", trainingCenter.Tracks, true)
	tcRoute("GET /api/training-center/curriculum", trainingCenter.Curriculum, true)
	tcRoute("POST /api/training-center/curriculum", trainingCenter.AddCurriculum, true)
	tcRoute("DELETE /api/training-center/curriculum/{id}", trainingCenter.DeleteCurriculum, true)
	tcRoute("GET /api/training-center/questions", trainingCenter.Questions, true)
	tcRoute("POST /api/training-center/questions", trainingCenter.AddQuestion, true)
	tcRoute("POST /api/training-center/questions/generate", trainingCenter.GenerateQuestions, true)
	tcRoute("POST /api/training-center/questions/{id}/approve", trainingCenter.ApproveQuestion, true)
	tcRoute("DELETE /api/training-center/questions/{id}", trainingCenter.DeleteQuestion, true)
	tcRoute("POST /api/training-center/exams", trainingCenter.CreateExam, true)
	tcRoute("GET /api/training-center/exams/report", trainingCenter.ExamReport, true)
	tcRoute("GET /api/training-center/plans", trainingCenter.Plans, true)
	tcRoute("POST /api/training-center/plans", trainingCenter.AddPlan, true)
	tcRoute("POST /api/training-center/plans/{id}/inspect", trainingCenter.InspectPlan, true)
	tcRoute("DELETE /api/training-center/plans/{id}", trainingCenter.DeletePlan, true)
	tcRoute("GET /api/training-center/mentors", trainingCenter.Mentors, true)
	tcRoute("PUT /api/training-center/mentors/{id}", trainingCenter.SetMentor, true)
	tcRoute("GET /api/training-center/my-exam", trainingCenter.MyExam, false)
	tcRoute("POST /api/training-center/my-exam/start", trainingCenter.StartExam, false)
	tcRoute("POST /api/training-center/my-exam/submit", trainingCenter.SubmitExam, false)

	secs["training_programs"].Set("POST /api/training-programs", "training_programs.post_root", trainingProgramHandler.Create)
	secs["training_programs"].Set("PUT /api/training-programs/{id}", "training_programs.put_by_id", trainingProgramHandler.Update)
	secs["training_programs"].Set("PUT /api/training-programs/{id}/complete", "training_programs.put_by_id_complete", trainingProgramHandler.Complete)
	secs["training_programs"].Set("DELETE /api/training-programs/{id}", "training_programs.delete_by_id", trainingProgramHandler.Delete)

	// تقرير ١ (B12): تركيبات الزبائن وأسعارها والمخزن — لأصحاب الطاقة الشمسية بس.
	// (كتالوج المنظومات /solar/systems يبقى مفتوح لأن الحجز وعرض السعر يقرونه.)
	// (ون بيس 11-10) requireSolarView انتقل لقاعدة بالنواة (internal/apps/sections)
	solarHandler.SetCostGate(func(r *http.Request) bool {
		switch middleware.RoleFromContext(r) {
		case "ADMIN", "OWNER", "MONITOR", "TECHNICIAN", "SERVICE_MANAGER":
			return true
		}
		ok, _ := permissionRepo.HasPermission(middleware.EmployeeIDFromContext(r), "solar_system")
		return ok
	})
	secs["solar"].Set("GET /api/solar/stats", "solar.get_stats", solarHandler.Stats)
	secs["solar"].Set("GET /api/solar/low-stock", "solar.get_low_stock", solarHandler.LowStock)

	secs["solar"].Set("GET /api/solar/components", "solar.get_components", solarHandler.ListComponents)
	secs["solar"].Set("POST /api/solar/components", "solar.post_components", solarHandler.CreateComponent)
	secs["solar"].Set("PUT /api/solar/components/{id}", "solar.put_components_by_id", solarHandler.UpdateComponent)
	secs["solar"].Set("DELETE /api/solar/components/{id}", "solar.delete_components_by_id", solarHandler.DeleteComponent)

	secs["solar"].Set("GET /api/solar/systems", "solar.get_systems", solarHandler.ListSystems)
	secs["solar"].Set("GET /api/solar/systems/{id}", "solar.get_systems_by_id", solarHandler.GetSystem)
	secs["solar"].Set("POST /api/solar/systems", "solar.post_systems", solarHandler.CreateSystem)
	secs["solar"].Set("PUT /api/solar/systems/{id}", "solar.put_systems_by_id", solarHandler.UpdateSystem)
	secs["solar"].Set("DELETE /api/solar/systems/{id}", "solar.delete_systems_by_id", solarHandler.DeleteSystem)

	// التجهيز يخصم من المخزن فعلياً — صلاحية إجبارية
	secs["solar"].Set("POST /api/solar/systems/{id}/process", "solar.post_systems_by_id_process", solarHandler.ProcessSystem)
	secs["solar"].Set("GET /api/solar/installations", "solar.get_installations", solarHandler.ListInstallations)
	secs["solar"].Set("PUT /api/solar/installations/{id}/contacted", "solar.put_installations_by_id_contacted", solarHandler.MarkContacted)

	// ── مراقبة النسخ الاحتياطية — للمالك وحده ──
	// requireOwner يرجّع 404 مو 403: أي حساب ثاني (حتى ADMIN) ما يعرف
	// إن هذا المسار موجود أصلاً. لا تضيف هنا أي middleware ثاني.
	secs["owner"].Set("GET /api/owner/backups", "owner.get_backups", backupHandler.Overview)

	// ═══ مختبر المحاكاة — للمالك وحده ═══
	//
	financeApp.Set("GET /api/funds", "finance.funds", revolvingFundHandler.ListFunds)
	financeApp.Set("PUT /api/funds/{id}", "finance.fund_update", revolvingFundHandler.UpdateFund)
	financeApp.Set("POST /api/funds/{id}/topup", "finance.fund_topup", revolvingFundHandler.Topup, idem)
	financeApp.Set("POST /api/funds/disburse", "finance.fund_disburse", revolvingFundHandler.Disburse, idem)
	financeApp.Set("GET /api/funds/balances", "finance.fund_balances", revolvingFundHandler.Balances)
	financeApp.Set("GET /api/funds/transactions", "finance.fund_txns", revolvingFundHandler.Txns)
	financeApp.Set("GET /api/funds/transactions/{id}/receipt", "finance.fund_receipt", revolvingFundHandler.Receipt)
	financeApp.Set("POST /api/funds/settlements/for/{employeeId}", "finance.settle_for", revolvingFundHandler.SubmitSettlementFor, idem)
	financeApp.Set("PUT /api/funds/settlements/{id}/review", "finance.settle_review", revolvingFundHandler.ReviewSettlement, idem)
	financeApp.Set("PUT /api/funds/settlements/{id}/discharge", "finance.settle_discharge", revolvingFundHandler.Discharge, idem)
	financeApp.Set("GET /api/funds/discharge-accounts", "finance.discharge_accts", revolvingFundHandler.DischargeAccounts)

	financeApp.Set("GET /api/funds/my-balance", "finance.my_balance", revolvingFundHandler.MyBalance)
	financeApp.Set("GET /api/funds/my-transactions", "finance.my_txns", revolvingFundHandler.MyTxns)
	financeApp.Set("POST /api/funds/settlements", "finance.my_settle", revolvingFundHandler.SubmitSettlement, idem)

	secs["gps"].Set("GET /api/gps/sims", "gps.get_sims", gpsHandler.ListSims)
	secs["gps"].Set("POST /api/gps/sims", "gps.post_sims", gpsHandler.CreateSim)
	secs["gps"].Set("PUT /api/gps/sims/{id}", "gps.put_sims_by_id", gpsHandler.UpdateSim)

	// دورة حياة الشريحة: الربط والتحرير والحرق كلها شغل مسؤول الجي بي اس.
	secs["gps"].Set("GET /api/gps/sims/available", "gps.get_sims_available", gpsHandler.ListAvailableSims)
	secs["gps"].Set("POST /api/gps/sims/{id}/assign", "gps.post_sims_by_id_assign", gpsHandler.AssignSim)
	secs["gps"].Set("POST /api/gps/sims/{id}/release", "gps.post_sims_by_id_release", gpsHandler.ReleaseSim)
	secs["gps"].Set("POST /api/gps/sims/{id}/burn", "gps.post_sims_by_id_burn", gpsHandler.BurnSim)

	// متابعة التجديد: القائمة يشوفها الاثنين (مهندس الجودة يتصل، ومسؤول
	// الجي بي اس يشوف منو يحتاج حرق)، وتسجيل نتيجة الاتصال لمهندس الجودة.
	// نتائج الجي بي اس للمراقب — قراءة فقط، بحارس المراقب نفسه.
	secs["gps"].Set("GET /api/gps/monitor-snapshot", "gps.get_monitor_snapshot", gpsHandler.MonitorSnapshot)
	secs["gps"].Set("GET /api/gps/subscriptions/follow-up", "gps.get_subscriptions_follow_up", gpsHandler.SubscriptionFollowUps)
	secs["gps"].Set("GET /api/gps/devices/{id}/follow-up", "gps.get_devices_by_id_follow_up", gpsHandler.ListFollowUps)
	secs["gps"].Set("POST /api/gps/devices/{id}/follow-up", "gps.post_devices_by_id_follow_up", gpsHandler.CreateFollowUp)

	// نفس الشي هنا: صف GpsDeviceRequest الجديد يبدأ status='PENDING' افتراضياً
	// بالمخطط نفسه (schema_base.go) — هذا "طلب" بانتظار مراجعة إداري GPS، مو جهاز
	// مفعّل فعلياً، فتقييد الإنشاء بصلاحية gps_system كان يمنع بالضبط سيناريو تقديم
	// الطلب من موظف مبيعات ما عنده هذي الصلاحية. الموافقة (PUT) تبقى محمية.
	secs["gps"].Set("GET /api/gps/devices", "gps.get_devices", gpsHandler.ListDevices)
	secs["gps"].Set("POST /api/gps/devices", "gps.post_devices", gpsHandler.CreateDevice)
	secs["gps"].Set("PUT /api/gps/devices/{id}", "gps.put_devices_by_id", gpsHandler.UpdateDevice)

	secs["gps"].Set("GET /api/gps/renewals", "gps.get_renewals", gpsHandler.ListRenewals)
	secs["gps"].Set("POST /api/gps/renewals", "gps.post_renewals", gpsHandler.CreateRenewal)
	secs["gps"].Set("PUT /api/gps/renewals/{id}", "gps.put_renewals_by_id", gpsHandler.UpdateRenewal)

	secs["gps"].Set("GET /api/gps/maintenance", "gps.get_maintenance", gpsHandler.ListMaintenance)
	secs["gps"].Set("POST /api/gps/maintenance", "gps.post_maintenance", gpsHandler.CreateMaintenance)
	secs["gps"].Set("PUT /api/gps/maintenance/{id}", "gps.put_maintenance_by_id", gpsHandler.UpdateMaintenance)

	secs["gps"].Set("GET /api/gps/settings", "gps.get_settings", gpsHandler.ListSettings)
	secs["gps"].Set("PUT /api/gps/settings", "gps.put_settings", gpsHandler.UpsertSettings)

	secs["gps"].Set("GET /api/gps/stats", "gps.get_stats", gpsHandler.Stats)

	// الإحصائيات العامة (stats) — إيرادات وأرباح الشركة وتفصيل أداء كل
	// موظف بالاسم. ⚠️⚠️ كان بلا `requireAdmin` — أي موظف مسجّل دخول
	// (حتى فني بالميدان) يقدر يستدعيه مباشرة ويحصل على كل هذا رغم إن
	// الواجهة (`StatsPage.tsx`) تقفل على غير ADMIN. الحارس هنا صار
	// يطابق القفل الفعلي بالواجهة.
	secs["stats"].Set("GET /api/stats", "stats.get_root", statsHandler.Overview)

	// إدارة المركبات — وقود/تنظيف/تبديل زيت، أعطال وأضرار، حالة شهرية
	// قائمة مبسطة لاختيار سيارة لحجز — بلا صلاحية إدارة الأسطول
	secs["vehicles"].Set("GET /api/vehicles/options", "vehicles.get_options", vehicleHandler.Options)
	secs["vehicles"].Set("GET /api/vehicles", "vehicles.get_root", vehicleHandler.List)
	secs["vehicles"].Set("POST /api/vehicles", "vehicles.post_root", vehicleHandler.Create)
	secs["vehicles"].Set("GET /api/vehicles/{id}/logs", "vehicles.get_by_id_logs", vehicleHandler.ListLogs)
	secs["vehicles"].Set("POST /api/vehicles/{id}/logs", "vehicles.post_by_id_logs", vehicleHandler.CreateLog)
	secs["vehicles"].Set("PUT /api/vehicles/{id}/logs/{logId}", "vehicles.put_by_id_logs_by_logid", vehicleHandler.UpdateLog)
	secs["vehicles"].Set("DELETE /api/vehicles/{id}/logs/{logId}", "vehicles.delete_by_id_logs_by_logid", vehicleHandler.DeleteLog)
	// صورة الوصل بمسار مستقل — القوائم ترجع علم وجودها فقط حتى ما تنبلع
	// ميغابايتات base64 بكل جلب لسجلات السيارة.
	secs["vehicles"].Set("GET /api/vehicles/{id}/logs/{logId}/receipt-photo", "vehicles.get_by_id_logs_by_logid_receipt_photo", vehicleHandler.LogReceiptPhoto)
	// كم مرة عبّأ كل موظف بالشهر
	secs["vehicles"].Set("GET /api/vehicles/fuel-stats/by-employee", "vehicles.get_fuel_stats_by_employee", vehicleHandler.EmployeeFuelStats)
	secs["vehicles"].Set("GET /api/vehicles/{id}/incidents", "vehicles.get_by_id_incidents", vehicleHandler.ListIncidents)
	secs["vehicles"].Set("POST /api/vehicles/{id}/incidents", "vehicles.post_by_id_incidents", vehicleHandler.CreateIncident)
	secs["vehicle_incidents"].Set("PUT /api/vehicle-incidents/{id}", "vehicle_incidents.put_by_id", vehicleHandler.UpdateIncident)
	secs["vehicles"].Set("GET /api/vehicles/{id}/monthly-status", "vehicles.get_by_id_monthly_status", vehicleHandler.ListMonthlyStatus)
	secs["vehicles"].Set("POST /api/vehicles/{id}/monthly-status", "vehicles.post_by_id_monthly_status", vehicleHandler.SetMonthlyStatus)
	secs["vehicles"].Set("POST /api/vehicles/{id}/ratings", "vehicles.post_by_id_ratings", vehicleHandler.CreateDailyRating)
	secs["vehicles"].Set("GET /api/vehicles/{id}/ratings", "vehicles.get_by_id_ratings", vehicleHandler.ListDailyRatings)
	// تعديل/حذف بيانات السيارة الأساسية محصور بالمالك/الأدمن (requireAdmin) فقط —
	// أضيق من صلاحية "إدارة المركبات" العامة (vehicle_management) المستخدمة لبقية
	// عمليات السيارة (سجلات، وثائق، صور...)، بناءً على طلب صريح من مالك النظام.
	secs["vehicles"].Set("PUT /api/vehicles/{id}", "vehicles.put_by_id", vehicleHandler.Update)
	secs["vehicles"].Set("DELETE /api/vehicles/{id}", "vehicles.delete_by_id", vehicleHandler.Delete)

	// ملف السيارة الكامل: وثائق وصور
	secs["vehicles"].Set("GET /api/vehicles/{id}/documents", "vehicles.get_by_id_documents", vehicleHandler.ListDocuments)
	secs["vehicles"].Set("POST /api/vehicles/{id}/documents", "vehicles.post_by_id_documents", vehicleHandler.CreateDocument)
	secs["vehicles"].Set("DELETE /api/vehicles/{id}/documents/{docId}", "vehicles.delete_by_id_documents_by_docid", vehicleHandler.DeleteDocument)
	secs["vehicles"].Set("GET /api/vehicles/{id}/photos", "vehicles.get_by_id_photos", vehicleHandler.ListPhotos)
	secs["vehicles"].Set("POST /api/vehicles/{id}/photos", "vehicles.post_by_id_photos", vehicleHandler.CreatePhoto)
	secs["vehicles"].Set("DELETE /api/vehicles/{id}/photos/{photoId}", "vehicles.delete_by_id_photos_by_photoid", vehicleHandler.DeletePhoto)

	// نظام المهمة: كل خروج سيارة يصير سجل مهمة متابَع (سبب، وجهة، عداد، ركاب)
	secs["vehicle_missions"].Set("POST /api/vehicle-missions", "vehicle_missions.post_root", vehicleMissionHandler.Start)
	secs["vehicle_missions"].Set("PUT /api/vehicle-missions/{id}/end", "vehicle_missions.put_by_id_end", vehicleMissionHandler.End)
	secs["vehicle_missions"].Set("GET /api/vehicle-missions", "vehicle_missions.get_root", vehicleMissionHandler.List)
	secs["vehicle_missions"].Set("GET /api/vehicle-missions/{id}", "vehicle_missions.get_by_id", vehicleMissionHandler.Get)
	secs["vehicle_missions"].Set("POST /api/vehicle-missions/{id}/rating", "vehicle_missions.post_by_id_rating", vehicleMissionHandler.CreateRating)
	// فحص أدوات المركبة العامة عند بدء مهمة — حصراً لليدر (نفس requireLeader
	// المستخدم بصيانة الأجهزة)، الموظف العادي ما يشوف/يستخدم هالراوت إطلاقاً.
	secs["vehicle_missions"].Set("POST /api/vehicle-missions/{id}/tool-check", "vehicle_missions.post_by_id_tool_check", vehicleMissionHandler.CreateToolCheck, requireLeader)
	peopleApp.Set("GET /api/employees/{id}/driver-rating-summary", "people.vehicle_rating", vehicleMissionHandler.DriverRatingSummary)

	// نظام حجز المركبات (مسبق) — منفصل عن بدء المهمة الفعلي أعلاه.
	secs["vehicle_bookings"].Set("POST /api/vehicle-bookings", "vehicle_bookings.post_root", vehicleBookingHandler.Create)
	secs["vehicle_bookings"].Set("PUT /api/vehicle-bookings/{id}/decide", "vehicle_bookings.put_by_id_decide", vehicleBookingHandler.Decide)
	secs["vehicle_bookings"].Set("PUT /api/vehicle-bookings/{id}/cancel", "vehicle_bookings.put_by_id_cancel", vehicleBookingHandler.Cancel)
	secs["vehicle_bookings"].Set("GET /api/vehicle-bookings", "vehicle_bookings.get_root", vehicleBookingHandler.List)
	// ملخصات التذكير — للمراقب بس (نظرة شاملة على كل السيارات والفنيين)
	// ملاحظة: هذولا فقط لوحة المراقبة (MonitorDashboard) تستدعيهم — عمداً مقيدين
	// بدور المراقب/الأدمن حصراً، لأنهم يعرضون راتب مقترح للفنيين (بيانات حساسة
	// ماريد الفني نفسه يشوفها). لا تحولها لصلاحية "vehicle_management" العامة —
	// هذي الصلاحية ممكن تنمنح لفنيين لتسجيل صيانة مركبات عادية بدون قصد كشف رواتب.
	// متابعة السيارات الشهرية (ملف الإكسل) — أبو الكميات.
	vehicleTrackingHandler := handler.NewVehicleTrackingHandler(db)
	// 🧽 غسل السيارات: عامل الغسل يأشّر، وصاحب إدارة المركبات يشوف ويتراجع
	carWashHandler := handler.NewCarWashHandler(db)
	// (ون بيس 11-10) requireCarWash انتقل لقاعدة بالنواة (internal/apps/sections)
	secs["car_wash"].Set("GET /api/car-wash", "car_wash.get_root", carWashHandler.List)
	secs["car_wash"].Set("POST /api/car-wash/{vehicleId}", "car_wash.post_by_vehicleid", carWashHandler.Mark)
	secs["car_wash"].Set("DELETE /api/car-wash/{vehicleId}", "car_wash.delete_by_vehicleid", carWashHandler.Unmark)
	secs["vehicles"].Set("GET /api/vehicles/tracking", "vehicles.get_tracking", vehicleTrackingHandler.Get)
	secs["vehicles"].Set("POST /api/vehicles/tracking/merge", "vehicles.post_tracking_merge", vehicleTrackingHandler.Merge)
	secs["vehicles"].Set("GET /api/vehicles/tracking/export", "vehicles.get_tracking_export", vehicleTrackingHandler.Export)
	secs["vehicles"].Set("PUT /api/vehicles/tracking/day", "vehicles.put_tracking_day", vehicleTrackingHandler.SaveDay)
	secs["vehicles"].Set("GET /api/vehicles/ratings/vehicle-summary", "vehicles.get_ratings_vehicle_summary", vehicleHandler.VehicleScoreSummaries)
	secs["vehicles"].Set("GET /api/vehicles/ratings/technician-summary", "vehicles.get_ratings_technician_summary", vehicleHandler.TechnicianWashSummaries)
	// إحصاء الغسل لكل سيارة بشهر: كم مرة انغسلت ومنو غسلها.
	// نفس حارس إدارة المركبات — أبو الكميات والإداري يفتحونها.
	secs["vehicles"].Set("GET /api/vehicles/ratings/wash-monthly", "vehicles.get_ratings_wash_monthly", vehicleHandler.VehicleWashMonthly)
	// تصدير نفس الإحصاء إكسل — نفس الحارس بالضبط، التصدير ما يوسّع الوصول.
	secs["vehicles"].Set("GET /api/vehicles/ratings/wash-monthly/export", "vehicles.get_ratings_wash_monthly_export", vehicleHandler.ExportVehicleWashMonthly)

	// المرحلة 2: مرفقات الأعطال/الأضرار، متابعة الإطارات والبطاريات، تنبيهات الاستحقاق
	secs["vehicle_incidents"].Set("GET /api/vehicle-incidents/{id}/attachments", "vehicle_incidents.get_by_id_attachments", vehicleHandler.ListIncidentAttachments)
	secs["vehicle_incidents"].Set("POST /api/vehicle-incidents/{id}/attachments", "vehicle_incidents.post_by_id_attachments", vehicleHandler.CreateIncidentAttachment)
	secs["vehicle_incidents"].Set("DELETE /api/vehicle-incidents/{id}/attachments/{attachmentId}", "vehicle_incidents.delete_by_id_attachments_by_attachmentid", vehicleHandler.DeleteIncidentAttachment)
	secs["vehicles"].Set("GET /api/vehicles/{id}/parts", "vehicles.get_by_id_parts", vehicleHandler.ListParts)
	secs["vehicles"].Set("POST /api/vehicles/{id}/parts", "vehicles.post_by_id_parts", vehicleHandler.CreatePart)
	secs["vehicle_parts"].Set("PUT /api/vehicle-parts/{id}/replace", "vehicle_parts.put_by_id_replace", vehicleHandler.ReplacePart)
	secs["vehicles"].Set("GET /api/vehicles/alerts", "vehicles.get_alerts", vehicleHandler.Alerts)
	secs["vehicles"].Set("GET /api/vehicles/dashboard", "vehicles.get_dashboard", vehicleHandler.FleetDashboard)
	secs["vehicles"].Set("GET /api/vehicles/{id}/expense-summary", "vehicles.get_by_id_expense_summary", vehicleHandler.ExpenseSummary)

	// الجودة — مشاكل تنفيذية ميدانية + مشاكل رقابية/إدارية
	// تقارير العمل — الفني يرسل تقرير عن حجزه، المراقب/الجودة يشوفون كل التقارير
	// (دازاي 10-26) ورق الحجوزات القديمة لليدر السابق — قائمته، وتنشال الصلاحية لحالها من يخلّص
	secs["paperwork"].Set("GET /api/paperwork/carryover", "paperwork.get_carryover", func(w http.ResponseWriter, r *http.Request) {
		rows, err := paperworkGuard.CarryoverList(middleware.EmployeeIDFromContext(r))
		if err != nil {
			handler.WriteError(w, http.StatusInternalServerError, "تعذر")
			return
		}
		handler.WriteJSON(w, http.StatusOK, rows)
	})
	// (دازاي 10-28) حجوزاتي المرحّلة لإدارة التقنيين الي بلا فاتورة — للتقني (أو ليدر الكادر)
	secs["paperwork"].Set("GET /api/paperwork/handover-bookings", "paperwork.get_handover_bookings", func(w http.ResponseWriter, r *http.Request) {
		ids, err := paperworkGuard.HandoverMissing(middleware.EmployeeIDFromContext(r))
		if err != nil {
			handler.WriteError(w, http.StatusInternalServerError, "تعذر")
			return
		}
		out := []*model.Booking{}
		if len(ids) > 0 {
			m, err := bookingRepo.FindByIDs(ids)
			if err != nil {
				handler.WriteError(w, http.StatusInternalServerError, "تعذر")
				return
			}
			for _, id := range ids {
				if b, ok := m[id]; ok {
					out = append(out, b)
				}
			}
		}
		handler.WriteJSON(w, http.StatusOK, out)
	})
	// نفس القائمة بس كحجوزات كاملة — لقائمة الحجوزات بشاشة الفاتورة
	secs["paperwork"].Set("GET /api/paperwork/carryover/bookings", "paperwork.get_carryover_bookings", func(w http.ResponseWriter, r *http.Request) {
		items, err := paperworkGuard.CarryoverList(middleware.EmployeeIDFromContext(r))
		if err != nil {
			handler.WriteError(w, http.StatusInternalServerError, "تعذر")
			return
		}
		ids := make([]string, 0, len(items))
		for _, it := range items {
			if !it.HasInvoice {
				ids = append(ids, it.ID)
			}
		}
		out := []*model.Booking{}
		if len(ids) > 0 {
			m, err := bookingRepo.FindByIDs(ids)
			if err != nil {
				handler.WriteError(w, http.StatusInternalServerError, "تعذر")
				return
			}
			for _, id := range ids {
				if b := m[id]; b != nil {
					out = append(out, b)
				}
			}
		}
		handler.WriteJSON(w, http.StatusOK, out)
	})
	safeguard.Loop("تسوية حجوزات النظام القديم", 7*time.Minute, time.Hour, func() {
		if n, err := bookingRepo.SettleOldImports(); err != nil {
			log.Printf("settle OLD-: %v", err)
		} else if n > 0 {
			log.Printf("settle OLD-: %d حجز قديم انحسب «تم»", n)
		}
	})
	// 👁️ متابعة الموظفين من «مذكراتي» (دازاي 10-29)
	safeguard.Loop("متابعات الموظفين", 9*time.Minute, time.Hour, func() {
		service.SweepWatches(db, func(owner, msg string) { _ = notificationRepo.Create(owner, "AI_WATCH", msg) })
	})
	// ⏰ منبّه كمال (دازاي 10-29)
	safeguard.Loop("تذكيرات كمال", 20*time.Second, 30*time.Second, func() {
		service.SendDueReminders(db, func(emp, msg string) { _ = notificationRepo.Create(emp, "AI_REMINDER", msg) })
	})
	safeguard.Loop("صلاحية ورق الحجوزات القديمة", 4*time.Minute, time.Hour, func() {
		done, err := paperworkGuard.SweepCarryover()
		if err != nil {
			log.Printf("carryover sweep: %v", err)
			return
		}
		for _, d := range done {
			_ = notificationRepo.Create(d.ID, "PAPERWORK_DONE", "✅ خلّصت كل فواتير وتقارير حجوزاتك القديمة — عاشت ايدك. صلاحية «إكمال الورق القديم» انشالت تلقائياً.")
			_ = notificationRepo.CreateForRole("OWNER", "PAPERWORK_DONE", "✅ "+d.Name+" خلّص ورق حجوزاته القديمة، وانشالت منه صلاحية «إكمال الورق القديم» تلقائياً.")
		}
	})
	fieldApp.Set("POST /api/work-reports", "field.work_report", workReportHandler.Create)
	secs["work_reports"].Set("GET /api/work-reports", "work_reports.get_root", workReportHandler.List)
	secs["work_reports"].Set("GET /api/work-reports/counts", "work_reports.get_counts", workReportHandler.Counts)
	// ✍️ توقيع الزبون عند التسليم (ع) 11-07
	fieldApp.Set("POST /api/bookings/{id}/customer-signature", "field.customer_signature", workReportHandler.SaveSignature)
	secs["bookings"].Set("GET /api/bookings/{id}/customer-signature", "bookings.get_by_id_customer_signature", workReportHandler.GetSignature)
	secs["work_reports"].Set("GET /api/work-reports/by-booking", "work_reports.get_by_booking", workReportHandler.ByBooking)
	secs["work_reports"].Set("PUT /api/work-reports/{id}/review", "work_reports.put_by_id_review", workReportHandler.Review)
	secs["bookings"].Set("PUT /api/bookings/{id}/progress/review", "bookings.put_by_id_progress_review", bookingProgressHandler.Review)
	// تقرير زيارة معاينة («كشف») — أي كادر مكلّف بالحجز يسلّمه، بلا حارس ورق محاسبي.
	fieldApp.Set("POST /api/booking-survey-reports", "field.survey_report", surveyReportHandler.Create)
	secs["booking_survey_reports"].Set("GET /api/booking-survey-reports", "booking_survey_reports.get_root", surveyReportHandler.List)
	// صور الكشف (حجز أو مشروع): الفحص بالهاندلر نفسه — العرض والتنزيل
	// للمراقب والمحاسب والمدير وإداري الحجوزات والكادر الي رفعها.
	// ═══ وحدة تقنية المعلومات ═══ — الجرد لصاحب it_assets، والإحصائيات
	// لصاحب it_stats أو it_assets (الي يدير الأجهزة يشوف أرقامها).
	itAssetHandler := handler.NewItAssetHandler(repository.NewItAssetRepository(db))
	// (ون بيس 11-10) requireItAssets انتقل لقاعدة بالنواة (internal/apps/sections)
	// (ون بيس 11-10) requireItStats انتقل لقاعدة بالنواة (internal/apps/sections)
	secs["it"].Set("GET /api/it/assets", "it.get_assets", itAssetHandler.List)
	secs["it"].Set("POST /api/it/assets", "it.post_assets", itAssetHandler.Create)
	secs["it"].Set("PUT /api/it/assets/{id}", "it.put_assets_by_id", itAssetHandler.Update)
	secs["it"].Set("DELETE /api/it/assets/{id}", "it.delete_assets_by_id", itAssetHandler.Delete)
	secs["it"].Set("GET /api/it/assets/{id}/logs", "it.get_assets_by_id_logs", itAssetHandler.Logs)
	secs["it"].Set("POST /api/it/assets/{id}/logs", "it.post_assets_by_id_logs", itAssetHandler.AddLog)
	secs["it"].Set("GET /api/it/stats", "it.get_stats", itAssetHandler.Stats)
	secs["survey_photos"].Set("GET /api/survey-photos", "survey_photos.get_root", surveyPhotoHandler.List)
	fieldApp.Set("POST /api/survey-photos", "field.survey_photo_upload", surveyPhotoHandler.Upload)
	secs["survey_photos"].Set("GET /api/survey-photos/{id}/file", "survey_photos.get_by_id_file", surveyPhotoHandler.File)
	fieldApp.Set("DELETE /api/survey-photos/{id}", "field.survey_photo_delete", surveyPhotoHandler.Delete)
	// تدقيق التكرار (ماتركس) — عرض وقرار "مو تكرار" حصراً، بلا حذف/دمج.
	// تدقيق التكرار: المدير/المالك أو صلاحية duplicate_review (طلب (ع): موظف يدقق بيهن).
	// (ون بيس 11-10) requireDuplicateReview انتقل لقاعدة بالنواة (internal/apps/sections)
	secs["duplicate_candidates"].Set("GET /api/duplicate-candidates", "duplicate_candidates.get_root", duplicateCandidateHandler.List)
	secs["duplicate_candidates"].Set("PUT /api/duplicate-candidates/{id}/dismiss", "duplicate_candidates.put_by_id_dismiss", duplicateCandidateHandler.Dismiss)
	duplicateCandidateHandler.SetResolvers(duplicateCandidateRepo, bookingDeleteRepo, notificationRepo, employeeRepo)
	secs["duplicate_candidates"].Set("GET /api/duplicate-candidates/report", "duplicate_candidates.get_report", duplicateCandidateHandler.Report)
	secs["duplicate_candidates"].Set("GET /api/duplicate-candidates/{id}/merge-preview", "duplicate_candidates.get_by_id_merge_preview", duplicateCandidateHandler.MergePreview)
	secs["duplicate_candidates"].Set("POST /api/duplicate-candidates/{id}/merge", "duplicate_candidates.post_by_id_merge", duplicateCandidateHandler.Merge)
	secs["duplicate_candidates"].Set("POST /api/duplicate-candidates/{id}/request-delete", "duplicate_candidates.post_by_id_request_delete", duplicateCandidateHandler.RequestDelete)

	secs["quality"].Set("GET /api/quality/issues", "quality.get_issues", qualityHandler.List)
	secs["quality"].Set("POST /api/quality/issues", "quality.post_issues", qualityHandler.Create)
	secs["quality"].Set("PUT /api/quality/issues/{id}", "quality.put_issues_by_id", qualityHandler.Update)

	// صيانة الأجهزة العامة (شيت "صيانة الاجهزة") — حصراً للـليدر (isLeader فريش من
	// قاعدة البيانات بكل طلب، مو من التوكن)
	secs["device_maintenance"].Set("GET /api/device-maintenance", "device_maintenance.get_root", deviceMaintenanceHandler.List)
	secs["device_maintenance"].Set("POST /api/device-maintenance", "device_maintenance.post_root", deviceMaintenanceHandler.Create)
	secs["device_maintenance"].Set("PUT /api/device-maintenance/{id}", "device_maintenance.put_by_id", deviceMaintenanceHandler.Update)

	// جرد الفريق ("جرد العدد") — حصراً للـليدر أيضاً
	// ═══ قراءة الجرد: الليدر أو منو انمنح متابعة الجرد ═══
	//
	// ⚠️ چانت مقفولة على علامة «ليدر» بس — بلا أي منفذ صلاحية. يعني
	// حتى لو انمنح إداري الكوادر أو المراقب «متابعة الجرد»، الخادم
	// يرفضه لأنه مو ليدر مسجّل. وهذا ضد «اعزل كل شغلة بصلاحية».
	//
	// ⚠️ و`inventory` مذكورة معاها لأنها صلاحية المراقب الافتراضية —
	// بدونها يخسر وصوله للجرد لمن ننشل البند من قائمته.
	//
	// ⚠️ والكتابة (POST) تبقى `requireLeader`: المتابِع **يتابع**،
	// الي يجرد هو الليدر.
	// ⚠️ الدور مذكور صراحةً — الصلاحيات تنقرأ من صفوف EmployeePermission
	// فقط، فمراقب حسابه ماكو إله صف `inventory` چان يخسر وصوله كلياً
	// لمن انشلنا البند من قائمته. الحارس يطابق شرط العرض بمكتبه بالضبط.
	requireInventoryFollow := middleware.RequireLeaderOrRoleOrAnyPermission(
		permissionRepo, employeeRepo, notificationRepo,
		[]string{"MONITOR", "HR_COORDINATOR", "PROCUREMENT_ADMIN"},
		"inventory_follow", "inventory")
	secs["team_inventory"].Set("GET /api/team-inventory/tools", "team_inventory.get_tools", teamInventoryCheckHandler.ListTools, requireInventoryFollow)
	secs["team_inventory"].Set("POST /api/team-inventory/tools", "team_inventory.post_tools", teamInventoryCheckHandler.CreateTool, requireLeader)
	secs["team_inventory"].Set("GET /api/team-inventory/checks", "team_inventory.get_checks", teamInventoryCheckHandler.List, requireInventoryFollow)
	// (ع) 10-16: الجرد الحقيقي (InventoryCheck من «جردي»/«مهامي») — نفس حارس تبويب المراقب
	secs["inventory"].Set("GET /api/inventory/checks/follow", "inventory.get_checks_follow", handler.NewInventoryFollowHandler(db).List, requireInventoryFollow)
	secs["team_inventory"].Set("POST /api/team-inventory/checks", "team_inventory.post_checks", teamInventoryCheckHandler.Create, requireLeader)

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
	//
	// ⚠️ **وصاحب صلاحية فاتورة الخدمة** (جي بي اس / داش كام): بلاه
	// يحفظ الفاتورة ويشوف رقمها **مرة وحدة** وخلاص — ما يكدر يرجعلها
	// ولا يعرف وين وصلت بالتدقيق. وهاي انكشفت لمّا منح المالك
	// الصلاحيتين لموظفين اثنين وما طلع عدهم شي.
	//
	// 🔴 **وما يشوف فواتير غيره**: `List` يفرض معرّف الطالب لكل من
	// `canReviewInvoices` ما يرجّع صح له — والأدوار الي تشوف الكل هي
	// ADMIN/OWNER/FINANCE/MONITOR وبس. فالقيد بالمعالج مو بالحارس،
	// ويصمد لو حاول يحط معرّف زميله بالرابط.
	//
	// ⚠️ والإضافة **زيادة مو تبديل**: ولا حارس ينضيّق، فما تنكسر
	// صلاحية ولا موظف شغّال.
	// (ون بيس 11-10) سلة الليدر صارت invoices.list/page/view/by_number بالنواة (basketPerms)
	// أسباب الشغل المجاني — يقراها أي موظف يسوي فاتورة
	secs["free_work_reasons"].Set("GET /api/free-work-reasons", "free_work_reasons.get_root", leaderInvoiceHandler.FreeReasons)
	secs["system_price_catalog"].Set("GET /api/system-price-catalog", "system_price_catalog.get_root", leaderInvoiceHandler.ListCatalog)
	secs["materials"].Set("GET /api/materials", "materials.get_root", leaderInvoiceHandler.ListMaterials)
	invoicesApp.Set("GET /api/leader-invoices", "invoices.list", leaderInvoiceHandler.List)
	invoicesApp.Set("GET /api/leader-invoices/page", "invoices.page", leaderInvoiceHandler.Page)
	invoicesApp.Set("GET /api/leader-invoices/{id}", "invoices.view", leaderInvoiceHandler.Get)
	// ⚠️ الليدر مثل اليوم، **وزيادةً** مسؤول خدمة الحجز: بخدمات مثل
	// الجي بي اس والداش كام الفاتورة على مسؤول الخدمة، وهو مو
	// بالضرورة ليدر. والي مو الاثنين تنسجّل مخالفته مثل ما چان.
	invoicesApp.Set("POST /api/leader-invoices", "invoices.create", leaderInvoiceHandler.Create, idem)
	// فاتورة الجي بي اس / الداش كام بسعر يدوي — الحارس هنا «أي وحدة
	// من الصلاحيتين»، والمعالج يفحص **أي** وحدة بالضبط حسب النوع
	// المطلوب، حتى مسؤول الجي بي اس ما يفوتر داش كام.
	invoicesApp.Set("POST /api/leader-invoices/service", "invoices.create_service", leaderInvoiceHandler.CreateServiceInvoice)
	// فاتورة بكلفة يدوية — شغل برّا جدول الكلفة. صلاحية وحدة معزولة
	// تنمنح فرد-فرد، لأن السعر الحر يشيل الحارس الوحيد على التسعير.
	invoicesApp.Set("POST /api/leader-invoices/manual", "invoices.create_manual", leaderInvoiceHandler.CreateManualInvoice)
	// ═══ إكسل حجوزات داخل الشركة ═══
	// المالك والمدير والمحاسب والمراقب (قرار (ع)). و`finance_audit`
	// جنبهم حتى الي يكمل شغل المحاسب يطلعه الملف بعد.
	internalExportHandler := handler.NewInternalBookingExportHandler(repository.NewInternalBookingReportRepository(db))
	secs["internal_bookings"].Set("GET /api/internal-bookings/export", "internal_bookings.get_export", internalExportHandler.ExportMonth)
	// فاتورة الشغل داخل الشركة: صلاحية وحدة ينطيها المالك للليدر
	// **أو** لإداري الكوادر — إداري الكوادر مو ليدر فحارس الفاتورة
	// العادية يرفضه، والفاتورة اليدوية تفتح له كل الفواتير الحرة.
	// قرار (ع) 10-07: تسعير الأعمال الداخلية القديمة (مؤقت) + طابور ورق الإداري.
	internalBackfillHandler := handler.NewInternalBackfillHandler(db, leaderInvoiceService)
	// (ون بيس 11-10) requireBackfill انتقل لقاعدة بالنواة (internal/apps/sections)
	secs["internal_backfill"].Set("GET /api/internal-backfill", "internal_backfill.get_root", internalBackfillHandler.List)
	secs["internal_backfill"].Set("POST /api/internal-backfill/{id}", "internal_backfill.post_by_id", internalBackfillHandler.Price)
	secs["bookings"].Set("GET /api/bookings/internal-paperwork", "bookings.get_internal_paperwork", internalBackfillHandler.CoordinatorQueue)
	invoicesApp.Set("POST /api/leader-invoices/internal", "invoices.create_internal", leaderInvoiceHandler.CreateInternalInvoice)
	// حساب تقريبي بدون حفظ لما زبون يستفسر — نفس صلاحية إنشاء الفاتورة (الليدر)
	// حساب تكلفة التنصيب للتنفيذ: فقرة رئيسية بكل الحسابات وكل الأدوار،
	// فما بيها قيد غير تسجيل الدخول — هي حاسبة ما تكشف بيانات أحد
	// (دازاي 10-28) التقني الي عنده حجز مرحّل بلا فاتورة يحسب كلفته هم — بدون execution_cost
	requireEstimate := func(next http.Handler) http.Handler {
		guarded := requireExecutionCost(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ids, err := paperworkGuard.HandoverMissing(middleware.EmployeeIDFromContext(r)); err == nil && len(ids) > 0 {
				next.ServeHTTP(w, r)
				return
			}
			guarded.ServeHTTP(w, r)
		})
	}
	// شرط «التسليم الناقص» يبقى طبقة بعد القاعدة (requireEstimate يعيد فحص الكلفة — بلا ضرر)
	invoicesApp.Set("POST /api/leader-invoices/estimate", "invoices.estimate", leaderInvoiceHandler.Estimate, requireEstimate)
	invoicesApp.Set("POST /api/leader-invoices/camera-cost", "invoices.camera_cost", leaderInvoiceHandler.CameraCost)
	// ── صندوق المراقب ──
	// شغل المراقب: يشوف ويأشّر. requireMonitor = ADMIN أو MONITOR،
	// والمالك داخل بالأدمن.
	secs["monitor_reviews"].Set("GET /api/monitor-reviews", "monitor_reviews.get_root", monitorReviewHandler.List)
	secs["monitor_reviews"].Set("GET /api/monitor-reviews/counts", "monitor_reviews.get_counts", monitorReviewHandler.Counts)
	secs["monitor_reviews"].Set("GET /api/monitor-reviews/cards", "monitor_reviews.get_cards", monitorReviewHandler.Cards)
	monitorActionHandler := handler.NewMonitorActionHandler(monitorReviewRepo, notificationRepo, extraTaskRepo)
	justHandler := handler.NewJustificationHandler(db, notificationRepo)
	monitorActionHandler.Just = justHandler
	secs["justifications"].Set("GET /api/justifications/mine", "justifications.get_mine", justHandler.Mine)
	meApp.Set("POST /api/justifications/{id}/answer", "me.justification_answer", justHandler.Answer)
	secs["justifications"].Set("GET /api/justifications", "justifications.get_root", justHandler.List)
	secs["justifications"].Set("POST /api/justifications/direct", "justifications.post_direct", justHandler.Direct)
	// (ع) 10-16: «المحاسب متأخر إذا تأخر على التدقيق ٢٤ ساعة» — خانة «المتأخر»
	invoicesApp.Set("GET /api/late-invoices", "invoices.late", handler.NewLateInvoicesHandler(db).List)
	secs["justifications"].Set("POST /api/justifications/{id}/decide", "justifications.post_by_id_decide", justHandler.Decide)
	secs["monitor_reviews"].Set("POST /api/monitor-reviews/{id}/action", "monitor_reviews.post_by_id_action", monitorActionHandler.Act)
	secs["monitor_reviews"].Set("POST /api/monitor-reviews/{id}/decide", "monitor_reviews.post_by_id_decide", monitorReviewHandler.Decide)
	// «منو فاتح النظام هسه» — المدير/المالك والمراقب.
	secs["presence"].Set("GET /api/presence/online", "presence.get_online", func(w http.ResponseWriter, r *http.Request) {
		rows, err := presenceRepo.Recent(60)
		if err != nil {
			http.Error(w, `{"error":"تعذر جلب المتصلين"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(rows)
	})
	secs["monitor_desk"].Set("GET /api/monitor-desk/counts", "monitor_desk.get_counts", monitorDeskHandler.Counts)
	// ── ماتركس ٢٠٥٠: سلسلة الحجز (١٤ محطة) + تقارير الأدوار + تقييم الليدر لفنيّيه ──
	matrixChainService := service.NewMatrixChainService(repository.NewMatrixChainRepository(db), aiActionRepo.EmployeeName)
	// (ك) 11-08: بلاغ #21 — عدّ «تجهيز المواد» يبدي من بداية دوام يوم الموعد.
	matrixChainService.SetWorkStartHour(func() int {
		if w, err := aiRepo.WorkWindow(); err == nil && w != nil {
			return w.StartHour
		}
		return 9
	})
	matrixChainHandler := handler.NewMatrixChainHandler(matrixChainService, matrixAutopilotService.ResolveFor)
	secs["ai"].Set("GET /api/ai/chain/{bookingId}", "ai.get_chain_by_bookingid", matrixChainHandler.Chain)
	// تقارير الأدوار فيها «ماتركس على المراقب» — للمدير والمالك بس.
	secs["ai"].Set("GET /api/ai/role-chain", "ai.get_role_chain", matrixChainHandler.RoleChain)
	secs["ai"].Set("GET /api/ai/role-chain/roles", "ai.get_role_chain_roles", matrixChainHandler.Roles)
	secs["ai"].Set("GET /api/ai/chain-learning", "ai.get_chain_learning", matrixChainHandler.Learning)
	secs["crew_ratings"].Set("GET /api/crew-ratings/pending", "crew_ratings.get_pending", matrixChainHandler.Pending)
	fieldApp.Set("POST /api/crew-ratings", "field.crew_rating", matrixChainHandler.Rate)
	// جرد العدّة بعد الحجز — الفني على حجزه هو بس، والنقص ينبّه صاحب صلاحية المخزن.
	matrixChainHandler.SetShortageNotifier(func(msg string) { _ = notificationRepo.CreateForPermission("inventory", "INVENTORY", msg) })
	secs["inventory"].Set("GET /api/inventory/after-checks/pending", "inventory.get_after_checks_pending", matrixChainHandler.AfterPending)
	fieldApp.Set("POST /api/inventory/after-checks", "field.after_check", matrixChainHandler.AfterSave)
	// المرحلة الثانية: ماتركس يقترح الموعد والكادر، والمنسق يقرر. كل ساعة يحسم
	// الاقتراحات الي صار بحجزها قرار (حتى تنقاس دقته).
	matrixSuggestService := service.NewMatrixSuggestService(repository.NewMatrixSuggestRepository(db))
	go func() {
		for {
			time.Sleep(time.Hour)
			matrixSuggestService.Evaluate()
		}
	}()
	// (ع) 10-17: كل موظف ياخذ صلاحيات منصبه مسجّلة تلقائياً — أول ما يشتغل
	// الخادم وبعدين كل ٦ ساعات (موظف جديد، منصب تغيّر…). إضافة بس.
	go func() {
		time.Sleep(time.Minute)
		for {
			if rows, err := permissionService.ApplyDefaultsAll(); err == nil && len(rows) > 0 {
				log.Printf("[permissions] انضافت صلاحيات المنصب لـ%d موظف", len(rows))
			}
			time.Sleep(6 * time.Hour)
		}
	}()
	matrixSuggestHandler := handler.NewMatrixSuggestHandler(matrixSuggestService)
	secs["ai"].Set("GET /api/ai/suggest/{bookingId}", "ai.get_suggest_by_bookingid", matrixSuggestHandler.Suggest)
	secs["ai"].Set("GET /api/ai/suggest-accuracy", "ai.get_suggest_accuracy", matrixSuggestHandler.Accuracy)
	// المرحلة الثالثة: ماتركس ينفّذ البسيط — مفاتيح المالك مطفية افتراضياً + بوابة دقة + تراجع.
	matrixAutonomyService := service.NewMatrixAutonomyService(matrixAutopilotService, aiActionRepo, notificationRepo,
		repository.NewSystemSwitchRepository(db), matrixSuggestService, bookingService)
	matrixAutonomyService.Start()
	matrixAutonomyHandler := handler.NewMatrixAutonomyHandler(matrixAutonomyService)
	secs["ai"].Set("GET /api/ai/autonomy", "ai.get_autonomy", matrixAutonomyHandler.Status)
	// عيون الرقابة: المخزن والسيارات ورأي الزبائن — للمراقب كمان (قرار (ع)).
	matrixWatchService := service.NewMatrixWatchService(repository.NewMatrixWatchRepository(db), matrixAutopilotService, notificationRepo)
	matrixWatchService.SetProjectMoney(repository.NewProjectPaymentRepository(db))
	go func() {
		for {
			time.Sleep(30 * time.Minute)
			now := time.Now().In(service.BaghdadLoc())
			if now.Hour() >= 11 {
				today := now.Format("2006-01-02")
				matrixWatchService.Remind(today, service.BaghdadMidnight(today))
			}
		}
	}()
	// ── الإعلام والعلاقات العامة (قرار (ع) 10-05) ──
	// التحويل: من إدارة المشاريع. والقراءة والتحديث: دور الإعلام أو صلاحية media
	// (والمدير والمالك). ومدير المشاريع يشوف التحويلات حتى يتابع.
	// (ون بيس 11-10) requireMediaView انتقل لقاعدة بالنواة (internal/apps/sections)
	// (ون بيس 11-10) requireMedia انتقل لقاعدة بالنواة (internal/apps/sections)
	mediaHandler := handler.NewMediaHandler(repository.NewMediaRepository(db), func(msg string) {
		_ = notificationRepo.CreateForRolesOrPermission([]string{"MEDIA", "PUBLIC_RELATIONS"}, "media", "MEDIA", msg)
	})
	// ── دفعات المشاريع (قرار (ع) 10-06): الصلاحية بالمعالج — المحاسب، إدارة المشاريع، ومشرف المشروع ──
	projectPaymentRepo := repository.NewProjectPaymentRepository(db)
	projectPaymentHandler := handler.NewProjectPaymentHandler(projectPaymentRepo, permissionRepo)
	fieldApp.Use(func(d *field.Deps) { d.CanContractValue = projectPaymentHandler.CanValue })
	projectService.SetPaymentChecker(projectPaymentRepo)
	invoicesApp.Set("GET /api/projects/{id}/payments", "invoices.project_payments", projectPaymentHandler.List)
	invoicesApp.Set("POST /api/projects/{id}/payments", "invoices.project_payment_add", projectPaymentHandler.Add, idem)
	fieldApp.Set("PUT /api/projects/{id}/contract-value", "field.project_contract_value", projectPaymentHandler.SetValue)
	invoicesApp.Set("POST /api/project-payments/{id}/verify", "invoices.project_payment_ok", projectPaymentHandler.Verify, idem)
	invoicesApp.Set("POST /api/project-payments/{id}/cancel", "invoices.project_payment_void", projectPaymentHandler.Cancel, idem)
	invoicesApp.Set("GET /api/project-payments/overview", "invoices.payments_overview", projectPaymentHandler.Overview)
	invoicesApp.Set("GET /api/project-payments/audit", "invoices.payments_audit", projectPaymentHandler.Audit)
	invoicesApp.Set("PUT /api/project-payments/{id}/amount", "invoices.payment_amount", projectPaymentHandler.SetAmount, idem)
	secs["projects"].Set("GET /api/projects/media/open", "projects.get_media_open", mediaHandler.OpenProjects)
	secs["projects"].Set("GET /api/projects/{id}/media", "projects.get_by_id_media", mediaHandler.ForProject)
	secs["projects"].Set("POST /api/projects/{id}/media", "projects.post_by_id_media", mediaHandler.Transfer)
	// سجل الانضباط الوظيفي — المدير والمالك بس (قرار (ع) 10-05).
	disciplineRecordHandler := handler.NewDisciplineRecordHandler(service.NewDisciplineRecordService(repository.NewDisciplineRecordRepository(db)))
	secs["discipline_record"].Set("GET /api/discipline-record", "discipline_record.get_root", disciplineRecordHandler.Get)
	// ── صوت الموظفين + المشاكل الوظيفية (قرارات (ع) 10-05) ──
	peerVoiceService := service.NewPeerVoiceService(repository.NewPeerVoiceRepository(db),
		repository.NewAchievementVoiceRepository(db).Names,
		func(roles []string, perm, msg string) {
			_ = notificationRepo.CreateForRolesOrPermission(roles, perm, "MATRIX_PEER", msg)
		})
	peerVoiceService.EnableModel(cfg.AIKeyMatrix, cfg.AIModel)
	peerHandler := handler.NewPeerVoiceHandler(peerVoiceService)
	safeguard.Loop("متابعة احتياج التدريب", 19*time.Minute, 6*time.Hour, func() {
		up, down, err := service.FollowUpTraining(db)
		if err != nil {
			log.Printf("training follow-up: %v", err)
			return
		}
		if down > 0 {
			_ = notificationRepo.CreateForRolesOrPermission([]string{"MONITOR"}, "monitoring", "training_followup",
				fmt.Sprintf("🎓 ماتركس — متابعة ٣٠ يوم: %d موظف ما تحسّن بعد احتياج التدريب. شوف «المشاكل الوظيفية».", down))
		}
		if up+down > 0 {
			log.Printf("training follow-up: %d تحسّن، %d لا", up, down)
		}
	})
	secs["peer"].Set("GET /api/peer/me", "peer.get_me", peerHandler.Mine)
	meApp.Set("POST /api/peer/me", "me.peer_note", peerHandler.Submit)
	meApp.Set("POST /api/peer/notes/{id}/followup", "me.peer_followup", peerHandler.Followup)
	meApp.Set("POST /api/peer/notes/{id}/accept", "me.peer_accept", peerHandler.Accept)
	secs["peer"].Set("GET /api/peer/voice", "peer.get_voice", peerHandler.Report)
	secs["workplace_issues"].Set("GET /api/workplace-issues", "workplace_issues.get_root", peerHandler.Issues)
	secs["workplace_issues"].Set("POST /api/workplace-issues", "workplace_issues.post_root", peerHandler.OpenIssue)
	secs["workplace_issues"].Set("GET /api/workplace-issues/{id}", "workplace_issues.get_by_id", peerHandler.Issue)
	secs["workplace_issues"].Set("POST /api/workplace-issues/{id}/entries", "workplace_issues.post_by_id_entries", peerHandler.AddEntry)
	secs["workplace_issues"].Set("POST /api/workplace-issues/{id}/analyze", "workplace_issues.post_by_id_analyze", peerHandler.Analyze)
	secs["workplace_issues"].Set("POST /api/workplace-issues/{id}/decide", "workplace_issues.post_by_id_decide", peerHandler.Decide)
	meApp.Set("POST /api/workplace-issues/{id}/party-followup", "me.workplace_party_followup", peerHandler.PartyFollowup)
	secs["media"].Set("GET /api/media/briefs", "media.get_briefs", mediaHandler.List)
	secs["media"].Set("PUT /api/media/briefs/{id}", "media.put_briefs_by_id", mediaHandler.Update)
	secs["ai"].Set("GET /api/ai/watch", "ai.get_watch", handler.NewMatrixWatchHandler(matrixWatchService).Report)
	// عدّاد كلفة الذكاء الاصطناعي — المدير والمالك.
	secs["ai"].Set("GET /api/ai/usage", "ai.get_usage", func(w http.ResponseWriter, r *http.Request) {
		rep, err := service.AIUsageSummary()
		if err != nil {
			http.Error(w, `{"error":"تعذر حساب الكلفة"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(rep)
	})
	secs["ai"].Set("GET /api/ai/autonomy/crew", "ai.get_autonomy_crew", matrixAutonomyHandler.ActiveCrew)
	secs["ai"].Set("POST /api/ai/autonomy/crew/{id}/undo", "ai.post_autonomy_crew_by_id_undo", matrixAutonomyHandler.UndoCrew)
	// عين ماتركس على المشاريع — المدير والمالك.
	matrixProjectService := service.NewMatrixProjectService(repository.NewMatrixProjectRepository(db))
	matrixAutopilotService.SetProjectDuties(matrixProjectService.Report)
	bookingService.StageLimits = matrixProjectService.Limits
	matrixProjectHandler := handler.NewMatrixProjectHandler(matrixProjectService)
	// ⏱️ تأخير المشاريع (قرار (ع) 10-10)
	projectDelayHandler := handler.NewProjectDelayHandler(db, matrixProjectService)
	// (ع) 10-17: المتأخر بمرحلة مشروع ينطلب منه تفسير تلقائي (مرة كل ٢٤ ساعة لكل مرحلة)
	service.StartProjectDelayEnforcer(db, matrixProjectService.Report, notificationRepo)
	// ⏳ صلاحية عروض الأسعار (ع) 11-07
	service.StartQuotationExpiryReminder(quotationRepo, notificationRepo)
	secs["projects"].Set("GET /api/projects/stage-limits", "projects.get_stage_limits", projectDelayHandler.StageLimits)
	secs["projects"].Set("PUT /api/projects/stage-limits", "projects.put_stage_limits", projectDelayHandler.SetStageLimit)
	fieldApp.Set("POST /api/projects/{id}/delay-reason", "field.project_delay_reason", projectDelayHandler.AddReason)
	secs["projects"].Set("GET /api/projects/my-delays", "projects.get_my_delays", projectDelayHandler.MyDelays)
	procurementWatchHandler := handler.NewProcurementWatchHandler(db)
	secs["monitor"].Set("GET /api/monitor/procurement-watch", "monitor.get_procurement_watch", procurementWatchHandler.Get)
	secs["projects"].Set("GET /api/projects/delays", "projects.get_delays", projectDelayHandler.Delays)
	secs["ai"].Set("GET /api/ai/projects", "ai.get_projects", matrixProjectHandler.Report)
	secs["ai"].Set("GET /api/ai/projects/{id}", "ai.get_projects_by_id", matrixProjectHandler.Chain)

	// ── تقييم الموظفين: ماتركس ٦٠٪ + البشر ٤٠٪ (قرار (ع) 10-05) ──
	// ⚠️ درجة تقييم بس — بلا فلوس. المفتاح matrix_scoring مطفي لحد ما المالك يشغّله.
	matrixScoreService := service.NewMatrixScoreService(repository.NewMatrixScoreRepository(db), matrixChainService,
		repository.NewSystemSwitchRepository(db), matrixProjectService.Report)
	safeguard.Loop("تقييم ماتركس", 13*time.Minute, 20*time.Minute, func() {
		if err := matrixScoreService.RunIfDue(); err != nil {
			log.Printf("matrix score: %v", err)
		}
	})
	staffScoreHandler := handler.NewStaffScoreHandler(matrixScoreService)
	secs["staff_score"].Set("GET /api/staff-score", "staff_score.get_root", staffScoreHandler.Board)
	// 🏅 رصيد نقاط الشهر — تصاعدي بلا حد (قرار (ع) 10-11)
	staffPointsHandler := handler.NewStaffPointsHandler(service.NewStaffPointsService(db, matrixScoreService.Repo()))
	secs["staff_points"].Set("GET /api/staff-points", "staff_points.get_root", staffPointsHandler.Month)
	secs["staff_score"].Set("GET /api/staff-score/me", "staff_score.get_me", staffScoreHandler.Mine)
	secs["staff_score"].Set("GET /api/staff-score/peers", "staff_score.get_peers", staffScoreHandler.Peers)
	secs["staff_score"].Set("GET /api/staff-score/{id}", "staff_score.get_by_id", staffScoreHandler.Detail)
	secs["staff_score"].Set("POST /api/staff-score/run", "staff_score.post_run", staffScoreHandler.Run)
	secs["matrix_score"].Set("POST /api/matrix-score/{id}/cancel", "matrix_score.post_by_id_cancel", staffScoreHandler.Cancel)
	secs["staff_ratings"].Set("GET /api/staff-ratings/coord-pending", "staff_ratings.get_coord_pending", staffScoreHandler.CoordPending)
	secs["staff_ratings"].Set("POST /api/staff-ratings/coord", "staff_ratings.post_coord", staffScoreHandler.RateCoord().ServeHTTP)
	secs["staff_ratings"].Set("GET /api/staff-ratings/audit-pending", "staff_ratings.get_audit_pending", staffScoreHandler.AuditPending)
	secs["staff_ratings"].Set("POST /api/staff-ratings/audit", "staff_ratings.post_audit", staffScoreHandler.RateAudit().ServeHTTP)
	secs["staff_ratings"].Set("POST /api/staff-ratings/quality", "staff_ratings.post_quality", staffScoreHandler.RateQuality().ServeHTTP)
	secs["staff_ratings"].Set("GET /api/staff-ratings/periodic", "staff_ratings.get_periodic", staffScoreHandler.PeriodicState)
	secs["staff_ratings"].Set("POST /api/staff-ratings/periodic", "staff_ratings.post_periodic", staffScoreHandler.RatePeriodic().ServeHTTP)
	// (ون بيس 11-10) requireRater انتقل لقاعدة بالنواة (internal/apps/sections)
	secs["staff_ratings"].Set("GET /api/staff-ratings/booking/{id}", "staff_ratings.get_booking_by_id", staffScoreHandler.BookingState)
	// ═══ ماتركس الوكيل (قرار (ع) 10-10): هايكو يفتش بنفسه بأدوات قراءة ═══
	{
		client, model := matrixCommandService.Client()
		agent := service.NewMatrixAgent(db, client, model)
		agent.RegisterReadTools(service.AgentDeps{
			Command: matrixCommandService, Projects: matrixProjectService.Report,
			Scores: matrixScoreService.Board, ScoreRepo: matrixScoreService.Repo(),
		})
		matrixCommandService.SetAgent(agent)
		trainingCenter.Agent = agent
		// 🗂️ السكرتير (قرار (ع) 10-22): المالك والمدير بس، كل واحد بذاكرته — docs/secretary.md
		secRepo := repository.NewSecretaryRepository(db)
		secHandler := handler.NewSecretaryHandler(secRepo, employeeRepo, extraTaskRepo, notificationRepo, voiceAnalysisService)
		// السكرتير: نفس الدماغ الداخلي ويا ماتركس (دازاي 10-28) — المفتاح AI_KEY_INTERNAL والسقف للدماغ.
		secClient := service.NewAnthropicClient(cfg.AIKeySecretary)
		irisBrain := service.NewSecretaryBrainFor(service.SecretaryDeps{
			DB: db, Client: secClient, Model: cfg.AIModelSecretary, Repo: secRepo,
			Matrix: service.AgentDeps{Command: matrixCommandService, Projects: matrixProjectService.Report,
				Scores: matrixScoreService.Board, ScoreRepo: matrixScoreService.Repo()},
			NotifyOwner: func(msg string) { _ = notificationRepo.CreateForRole("OWNER", "AI_SUGGESTION", msg) },
			Quotes:      quotationService,
		}, cfg.AISecretaryDailyCap) // حارس فاتورة
		irisBrain.SetWebSearch(cfg.AIIrisWebSearch)
		secHandler.SetBrain(irisBrain)
		// 👂 أذن تتعلم (دازاي 10-29): ورقة المصطلحات + كلمات النظام + تصحيحات المالك
		service.DefaultVoiceLexicon = service.NewVoiceLexicon(db)
		service.StartVoiceLexicon(service.DefaultVoiceLexicon)
		voiceLex := handler.NewVoiceLexiconHandler(db)
		secs["voice"].Set("GET /api/voice/terms", "voice.get_terms", voiceLex.List)
		secs["voice"].Set("POST /api/voice/terms", "voice.post_terms", voiceLex.Add)
		secs["voice"].Set("DELETE /api/voice/terms", "voice.delete_terms", voiceLex.Delete)
		// 🎙️ أذن كمال (دازاي 10-29): كل المزوّدات المفعّلة بمفاتيحها، والأساسي AI_IRIS_VOICE.
		// Whisper ما يدخل بمسار كمال (بطيء وضعيف بالعراقي) — يبقى للإنجازات بس.
		var sttAll []service.NamedSTT
		if g := service.NewGeminiTranscriberModel(cfg.GeminiAPIKey, cfg.GeminiSTTModel); g != nil {
			slow := strings.Contains(cfg.GeminiSTTModel, "-pro")
			t := 20 * time.Second
			if slow {
				t = 35 * time.Second
			}
			sttAll = append(sttAll, service.NamedSTT{Name: "gemini", P: g, Timeout: t})
			if cfg.GeminiSTTModel != "" && cfg.GeminiSTTModel != "gemini-flash-latest" {
				sttAll = append(sttAll, service.NamedSTT{Name: "gemini-flash", P: service.NewGeminiTranscriber(cfg.GeminiAPIKey), Timeout: 20 * time.Second})
			}
		}
		if d := service.NewDeepgramSTT(cfg.DeepgramKey, cfg.DeepgramModel); d != nil {
			sttAll = append(sttAll, service.NamedSTT{Name: "deepgram", P: d})
		}
		if s := service.NewSonioxSTT(cfg.SonioxKey, cfg.SonioxModel); s != nil {
			sttAll = append(sttAll, service.NamedSTT{Name: "soniox", P: s})
		}
		if e := service.NewElevenLabsSTT(cfg.ElevenLabsKey, cfg.ElevenSTTModel); e != nil {
			sttAll = append(sttAll, service.NamedSTT{Name: "elevenlabs", P: e})
		}
		var polish func(ctx context.Context, text string) string
		if cfg.VoicePolish && secClient != nil {
			polish = func(ctx context.Context, text string) string {
				return service.PolishTranscript(ctx, secClient, cfg.AIModel, text)
			}
		}
		primary := ""
		for _, p := range sttAll {
			if p.Name == cfg.AIIrisVoice || (primary == "" && cfg.AIIrisVoice != "whisper") {
				primary = p.Name
				secHandler.SetSTT(p.P)
			}
		}
		secHandler.SetVoiceProviders(primary, sttAll, polish)
		log.Printf("🎙️ أذن كمال: الأساسي «%s» من %d مزوّد (تصحيح القاموس: %v)", primary, len(sttAll), polish != nil)
		// 🏛️ الهيكل الجديد (دازاي 10-29): إدارات ووحدات ومناصب — فوك الأدوار
		orgService := service.NewOrgService(db)
		// الكتالوج أول — وإلا حزم الوحدات تنزرع ناقصة على قاعدة جديدة
		if err := permissionRepo.EnsureSeeded(); err != nil {
			log.Printf("permission seed: %v", err)
		}
		if err := orgService.Seed(); err != nil {
			log.Printf("org seed: %v", err)
		}
		// (دازاي 11-06) ماتركس يكتب مسودة خطة الوحدة، والبنود المعتمدة تصير مهام
		orgService.SetPlanDrafter(func(system, task string) (string, error) {
			run, err := agent.Run("CHAT", system, task)
			if err != nil {
				return "", err
			}
			return run.Answer, nil
		})
		orgService.SetTaskCreator(extraTaskRepo.CreateFor)
		specsSuggest.Ask = func(system, task string) (string, error) {
			// (مراجعة 11-06) بلا أي أداة: نص المنتج من موظف — ما ينفّذ أوامر ولا يقرا بيانات
			run, err := agent.ChatStream("CHAT", system, nil, task, nil, service.ChatOpts{AllowTool: func(string) bool { return false }})
			if err != nil {
				return "", err
			}
			return run.Answer, nil
		}
		orgHandler := handler.NewOrgHandler(orgService, notificationRepo)
		orgHandler.PersonalScore = func(empID, period string) (*float64, []service.EvidenceLine) {
			d, err := matrixScoreService.Detail(empID, period)
			if err != nil || d.Final == nil {
				return nil, nil
			}
			lines := []service.EvidenceLine{}
			if d.MatrixPct != nil {
				lines = append(lines, service.EvidenceLine{Label: "نقاط ماتركس", Value: strconv.Itoa(d.Earned) + "/" + strconv.Itoa(d.Max)})
			}
			if d.HumanAvg != nil {
				lines = append(lines, service.EvidenceLine{Label: "التقييم البشري", Value: strconv.FormatFloat(*d.HumanAvg, 'f', 1, 64), Note: strconv.Itoa(d.HumanCount) + " تقييم"})
			}
			for _, l := range d.TopLosses {
				lines = append(lines, service.EvidenceLine{Label: "خسر بـ: " + l.Title, Value: "−" + strconv.Itoa(l.Lost), Note: strconv.Itoa(l.Count) + " مرة"})
			}
			return d.Final, lines
		}
		// 🏛️ ماتركس الحاكم (دازاي 10-29): الغرامات (كل ساعة) والتقييم الشهري (أول ٣ أيام).
		// مفتاح الطوارئ org_governor، والتراجع من صفحة قرارات ماتركس يرجّع النقاط/المنصب.
		orgGovernor := service.NewOrgGovernor(db, orgService, matrixScoreService, disciplineRepo, aiActionRepo, aiRepo,
			repository.NewSystemSwitchRepository(db), notificationRepo)
		orgGovernor.RegisterUndo(matrixAutopilotService.OnUndo)
		// 🔎 التشخيص الذكي (النقطة ٤): ماتركس يفتح/يسكّر مشاكل الوحدات من أرقامها كل ٦ ساعات
		safeguard.Loop("org-diagnosis", 6*time.Minute, 6*time.Hour, func() {
			orgService.RunDiagnosis(repository.NewSystemSwitchRepository(db), func(emp, msg string) { _ = notificationRepo.Create(emp, "ORG_DIAGNOSIS", msg) })
		})
		// المناصب المؤقتة بالإجازات (المبدأ ٢٠) — مستقلة عن مفتاح الحاكم
		safeguard.Loop("org-acting", 3*time.Minute, time.Hour, func() {
			orgService.RunActing(func(emp, msg string) { _ = notificationRepo.Create(emp, "ORG_ACTING", msg) })
		})
		safeguard.Loop("org-governor", 4*time.Minute, time.Hour, func() {
			orgGovernor.RunFines()
			orgGovernor.RunMonthlyIfDue()
		})
		secs["org_x"].Set("GET /api/org/chart", "org_x.get_chart", orgHandler.Chart)
		// 🏴‍☠️ ون بيس: كتابات مكتب الإدارات صارت أفعالاً بالنواة (الفحص على الشي بالخدمة)
		coreRouter.Mount(org.New(org.Deps{
			AddIssue: orgHandler.AddIssue, UpdateIssue: orgHandler.UpdateIssue, SetSolution: orgHandler.SetSolution,
			CreatePlan: orgHandler.CreatePlan, AddPlanItem: orgHandler.AddPlanItem, UpdatePlanItem: orgHandler.UpdatePlanItem,
			SubmitPlan: orgHandler.SubmitPlan, ReviewPlan: orgHandler.ReviewPlan, DraftPlanAI: orgHandler.DraftPlanAI,
			AddComment: orgHandler.AddComment, IssueDirective: orgHandler.IssueDirective,
			AckDirective: orgHandler.AckDirective, ArchiveDirective: orgHandler.ArchiveDirective,
		}))
		secs["org_x"].Set("GET /api/org/me", "org_x.get_me", orgHandler.Me)
		secs["org_x"].Set("POST /api/org/assign", "org_x.post_assign", orgHandler.Assign)
		secs["org_x"].Set("POST /api/org/assignments/{id}/end", "org_x.post_assignments_by_id_end", orgHandler.End)
		// مكتب المدير (المرحلة ٢): الوصول يتحقق بالخدمة حسب المنصب — الي ما يحق له ياخذ 404 مو 403
		secs["org_x"].Set("GET /api/org/desk/{division}", "org_x.get_desk_by_division", orgHandler.Desk)
		// 📋 (دازاي 11-08) كل الخطط بمكان واحد + إرجاع حتى بعد الاعتماد — الخدمة تحصرها بالمالك ومدير النظام
		secs["org_x"].Set("GET /api/org/plans/overview", "org_x.get_plans_overview", orgHandler.PlansOverview)
		secs["org_x"].Set("POST /api/org/plans/{id}/reopen", "org_x.post_plans_by_id_reopen", orgHandler.ReopenPlan)
		secs["org_x"].Set("GET /api/org/people/{id}/executive", "org_x.get_people_by_id_executive", orgHandler.Executive)
		secs["org_x"].Set("GET /api/org/directives", "org_x.get_directives", orgHandler.ManagedDirectives)
		secs["org_x"].Set("GET /api/org/directives/mine", "org_x.get_directives_mine", orgHandler.MyDirectives)
		service.SetDirectiveDeps(orgService, func(emp, msg string) { _ = notificationRepo.Create(emp, "ORG_DIRECTIVE", msg) })
		secs["org_x"].Set("GET /api/org/outputs", "org_x.get_outputs", orgHandler.Outputs)
		secs["org_x"].Set("GET /api/org/directors-ranking", "org_x.get_directors_ranking", orgHandler.DirectorsRanking)
		secs["org_x"].Set("PUT /api/org/tasks/{id}/clause", "org_x.put_tasks_by_id_clause", orgHandler.TaskClause)
		secs["org_x"].Set("PUT /api/org/kpi-criteria/{id}/clause", "org_x.put_kpi_criteria_by_id_clause", orgHandler.CriterionClause)
		secs["org_x"].Set("GET /api/org/my-items", "org_x.get_my_items", orgHandler.MyItems)
		secs["org_x"].Set("GET /api/org/regulations", "org_x.get_regulations", orgHandler.Regulations)
		secs["org_x"].Set("PUT /api/org/authority/{division}", "org_x.put_authority_by_division", orgHandler.SetAuthority)
		secs["voice"].Set("POST /api/voice/compare/{id}", "voice.post_compare_by_id", secHandler.CompareVoice)
		secs["secretary"].Set("POST /api/secretary/message", "secretary.post_message", secHandler.Message)
		secs["secretary"].Set("POST /api/secretary/actions/{id}/confirm", "secretary.post_actions_by_id_confirm", secHandler.Confirm)
		secs["secretary"].Set("POST /api/secretary/actions/{id}/cancel", "secretary.post_actions_by_id_cancel", secHandler.Cancel)
		secs["secretary"].Set("GET /api/secretary/today", "secretary.get_today", secHandler.Today)
		secs["secretary"].Set("GET /api/secretary/audio/{id}", "secretary.get_audio_by_id", secHandler.Audio)
		secs["secretary"].Set("GET /api/secretary/files/{id}", "secretary.get_files_by_id", secHandler.File)
		secs["secretary"].Set("GET /api/secretary/history", "secretary.get_history", secHandler.History)
		// 💬 محادثات ايريس المتعددة (دازاي 10-23)
		secs["secretary"].Set("GET /api/secretary/conversations", "secretary.get_conversations", secHandler.Conversations)
		secs["secretary"].Set("POST /api/secretary/conversations", "secretary.post_conversations", secHandler.NewConversation)
		secs["secretary"].Set("PATCH /api/secretary/conversations/{id}", "secretary.patch_conversations_by_id", secHandler.RenameConversation)
		secs["secretary"].Set("DELETE /api/secretary/conversations/{id}", "secretary.delete_conversations_by_id", secHandler.DeleteConversation)
		// 🗓️ «تذكيراتي» (دازاي 10-22): أي موظف يطلب مقابلة أو ضيف، والمالك والمدير يحددون موعد أو يرفضون
		meetingRepo := repository.NewMeetingRequestRepository(db)
		secHandler.SetMeetings(meetingRepo)
		meetingHandler := handler.NewMeetingRequestHandler(meetingRepo, notificationRepo)
		requestsApp.Set("POST /api/meeting-requests", "requests.meeting", meetingHandler.Create)
		secs["meeting_requests"].Set("GET /api/meeting-requests/mine", "meeting_requests.get_mine", meetingHandler.Mine)
		secs["secretary"].Set("GET /api/secretary/reminders", "secretary.get_reminders", secHandler.Reminders)
		secs["secretary"].Set("POST /api/secretary/reminders/{id}/schedule", "secretary.post_reminders_by_id_schedule", secHandler.Schedule)
		secs["secretary"].Set("POST /api/secretary/reminders/{id}/decline", "secretary.post_reminders_by_id_decline", secHandler.Decline)
		// 📔 «مذكراتي» و«جدول أعمالي» (دازاي 10-23)
		secs["secretary"].Set("GET /api/secretary/notes", "secretary.get_notes", secHandler.Notes)
		secs["secretary"].Set("POST /api/secretary/notes", "secretary.post_notes", secHandler.AddNote)
		secs["secretary"].Set("PUT /api/secretary/notes/{id}/done", "secretary.put_notes_by_id_done", secHandler.NoteDone)
		secs["secretary"].Set("DELETE /api/secretary/notes/{id}", "secretary.delete_notes_by_id", secHandler.DeleteNote)
		// 💬 وكيلي الواتساب: الجودة والحجوزات (دازاي 10-23 — docs/ai-quality-booking-plan.md).
		// مطفيين لحد ما المالك يشغّلهم والواتساب ينربط؛ وضع التجربة للمالك يشتغل هسه.
		{
			service.SetAIBudget(service.WaAgentQuality, cfg.AIQualityDailyUSD)
			service.SetAIBudget(service.WaAgentBooking, cfg.AIBookingDailyUSD)
			waClient := whatsapp.New(cfg.WaToken, cfg.WaPhoneID, cfg.WaAppSecret, cfg.WaVerifyToken)
			waSvc := service.NewWaAgentService(db, repository.NewWaRepository(db), waClient, bookingService,
				qualityFollowUpRepo, repository.NewMatrixProposalRepository(db), bookingDeleteRepo,
				repository.NewSystemSwitchRepository(db), notificationRepo, aiAgentNames.AgentName)
			waSvc.DailyNew, waSvc.TemplateLang = cfg.WaDailyNew, cfg.WaTemplateLng
			waSvc.Templates = map[string]string{service.WaAgentQuality: cfg.WaTemplateQuality, service.WaAgentBooking: cfg.WaTemplateBooking}
			waSvc.PrivacyApproved = cfg.WaPrivacyApproved
			waSvc.SetBrain(service.WaAgentQuality, service.NewLuffyWaBrain(service.WaAgentQuality,
				service.NewAnthropicClientFor(service.WaAgentQuality, cfg.AIKeyQuality), cfg.AIModel, 200))
			waSvc.SetBrain(service.WaAgentBooking, service.NewLuffyWaBrain(service.WaAgentBooking,
				service.NewAnthropicClientFor(service.WaAgentBooking, cfg.AIKeyBooking), cfg.AIModel, 200))
			aiAgentNames.SetEnabled(waSvc.Enabled)
			waH := handler.NewWaHandler(waSvc, waClient, permissionRepo)
			requireWa := authz.Strict(permissionRepo, authz.WhatsAppChats) // (دازاي 11-04) صلاحية بيد المالك
			mux.Handle("GET /api/wa/webhook", http.HandlerFunc(waH.Verify))
			mux.Handle("POST /api/wa/webhook", http.HandlerFunc(waH.Receive))
			secs["wa"].Set("GET /api/wa/conversations", "wa.get_conversations", waH.List, requireWa)
			secs["wa"].Set("GET /api/wa/conversations/{id}", "wa.get_conversations_by_id", waH.Get, requireWa)
			secs["wa"].Set("POST /api/wa/conversations/{id}/takeover", "wa.post_conversations_by_id_takeover", waH.Takeover, requireWa)
			secs["wa"].Set("POST /api/wa/conversations/{id}/release", "wa.post_conversations_by_id_release", waH.Release, requireWa)
			secs["wa"].Set("POST /api/wa/conversations/{id}/send", "wa.post_conversations_by_id_send", waH.Send, requireWa)
			secs["wa"].Set("POST /api/wa/conversations/{id}/close", "wa.post_conversations_by_id_close", waH.Close, requireWa)
			secs["wa"].Set("GET /api/wa/actions", "wa.get_actions", waH.Actions, requireWa)
			secs["wa"].Set("POST /api/wa/actions/{id}/undo", "wa.post_actions_by_id_undo", waH.Undo, requireWa)
			secs["ai"].Set("PUT /api/ai/agents/{agent}/enabled", "ai.put_agents_by_agent_enabled", waH.SetEnabled)
			secs["wa"].Set("POST /api/wa/sandbox/start", "wa.post_sandbox_start", waH.SandboxStart)
			secs["wa"].Set("POST /api/wa/sandbox/{id}/say", "wa.post_sandbox_by_id_say", waH.SandboxSay)
			log.Printf("💬 وكلاء الواتساب: الواتساب %s · ابو الجودة %s · ابو الحجوزات %s",
				map[bool]string{true: "مربوط", false: "مو مربوط"}[waClient.On()],
				map[bool]string{true: "شغّال", false: "مطفي"}[waSvc.Enabled(service.WaAgentQuality)],
				map[bool]string{true: "شغّال", false: "مطفي"}[waSvc.Enabled(service.WaAgentBooking)])
			go func() {
				time.Sleep(2 * time.Minute)
				for {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
					waSvc.Scan(ctx)
					cancel()
					time.Sleep(10 * time.Minute)
				}
			}()
		}
		secs["secretary"].Set("GET /api/secretary/agenda", "secretary.get_agenda", secHandler.Agenda)
		// ☀️ رسالة الصبح (٨ص) + 🏆 إشادة الأسبوع (الخميس) — قرار (ع) 10-11
		morningService := service.NewMatrixMorningService(db, aiRepo, notificationRepo, announcementRepo, agent)
		safeguard.Loop("رسالة الصبح وإشادة الأسبوع", 3*time.Minute, 15*time.Minute, func() {
			if err := morningService.RunMorningIfDue(); err != nil {
				log.Printf("morning brief: %v", err)
			}
			if err := morningService.RunKudosIfDue(); err != nil {
				log.Printf("weekly kudos: %v", err)
			}
			if err := morningService.RunMonitorIfDue(); err != nil {
				log.Printf("monitor daily: %v", err)
			}
		})
		// صوت ماتركس: كل تنبيه بأسلوب جديد (وأي فشل يرجّع الأصل)
		repository.SetNotificationRewriter(func(emp, msg string) string {
			return agent.Rewrite(notificationRepo.RecentAutopilotMessages(emp, 3), msg)
		})
		// 💬 دردشة ماتركس — المالك والمدير
		chatHandler := handler.NewMatrixChatHandler(db, agent)
		chatHandler.SetChatModel(cfg.AIModelChat)
		chatHandler.SetPermissions(func(emp string) []string {
			rows, _ := permissionRepo.ListForEmployee(emp)
			out := make([]string, 0, len(rows))
			for _, p := range rows {
				out = append(out, p.Name)
			}
			return out
		})
		// (دازاي 10-29) محادثة ماتركس لصاحب agent_matrix هم — أدواته حسب صلاحياته
		// (ون بيس 11-10) requireMatrixChat انتقل لقاعدة بالنواة (internal/apps/sections)
		// 🧠 اقتراحات الذكاء (دازاي 10-28): ماتركس يقترح تطوير نفسه
		agent.RegisterSuggestTool(db, "MATRIX", func(msg string) {
			_ = notificationRepo.CreateForRole("OWNER", "AI_SUGGESTION", msg)
		})
		agent.RegisterChangelogTool() // 📜 سجل التحديثات (دازاي 11-06)
		// 🌱 التطوير الذاتي: قياس الأدوات، الفجوات، الدروس المقترحة (دازاي 11-06)
		agent.RegisterGrowthTools(db, "MATRIX", func(msg string) {
			_ = notificationRepo.CreateForRole("OWNER", "AI_SUGGESTION", msg)
		})
		// 🌙 التأكد الليلي من اقتراحات ماتركس الي انگال عليها «تم» (دازاي 11-06)
		selfReviewer := service.NewSelfReviewer(db, agent, "MATRIX") // 🪞 مراجعة ذاتية كل جمعة
		safeguard.Loop("مراجعة ماتركس الذاتية", 6*time.Minute, 15*time.Minute, func() {
			if err := selfReviewer.RunIfDue(); err != nil {
				log.Printf("self review: %v", err)
			}
		})
		// 🧪 امتحان ماتركس (دازاي 11-06): أسئلة ثابتة بعد كل نشر، والأدوات الكاتبة ناشفة
		matrixEval := service.NewAgentEvaluator(db, agent, "MATRIX", chatHandler.EvalSetup, func(msg string) {
			_ = notificationRepo.CreateForRole("OWNER", "AI_SUGGESTION", msg)
		})
		safeguard.Loop("امتحان ماتركس", 7*time.Minute, 20*time.Minute, func() {
			if err := matrixEval.RunIfDue(); err != nil {
				log.Printf("agent eval: %v", err)
			}
		})
		suggestVerifier := service.NewSuggestVerifier(db, agent, "MATRIX")
		safeguard.Loop("تأكد اقتراحات ماتركس", 5*time.Minute, 15*time.Minute, func() {
			if err := suggestVerifier.RunIfDue(); err != nil {
				log.Printf("suggest verify: %v", err)
			}
		})
		secs["matrix"].Set("GET /api/matrix/chats", "matrix.get_chats", chatHandler.List)
		secs["matrix"].Set("POST /api/matrix/chats", "matrix.post_chats", chatHandler.Create)
		secs["matrix"].Set("GET /api/matrix/chats/{id}", "matrix.get_chats_by_id", chatHandler.Messages)
		secs["matrix"].Set("DELETE /api/matrix/chats/{id}", "matrix.delete_chats_by_id", chatHandler.Delete)
		secs["matrix"].Set("POST /api/matrix/chats/{id}/messages", "matrix.post_chats_by_id_messages", chatHandler.Send)
		secs["matrix"].Set("POST /api/matrix/chats/{id}/stream", "matrix.post_chats_by_id_stream", chatHandler.Stream)
		secs["matrix"].Set("POST /api/matrix/chats/{id}/messages/{msgId}/correct", "matrix.post_chats_by_id_messages_by_msgid_correct", chatHandler.Correct)
		secs["matrix"].Set("GET /api/matrix/lessons", "matrix.get_lessons", chatHandler.Lessons)
		secs["matrix"].Set("DELETE /api/matrix/lessons/{id}", "matrix.delete_lessons_by_id", chatHandler.DisableLesson)
		secs["matrix"].Set("POST /api/matrix/lessons/{id}/approve", "matrix.post_lessons_by_id_approve", chatHandler.ApproveLesson)
		// 🌱 👍/👎 على أجوبة الوكلاء وصحتهم (دازاي 11-06)
		growthHandler := handler.NewAgentGrowthHandler(db)
		secs["matrix"].Set("POST /api/matrix/chats/{id}/messages/{msgId}/feedback", "matrix.post_chats_by_id_messages_by_msgid_feedback", growthHandler.MatrixFeedback)
		secs["secretary"].Set("POST /api/secretary/messages/{msgId}/feedback", "secretary.post_messages_by_msgid_feedback", growthHandler.SecretaryFeedback)
		secs["ai"].Set("GET /api/ai/health", "ai.get_health", growthHandler.Health)
		secs["ai"].Set("GET /api/ai/actions", "ai.get_actions", growthHandler.Actions)
		growthHandler.SetEvaluator(matrixEval.RunNow, func() bool { return agent.Enabled("LEARNING") })
		secs["ai"].Set("POST /api/ai/evals/run", "ai.post_evals_run", growthHandler.RunEval)
		// 🎯 «راقب فلان» — يومية كل ليلة وخلاصة بالنهاية
		missionService := service.NewMatrixMissionService(db, aiRepo, notificationRepo, agent)
		safeguard.Loop("مهام ماتركس", 5*time.Minute, 30*time.Minute, func() {
			if err := missionService.RunIfDue(); err != nil {
				log.Printf("matrix missions: %v", err)
			}
		})
		missionHandler := handler.NewMatrixMissionHandler(missionService)
		secs["matrix"].Set("GET /api/matrix/missions", "matrix.get_missions", missionHandler.List)
		secs["matrix"].Set("POST /api/matrix/missions", "matrix.post_missions", missionHandler.Create)
		secs["matrix"].Set("GET /api/matrix/missions/{id}", "matrix.get_missions_by_id", missionHandler.Get)
		secs["matrix"].Set("DELETE /api/matrix/missions/{id}", "matrix.delete_missions_by_id", missionHandler.Stop)
	}

	// ── تكلفة الشبكات ──
	// الاستمارة والحساب: نفس قيد حاسبة الكاميرات (صلاحية حساب التنفيذ).
	// أما تعديل الأسعار فمحصور بالمالك ومدير النظام — سعر يتغيّر يعني
	// كل فاتورة جاية تتغير معاه.
	secs["network_cost"].Set("GET /api/network-cost/items", "network_cost.get_items", networkCostHandler.ListActive)
	secs["network_cost"].Set("POST /api/network-cost/calculate", "network_cost.post_calculate", networkCostHandler.Calculate)
	// تقرير ١ (B3): صلاحية «أسعار الشبكات» تفتح الشاشة — الخادم يقبلها.
	// (ون بيس 11-10) requireNetPrices انتقل لقاعدة بالنواة (internal/apps/sections)
	secs["network_cost"].Set("GET /api/network-cost/prices", "network_cost.get_prices", networkCostHandler.ListAll)
	secs["network_cost"].Set("POST /api/network-cost/prices", "network_cost.post_prices", networkCostHandler.CreatePrice)
	secs["network_cost"].Set("PUT /api/network-cost/prices/{id}", "network_cost.put_prices_by_id", networkCostHandler.UpdatePrice)
	secs["network_cost"].Set("DELETE /api/network-cost/prices/{id}", "network_cost.delete_prices_by_id", networkCostHandler.DeactivatePrice)
	invoicesApp.Set("GET /api/leader-invoices/camera-cost/options", "invoices.camera_options", leaderInvoiceHandler.CameraCostOptions)
	// الاعتماد محصور بمدير/محاسب فقط — الليدر ما يقدر يعتمد فاتورته بنفسه
	// البحث بالفاتورة المحاسبية — لازم يجي قبل مسار {id} حتى ما ينحسب
	// "by-number" معرّف فاتورة
	invoicesApp.Set("GET /api/leader-invoices/by-number", "invoices.by_number", leaderInvoiceHandler.FindByExternalNumber)
	invoicesApp.Set("PUT /api/leader-invoices/{id}/approve", "invoices.approve", leaderInvoiceHandler.Approve)
	// ربط رقم فاتورة محاسبية بفاتورة معتمدة أصلاً، وتعديل المبالغ —
	// الاثنين للمحاسب/المدير حصراً
	// ═══ تدقيق الفاتورة ═══
	// الحكم قبل الاعتماد، وسحب الاعتماد لما ينصار بالغلط.
	invoicesApp.Set("PUT /api/leader-invoices/{id}/audit", "invoices.audit", leaderInvoiceHandler.SetAuditVerdict)
	// ⚠️ المحاسب يرسل للمراقب — نفس حارس بقية عمليات المحاسب.
	invoicesApp.Set("PUT /api/leader-invoices/{id}/monitor-request", "invoices.monitor_request", leaderInvoiceHandler.RequestMonitorReview)
	// ⚠️ والمراقب يبتّ — حارس المراقب مو المحاسب: المحاسب ما يبتّ
	// بطلبه بنفسه، وإلا الإرسال للمراقب يصير شكلياً.
	invoicesApp.Set("PUT /api/leader-invoices/{id}/monitor-decide", "invoices.monitor_decide", leaderInvoiceHandler.DecideMonitorReview)
	// ⚠️⚠️ الإرجاع للمحاسب: **المالك وحده ولا حتى الإداري**. هاي تشيل
	// اعتماداً صار وترجّع فاتورة لأول الطابور — صلاحية بهالوزن بيد
	// أكثر من واحد تعني إن المالك ما يعرف منو غيّر شنو.
	invoicesApp.Set("PUT /api/leader-invoices/{id}/return", "invoices.return", leaderInvoiceHandler.ReturnToAccountant)
	invoicesApp.Set("PUT /api/leader-invoices/{id}/revoke", "invoices.revoke", leaderInvoiceHandler.RevokeApproval)
	// (دازاي 10-25) طلب حذف فاتورة: المحاسب يطلب، والمالك وحده يوافق (نسخة كاملة قبل الحذف)
	invoiceDeleteHandler := handler.NewInvoiceDeleteHandler(db, notificationRepo)
	invoicesApp.Set("POST /api/leader-invoices/{id}/delete-request", "invoices.delete_request", invoiceDeleteHandler.Request)
	invoicesApp.Set("GET /api/leader-invoices/delete-requests", "invoices.delete_requests", invoiceDeleteHandler.List)
	invoicesApp.Set("POST /api/leader-invoices/delete-requests/{id}/decide", "invoices.delete_decide", invoiceDeleteHandler.Decide)
	invoicesApp.Set("GET /api/leader-invoices/approved-without-number", "invoices.no_number", leaderInvoiceHandler.ApprovedWithoutNumber)
	invoicesApp.Set("PUT /api/leader-invoices/{id}/external-number", "invoices.set_number", leaderInvoiceHandler.SetExternalNumber)
	invoicesApp.Set("PUT /api/leader-invoices/{id}/adjust", "invoices.adjust", leaderInvoiceHandler.Adjust)
	// ⚠️ الحارس يطابق حارس شاشة فواتير الليدر بالضبط
	// (`navTree.tsx`: ADMIN·FINANCE·MONITOR + `leader_invoices_view`).
	// حارس المراقبة وحده **يكسر المحاسب** — وهو الي يدقّق المبالغ
	// ويطبع الفاتورة بتعديلاتها.
	invoicesApp.Set("GET /api/leader-invoices/{id}/adjustments", "invoices.adjustments", leaderInvoiceHandler.Adjustments)

	// إحصائيات الموظفين الشهرية — حصراً للمالك/الأدمن (requireAdmin يسمح OWNER
	// تلقائياً لأنه يتخطى أي قيد أدوار بـRequireRole).
	secs["employee_stats"].Set("GET /api/employee-stats/monthly", "employee_stats.get_monthly", employeeStatsHandler.Monthly)
	secs["employee_stats"].Set("GET /api/employee-stats/monthly/export", "employee_stats.get_monthly_export", employeeStatsHandler.MonthlyExport)
	secs["employee_stats"].Set("GET /api/employee-stats/range", "employee_stats.get_range", employeeStatsHandler.Range)
	secs["employee_stats"].Set("GET /api/employee-stats/curve/{employeeId}", "employee_stats.get_curve_by_employeeid", employeeStatsHandler.Curve)

	// إدارة الإحصائيات: يومية/أسبوعية/مشاريع — حصراً لمدير النظام.
	secs["stats_management"].Set("GET /api/stats-management/daily", "stats_management.get_daily", statsManagementHandler.Daily)
	secs["stats_management"].Set("GET /api/stats-management/weekly", "stats_management.get_weekly", statsManagementHandler.Weekly)
	secs["stats_management"].Set("GET /api/stats-management/projects", "stats_management.get_projects", statsManagementHandler.ProjectStages)
	secs["stats_management"].Set("GET /api/stats-management/projects/monthly", "stats_management.get_projects_monthly", statsManagementHandler.ProjectMonths)
	monthCompareHandler := handler.NewMonthCompareHandler(db)
	secs["stats_management"].Set("GET /api/stats-management/month-compare", "stats_management.get_month_compare", monthCompareHandler.Compare)
	secs["stats_management"].Set("GET /api/stats-management/internal-works", "stats_management.get_internal_works", statsManagementHandler.InternalWorks)

	// تقدير مدة العمل المتعلَّم (learned baseline) — قراءة فقط، متاح لأي مستخدم
	// مسجّل دخول (يحتاجها المنسق قبل تثبيت موعد/فريق).
	secs["job_duration_estimate"].Set("GET /api/job-duration-estimate", "job_duration_estimate.get_root", jobDurationHandler.Estimate)

	// 🏴‍☠️ ون بيس: التطبيقات المنقولة — الحارس صار قاعدة بالنواة (الرفض 404)
	coreRouter.Mount(
		meApp, requestsApp, fieldApp, accessApp,
		system.New(system.Deps{SetSwitch: systemSwitchHandler.Set}),
		peopleApp, financeApp, invoicesApp,
	)
	coreRouter.Mount(append(sections.Apps(secs), designsApp, trainingCenterApp)...)

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

// aiBrainsStatus صرف وسقف الدماغين لصفحة الوكلاء (دازاي 10-28).
func aiBrainsStatus(model, chatModel string) map[string]any {
	if chatModel == "" {
		chatModel = model
	}
	out := map[string]any{}
	for _, b := range []string{service.AIBrainInternal, service.AIBrainExternal} {
		spent, limit := service.AIBrainState(b)
		m := map[string]any{"model": model, "spentTodayUSD": spent, "dailyLimitUSD": limit}
		if b == service.AIBrainInternal {
			m["chatModel"] = chatModel
		}
		out[b] = m
	}
	return out
}

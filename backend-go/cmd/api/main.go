package main

import (
	"log"
	"net/http"

	"staffmange-api/internal/config"
	"staffmange-api/internal/database"
	"staffmange-api/internal/handler"
	"staffmange-api/internal/middleware"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/service"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Repositories
	employeeRepo := repository.NewEmployeeRepository(db)
	permissionRepo := repository.NewPermissionRepository(db)
	serviceRepo := repository.NewServiceRepository(db)
	customerRepo := repository.NewCustomerRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	cartRepo := repository.NewCartRepository(db)
	expenseRepo := repository.NewExpenseRepository(db)
	inventoryRepo := repository.NewInventoryRepository(db)
	kpiRepo := repository.NewKpiRepository(db)
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
	statsRepo := repository.NewStatsRepository(db)

	// Services
	authService := service.NewAuthService(employeeRepo, cfg.JWTSecret)
	employeeService := service.NewEmployeeService(employeeRepo)
	permissionService := service.NewPermissionService(permissionRepo, employeeRepo)
	serviceCatalogService := service.NewServiceCatalogService(serviceRepo)
	customerService := service.NewCustomerService(customerRepo)
	bookingService := service.NewBookingService(bookingRepo, employeeRepo)
	cartService := service.NewCartService(cartRepo)
	expenseService := service.NewExpenseService(expenseRepo)
	inventoryService := service.NewInventoryService(inventoryRepo)
	kpiService := service.NewKpiService(kpiRepo)
	smartKpiService := service.NewSmartKpiService(smartKpiRepo)
	complaintService := service.NewComplaintService(complaintRepo)
	trainingService := service.NewTrainingService(trainingRepo)
	missionService := service.NewMissionService(missionRepo)
	projectService := service.NewProjectService(projectRepo)
	procurementService := service.NewProcurementService(procurementRepo)
	supplierService := service.NewSupplierService(supplierRepo)
	quotationService := service.NewQuotationService(quotationRepo)
	productService := service.NewProductService(productRepo)
	gpsService := service.NewGpsService(gpsRepo)
	statsService := service.NewStatsService(statsRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	employeeHandler := handler.NewEmployeeHandler(employeeService)
	permissionHandler := handler.NewPermissionHandler(permissionService)
	serviceHandler := handler.NewServiceHandler(serviceCatalogService)
	customerHandler := handler.NewCustomerHandler(customerService)
	bookingHandler := handler.NewBookingHandler(bookingService)
	cartHandler := handler.NewCartHandler(cartService)
	expenseHandler := handler.NewExpenseHandler(expenseService)
	inventoryHandler := handler.NewInventoryHandler(inventoryService)
	kpiHandler := handler.NewKpiHandler(kpiService)
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
	statsHandler := handler.NewStatsHandler(statsService)

	requireAuth := middleware.RequireAuth(authService)
	requireAdmin := middleware.RequireRole("ADMIN")
	requireFinance := middleware.RequireRole("ADMIN", "FINANCE")
	requireHR := middleware.RequireRole("ADMIN", "HR_COORDINATOR")
	requireMonitor := middleware.RequireRole("ADMIN", "MONITOR")
	requireProjectManager := middleware.RequireRole("ADMIN", "PROJECT_MANAGER")
	requireGpsAdmin := middleware.RequireRole("ADMIN", "GPS_ADMIN")

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)

	// موظفين — القراءة تحتاج تسجيل دخول فقط، الإنشاء/التعديل الحساس محمي بدور ADMIN
	mux.Handle("GET /api/v1/employees", middleware.Chain(http.HandlerFunc(employeeHandler.List), requireAuth))
	mux.Handle("GET /api/v1/employees/{id}", middleware.Chain(http.HandlerFunc(employeeHandler.Get), requireAuth))
	mux.Handle("POST /api/v1/employees", middleware.Chain(http.HandlerFunc(employeeHandler.Create), requireAuth, requireAdmin))
	mux.Handle("PUT /api/v1/employees/{id}", middleware.Chain(http.HandlerFunc(employeeHandler.Update), requireAuth, requireAdmin))

	// الصلاحيات — العرض متاح لأي مسجل دخول، التعديل والتطبيق التلقائي محصور بمدير النظام فقط
	mux.Handle("GET /api/v1/permissions", middleware.Chain(http.HandlerFunc(permissionHandler.ListAll), requireAuth))
	mux.Handle("GET /api/v1/permissions/role-defaults", middleware.Chain(http.HandlerFunc(permissionHandler.RoleDefaults), requireAuth))
	mux.Handle("GET /api/v1/permissions/employee/{id}", middleware.Chain(http.HandlerFunc(permissionHandler.ListForEmployee), requireAuth))
	mux.Handle("PUT /api/v1/permissions/employee/{id}", middleware.Chain(http.HandlerFunc(permissionHandler.SetForEmployee), requireAuth, requireAdmin))
	mux.Handle("POST /api/v1/permissions/employee/{id}/apply-defaults", middleware.Chain(http.HandlerFunc(permissionHandler.ApplyDefaults), requireAuth, requireAdmin))

	// الخدمات والمهارات — القراءة لأي مسجل دخول، الإضافة لمدير النظام فقط
	mux.Handle("GET /api/v1/services", middleware.Chain(http.HandlerFunc(serviceHandler.List), requireAuth))
	mux.Handle("POST /api/v1/services", middleware.Chain(http.HandlerFunc(serviceHandler.Create), requireAuth, requireAdmin))

	// العملاء — أي مسجل دخول يقدر يبحث وينشئ عميل (يطابق سلوك المبيعات بالباك إند القديم)
	mux.Handle("GET /api/v1/customers", middleware.Chain(http.HandlerFunc(customerHandler.List), requireAuth))
	mux.Handle("GET /api/v1/customers/lookup", middleware.Chain(http.HandlerFunc(customerHandler.Lookup), requireAuth))
	mux.Handle("POST /api/v1/customers", middleware.Chain(http.HandlerFunc(customerHandler.FindOrCreate), requireAuth))

	// الحجوزات — دورة حياة الحجز الكاملة، كل خطوة تتطلب تسجيل دخول فقط (الصلاحية الدقيقة تُفرض بالواجهة حالياً)
	mux.Handle("GET /api/v1/bookings", middleware.Chain(http.HandlerFunc(bookingHandler.List), requireAuth))
	mux.Handle("POST /api/v1/bookings", middleware.Chain(http.HandlerFunc(bookingHandler.Create), requireAuth))
	mux.Handle("PUT /api/v1/bookings/{id}/confirm", middleware.Chain(http.HandlerFunc(bookingHandler.Confirm), requireAuth))
	mux.Handle("PUT /api/v1/bookings/{id}/details", middleware.Chain(http.HandlerFunc(bookingHandler.UpdateDetails), requireAuth))
	mux.Handle("PUT /api/v1/bookings/{id}/schedule", middleware.Chain(http.HandlerFunc(bookingHandler.Schedule), requireAuth))
	mux.Handle("GET /api/v1/bookings/{id}/schedule-log", middleware.Chain(http.HandlerFunc(bookingHandler.ScheduleLog), requireAuth))
	mux.Handle("PUT /api/v1/bookings/{id}/assign", middleware.Chain(http.HandlerFunc(bookingHandler.Assign), requireAuth))
	mux.Handle("PUT /api/v1/bookings/{id}/supervisor", middleware.Chain(http.HandlerFunc(bookingHandler.Supervisor), requireAuth))
	mux.Handle("PUT /api/v1/bookings/{id}/start", middleware.Chain(http.HandlerFunc(bookingHandler.Start), requireAuth))
	mux.Handle("PUT /api/v1/bookings/{id}/complete", middleware.Chain(http.HandlerFunc(bookingHandler.Complete), requireAuth))
	mux.Handle("PUT /api/v1/bookings/{id}/verify", middleware.Chain(http.HandlerFunc(bookingHandler.Verify), requireAuth, requireFinance))

	// سلة الحجز
	mux.Handle("GET /api/v1/cart/booking/{bookingId}", middleware.Chain(http.HandlerFunc(cartHandler.ListForBooking), requireAuth))
	mux.Handle("POST /api/v1/cart/booking/{bookingId}", middleware.Chain(http.HandlerFunc(cartHandler.Create), requireAuth))
	mux.Handle("PUT /api/v1/cart/{id}", middleware.Chain(http.HandlerFunc(cartHandler.Update), requireAuth))
	mux.Handle("DELETE /api/v1/cart/{id}", middleware.Chain(http.HandlerFunc(cartHandler.Delete), requireAuth))

	// المصاريف — أي موظف يقدر يرسل مصروف، الموافقة/الرفض للمحاسب ومدير النظام فقط
	mux.Handle("GET /api/v1/expenses", middleware.Chain(http.HandlerFunc(expenseHandler.List), requireAuth))
	mux.Handle("POST /api/v1/expenses", middleware.Chain(http.HandlerFunc(expenseHandler.Create), requireAuth))
	mux.Handle("PUT /api/v1/expenses/{id}/status", middleware.Chain(http.HandlerFunc(expenseHandler.UpdateStatus), requireAuth, requireFinance))

	// المخزون — أدوات شخصية / مركبات / أدوات مشتركة / طلبات الأدوات
	mux.Handle("GET /api/v1/inventory/personal", middleware.Chain(http.HandlerFunc(inventoryHandler.ListPersonalTools), requireAuth))
	mux.Handle("POST /api/v1/inventory/personal", middleware.Chain(http.HandlerFunc(inventoryHandler.CreatePersonalTool), requireAuth))
	mux.Handle("PUT /api/v1/inventory/personal/{id}", middleware.Chain(http.HandlerFunc(inventoryHandler.UpdatePersonalTool), requireAuth))
	mux.Handle("DELETE /api/v1/inventory/personal/{id}", middleware.Chain(http.HandlerFunc(inventoryHandler.DeletePersonalTool), requireAuth, requireHR))

	mux.Handle("GET /api/v1/inventory/vehicle", middleware.Chain(http.HandlerFunc(inventoryHandler.ListVehicleTools), requireAuth))
	mux.Handle("POST /api/v1/inventory/vehicle", middleware.Chain(http.HandlerFunc(inventoryHandler.CreateVehicleTool), requireAuth))
	mux.Handle("PUT /api/v1/inventory/vehicle/{id}", middleware.Chain(http.HandlerFunc(inventoryHandler.UpdateVehicleTool), requireAuth))
	mux.Handle("DELETE /api/v1/inventory/vehicle/{id}", middleware.Chain(http.HandlerFunc(inventoryHandler.DeleteVehicleTool), requireAuth))

	mux.Handle("GET /api/v1/inventory/ondemand", middleware.Chain(http.HandlerFunc(inventoryHandler.ListOnDemandTools), requireAuth))
	mux.Handle("POST /api/v1/inventory/ondemand", middleware.Chain(http.HandlerFunc(inventoryHandler.CreateOnDemandTool), requireAuth))
	mux.Handle("PUT /api/v1/inventory/ondemand/{id}", middleware.Chain(http.HandlerFunc(inventoryHandler.UpdateOnDemandTool), requireAuth))

	mux.Handle("GET /api/v1/inventory/requests", middleware.Chain(http.HandlerFunc(inventoryHandler.ListToolRequests), requireAuth))
	mux.Handle("POST /api/v1/inventory/requests", middleware.Chain(http.HandlerFunc(inventoryHandler.CreateToolRequest), requireAuth))
	mux.Handle("PUT /api/v1/inventory/requests/{id}/approve", middleware.Chain(http.HandlerFunc(inventoryHandler.ApproveToolRequest), requireAuth, requireHR))
	mux.Handle("PUT /api/v1/inventory/requests/{id}/reject", middleware.Chain(http.HandlerFunc(inventoryHandler.RejectToolRequest), requireAuth, requireHR))
	mux.Handle("PUT /api/v1/inventory/requests/{id}/return", middleware.Chain(http.HandlerFunc(inventoryHandler.ReturnToolRequest), requireAuth))

	// تقييم الأداء اليدوي (KPI)
	mux.Handle("GET /api/v1/kpi", middleware.Chain(http.HandlerFunc(kpiHandler.List), requireAuth))
	mux.Handle("GET /api/v1/kpi/employee/{employeeId}", middleware.Chain(http.HandlerFunc(kpiHandler.ListForEmployee), requireAuth))
	mux.Handle("POST /api/v1/kpi", middleware.Chain(http.HandlerFunc(kpiHandler.Create), requireAuth, requireMonitor))
	mux.Handle("DELETE /api/v1/kpi/{id}", middleware.Chain(http.HandlerFunc(kpiHandler.Delete), requireAuth, requireAdmin))

	// تقييم الأداء التلقائي (Smart KPI) — الرانك الأسبوعي/الشهري للفنيين
	mux.Handle("GET /api/v1/smart-kpi/technician/{employeeId}", middleware.Chain(http.HandlerFunc(smartKpiHandler.Technician), requireAuth))
	mux.Handle("GET /api/v1/smart-kpi/leaderboard", middleware.Chain(http.HandlerFunc(smartKpiHandler.Leaderboard), requireAuth))

	// الشكاوى
	mux.Handle("GET /api/v1/complaints", middleware.Chain(http.HandlerFunc(complaintHandler.List), requireAuth))
	mux.Handle("POST /api/v1/complaints", middleware.Chain(http.HandlerFunc(complaintHandler.Create), requireAuth))
	mux.Handle("PUT /api/v1/complaints/{id}", middleware.Chain(http.HandlerFunc(complaintHandler.Update), requireAuth))
	mux.Handle("PUT /api/v1/complaints/{id}/resolve", middleware.Chain(http.HandlerFunc(complaintHandler.Resolve), requireAuth))

	// التدريب — عرض متاح لأي مسجل دخول، التعيين وإدارة المواد لمدير النظام فقط
	mux.Handle("GET /api/v1/training/materials/mine", middleware.Chain(http.HandlerFunc(trainingHandler.MaterialsMine), requireAuth))
	mux.Handle("GET /api/v1/training/assignments/{employeeId}", middleware.Chain(http.HandlerFunc(trainingHandler.Assignments), requireAuth))
	mux.Handle("PUT /api/v1/training/assignments/{employeeId}", middleware.Chain(http.HandlerFunc(trainingHandler.SetAssignments), requireAuth, requireAdmin))
	mux.Handle("GET /api/v1/training/materials", middleware.Chain(http.HandlerFunc(trainingHandler.ListMaterials), requireAuth))
	mux.Handle("POST /api/v1/training/materials", middleware.Chain(http.HandlerFunc(trainingHandler.CreateMaterial), requireAuth, requireAdmin))
	mux.Handle("PUT /api/v1/training/materials/{id}", middleware.Chain(http.HandlerFunc(trainingHandler.UpdateMaterial), requireAuth, requireAdmin))
	mux.Handle("DELETE /api/v1/training/materials/{id}", middleware.Chain(http.HandlerFunc(trainingHandler.DeleteMaterial), requireAuth, requireAdmin))

	// تتبع المهام (missions)
	mux.Handle("GET /api/v1/missions", middleware.Chain(http.HandlerFunc(missionHandler.List), requireAuth))
	mux.Handle("GET /api/v1/missions/monitor/live", middleware.Chain(http.HandlerFunc(missionHandler.MonitorLive), requireAuth))
	mux.Handle("GET /api/v1/missions/reports/performance", middleware.Chain(http.HandlerFunc(missionHandler.PerformanceReport), requireAuth))
	mux.Handle("GET /api/v1/missions/my/{employeeId}", middleware.Chain(http.HandlerFunc(missionHandler.ListForEmployee), requireAuth))
	mux.Handle("GET /api/v1/missions/{id}", middleware.Chain(http.HandlerFunc(missionHandler.Get), requireAuth))
	mux.Handle("POST /api/v1/missions", middleware.Chain(http.HandlerFunc(missionHandler.Create), requireAuth))
	mux.Handle("PUT /api/v1/missions/{id}/stage", middleware.Chain(http.HandlerFunc(missionHandler.UpdateStage), requireAuth))

	// إدارة المشاريع (projects)
	mux.Handle("GET /api/v1/projects", middleware.Chain(http.HandlerFunc(projectHandler.List), requireAuth))
	mux.Handle("POST /api/v1/projects", middleware.Chain(http.HandlerFunc(projectHandler.Create), requireAuth))
	mux.Handle("PUT /api/v1/projects/{id}", middleware.Chain(http.HandlerFunc(projectHandler.Update), requireAuth))
	mux.Handle("DELETE /api/v1/projects/{id}", middleware.Chain(http.HandlerFunc(projectHandler.Delete), requireAuth, requireProjectManager))

	// المشتريات (procurement)
	mux.Handle("GET /api/v1/procurement", middleware.Chain(http.HandlerFunc(procurementHandler.List), requireAuth))
	mux.Handle("GET /api/v1/procurement/stats", middleware.Chain(http.HandlerFunc(procurementHandler.Stats), requireAuth))
	mux.Handle("POST /api/v1/procurement", middleware.Chain(http.HandlerFunc(procurementHandler.Create), requireAuth))
	mux.Handle("PUT /api/v1/procurement/{id}/status", middleware.Chain(http.HandlerFunc(procurementHandler.UpdateStatus), requireAuth, requireFinance))
	mux.Handle("PUT /api/v1/procurement/{id}/fulfill", middleware.Chain(http.HandlerFunc(procurementHandler.Fulfill), requireAuth, requireFinance))

	// الموردون (suppliers)
	mux.Handle("GET /api/v1/suppliers/specialties", middleware.Chain(http.HandlerFunc(supplierHandler.ListSpecialties), requireAuth))
	mux.Handle("POST /api/v1/suppliers/specialties", middleware.Chain(http.HandlerFunc(supplierHandler.CreateSpecialty), requireAuth, requireAdmin))
	mux.Handle("DELETE /api/v1/suppliers/specialties/{id}", middleware.Chain(http.HandlerFunc(supplierHandler.DeleteSpecialty), requireAuth, requireAdmin))
	mux.Handle("GET /api/v1/suppliers", middleware.Chain(http.HandlerFunc(supplierHandler.List), requireAuth))
	mux.Handle("POST /api/v1/suppliers", middleware.Chain(http.HandlerFunc(supplierHandler.Create), requireAuth, requireAdmin))
	mux.Handle("PUT /api/v1/suppliers/{id}", middleware.Chain(http.HandlerFunc(supplierHandler.Update), requireAuth, requireAdmin))
	mux.Handle("DELETE /api/v1/suppliers/{id}", middleware.Chain(http.HandlerFunc(supplierHandler.Delete), requireAuth, requireAdmin))
	mux.Handle("POST /api/v1/suppliers/{id}/rate", middleware.Chain(http.HandlerFunc(supplierHandler.Rate), requireAuth))

	// عروض الأسعار (quotations)
	mux.Handle("GET /api/v1/quotations", middleware.Chain(http.HandlerFunc(quotationHandler.List), requireAuth))
	mux.Handle("GET /api/v1/quotations/{id}", middleware.Chain(http.HandlerFunc(quotationHandler.Get), requireAuth))
	mux.Handle("POST /api/v1/quotations", middleware.Chain(http.HandlerFunc(quotationHandler.Create), requireAuth))
	mux.Handle("PUT /api/v1/quotations/{id}", middleware.Chain(http.HandlerFunc(quotationHandler.Update), requireAuth))
	mux.Handle("DELETE /api/v1/quotations/{id}", middleware.Chain(http.HandlerFunc(quotationHandler.Delete), requireAuth, requireAdmin))

	// المنتجات (products)
	mux.Handle("GET /api/v1/products", middleware.Chain(http.HandlerFunc(productHandler.List), requireAuth))
	mux.Handle("POST /api/v1/products", middleware.Chain(http.HandlerFunc(productHandler.Create), requireAuth))
	mux.Handle("PUT /api/v1/products/{id}", middleware.Chain(http.HandlerFunc(productHandler.Update), requireAuth, requireAdmin))
	mux.Handle("DELETE /api/v1/products/{id}", middleware.Chain(http.HandlerFunc(productHandler.Delete), requireAuth, requireAdmin))

	// نظام GPS — عملاء / شرائح SIM / طلبات الأجهزة / التجديد / الصيانة / الأسعار / الإحصائيات
	mux.Handle("GET /api/v1/gps/customers", middleware.Chain(http.HandlerFunc(gpsHandler.ListCustomers), requireAuth))
	mux.Handle("POST /api/v1/gps/customers", middleware.Chain(http.HandlerFunc(gpsHandler.CreateCustomer), requireAuth))
	mux.Handle("PUT /api/v1/gps/customers/{id}", middleware.Chain(http.HandlerFunc(gpsHandler.UpdateCustomer), requireAuth))

	mux.Handle("GET /api/v1/gps/sims", middleware.Chain(http.HandlerFunc(gpsHandler.ListSims), requireAuth))
	mux.Handle("POST /api/v1/gps/sims", middleware.Chain(http.HandlerFunc(gpsHandler.CreateSim), requireAuth))
	mux.Handle("PUT /api/v1/gps/sims/{id}", middleware.Chain(http.HandlerFunc(gpsHandler.UpdateSim), requireAuth))

	mux.Handle("GET /api/v1/gps/devices", middleware.Chain(http.HandlerFunc(gpsHandler.ListDevices), requireAuth))
	mux.Handle("POST /api/v1/gps/devices", middleware.Chain(http.HandlerFunc(gpsHandler.CreateDevice), requireAuth))
	mux.Handle("PUT /api/v1/gps/devices/{id}", middleware.Chain(http.HandlerFunc(gpsHandler.UpdateDevice), requireAuth))

	mux.Handle("GET /api/v1/gps/renewals", middleware.Chain(http.HandlerFunc(gpsHandler.ListRenewals), requireAuth))
	mux.Handle("POST /api/v1/gps/renewals", middleware.Chain(http.HandlerFunc(gpsHandler.CreateRenewal), requireAuth))
	mux.Handle("PUT /api/v1/gps/renewals/{id}", middleware.Chain(http.HandlerFunc(gpsHandler.UpdateRenewal), requireAuth))

	mux.Handle("GET /api/v1/gps/maintenance", middleware.Chain(http.HandlerFunc(gpsHandler.ListMaintenance), requireAuth))
	mux.Handle("POST /api/v1/gps/maintenance", middleware.Chain(http.HandlerFunc(gpsHandler.CreateMaintenance), requireAuth))
	mux.Handle("PUT /api/v1/gps/maintenance/{id}", middleware.Chain(http.HandlerFunc(gpsHandler.UpdateMaintenance), requireAuth))

	mux.Handle("GET /api/v1/gps/settings", middleware.Chain(http.HandlerFunc(gpsHandler.ListSettings), requireAuth))
	mux.Handle("PUT /api/v1/gps/settings", middleware.Chain(http.HandlerFunc(gpsHandler.UpsertSettings), requireAuth, requireGpsAdmin))

	mux.Handle("GET /api/v1/gps/stats", middleware.Chain(http.HandlerFunc(gpsHandler.Stats), requireAuth))

	// الإحصائيات العامة (stats) — لوحة معلومات المدير/المشرف
	mux.Handle("GET /api/v1/stats", middleware.Chain(http.HandlerFunc(statsHandler.Overview), requireAuth))

	handlerChain := middleware.Chain(mux, middleware.Recovery, middleware.Logging, middleware.CORS)

	log.Printf("staffmange-api listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, handlerChain); err != nil {
		log.Fatal(err)
	}
}

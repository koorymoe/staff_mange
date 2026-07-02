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

	// Services
	authService := service.NewAuthService(employeeRepo, cfg.JWTSecret)
	employeeService := service.NewEmployeeService(employeeRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	employeeHandler := handler.NewEmployeeHandler(employeeService)

	requireAuth := middleware.RequireAuth(authService)
	requireAdmin := middleware.RequireRole("ADMIN")

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

	handlerChain := middleware.Chain(mux, middleware.Recovery, middleware.Logging, middleware.CORS)

	log.Printf("staffmange-api listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, handlerChain); err != nil {
		log.Fatal(err)
	}
}

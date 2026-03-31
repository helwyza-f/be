package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	// Import semua domain baru
	"github.com/helwiza/saas/internal/auth"
	"github.com/helwiza/saas/internal/customer"
	"github.com/helwiza/saas/internal/reservation"
	"github.com/helwiza/saas/internal/resource"
	"github.com/helwiza/saas/internal/tenant"

	"github.com/helwiza/saas/internal/platform/database"
	"github.com/helwiza/saas/internal/platform/http"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env untuk koneksi DB dan AWS S3 keys
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// 1. Database Connection (sqlx)
	db, err := database.NewPostgres(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)
	if err != nil {
		log.Fatalf("DB Connection Error: %v", err)
	}
	defer db.Close()

	// 2. Database Migration (v4)
	runMigration(db.DB)

	// 3. Dependency Injection (Wiring)
	
	// --- AUTH DOMAIN ---
	authSvc := auth.NewService()
	authHdl := auth.NewHandler(authSvc)

	// --- TENANT DOMAIN ---
	tenantRepo := tenant.NewRepository(db)
	tenantSvc  := tenant.NewService(tenantRepo, authSvc) // Butuh authSvc buat Login (JWT)
	tenantHdl  := tenant.NewHandler(tenantSvc)

	// --- RESOURCE DOMAIN ---
	resourceRepo := resource.NewRepository(db)
	resourceSvc  := resource.NewService(resourceRepo)
	resourceHdl  := resource.NewHandler(resourceSvc)

	// --- RESERVATION DOMAIN ---
	reservationRepo := reservation.NewRepository(db)
	reservationSvc  := reservation.NewService(reservationRepo)
	reservationHdl  := reservation.NewHandler(reservationSvc)

	// --- CUSTOMER DOMAIN ---
	customerRepo := customer.NewRepository(db)
	customerSvc  := customer.NewService(customerRepo)
	customerHdl  := customer.NewHandler(customerSvc)

	// 4. Setup Router dengan Config Baru
	routerConfig := http.Config{
		TenantHandler:      tenantHdl,
		ResourceHandler:    resourceHdl,
		ReservationHandler: reservationHdl,
		CustomerHandler:    customerHdl,
		AuthHandler:        authHdl,
	}
	
	r := http.NewRouter(routerConfig)

	// 5. Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("🚀 Server bookinaja-api running on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

func runMigration(db *sql.DB) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Migration driver error: %v", err)
	}
	
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver,
	)
	if err != nil {
		log.Fatalf("Migration init error: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration failed: %v", err)
	}
	
	log.Println("✅ Database migration completed.")
}
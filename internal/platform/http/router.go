package http

import (
	"github.com/gin-gonic/gin"
	"github.com/helwiza/saas/internal/auth"
	"github.com/helwiza/saas/internal/customer"
	"github.com/helwiza/saas/internal/middleware"
	"github.com/helwiza/saas/internal/reservation"
	"github.com/helwiza/saas/internal/resource"
	"github.com/helwiza/saas/internal/tenant"
)

type Config struct {
    TenantHandler      *tenant.Handler
    ResourceHandler    *resource.Handler
    ReservationHandler *reservation.Handler
    CustomerHandler    *customer.Handler
    AuthHandler        *auth.Handler
}

func NewRouter(cfg Config) *gin.Engine {
    r := gin.Default()

    r.RedirectTrailingSlash = false
    r.RedirectFixedPath = false

    r.Use(middleware.CORSMiddleware())
    r.Use(gin.Recovery())
    r.Use(gin.Logger())

    r.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "pong"})
    })

    v1 := r.Group("/api/v1")
    {
        // --- PUBLIC ROUTES (No Auth Required) ---
        v1.GET("/public/landing", cfg.TenantHandler.GetPublicLandingData)
        v1.GET("/public/resources/:id", cfg.ResourceHandler.GetPublicDetail)
        v1.POST("/public/bookings", cfg.ReservationHandler.Create)
        
        // PINDAH KE SINI: Agar customer publik bisa dapet live validation WA
        v1.GET("/validate-phone", cfg.CustomerHandler.ValidatePhone) 

        v1.POST("/register", cfg.TenantHandler.Register)
        v1.POST("/login", cfg.TenantHandler.Login)

        // --- GUEST ROUTES (Magic Link Access) ---
        guest := v1.Group("/guest")
        {
            guest.GET("/availability/:resource_id", cfg.ReservationHandler.Availability)
            guest.GET("/status/:token", cfg.ReservationHandler.Status)
        }

        // --- PROTECTED ROUTES (Admin/Tenant Only) ---
        protected := v1.Group("/")
        protected.Use(middleware.AuthMiddleware())
        {
            // Admin Profile & Settings
            admin := protected.Group("/admin")
            {
                admin.GET("/profile", cfg.TenantHandler.GetProfile)
                admin.PUT("/profile", cfg.TenantHandler.UpdateProfile)
                admin.POST("/upload", cfg.TenantHandler.UploadImage)
                // (ValidatePhone sudah dihapus dari sini)
            }

            // RESOURCE MANAGEMENT
            resources := protected.Group("/resources-all")
            {
                resources.GET("", cfg.ResourceHandler.List)
                resources.POST("", cfg.ResourceHandler.Create)
                resources.DELETE("/:id", cfg.ResourceHandler.Delete)

                resources.GET("/:id/items", cfg.ResourceHandler.ListItems)
                resources.POST("/:id/items", cfg.ResourceHandler.AddItem)
                resources.PUT("/items/:id", cfg.ResourceHandler.UpdateItem)
                resources.DELETE("/items/:id", cfg.ResourceHandler.DeleteItem)
            }

            // BOOKING MANAGEMENT
            bookings := protected.Group("/bookings")
            {
                bookings.GET("", cfg.ReservationHandler.ListAll)
                bookings.GET("/:id", cfg.ReservationHandler.GetDetail)
                bookings.PUT("/:id/status", cfg.ReservationHandler.UpdateStatus)
                bookings.POST("/manual", cfg.ReservationHandler.Create)
            }

            // CUSTOMER MANAGEMENT
            customers := protected.Group("/customers")
            {
                customers.GET("", cfg.CustomerHandler.List)
                customers.POST("", cfg.CustomerHandler.Create)
            }

            protected.GET("/me", cfg.AuthHandler.CheckMe)
        }
    }

    return r
}
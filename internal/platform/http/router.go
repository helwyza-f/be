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
	
	// FIX: Matikan auto-redirect trailing slash agar tidak memicu CORS error
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
		v1.GET("/public/landing", cfg.TenantHandler.GetPublicLandingData)
		v1.POST("/register", cfg.TenantHandler.Register)
		v1.POST("/login", cfg.TenantHandler.Login)

		guest := v1.Group("/guest")
		{
			guest.GET("/availability/:resource_id", cfg.ReservationHandler.Availability)
			guest.GET("/status/:token", cfg.ReservationHandler.Status)
		}

		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware()) 
		{
			admin := protected.Group("/admin")
			{   
				admin.GET("/profile", cfg.TenantHandler.GetProfile)
				admin.PUT("/profile", cfg.TenantHandler.UpdateProfile)
				admin.POST("/upload", cfg.TenantHandler.UploadImage)
			}

			// RESOURCE ROUTES
			// Menggunakan path langsung tanpa slash di awal agar match dengan /api/v1/resources
			resources := protected.Group("/resources-all")
			{
			resources.GET("", cfg.ResourceHandler.List)
			resources.POST("", cfg.ResourceHandler.Create)
			resources.DELETE("/:id", cfg.ResourceHandler.Delete)
			
			// Items Management
			resources.GET("/:id/items", cfg.ResourceHandler.ListItems) 
			resources.POST("/:id/items", cfg.ResourceHandler.AddItem)
			
			// Perhatikan path-nya: /api/v1/resources-all/items/:id
			resources.PUT("/items/:id", cfg.ResourceHandler.UpdateItem)   // Tambahkan ini
			resources.DELETE("/items/:id", cfg.ResourceHandler.DeleteItem) // Tambahkan ini (opsional)
		}

			// RESERVATION ROUTES
			bookings := protected.Group("/bookings")
			{
				bookings.POST("", cfg.ReservationHandler.Create)
			}

			// CUSTOMER ROUTES
			customers := protected.Group("/customers")
			{
				customers.POST("", cfg.CustomerHandler.Create)
				customers.GET("", cfg.CustomerHandler.List)
			}

			protected.GET("/me", cfg.AuthHandler.CheckMe)
		}
	}

	return r
}
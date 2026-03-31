package tenant

import (
	"time"

	"github.com/google/uuid"
	"github.com/helwiza/saas/internal/booking" // Import shared models
	"github.com/lib/pq"
)

// Request untuk pendaftaran tenant baru
type RegisterReq struct {
	TenantName   string `json:"tenant_name" binding:"required"`
	TenantSlug   string `json:"tenant_slug" binding:"required"`
	BusinessType string `json:"business_type"`
	AdminName    string `json:"admin_name" binding:"required"`
	AdminEmail   string `json:"admin_email" binding:"required,email"`
	AdminPass    string `json:"admin_password" binding:"required,min=6"`
}

type LoginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  booking.User `json:"user"`
}

// Tenant mewakili tabel 'tenants'
type Tenant struct {
	ID           uuid.UUID      `db:"id" json:"id"`
	Name         string         `db:"name" json:"name"`
	Slug         string         `db:"slug" json:"slug"`
	BusinessType string         `db:"business_type" json:"business_type"`
	Slogan       string         `db:"slogan" json:"slogan"`
	Address      string         `db:"address" json:"address"`
	OpenTime     string         `db:"open_time" json:"open_time"`
	CloseTime    string         `db:"close_time" json:"close_time"`
	LogoURL      string         `db:"logo_url" json:"logo_url"`
	BannerURL    string         `db:"banner_url" json:"banner_url"`
	Gallery      pq.StringArray `db:"gallery" json:"gallery"` // Array string untuk galeri
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
}
package customer

import (
	"time"

	"github.com/google/uuid"
)

// Customer adalah entitas pelanggan per tenant
type Customer struct {
	ID            uuid.UUID `db:"id" json:"id"`
	TenantID      uuid.UUID `db:"tenant_id" json:"tenant_id"`
	Name          string    `db:"name" json:"name"`
	Phone         string    `db:"phone" json:"phone"`
	Email         *string   `db:"email" json:"email"`
	LoyaltyPoints int       `db:"loyalty_points" json:"loyalty_points"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

// RegisterReq DTO untuk pendaftaran pelanggan baru
type RegisterReq struct {
	Name  string  `json:"name" binding:"required"`
	Phone string  `json:"phone" binding:"required"`
	Email *string `json:"email"`
}
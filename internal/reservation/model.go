package reservation

import (
	"time"

	"github.com/google/uuid"
)

type Booking struct {
	ID          uuid.UUID `db:"id" json:"id"`
	TenantID    uuid.UUID `db:"tenant_id" json:"tenant_id"`
	CustomerID  uuid.UUID `db:"customer_id" json:"customer_id"`
	ResourceID  uuid.UUID `db:"resource_id" json:"resource_id"`
	StartTime   time.Time `db:"start_time" json:"start_time"`
	EndTime     time.Time `db:"end_time" json:"end_time"`
	AccessToken uuid.UUID `db:"access_token" json:"access_token"`
	Status      string    `db:"status" json:"status"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type BookingOption struct {
	ID             uuid.UUID `db:"id" json:"id"`
	BookingID      uuid.UUID `db:"booking_id" json:"booking_id"`
	ResourceItemID uuid.UUID `db:"resource_item_id" json:"resource_item_id"`
	PriceAtBooking float64   `db:"price_at_booking" json:"price_at_booking"`
}

type CreateBookingReq struct {
	TenantID      string   `json:"tenant_id" binding:"required"`
	ResourceID    string   `json:"resource_id" binding:"required"`
	CustomerName  string   `json:"customer_name" binding:"required"`
	CustomerPhone string   `json:"customer_phone" binding:"required"`
	ItemIDs       []string `json:"item_ids"`
	StartTime     string   `json:"start_time" binding:"required"`
	Duration      int      `json:"duration" binding:"required"`
}

type BookingDetail struct {
	Booking
	CustomerName  string          `db:"customer_name" json:"customer_name"`
	CustomerPhone string          `db:"customer_phone" json:"customer_phone"`
	ResourceName  string          `db:"resource_name" json:"resource_name"`
	TotalAmount   float64         `db:"total_amount" json:"total_amount"`
	Options       []BookingOptionDetail `json:"options"`
}

type BookingOptionDetail struct {
	ID             uuid.UUID `json:"id"`
	ItemName       string    `db:"item_name" json:"item_name"`
	ItemType       string    `db:"item_type" json:"item_type"`
	PriceAtBooking float64   `db:"price_at_booking" json:"price_at_booking"`
}
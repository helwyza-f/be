package reservation

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetOrCreateCustomer(ctx context.Context, tenantID uuid.UUID, name, phone string) (uuid.UUID, error) {
	var customerID uuid.UUID
	queryFind := `SELECT id FROM customers WHERE tenant_id = $1 AND phone = $2 LIMIT 1`
	err := r.db.GetContext(ctx, &customerID, queryFind, tenantID, phone)

	if err == sql.ErrNoRows {
		customerID = uuid.New()
		queryInsert := `INSERT INTO customers (id, tenant_id, name, phone) VALUES ($1, $2, $3, $4)`
		_, err = r.db.ExecContext(ctx, queryInsert, customerID, tenantID, name, phone)
		if err != nil {
			return uuid.Nil, err
		}
	}
	return customerID, err
}

func (r *Repository) CheckAvailability(ctx context.Context, resourceID uuid.UUID, start, end time.Time) (bool, error) {
	var count int
	// Menggunakan OVERLAPS untuk memastikan tidak ada tabrakan jadwal
	query := `
		SELECT COUNT(*) 
		FROM bookings 
		WHERE resource_id = $1 
		AND status NOT IN ('cancelled', 'rejected')
		AND (start_time, end_time) OVERLAPS ($2, $3)`

	err := r.db.GetContext(ctx, &count, query, resourceID, start, end)
	return count == 0, err
}

// CreateWithItems Update: Menambahkan parameter duration untuk hitung harga
func (r *Repository) CreateWithItems(ctx context.Context, b Booking, itemIDs []uuid.UUID, duration int) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("repo: gagal memulai transaksi: %w", err)
	}
	defer tx.Rollback()

	// 1. Insert Master Booking
	queryBooking := `
		INSERT INTO bookings (id, tenant_id, customer_id, resource_id, start_time, end_time, access_token, status, created_at)
		VALUES (:id, :tenant_id, :customer_id, :resource_id, :start_time, :end_time, :access_token, :status, :created_at)`

	_, err = tx.NamedExecContext(ctx, queryBooking, b)
	if err != nil {
		return fmt.Errorf("repo: gagal insert booking: %w", err)
	}

	// 2. Insert Booking Options dengan Harga x Durasi
	if len(itemIDs) > 0 {
		// FIX: price_per_hour dikalikan dengan duration ($4)
		queryItem := `
			INSERT INTO booking_options (id, booking_id, resource_item_id, price_at_booking)
			SELECT uuid_generate_v4(), $1, id, (price_per_hour * $4) 
			FROM resource_items 
			WHERE id = $2 AND resource_id = $3`

		for _, itemID := range itemIDs {
			_, err = tx.ExecContext(ctx, queryItem, b.ID, itemID, b.ResourceID, duration)
			if err != nil {
				return fmt.Errorf("repo: gagal insert booking option [%s]: %w", itemID, err)
			}
		}
	}

	return tx.Commit()
}

func (r *Repository) GetByToken(ctx context.Context, token uuid.UUID) (*BookingDetail, error) {
	var b BookingDetail

	query := `
		SELECT 
			b.*, 
			c.name as customer_name, 
			c.phone as customer_phone, 
			r.name as resource_name,
			COALESCE((SELECT SUM(price_at_booking) FROM booking_options WHERE booking_id = b.id), 0) as total_amount
		FROM bookings b
		JOIN customers c ON b.customer_id = c.id
		JOIN resources r ON b.resource_id = r.id
		WHERE b.access_token = $1 
		LIMIT 1`

	err := r.db.GetContext(ctx, &b, query, token)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var options []BookingOptionDetail
	queryOptions := `
		SELECT 
			bo.id, 
			ri.name as item_name, 
			ri.item_type, 
			bo.price_at_booking
		FROM booking_options bo
		JOIN resource_items ri ON bo.resource_item_id = ri.id
		WHERE bo.booking_id = $1`

	err = r.db.SelectContext(ctx, &options, queryOptions, b.ID)
	if err == nil {
		b.Options = options
	}

	return &b, nil
}

func (r *Repository) ListUpcoming(ctx context.Context, resourceID uuid.UUID, from time.Time) ([]Booking, error) {
	var bookings []Booking
	// Query tetap sama, tapi end_time > from memastikan jam di tengah durasi juga kena tarik
	query := `SELECT * FROM bookings WHERE resource_id = $1 AND end_time > $2 AND status != 'cancelled' ORDER BY start_time ASC`
	err := r.db.SelectContext(ctx, &bookings, query, resourceID, from)
	return bookings, err
}

func (r *Repository) FindAllByTenant(ctx context.Context, tenantID uuid.UUID, status string) ([]BookingDetail, error) {
	var res []BookingDetail
	query := `
		SELECT b.*, c.name as customer_name, c.phone as customer_phone, r.name as resource_name,
		COALESCE((SELECT SUM(price_at_booking) FROM booking_options WHERE booking_id = b.id), 0) as total_amount
		FROM bookings b
		JOIN customers c ON b.customer_id = c.id
		JOIN resources r ON b.resource_id = r.id
		WHERE b.tenant_id = $1`

	if status != "" {
		query += " AND b.status = $2 ORDER BY b.created_at DESC"
		err := r.db.SelectContext(ctx, &res, query, tenantID, status)
		return res, err
	}

	err := r.db.SelectContext(ctx, &res, query+" ORDER BY b.created_at DESC", tenantID)
	return res, err
}

func (r *Repository) UpdateStatus(ctx context.Context, id, tenantID uuid.UUID, status string) error {
	query := `UPDATE bookings SET status = $1 WHERE id = $2 AND tenant_id = $3`
	_, err := r.db.ExecContext(ctx, query, status, id, tenantID)
	return err
}

func (r *Repository) FindByID(ctx context.Context, id, tenantID uuid.UUID) (*BookingDetail, error) {
	var b BookingDetail
	query := `
		SELECT b.*, c.name as customer_name, c.phone as customer_phone, r.name as resource_name,
		COALESCE((SELECT SUM(price_at_booking) FROM booking_options WHERE booking_id = b.id), 0) as total_amount
		FROM bookings b
		JOIN customers c ON b.customer_id = c.id
		JOIN resources r ON b.resource_id = r.id
		WHERE b.id = $1 AND b.tenant_id = $2
		LIMIT 1`
	err := r.db.GetContext(ctx, &b, query, id, tenantID)

	var options []BookingOptionDetail
	queryOptions := `
		SELECT bo.id, ri.name as item_name, ri.item_type, bo.price_at_booking
		FROM booking_options bo
		JOIN resource_items ri ON bo.resource_item_id = ri.id
		WHERE bo.booking_id = $1`
	r.db.SelectContext(ctx, &options, queryOptions, b.ID)
	b.Options = options

	return &b, err
}

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

func (r *Repository) CheckAvailability(ctx context.Context, resourceID uuid.UUID, start, end time.Time) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*) 
		FROM bookings 
		WHERE resource_id = $1 
		AND status != 'cancelled'
		AND (start_time, end_time) OVERLAPS ($2, $3)`

	err := r.db.GetContext(ctx, &count, query, resourceID, start, end)
	if err != nil {
		return false, fmt.Errorf("repo: gagal cek ketersediaan: %w", err)
	}
	return count == 0, nil
}

func (r *Repository) CreateWithItems(ctx context.Context, b Booking, itemIDs []uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.NamedExecContext(ctx, `
		INSERT INTO bookings (id, tenant_id, customer_id, resource_id, start_time, end_time, access_token, status)
		VALUES (:id, :tenant_id, :customer_id, :resource_id, :start_time, :end_time, :access_token, :status)`, b)
	if err != nil { return err }

	for _, itemID := range itemIDs {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO booking_options (id, booking_id, resource_item_id, price_at_booking)
			SELECT uuid_generate_v4(), $1, id, price_per_hour 
			FROM resource_items WHERE id = $2`, b.ID, itemID)
		if err != nil { return err }
	}

	return tx.Commit()
}

func (r *Repository) GetByToken(ctx context.Context, token uuid.UUID) (*Booking, error) {
	var b Booking
	err := r.db.GetContext(ctx, &b, `SELECT * FROM bookings WHERE access_token = $1 LIMIT 1`, token)
	if err == sql.ErrNoRows { return nil, nil }
	return &b, err
}

func (r *Repository) ListUpcoming(ctx context.Context, resourceID uuid.UUID, from time.Time) ([]Booking, error) {
	var bookings []Booking
	query := `SELECT * FROM bookings WHERE resource_id = $1 AND end_time > $2 AND status != 'cancelled' ORDER BY start_time ASC`
	err := r.db.SelectContext(ctx, &bookings, query, resourceID, from)
	return bookings, err
}
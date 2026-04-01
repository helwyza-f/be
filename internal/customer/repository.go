package customer

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, c Customer) error {
	query := `INSERT INTO customers (id, tenant_id, name, phone, email, loyalty_points, created_at) 
              VALUES (:id, :tenant_id, :name, :phone, :email, :loyalty_points, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, c)
	if err != nil {
		return fmt.Errorf("repo: gagal simpan customer: %w", err)
	}
	return nil
}

func (r *Repository) FindByTenant(ctx context.Context, tenantID uuid.UUID) ([]Customer, error) {
	var customers []Customer
	query := `SELECT * FROM customers WHERE tenant_id = $1 ORDER BY name ASC`
	err := r.db.SelectContext(ctx, &customers, query, tenantID)
	return customers, err
}

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (*Customer, error) {
	var c Customer
	query := `SELECT * FROM customers WHERE id = $1 LIMIT 1`
	err := r.db.GetContext(ctx, &c, query, id)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *Repository) FindByPhone(ctx context.Context, tenantID uuid.UUID, phone string) (*Customer, error) {
	var c Customer
	query := `SELECT * FROM customers WHERE tenant_id = $1 AND phone = $2 LIMIT 1`
	err := r.db.GetContext(ctx, &c, query, tenantID, phone)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}
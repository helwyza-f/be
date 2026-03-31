package tenant

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/helwiza/saas/internal/booking"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateWithAdmin(ctx context.Context, t booking.Tenant, u booking.User) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil { return err }
	defer tx.Rollback()

	_, err = tx.NamedExecContext(ctx, `INSERT INTO tenants (id, name, slug, business_type) VALUES (:id, :name, :slug, :business_type)`, t)
	if err != nil { return err }

	_, err = tx.NamedExecContext(ctx, `INSERT INTO users (id, tenant_id, name, email, password, role) VALUES (:id, :tenant_id, :name, :email, :password, :role)`, u)
	if err != nil { return err }

	return tx.Commit()
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*booking.User, error) {
	var u booking.User
	err := r.db.GetContext(ctx, &u, `SELECT * FROM users WHERE email = $1 LIMIT 1`, email)
	if err == sql.ErrNoRows { return nil, nil }
	return &u, err
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*booking.Tenant, error) {
	var t booking.Tenant
	err := r.db.GetContext(ctx, &t, `SELECT * FROM tenants WHERE id = $1`, id)
	return &t, err
}

func (r *Repository) GetBySlug(ctx context.Context, slug string) (*booking.Tenant, error) {
	var t booking.Tenant
	err := r.db.GetContext(ctx, &t, `SELECT * FROM tenants WHERE slug = $1`, slug)
	return &t, err
}

func (r *Repository) Update(ctx context.Context, t booking.Tenant) error {
	query := `UPDATE tenants SET name=:name, slogan=:slogan, address=:address, open_time=:open_time, 
			  close_time=:close_time, logo_url=:logo_url, banner_url=:banner_url, gallery=:gallery WHERE id=:id`
	_, err := r.db.NamedExecContext(ctx, query, t)
	return err
}

func (r *Repository) Exists(ctx context.Context, slug, email string) (bool, bool, error) {
	var slugExists, emailExists bool
	err := r.db.GetContext(ctx, &slugExists, "SELECT EXISTS(SELECT 1 FROM tenants WHERE slug = $1)", slug)
	err = r.db.GetContext(ctx, &emailExists, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email)
	return slugExists, emailExists, err
}

func (r *Repository) ListResources(ctx context.Context, tenantID uuid.UUID) ([]booking.Resource, error) {
	var res []booking.Resource
	
	// Kita ambil semua resource yang statusnya bukan 'deleted'
	query := `SELECT * FROM resources WHERE tenant_id = $1 AND status != 'deleted' ORDER BY created_at DESC`
	
	err := r.db.SelectContext(ctx, &res, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("repo: gagal mengambil daftar resource: %w", err)
	}
	
	return res, nil
}
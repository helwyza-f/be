package tenant

import (
	"context"
	"database/sql"
	"encoding/json"

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

func (r *Repository) CreateWithAdmin(ctx context.Context, t Tenant, u booking.User) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil { return err }
	defer tx.Rollback()

	// Update query untuk menyertakan business_category
	_, err = tx.NamedExecContext(ctx, `
		INSERT INTO tenants (id, name, slug, business_category, business_type) 
		VALUES (:id, :name, :slug, :business_category, :business_type)`, t)
	if err != nil { return err }

	_, err = tx.NamedExecContext(ctx, `
		INSERT INTO users (id, tenant_id, name, email, password, role) 
		VALUES (:id, :tenant_id, :name, :email, :password, :role)`, u)
	if err != nil { return err }

	return tx.Commit()
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*booking.User, error) {
	var u booking.User
	err := r.db.GetContext(ctx, &u, `SELECT * FROM users WHERE email = $1 LIMIT 1`, email)
	if err == sql.ErrNoRows { return nil, nil }
	return &u, err
}

func (r *Repository) GetBySlug(ctx context.Context, slug string) (*Tenant, error) {
	var t Tenant
	err := r.db.GetContext(ctx, &t, `SELECT * FROM tenants WHERE slug = $1`, slug)
	if err == sql.ErrNoRows { return nil, nil }
	return &t, err
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Tenant, error) {
	var t Tenant
	err := r.db.GetContext(ctx, &t, `SELECT * FROM tenants WHERE id = $1`, id)
	return &t, err
}

func (r *Repository) Update(ctx context.Context, t Tenant) error {
	query := `UPDATE tenants SET 
		name=:name, slogan=:slogan, address=:address, open_time=:open_time, 
		close_time=:close_time, logo_url=:logo_url, banner_url=:banner_url, 
		gallery=:gallery, business_category=:business_category WHERE id=:id`
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
	query := `SELECT * FROM resources WHERE tenant_id = $1 AND status != 'deleted' ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &res, query, tenantID)
	return res, err
}

func (r *Repository) ListResourcesWithItems(ctx context.Context, tenantID uuid.UUID) ([]map[string]interface{}, error) {
    // Gunakan ::text untuk kolom id dan di dalam json_build_object
    query := `
        SELECT 
            r.id::text as id, 
            r.name, 
            r.category, 
            r.status,
            COALESCE(json_agg(json_build_object(
                'id', i.id::text, 
                'name', i.name,
                'item_type', i.item_type,
                'price_per_hour', i.price_per_hour
            )) FILTER (WHERE i.id IS NOT NULL), '[]') as items
        FROM resources r
        LEFT JOIN resource_items i ON r.id = i.resource_id
        WHERE r.tenant_id = $1 AND r.status != 'deleted'
        GROUP BY r.id
        ORDER BY r.created_at DESC`

    rows, err := r.db.QueryxContext(ctx, query, tenantID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var results []map[string]interface{}
    for rows.Next() {
        res := make(map[string]interface{})
        err := rows.MapScan(res)
        if err != nil {
            return nil, err
        }

        // --- PENTING: Fix untuk json_agg yang terbaca sebagai []byte ---
        // Kadang json_agg tetap dikirim sebagai []byte oleh driver
        if itemsBytes, ok := res["items"].([]byte); ok {
            var itemsArray []map[string]interface{}
            if err := json.Unmarshal(itemsBytes, &itemsArray); err == nil {
                res["items"] = itemsArray
            }
        }

        results = append(results, res)
    }

    return results, nil
}
package resource

import (
	"context"

	"github.com/google/uuid"
	"github.com/helwiza/saas/internal/booking"
)

type Service struct {
	repo *Repository
}

func NewService(r *Repository) *Service {
	return &Service{repo: r}
}

func (s *Service) CreateResource(ctx context.Context, tenantID, name, category string) (*booking.Resource, error) {
	tID, _ := uuid.Parse(tenantID)
	res := booking.Resource{
		ID:       uuid.New(),
		TenantID: tID,
		Name:     name,
		Category: category,
		Status:   "available",
		Metadata: []byte("{}"),

	}
	return s.repo.Create(ctx, res)
}

func (s *Service) ListResources(ctx context.Context, tenantID string) ([]booking.Resource, error) {
	tID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, err
	}
	// Repository sekarang sudah mengembalikan Resource + Items di dalamnya
	return s.repo.ListByTenant(ctx, tID)
}

func (s *Service) AddResourceItem(ctx context.Context, resID, name string, price float64, iType string, isDefault bool) (*booking.ResourceItem, error) {
	rID, _ := uuid.Parse(resID)
	item := booking.ResourceItem{
		ID:           uuid.New(),
		ResourceID:   rID,
		Name:         name,
		PricePerHour: price,
		ItemType:     iType,
		IsDefault:    isDefault,
		Metadata:     []byte("{}"),
	}
	return s.repo.CreateItem(ctx, item)
}

func (s *Service) GetItems(ctx context.Context, resourceID string) ([]booking.ResourceItem, error) {
	rID, err := uuid.Parse(resourceID)
	if err != nil { return nil, err }
	return s.repo.ListItemsByResource(ctx, rID)
}

func (s *Service) DeleteResource(ctx context.Context, id string) error {
	uID, err := uuid.Parse(id)
	if err != nil { return err }
	return s.repo.Delete(ctx, uID)
}

func (s *Service) UpdateItem(ctx context.Context, id string, item booking.ResourceItem) error {
	uID, err := uuid.Parse(id)
	if err != nil { return err }
	return s.repo.UpdateItem(ctx, uID, item)
}

func (s *Service) DeleteItem(ctx context.Context, id string) error {
	uID, err := uuid.Parse(id)
	if err != nil { return err }
	return s.repo.DeleteItem(ctx, uID)
}

func (s *Service) GetResourceDetail(ctx context.Context, id string) (*booking.Resource, error) {
    uID, err := uuid.Parse(id)
    if err != nil { return nil, err }
    
    // Gunakan repo yang sudah kita buat tadi
    return s.repo.GetOneWithItems(ctx, uID)
}
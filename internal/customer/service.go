package customer

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(r *Repository) *Service {
	return &Service{repo: r}
}

func (s *Service) Register(ctx context.Context, tenantID string, req RegisterReq) (*Customer, error) {
	tID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, fmt.Errorf("service: id tenant tidak valid")
	}

	cust := Customer{
		ID:            uuid.New(),
		TenantID:      tID,
		Name:          req.Name,
		Phone:         req.Phone,
		LoyaltyPoints: 0,
		CreatedAt:     time.Now(),
	}

	if req.Email != "" {
		cust.Email = &req.Email
	}

	if err := s.repo.Create(ctx, cust); err != nil {
		return nil, err
	}

	return &cust, nil
}

func (s *Service) ListByTenant(ctx context.Context, tenantID string) ([]Customer, error) {
	tID, _ := uuid.Parse(tenantID)
	return s.repo.FindByTenant(ctx, tID)
}
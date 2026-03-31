package reservation

import (
	"context"
	"errors"
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

func (s *Service) Create(ctx context.Context, tenantID string, req CreateBookingReq) (*Booking, error) {
	tID, err := uuid.Parse(tenantID)
	if err != nil { return nil, errors.New("ID tenant tidak valid") }
	
	cID, err := uuid.Parse(req.CustomerID)
	if err != nil { return nil, errors.New("ID customer tidak valid") }
	
	rID, err := uuid.Parse(req.ResourceID)
	if err != nil { return nil, errors.New("ID resource tidak valid") }

	start, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil { return nil, errors.New("format waktu salah, gunakan ISO8601/RFC3339") }
	
	end := start.Add(time.Duration(req.Duration) * time.Hour)
	if start.Before(time.Now()) { return nil, errors.New("tidak bisa booking untuk waktu lampau") }

	available, err := s.repo.CheckAvailability(ctx, rID, start, end)
	if err != nil { return nil, fmt.Errorf("gagal cek slot: %w", err) }
	if !available { return nil, errors.New("slot waktu sudah terisi") }

	var itemUUIDs []uuid.UUID
	for _, idStr := range req.ItemIDs {
		if uID, err := uuid.Parse(idStr); err == nil {
			itemUUIDs = append(itemUUIDs, uID)
		}
	}

	newBooking := Booking{
		ID:          uuid.New(),
		TenantID:    tID,
		CustomerID:  cID,
		ResourceID:  rID,
		StartTime:   start,
		EndTime:     end,
		AccessToken: uuid.New(),
		Status:      "pending",
		CreatedAt:   time.Now(),
	}

	if err := s.repo.CreateWithItems(ctx, newBooking, itemUUIDs); err != nil {
		return nil, fmt.Errorf("gagal menyimpan booking: %w", err)
	}

	return &newBooking, nil
}

func (s *Service) GetAvailability(ctx context.Context, resourceID string, date time.Time) ([]Booking, error) {
	rID, err := uuid.Parse(resourceID)
	if err != nil { return nil, errors.New("ID resource tidak valid") }
	
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	return s.repo.ListUpcoming(ctx, rID, startOfDay)
}

func (s *Service) GetStatusByToken(ctx context.Context, token string) (*Booking, error) {
	tkn, err := uuid.Parse(token)
	if err != nil { return nil, errors.New("format token tidak valid") }
	return s.repo.GetByToken(ctx, tkn)
}
package reservation

import (
	"context"
	"errors"
	"fmt"

	"time"

	"github.com/google/uuid"
	"github.com/helwiza/saas/internal/platform/fonnte"
)

type Service struct {
	repo *Repository
}

func NewService(r *Repository) *Service {
	return &Service{repo: r}
}

func (s *Service) Create(ctx context.Context, req CreateBookingReq) (*Booking, error) {
	// 1. VALIDASI WA VIA FONNTE
	isValid, err := fonnte.ValidateNumber(req.CustomerPhone)
	if err != nil || !isValid {
		return nil, errors.New("NOMOR WHATSAPP TIDAK TERDAFTAR ATAU TIDAK AKTIF")
	}

	tID, _ := uuid.Parse(req.TenantID)
	rID, _ := uuid.Parse(req.ResourceID)
	start, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil { return nil, errors.New("FORMAT WAKTU SALAH") }
	
	end := start.Add(time.Duration(req.Duration) * time.Hour)

	// 2. SILENT REGISTER CUSTOMER
	cID, err := s.repo.GetOrCreateCustomer(ctx, tID, req.CustomerName, req.CustomerPhone)
	if err != nil { return nil, fmt.Errorf("GAGAL MENGIDENTIFIKASI CUSTOMER") }

	// 3. CEK KETERSEDIAAN
	available, err := s.repo.CheckAvailability(ctx, rID, start, end)
	if err != nil || !available { return nil, errors.New("SLOT WAKTU SUDAH TERISI") }

	var itemUUIDs []uuid.UUID
	for _, idStr := range req.ItemIDs {
		if uID, err := uuid.Parse(idStr); err == nil { itemUUIDs = append(itemUUIDs, uID) }
	}

	newBooking := Booking{
		ID:          uuid.New(),
		TenantID:    tID,
		CustomerID:  cID,
		ResourceID:  rID,
		StartTime:   start.UTC(),
		EndTime:     end.UTC(),
		AccessToken: uuid.New(),
		Status:      "pending",
		CreatedAt:   time.Now(),
	}

	if err := s.repo.CreateWithItems(ctx, newBooking, itemUUIDs); err != nil {
		return nil, err
	}

	return &newBooking, nil
}

func (s *Service) GetAvailability(ctx context.Context, resourceID string, date time.Time) ([]Booking, error) {
	rID, err := uuid.Parse(resourceID)
	if err != nil { return nil, errors.New("ID resource tidak valid") }
	
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	return s.repo.ListUpcoming(ctx, rID, startOfDay)
}

func (s *Service) GetStatusByToken(ctx context.Context, token string) (*BookingDetail, error) {
	tkn, err := uuid.Parse(token)
	if err != nil { return nil, errors.New("FORMAT TOKEN TIDAK VALID") }
	return s.repo.GetByToken(ctx, tkn)
}

func (s *Service) ListByTenant(ctx context.Context, tenantID, status string) ([]BookingDetail, error) {
    tID, _ := uuid.Parse(tenantID)
    return s.repo.FindAllByTenant(ctx, tID, status)
}

func (s *Service) GetDetailForAdmin(ctx context.Context, id, tenantID string) (*BookingDetail, error) {
    bID, _ := uuid.Parse(id)
    tID, _ := uuid.Parse(tenantID)
    return s.repo.FindByID(ctx, bID, tID)
}

func (s *Service) UpdateStatus(ctx context.Context, id, tenantID, status string) error {
    bID, _ := uuid.Parse(id)
    tID, _ := uuid.Parse(tenantID)
    
    // Logic tambahan: Jika status diubah jadi 'ongoing', pastikan unit/resource set jadi 'busy'?
    // Bisa ditambah di sini nanti.
    
    return s.repo.UpdateStatus(ctx, bID, tID, status)
}
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

func (s *Service) Create(ctx context.Context, req CreateBookingReq) (*Booking, error) {
	// 1. PARSE WAKTU (FIX TIME DRIFT)
	// Kita tidak pakai RFC3339 (Z) lagi agar jam tidak loncat +7
	// Kita baca string "2026-04-02T16:00:00" sebagai waktu lokal
	layout := "2006-01-02T15:04:05"
	start, err := time.ParseInLocation(layout, req.StartTime, time.Local)
	if err != nil {
		// Fallback jika frontend masih kirim format lain
		start, err = time.Parse(time.RFC3339, req.StartTime)
		if err != nil {
			return nil, errors.New("FORMAT WAKTU SALAH, GUNAKAN YYYY-MM-DDTHH:MM:SS")
		}
	}

	tID, _ := uuid.Parse(req.TenantID)
	rID, _ := uuid.Parse(req.ResourceID)

	// Hitung End Time berdasarkan durasi
	end := start.Add(time.Duration(req.Duration) * time.Hour)

	// 2. SILENT REGISTER CUSTOMER
	cID, err := s.repo.GetOrCreateCustomer(ctx, tID, req.CustomerName, req.CustomerPhone)
	if err != nil {
		return nil, fmt.Errorf("GAGAL MENGIDENTIFIKASI CUSTOMER")
	}

	// 3. CEK KETERSEDIAAN (Tabrakan Jadwal)
	available, err := s.repo.CheckAvailability(ctx, rID, start, end)
	if err != nil || !available {
		return nil, errors.New("SLOT WAKTU SUDAH TERISI")
	}

	// 4. PREPARE ITEMS & BOOKING
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
		StartTime:   start, // Simpan waktu lokal
		EndTime:     end,   // Simpan waktu lokal
		AccessToken: uuid.New(),
		Status:      "pending",
		CreatedAt:   time.Now(),
	}

	// 5. SIMPAN KE DATABASE (PASTIKAN REPO MENGKALIKAN HARGA DENGAN DURATION)
	if err := s.repo.CreateWithItems(ctx, newBooking, itemUUIDs, req.Duration); err != nil {
		return nil, err
	}

	return &newBooking, nil
}

func (s *Service) GetAvailability(ctx context.Context, resourceID string, date time.Time) ([]string, error) {
	rID, err := uuid.Parse(resourceID)
	if err != nil {
		return nil, errors.New("ID RESOURCE TIDAK VALID")
	}

	// Ambil semua booking dari awal hari sampai akhir hari
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
	bookings, err := s.repo.ListUpcoming(ctx, rID, startOfDay)
	if err != nil {
		return nil, err
	}

	// Logic Gahar: Pecah booking durasi panjang menjadi per jam agar UI bisa tandai FULL
	busySlots := []string{}
	for _, b := range bookings {
		curr := b.StartTime
		for curr.Before(b.EndTime) {
			// Format "15:00" agar match dengan TIME_SLOTS di Frontend
			busySlots = append(busySlots, curr.Format("15:04"))
			curr = curr.Add(time.Hour)
		}
	}

	return busySlots, nil
}

func (s *Service) GetStatusByToken(ctx context.Context, token string) (*BookingDetail, error) {
	tkn, err := uuid.Parse(token)
	if err != nil {
		return nil, errors.New("FORMAT TOKEN TIDAK VALID")
	}
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
	return s.repo.UpdateStatus(ctx, bID, tID, status)
}

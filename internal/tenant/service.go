package tenant

import (
	"context"
	"errors"
	"strings"
	"time"
	"github.com/google/uuid"
	"github.com/helwiza/saas/internal/auth"
	"github.com/helwiza/saas/internal/booking"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo *Repository
	authService *auth.Service
}

func NewService(r *Repository, authService *auth.Service) *Service {
	return &Service{repo: r, authService: authService}
}

func (s *Service) Register(ctx context.Context, req RegisterReq) (*Tenant, error) {
	slug := strings.ToLower(strings.TrimSpace(req.TenantSlug))
	slugEx, emailEx, err := s.repo.Exists(ctx, slug, req.AdminEmail)
	if err != nil { return nil, err }
	if slugEx { return nil, errors.New("subdomain sudah digunakan") }
	if emailEx { return nil, errors.New("email sudah terdaftar") }

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.AdminPass), bcrypt.DefaultCost)
	if err != nil { return nil, err }
	
	tID := uuid.New()
	tenant := Tenant{
		ID:               tID,
		Name:             req.TenantName,
		Slug:             slug,
		BusinessCategory: req.BusinessCategory,
		BusinessType:     req.BusinessType,
		CreatedAt:        time.Now(),
	}
	
	user := booking.User{
		ID:        uuid.New(),
		TenantID:  tID,
		Name:      req.AdminName,
		Email:     req.AdminEmail,
		Password:  string(hashed),
		Role:      "owner",
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateWithAdmin(ctx, tenant, user); err != nil { return nil, err }
	return &tenant, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*LoginResponse, error) {
	u, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil { return nil, err }
	if u == nil { return nil, errors.New("email atau password salah") }

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return nil, errors.New("email atau password salah")
	}

	token, err := s.authService.GenerateToken(u.ID, u.TenantID, u.Role)
	if err != nil { return nil, err }

	return &LoginResponse{Token: token, User: *u}, nil
}

func (s *Service) GetProfile(ctx context.Context, id uuid.UUID) (*Tenant, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) UpdateProfile(ctx context.Context, id uuid.UUID, req Tenant) (*Tenant, error) {
	curr, err := s.repo.GetByID(ctx, id)
	if err != nil || curr == nil { return nil, errors.New("tenant tidak ditemukan") }
	
	req.ID = id
	req.Slug = curr.Slug // Protect slug
	if err := s.repo.Update(ctx, req); err != nil { return nil, err }
	return &req, nil
}
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

func (s *Service) Register(ctx context.Context, req RegisterReq) (*booking.Tenant, error) {
	slug := strings.ToLower(strings.TrimSpace(req.TenantSlug))
	slugEx, emailEx, _ := s.repo.Exists(ctx, slug, req.AdminEmail)
	if slugEx { return nil, errors.New("subdomain sudah digunakan") }
	if emailEx { return nil, errors.New("email sudah terdaftar") }

	hashed, _ := bcrypt.GenerateFromPassword([]byte(req.AdminPass), bcrypt.DefaultCost)
	tID := uuid.New()
	
	tenant := booking.Tenant{ID: tID, Name: req.TenantName, Slug: slug, BusinessType: req.BusinessType, CreatedAt: time.Now()}
	user := booking.User{ID: uuid.New(), TenantID: tID, Name: req.AdminName, Email: req.AdminEmail, Password: string(hashed), Role: "owner", CreatedAt: time.Now()}

	if err := s.repo.CreateWithAdmin(ctx, tenant, user); err != nil { return nil, err }
	return &tenant, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*LoginResponse, error) {
	u, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err // Return error database yang sebenarnya
	}
	if u == nil {
		return nil, errors.New("email atau password salah")
	}

	token, err := s.authService.GenerateToken(u.ID, u.TenantID, u.Role)
    if err != nil { return nil, err }

    return &LoginResponse{Token: token, User: *u}, nil
}

func (s *Service) GetProfile(ctx context.Context, id uuid.UUID) (*booking.Tenant, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) UpdateProfile(ctx context.Context, id uuid.UUID, req booking.Tenant) (*booking.Tenant, error) {
	curr, _ := s.repo.GetByID(ctx, id)
	req.ID = id
	req.Slug = curr.Slug // Protect slug from being changed via profile update
	if err := s.repo.Update(ctx, req); err != nil { return nil, err }
	return &req, nil
}
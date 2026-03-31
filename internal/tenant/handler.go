package tenant

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/helwiza/saas/internal/booking"
	"github.com/helwiza/saas/internal/platform/storage"
)

type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

// GetPublicLandingData melayani data untuk Landing Page publik (Tanpa Auth)
func (h *Handler) GetPublicLandingData(c *gin.Context) {
	slug := c.Query("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug is required"})
		return
	}

	// Bersihkan slug jika ada .localhost atau domain lain
	cleanSlug := strings.Split(slug, ".")[0]

	// 1. Ambil data profil tenant berdasarkan slug
	tenant, err := h.service.repo.GetBySlug(c.Request.Context(), cleanSlug)
	if err != nil || tenant.ID == uuid.Nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bisnis tidak ditemukan"})
		return
	}

	// 2. Ambil daftar resource (meja/studio) milik tenant tersebut
	// Kita bisa panggil repo langsung atau buatkan fungsi di service
	resources, _ := h.service.repo.ListResources(c.Request.Context(), tenant.ID)

	c.JSON(http.StatusOK, gin.H{
		"profile":   tenant,
		"resources": resources,
	})
}

// UploadImage menangani upload logo/banner ke S3
func (h *Handler) UploadImage(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tidak ada gambar yang diupload"})
		return
	}

	tenantID := c.MustGet("tenantID").(string)

	s3Provider, err := storage.NewS3Client()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Storage configuration error"})
		return
	}

	url, err := s3Provider.UploadFile(c.Request.Context(), file, "tenants/"+tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal upload ke S3: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Upload sukses",
		"url":     url,
	})
}

// --- AUTH & PROFILE HANDLERS ---

func (h *Handler) Register(c *gin.Context) {
	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	t, err := h.service.Register(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, t)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetProfile(c *gin.Context) {
	tIDRaw, _ := c.Get("tenantID")
	tID, _ := uuid.Parse(tIDRaw.(string))
	p, err := h.service.GetProfile(c.Request.Context(), tID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	tIDRaw, _ := c.Get("tenantID")
	tID, _ := uuid.Parse(tIDRaw.(string))
	var req booking.Tenant
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated, err := h.service.UpdateProfile(c.Request.Context(), tID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}
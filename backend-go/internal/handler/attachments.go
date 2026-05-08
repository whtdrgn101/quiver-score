package handler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/quiverscore/backend-go/internal/config"
	"github.com/quiverscore/backend-go/internal/imaging"
	"github.com/quiverscore/backend-go/internal/middleware"
	"github.com/quiverscore/backend-go/internal/repository"
	"github.com/quiverscore/backend-go/internal/storage"
)

// maxAttachmentBytes caps raw multipart uploads at 10 MB.
const maxAttachmentBytes = 10 * 1024 * 1024

// OwnerVerifier confirms an owner exists and belongs to the calling user.
// Each owner_type plugs in its own implementation via the registry on
// AttachmentsHandler.Owners.
type OwnerVerifier interface {
	OwnerBelongsToUser(ctx context.Context, ownerID, userID string) (bool, error)
}

// OwnerVerifierFunc adapts a plain function to the OwnerVerifier interface.
type OwnerVerifierFunc func(ctx context.Context, ownerID, userID string) (bool, error)

func (f OwnerVerifierFunc) OwnerBelongsToUser(ctx context.Context, ownerID, userID string) (bool, error) {
	return f(ctx, ownerID, userID)
}

// OwnerConfig is everything the handler needs to know about an owner_type:
// how to verify ownership, the per-user upload limiter for that type, and the
// per-owner attachment cap.
type OwnerConfig struct {
	Verifier    OwnerVerifier
	RateLimiter *middleware.RateLimiter // keyed by userID
	MaxPerOwner int                     // 0 = unlimited
}

// AttachmentRepository is the data-access surface the handler needs. Defined
// as an interface so unit tests can swap in a mock without a real Postgres.
type AttachmentRepository interface {
	Insert(ctx context.Context, a *repository.AttachmentRow) error
	Get(ctx context.Context, id, userID string) (*repository.AttachmentRow, error)
	ListByOwner(ctx context.Context, ownerType, ownerID, userID string) ([]repository.AttachmentRow, error)
	Delete(ctx context.Context, id, userID string) (*repository.AttachmentRow, error)
	CountByOwner(ctx context.Context, ownerType, ownerID string) (int, error)
}

type AttachmentsHandler struct {
	Repo    AttachmentRepository
	Owners  map[string]OwnerConfig
	Storage storage.ObjectStore
	Imaging *imaging.Processor
	Cfg     *config.Config
}

func (h *AttachmentsHandler) Routes(r chi.Router) {
	r.Use(middleware.RequireAuth(h.Cfg.SecretKey))
	r.Post("/", h.Upload)
	r.Get("/", h.List)
	r.Get("/{id}", h.GetFull)
	r.Get("/{id}/thumb", h.GetThumb)
	r.Delete("/{id}", h.Delete)
}

// fullKey/thumbKey define the GCS object layout. The users/{userID}/ prefix is
// load-bearing for account-deletion: a single ObjectStore.DeletePrefix wipes
// everything the user owns, regardless of owner_type.
func fullKey(userID, attachmentID string) string {
	return fmt.Sprintf("users/%s/attachments/%s/full.jpg", userID, attachmentID)
}

func thumbKey(userID, attachmentID string) string {
	return fmt.Sprintf("users/%s/attachments/%s/thumb.jpg", userID, attachmentID)
}

// ── Routes ───────────────────────────────────────────────────────────────

func (h *AttachmentsHandler) Upload(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	ownerType := r.URL.Query().Get("owner_type")
	ownerID := r.URL.Query().Get("owner_id")
	if ownerType == "" || ownerID == "" {
		Error(w, http.StatusBadRequest, "owner_type and owner_id query parameters are required")
		return
	}
	cfg, ok := h.Owners[ownerType]
	if !ok {
		Error(w, http.StatusBadRequest, "Unknown owner_type")
		return
	}
	if _, err := uuid.Parse(ownerID); err != nil {
		Error(w, http.StatusBadRequest, "owner_id must be a UUID")
		return
	}

	if cfg.RateLimiter != nil && !cfg.RateLimiter.Allow(userID) {
		w.Header().Set("Retry-After", "60")
		Error(w, http.StatusTooManyRequests, "Upload rate limit exceeded for this owner type")
		return
	}

	owns, err := cfg.Verifier.OwnerBelongsToUser(r.Context(), ownerID, userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to verify owner")
		return
	}
	if !owns {
		Error(w, http.StatusNotFound, "Owner not found")
		return
	}

	// Cap before reading the body so we reject without paying for the upload.
	count, err := h.Repo.CountByOwner(r.Context(), ownerType, ownerID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to check attachment count")
		return
	}
	if cfg.MaxPerOwner > 0 && count >= cfg.MaxPerOwner {
		Error(w, http.StatusConflict, fmt.Sprintf("Owner already has the maximum of %d attachments", cfg.MaxPerOwner))
		return
	}

	if err := r.ParseMultipartForm(maxAttachmentBytes); err != nil {
		Error(w, http.StatusBadRequest, "Request too large or invalid multipart form")
		return
	}
	file, header, err := r.FormFile("image")
	if err != nil {
		Error(w, http.StatusBadRequest, "Missing 'image' field in multipart form")
		return
	}
	defer file.Close()

	if header.Size > maxAttachmentBytes {
		Error(w, http.StatusRequestEntityTooLarge, "Image exceeds 10 MB limit")
		return
	}

	raw, err := io.ReadAll(file)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to read upload")
		return
	}

	processed, err := h.Imaging.Process(raw, header.Header.Get("Content-Type"))
	if err != nil {
		if errors.Is(err, imaging.ErrUnsupportedType) {
			Error(w, http.StatusBadRequest, err.Error())
			return
		}
		Error(w, http.StatusBadRequest, "Failed to process image")
		return
	}

	id := uuid.New().String()
	fk := fullKey(userID, id)
	tk := thumbKey(userID, id)

	// Storage first so a row pointing at missing bytes is impossible.
	if err := h.Storage.Put(r.Context(), fk, processed.ContentType, bytes.NewReader(processed.Full)); err != nil {
		Error(w, http.StatusInternalServerError, "Failed to store image")
		return
	}
	if err := h.Storage.Put(r.Context(), tk, processed.ContentType, bytes.NewReader(processed.Thumb)); err != nil {
		_ = h.Storage.Delete(r.Context(), fk)
		Error(w, http.StatusInternalServerError, "Failed to store thumbnail")
		return
	}

	row := &repository.AttachmentRow{
		ID:          id,
		UserID:      userID,
		OwnerType:   ownerType,
		OwnerID:     ownerID,
		StorageKey:  fk,
		ThumbKey:    tk,
		ContentType: processed.ContentType,
		FullSize:    len(processed.Full),
		ThumbSize:   len(processed.Thumb),
		Width:       processed.Width,
		Height:      processed.Height,
		CreatedAt:   time.Now(),
	}
	if err := h.Repo.Insert(r.Context(), row); err != nil {
		// Roll back GCS so a row insert failure doesn't leave orphan objects.
		_ = h.Storage.Delete(r.Context(), fk)
		_ = h.Storage.Delete(r.Context(), tk)
		Error(w, http.StatusInternalServerError, "Failed to save attachment")
		return
	}

	JSON(w, http.StatusCreated, row)
}

func (h *AttachmentsHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	ownerType := r.URL.Query().Get("owner_type")
	ownerID := r.URL.Query().Get("owner_id")
	if ownerType == "" || ownerID == "" {
		Error(w, http.StatusBadRequest, "owner_type and owner_id query parameters are required")
		return
	}
	cfg, ok := h.Owners[ownerType]
	if !ok {
		Error(w, http.StatusBadRequest, "Unknown owner_type")
		return
	}
	if _, err := uuid.Parse(ownerID); err != nil {
		Error(w, http.StatusBadRequest, "owner_id must be a UUID")
		return
	}

	owns, err := cfg.Verifier.OwnerBelongsToUser(r.Context(), ownerID, userID)
	if err != nil || !owns {
		// 404 for both "doesn't exist" and "not yours" — never reveal the difference.
		Error(w, http.StatusNotFound, "Owner not found")
		return
	}

	rows, err := h.Repo.ListByOwner(r.Context(), ownerType, ownerID, userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to list attachments")
		return
	}
	if rows == nil {
		rows = []repository.AttachmentRow{}
	}
	JSON(w, http.StatusOK, rows)
}

func (h *AttachmentsHandler) GetFull(w http.ResponseWriter, r *http.Request) {
	h.serveImage(w, r, false)
}

func (h *AttachmentsHandler) GetThumb(w http.ResponseWriter, r *http.Request) {
	h.serveImage(w, r, true)
}

func (h *AttachmentsHandler) serveImage(w http.ResponseWriter, r *http.Request, thumb bool) {
	userID := middleware.GetUserID(r.Context())
	id := chi.URLParam(r, "id")

	row, err := h.Repo.Get(r.Context(), id, userID)
	if err != nil {
		Error(w, http.StatusNotFound, "Attachment not found")
		return
	}

	key := row.StorageKey
	if thumb {
		key = row.ThumbKey
	}
	body, meta, err := h.Storage.Get(r.Context(), key)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			Error(w, http.StatusNotFound, "Image data missing")
			return
		}
		Error(w, http.StatusInternalServerError, "Failed to load image")
		return
	}
	defer body.Close()

	w.Header().Set("Content-Type", meta.ContentType)
	if meta.Size > 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(meta.Size, 10))
	}
	w.Header().Set("Cache-Control", "private, max-age=86400")
	w.Header().Set("ETag", `"`+row.ID+`"`)
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, body)
}

func (h *AttachmentsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	id := chi.URLParam(r, "id")

	row, err := h.Repo.Delete(r.Context(), id, userID)
	if err != nil {
		Error(w, http.StatusNotFound, "Attachment not found")
		return
	}

	// Best-effort GCS cleanup. If a delete fails, soft-delete reaps the
	// orphaned objects in 7 days, so we don't block the response on it.
	_ = h.Storage.Delete(r.Context(), row.StorageKey)
	_ = h.Storage.Delete(r.Context(), row.ThumbKey)

	w.WriteHeader(http.StatusNoContent)
}

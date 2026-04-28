package handler

import (
	"context"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/quiverscore/backend-go/internal/config"
	"github.com/quiverscore/backend-go/internal/middleware"
	"github.com/quiverscore/backend-go/internal/repository"
)

type EndImageRepository interface {
	Upload(ctx context.Context, id, endID, sessionID, userID, contentType string, fileSize int, imageData []byte) (*repository.EndImageOut, error)
	GetMeta(ctx context.Context, imageID, userID string) (*repository.EndImageOut, error)
	GetImageData(ctx context.Context, imageID, userID string) ([]byte, string, error)
	ListByEnd(ctx context.Context, endID, userID string) ([]repository.EndImageOut, error)
	ListBySession(ctx context.Context, sessionID, userID string) ([]repository.EndImageOut, error)
	Delete(ctx context.Context, imageID, userID string) error
	EndBelongsToSession(ctx context.Context, endID, sessionID string) (bool, error)
	SessionBelongsToUser(ctx context.Context, sessionID, userID string) (bool, error)
}

type EndImagesHandler struct {
	Images EndImageRepository
	Cfg    *config.Config
}

var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/heic": true,
}

const maxImageBytes = 10 * 1024 * 1024 // 10 MB

func (h *EndImagesHandler) Routes(r chi.Router) {
	r.Use(middleware.RequireAuth(h.Cfg.SecretKey))

	// Session-scoped routes
	r.Get("/{sessionId}/images", h.ListBySession)
	r.Post("/{sessionId}/ends/{endId}/images", h.Upload)
	r.Get("/{sessionId}/ends/{endId}/images", h.ListByEnd)
	r.Get("/{sessionId}/images/{imageId}", h.GetImage)
	r.Delete("/{sessionId}/images/{imageId}", h.Delete)
}

func (h *EndImagesHandler) Upload(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	endID := chi.URLParam(r, "endId")
	userID := middleware.GetUserID(r.Context())

	// Verify session belongs to user
	owns, err := h.Images.SessionBelongsToUser(r.Context(), sessionID, userID)
	if err != nil || !owns {
		Error(w, http.StatusNotFound, "Session not found")
		return
	}

	// Verify end belongs to session
	belongs, err := h.Images.EndBelongsToSession(r.Context(), endID, sessionID)
	if err != nil || !belongs {
		Error(w, http.StatusNotFound, "End not found in this session")
		return
	}

	// Parse multipart form (limit to maxImageBytes)
	if err := r.ParseMultipartForm(maxImageBytes); err != nil {
		Error(w, http.StatusBadRequest, "Request too large or invalid multipart form")
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		Error(w, http.StatusBadRequest, "Missing 'image' field in multipart form")
		return
	}
	defer file.Close()

	// Validate content type
	contentType := header.Header.Get("Content-Type")
	if !allowedImageTypes[contentType] {
		Error(w, http.StatusBadRequest, "Unsupported image type. Allowed: jpeg, png, webp, heic")
		return
	}

	// Validate file size
	if header.Size > maxImageBytes {
		Error(w, http.StatusBadRequest, "Image exceeds 10 MB limit")
		return
	}

	// Read image data
	imageData, err := io.ReadAll(file)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to read image data")
		return
	}

	id := uuid.New().String()
	out, err := h.Images.Upload(r.Context(), id, endID, sessionID, userID, contentType, len(imageData), imageData)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to save image")
		return
	}

	JSON(w, http.StatusCreated, out)
}

func (h *EndImagesHandler) GetImage(w http.ResponseWriter, r *http.Request) {
	imageID := chi.URLParam(r, "imageId")
	sessionID := chi.URLParam(r, "sessionId")
	userID := middleware.GetUserID(r.Context())

	// Verify session belongs to user
	owns, err := h.Images.SessionBelongsToUser(r.Context(), sessionID, userID)
	if err != nil || !owns {
		Error(w, http.StatusNotFound, "Session not found")
		return
	}

	data, contentType, err := h.Images.GetImageData(r.Context(), imageID, userID)
	if err != nil {
		Error(w, http.StatusNotFound, "Image not found")
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "private, max-age=86400")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (h *EndImagesHandler) ListByEnd(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	endID := chi.URLParam(r, "endId")
	userID := middleware.GetUserID(r.Context())

	owns, err := h.Images.SessionBelongsToUser(r.Context(), sessionID, userID)
	if err != nil || !owns {
		Error(w, http.StatusNotFound, "Session not found")
		return
	}

	images, err := h.Images.ListByEnd(r.Context(), endID, userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to list images")
		return
	}

	if images == nil {
		images = []repository.EndImageOut{}
	}
	JSON(w, http.StatusOK, images)
}

func (h *EndImagesHandler) ListBySession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	userID := middleware.GetUserID(r.Context())

	owns, err := h.Images.SessionBelongsToUser(r.Context(), sessionID, userID)
	if err != nil || !owns {
		Error(w, http.StatusNotFound, "Session not found")
		return
	}

	images, err := h.Images.ListBySession(r.Context(), sessionID, userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Failed to list images")
		return
	}

	if images == nil {
		images = []repository.EndImageOut{}
	}
	JSON(w, http.StatusOK, images)
}

func (h *EndImagesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	imageID := chi.URLParam(r, "imageId")
	sessionID := chi.URLParam(r, "sessionId")
	userID := middleware.GetUserID(r.Context())

	owns, err := h.Images.SessionBelongsToUser(r.Context(), sessionID, userID)
	if err != nil || !owns {
		Error(w, http.StatusNotFound, "Session not found")
		return
	}

	if err := h.Images.Delete(r.Context(), imageID, userID); err != nil {
		Error(w, http.StatusNotFound, "Image not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

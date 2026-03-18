package handler

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/quiverscore/backend-go/internal/config"
	"github.com/quiverscore/backend-go/internal/middleware"
	"github.com/quiverscore/backend-go/internal/repository"
)

type UsersHandler struct {
	Users *repository.UserRepo
	Cfg   *config.Config
}

var allowedAvatarTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

const maxAvatarBytes = 2 * 1024 * 1024 // 2 MB

func (h *UsersHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	u, err := h.Users.GetMe(r.Context(), userID)
	if err != nil {
		Error(w, http.StatusUnauthorized, "User not found")
		return
	}

	JSON(w, http.StatusOK, u)
}

// ── Profile Update ───────────────────────────────────────────────────

type profileUpdate struct {
	DisplayName    *string `json:"display_name"`
	BowType        *string `json:"bow_type"`
	Classification *string `json:"classification"`
	Bio            *string `json:"bio"`
	ProfilePublic  *bool   `json:"profile_public"`
}

func (h *UsersHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	// Parse raw JSON to detect which fields were explicitly set
	body, err := io.ReadAll(r.Body)
	if err != nil {
		ValidationError(w, "Invalid request body")
		return
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		ValidationError(w, "Invalid request body")
		return
	}

	var req profileUpdate
	if err := json.Unmarshal(body, &req); err != nil {
		ValidationError(w, "Invalid request body")
		return
	}

	_, displayNameSet := raw["display_name"]
	_, bowTypeSet := raw["bow_type"]
	_, classificationSet := raw["classification"]
	_, bioSet := raw["bio"]

	userID := middleware.GetUserID(r.Context())

	u, err := h.Users.UpdateProfile(r.Context(), userID,
		req.DisplayName, req.BowType, req.Classification, req.Bio,
		displayNameSet, bowTypeSet, classificationSet, bioSet,
		req.ProfilePublic,
	)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusOK, u)
}

// ── Avatar Upload ────────────────────────────────────────────────────

func (h *UsersHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxAvatarBytes + 1024); err != nil {
		Error(w, http.StatusBadRequest, "File must be under 2 MB")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		Error(w, http.StatusBadRequest, "File is required")
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if !allowedAvatarTypes[contentType] {
		Error(w, http.StatusBadRequest, "File must be JPEG, PNG, or WebP")
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if len(data) > maxAvatarBytes {
		Error(w, http.StatusBadRequest, "File must be under 2 MB")
		return
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	dataURI := "data:" + contentType + ";base64," + encoded

	userID := middleware.GetUserID(r.Context())
	u, err := h.Users.UpdateAvatar(r.Context(), userID, dataURI)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusOK, u)
}

// ── Avatar URL Upload ────────────────────────────────────────────────

type avatarURLRequest struct {
	URL string `json:"url"`
}

func (h *UsersHandler) UploadAvatarFromURL(w http.ResponseWriter, r *http.Request) {
	var req avatarURLRequest
	if !Decode(w, r, &req) {
		return
	}
	if req.URL == "" {
		ValidationError(w, "url is required")
		return
	}

	resp, err := http.Get(req.URL)
	if err != nil {
		Error(w, http.StatusBadRequest, "Could not fetch image from URL")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		Error(w, http.StatusBadRequest, "Could not fetch image from URL")
		return
	}

	contentType := resp.Header.Get("Content-Type")
	// Strip params like charset
	if idx := len(contentType); idx > 0 {
		for i, c := range contentType {
			if c == ';' {
				contentType = contentType[:i]
				break
			}
		}
	}
	if !allowedAvatarTypes[contentType] {
		Error(w, http.StatusBadRequest, "URL must point to a JPEG, PNG, or WebP image")
		return
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		Error(w, http.StatusBadRequest, "Could not fetch image from URL")
		return
	}
	if len(data) > maxAvatarBytes {
		Error(w, http.StatusBadRequest, "Image must be under 2 MB")
		return
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	dataURI := "data:" + contentType + ";base64," + encoded

	userID := middleware.GetUserID(r.Context())
	u, err := h.Users.UpdateAvatar(r.Context(), userID, dataURI)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusOK, u)
}

// ── Avatar Delete ────────────────────────────────────────────────────

func (h *UsersHandler) DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	u, err := h.Users.DeleteAvatar(r.Context(), userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	JSON(w, http.StatusOK, u)
}

// ── Public Profile ───────────────────────────────────────────────────

func (h *UsersHandler) GetPublicProfile(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	profile, err := h.Users.GetPublicProfile(r.Context(), username)
	if err != nil {
		Error(w, http.StatusNotFound, "Profile not found")
		return
	}

	JSON(w, http.StatusOK, profile)
}

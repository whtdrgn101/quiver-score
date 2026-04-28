package handler

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/quiverscore/backend-go/internal/config"
	"github.com/quiverscore/backend-go/internal/middleware"
	"github.com/quiverscore/backend-go/internal/repository"
)

// ── Mock ──────────────────────────────────────────────────────────────

type mockEndImageRepo struct {
	uploadResult    *repository.EndImageOut
	uploadErr       error
	getMetaResult   *repository.EndImageOut
	getMetaErr      error
	getDataResult   []byte
	getDataType     string
	getDataErr      error
	listByEndResult []repository.EndImageOut
	listByEndErr    error
	listBySessionResult []repository.EndImageOut
	listBySessionErr    error
	deleteErr       error

	endBelongsToSession    bool
	endBelongsToSessionErr error
	sessionBelongsToUser    bool
	sessionBelongsToUserErr error
}

func (m *mockEndImageRepo) Upload(_ context.Context, _, _, _, _, _ string, _ int, _ []byte) (*repository.EndImageOut, error) {
	return m.uploadResult, m.uploadErr
}

func (m *mockEndImageRepo) GetMeta(_ context.Context, _, _ string) (*repository.EndImageOut, error) {
	return m.getMetaResult, m.getMetaErr
}

func (m *mockEndImageRepo) GetImageData(_ context.Context, _, _ string) ([]byte, string, error) {
	return m.getDataResult, m.getDataType, m.getDataErr
}

func (m *mockEndImageRepo) ListByEnd(_ context.Context, _, _ string) ([]repository.EndImageOut, error) {
	return m.listByEndResult, m.listByEndErr
}

func (m *mockEndImageRepo) ListBySession(_ context.Context, _, _ string) ([]repository.EndImageOut, error) {
	return m.listBySessionResult, m.listBySessionErr
}

func (m *mockEndImageRepo) Delete(_ context.Context, _, _ string) error {
	return m.deleteErr
}

func (m *mockEndImageRepo) EndBelongsToSession(_ context.Context, _, _ string) (bool, error) {
	return m.endBelongsToSession, m.endBelongsToSessionErr
}

func (m *mockEndImageRepo) SessionBelongsToUser(_ context.Context, _, _ string) (bool, error) {
	return m.sessionBelongsToUser, m.sessionBelongsToUserErr
}

// ── Helpers ───────────────────────────────────────────────────────────

func endImagesHandler(mock *mockEndImageRepo) *EndImagesHandler {
	return &EndImagesHandler{
		Images: mock,
		Cfg:    &config.Config{SecretKey: "test-secret"},
	}
}

func endImageAuthedReq(method, path string, userID string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	return req.WithContext(ctx)
}

func withURLParams(r *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func createMultipartRequest(t *testing.T, userID string, imageData []byte, contentType, fieldName string) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create a form file part with explicit content type
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="`+fieldName+`"; filename="target.jpg"`)
	h.Set("Content-Type", contentType)
	part, err := writer.CreatePart(h)
	if err != nil {
		t.Fatal(err)
	}
	part.Write(imageData)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	return req.WithContext(ctx)
}

func sampleImageOut() *repository.EndImageOut {
	return &repository.EndImageOut{
		ID:          "img-1",
		EndID:       "end-1",
		SessionID:   "session-1",
		UserID:      "user-1",
		ContentType: "image/jpeg",
		FileSize:    1024,
		CreatedAt:   time.Now().UTC(),
	}
}

// ── Upload Tests ──────────────────────────────────────────────────────

func TestUpload_Success(t *testing.T) {
	out := sampleImageOut()
	mock := &mockEndImageRepo{
		sessionBelongsToUser: true,
		endBelongsToSession:  true,
		uploadResult:         out,
	}
	h := endImagesHandler(mock)

	// Create multipart request with a JPEG
	imageData := []byte{0xFF, 0xD8, 0xFF, 0xE0} // JPEG magic bytes
	req := createMultipartRequest(t, "user-1", imageData, "image/jpeg", "image")
	req = withURLParams(req, map[string]string{
		"sessionId": "session-1",
		"endId":     "end-1",
	})

	// Set the correct content type on the form file part
	rr := httptest.NewRecorder()
	h.Upload(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestUpload_SessionNotOwned(t *testing.T) {
	mock := &mockEndImageRepo{
		sessionBelongsToUser: false,
	}
	h := endImagesHandler(mock)

	imageData := []byte{0xFF, 0xD8, 0xFF, 0xE0}
	req := createMultipartRequest(t, "user-1", imageData, "image/jpeg", "image")
	req = withURLParams(req, map[string]string{
		"sessionId": "session-1",
		"endId":     "end-1",
	})

	rr := httptest.NewRecorder()
	h.Upload(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestUpload_EndNotInSession(t *testing.T) {
	mock := &mockEndImageRepo{
		sessionBelongsToUser: true,
		endBelongsToSession:  false,
	}
	h := endImagesHandler(mock)

	imageData := []byte{0xFF, 0xD8, 0xFF, 0xE0}
	req := createMultipartRequest(t, "user-1", imageData, "image/jpeg", "image")
	req = withURLParams(req, map[string]string{
		"sessionId": "session-1",
		"endId":     "end-1",
	})

	rr := httptest.NewRecorder()
	h.Upload(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestUpload_MissingImageField(t *testing.T) {
	mock := &mockEndImageRepo{
		sessionBelongsToUser: true,
		endBelongsToSession:  true,
	}
	h := endImagesHandler(mock)

	// Create multipart with wrong field name
	imageData := []byte{0xFF, 0xD8, 0xFF, 0xE0}
	req := createMultipartRequest(t, "user-1", imageData, "image/jpeg", "wrong_field")
	req = withURLParams(req, map[string]string{
		"sessionId": "session-1",
		"endId":     "end-1",
	})

	rr := httptest.NewRecorder()
	h.Upload(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

// ── GetImage Tests ────────────────────────────────────────────────────

func TestGetImage_Success(t *testing.T) {
	imageData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10}
	mock := &mockEndImageRepo{
		sessionBelongsToUser: true,
		getDataResult:        imageData,
		getDataType:          "image/jpeg",
	}
	h := endImagesHandler(mock)

	req := endImageAuthedReq(http.MethodGet, "/session-1/images/img-1", "user-1")
	req = withURLParams(req, map[string]string{
		"sessionId": "session-1",
		"imageId":   "img-1",
	})

	rr := httptest.NewRecorder()
	h.GetImage(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); ct != "image/jpeg" {
		t.Errorf("expected Content-Type image/jpeg, got %s", ct)
	}
	if rr.Body.Len() != len(imageData) {
		t.Errorf("expected body length %d, got %d", len(imageData), rr.Body.Len())
	}
}

func TestGetImage_NotFound(t *testing.T) {
	mock := &mockEndImageRepo{
		sessionBelongsToUser: true,
		getDataErr:           errNotFound,
	}
	h := endImagesHandler(mock)

	req := endImageAuthedReq(http.MethodGet, "/session-1/images/img-999", "user-1")
	req = withURLParams(req, map[string]string{
		"sessionId": "session-1",
		"imageId":   "img-999",
	})

	rr := httptest.NewRecorder()
	h.GetImage(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestGetImage_SessionNotOwned(t *testing.T) {
	mock := &mockEndImageRepo{
		sessionBelongsToUser: false,
	}
	h := endImagesHandler(mock)

	req := endImageAuthedReq(http.MethodGet, "/session-1/images/img-1", "user-1")
	req = withURLParams(req, map[string]string{
		"sessionId": "session-1",
		"imageId":   "img-1",
	})

	rr := httptest.NewRecorder()
	h.GetImage(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", rr.Code, rr.Body.String())
	}
}

// ── ListByEnd Tests ───────────────────────────────────────────────────

func TestListByEnd_Success(t *testing.T) {
	mock := &mockEndImageRepo{
		sessionBelongsToUser: true,
		listByEndResult: []repository.EndImageOut{
			*sampleImageOut(),
		},
	}
	h := endImagesHandler(mock)

	req := endImageAuthedReq(http.MethodGet, "/session-1/ends/end-1/images", "user-1")
	req = withURLParams(req, map[string]string{
		"sessionId": "session-1",
		"endId":     "end-1",
	})

	rr := httptest.NewRecorder()
	h.ListByEnd(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestListByEnd_EmptyReturnsEmptyArray(t *testing.T) {
	mock := &mockEndImageRepo{
		sessionBelongsToUser: true,
		listByEndResult:      nil,
	}
	h := endImagesHandler(mock)

	req := endImageAuthedReq(http.MethodGet, "/session-1/ends/end-1/images", "user-1")
	req = withURLParams(req, map[string]string{
		"sessionId": "session-1",
		"endId":     "end-1",
	})

	rr := httptest.NewRecorder()
	h.ListByEnd(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if rr.Body.String() != "[]\n" {
		t.Errorf("expected empty array, got %s", rr.Body.String())
	}
}

// ── ListBySession Tests ───────────────────────────────────────────────

func TestListBySession_Success(t *testing.T) {
	mock := &mockEndImageRepo{
		sessionBelongsToUser: true,
		listBySessionResult: []repository.EndImageOut{
			*sampleImageOut(),
		},
	}
	h := endImagesHandler(mock)

	req := endImageAuthedReq(http.MethodGet, "/session-1/images", "user-1")
	req = withURLParams(req, map[string]string{
		"sessionId": "session-1",
	})

	rr := httptest.NewRecorder()
	h.ListBySession(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestListBySession_NotOwned(t *testing.T) {
	mock := &mockEndImageRepo{
		sessionBelongsToUser: false,
	}
	h := endImagesHandler(mock)

	req := endImageAuthedReq(http.MethodGet, "/session-1/images", "user-1")
	req = withURLParams(req, map[string]string{
		"sessionId": "session-1",
	})

	rr := httptest.NewRecorder()
	h.ListBySession(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", rr.Code, rr.Body.String())
	}
}

// ── Delete Tests ──────────────────────────────────────────────────────

func TestDelete_Success(t *testing.T) {
	mock := &mockEndImageRepo{
		sessionBelongsToUser: true,
	}
	h := endImagesHandler(mock)

	req := endImageAuthedReq(http.MethodDelete, "/session-1/images/img-1", "user-1")
	req = withURLParams(req, map[string]string{
		"sessionId": "session-1",
		"imageId":   "img-1",
	})

	rr := httptest.NewRecorder()
	h.Delete(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestDelete_NotFound(t *testing.T) {
	mock := &mockEndImageRepo{
		sessionBelongsToUser: true,
		deleteErr:            repository.ErrNotFound,
	}
	h := endImagesHandler(mock)

	req := endImageAuthedReq(http.MethodDelete, "/session-1/images/img-999", "user-1")
	req = withURLParams(req, map[string]string{
		"sessionId": "session-1",
		"imageId":   "img-999",
	})

	rr := httptest.NewRecorder()
	h.Delete(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestDelete_SessionNotOwned(t *testing.T) {
	mock := &mockEndImageRepo{
		sessionBelongsToUser: false,
	}
	h := endImagesHandler(mock)

	req := endImageAuthedReq(http.MethodDelete, "/session-1/images/img-1", "user-1")
	req = withURLParams(req, map[string]string{
		"sessionId": "session-1",
		"imageId":   "img-1",
	})

	rr := httptest.NewRecorder()
	h.Delete(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", rr.Code, rr.Body.String())
	}
}

package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/quiverscore/backend-go/internal/config"
	"github.com/quiverscore/backend-go/internal/imaging"
	"github.com/quiverscore/backend-go/internal/middleware"
	"github.com/quiverscore/backend-go/internal/repository"
	"github.com/quiverscore/backend-go/internal/storage"
)

// ── Fixtures ───────────────────────────────────────────────────────────

const (
	testUserID    = "11111111-1111-1111-1111-111111111111"
	testOwnerType = "session_end"
	testOwnerID   = "22222222-2222-2222-2222-222222222222"
)

// realJPEG produces a small but decodable JPEG so the imaging processor has
// real bytes to work with. Tests that only need the upload to fail before
// processing (validation, rate limit) can still pass arbitrary bytes.
func realJPEG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 100, 80))
	for y := 0; y < 80; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 2), G: uint8(y * 3), B: 128, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatalf("encode jpeg: %v", err)
	}
	return buf.Bytes()
}

// ── Mocks ──────────────────────────────────────────────────────────────

type mockAttachmentRepo struct {
	mu       sync.Mutex
	rows     map[string]*repository.AttachmentRow // by id
	insertEr error
	countErr error
	getErr   error
	listErr  error
	delErr   error
}

func newMockRepo() *mockAttachmentRepo {
	return &mockAttachmentRepo{rows: map[string]*repository.AttachmentRow{}}
}

func (m *mockAttachmentRepo) Insert(_ context.Context, a *repository.AttachmentRow) error {
	if m.insertEr != nil {
		return m.insertEr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	clone := *a
	m.rows[a.ID] = &clone
	return nil
}

func (m *mockAttachmentRepo) Get(_ context.Context, id, userID string) (*repository.AttachmentRow, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	row, ok := m.rows[id]
	if !ok || row.UserID != userID {
		return nil, repository.ErrNotFound
	}
	clone := *row
	return &clone, nil
}

func (m *mockAttachmentRepo) ListByOwner(_ context.Context, ownerType, ownerID, userID string) ([]repository.AttachmentRow, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []repository.AttachmentRow
	for _, row := range m.rows {
		if row.OwnerType == ownerType && row.OwnerID == ownerID && row.UserID == userID {
			out = append(out, *row)
		}
	}
	return out, nil
}

func (m *mockAttachmentRepo) Delete(_ context.Context, id, userID string) (*repository.AttachmentRow, error) {
	if m.delErr != nil {
		return nil, m.delErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	row, ok := m.rows[id]
	if !ok || row.UserID != userID {
		return nil, repository.ErrNotFound
	}
	clone := *row
	delete(m.rows, id)
	return &clone, nil
}

func (m *mockAttachmentRepo) CountByOwner(_ context.Context, ownerType, ownerID string) (int, error) {
	if m.countErr != nil {
		return 0, m.countErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	n := 0
	for _, row := range m.rows {
		if row.OwnerType == ownerType && row.OwnerID == ownerID {
			n++
		}
	}
	return n, nil
}

type stubVerifier struct {
	owns bool
	err  error
}

func (s *stubVerifier) OwnerBelongsToUser(_ context.Context, _, _ string) (bool, error) {
	return s.owns, s.err
}

// ── Handler builder ───────────────────────────────────────────────────

type harness struct {
	h        *AttachmentsHandler
	repo     *mockAttachmentRepo
	store    storage.ObjectStore
	verifier *stubVerifier
	limiter  *middleware.RateLimiter
}

func newHarness() *harness {
	repo := newMockRepo()
	store := storage.NewMemory()
	verifier := &stubVerifier{owns: true}
	// Generous limiter so tests don't hit it unless they explicitly drain it.
	limiter := middleware.NewRateLimiterPerHour(1000, 1000)

	h := &AttachmentsHandler{
		Repo:    repo,
		Storage: store,
		Imaging: imaging.NewProcessor(),
		Cfg:     &config.Config{SecretKey: "test-secret"},
		Owners: map[string]OwnerConfig{
			testOwnerType: {
				Verifier:    verifier,
				RateLimiter: limiter,
				MaxPerOwner: 3,
			},
		},
	}
	return &harness{h: h, repo: repo, store: store, verifier: verifier, limiter: limiter}
}

func authedReq(method, target string) *http.Request {
	req := httptest.NewRequest(method, target, nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, testUserID)
	return req.WithContext(ctx)
}

func withChiParam(r *http.Request, key, val string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, val)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func multipartUpload(t *testing.T, target string, body []byte, contentType string) *http.Request {
	t.Helper()
	buf := &bytes.Buffer{}
	mw := multipart.NewWriter(buf)
	hdr := make(textproto.MIMEHeader)
	hdr.Set("Content-Disposition", `form-data; name="image"; filename="upload.dat"`)
	hdr.Set("Content-Type", contentType)
	part, err := mw.CreatePart(hdr)
	if err != nil {
		t.Fatal(err)
	}
	part.Write(body)
	mw.Close()

	req := httptest.NewRequest(http.MethodPost, target, buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, testUserID)
	return req.WithContext(ctx)
}

// ── Upload tests ───────────────────────────────────────────────────────

func TestAttachmentUpload_Success(t *testing.T) {
	hh := newHarness()
	req := multipartUpload(t, "/?owner_type="+testOwnerType+"&owner_id="+testOwnerID, realJPEG(t), "image/jpeg")

	rr := httptest.NewRecorder()
	hh.h.Upload(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rr.Code, rr.Body.String())
	}

	var resp repository.AttachmentRow
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.OwnerType != testOwnerType || resp.OwnerID != testOwnerID || resp.UserID != testUserID {
		t.Errorf("response fields wrong: %+v", resp)
	}
	if resp.FullSize == 0 || resp.ThumbSize == 0 {
		t.Error("expected non-zero rendition sizes")
	}
	if resp.Width != 100 || resp.Height != 80 {
		t.Errorf("dimensions = %dx%d, want 100x80", resp.Width, resp.Height)
	}

	// Storage should now hold both renditions under the user-prefixed layout.
	if _, _, err := hh.store.Get(req.Context(), fullKey(testUserID, resp.ID)); err != nil {
		t.Errorf("full not in storage: %v", err)
	}
	if _, _, err := hh.store.Get(req.Context(), thumbKey(testUserID, resp.ID)); err != nil {
		t.Errorf("thumb not in storage: %v", err)
	}
}

func TestAttachmentUpload_MissingOwnerParams(t *testing.T) {
	hh := newHarness()
	req := multipartUpload(t, "/", realJPEG(t), "image/jpeg")
	rr := httptest.NewRecorder()
	hh.h.Upload(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d", rr.Code)
	}
}

func TestAttachmentUpload_UnknownOwnerType(t *testing.T) {
	hh := newHarness()
	req := multipartUpload(t, "/?owner_type=unknown&owner_id="+testOwnerID, realJPEG(t), "image/jpeg")
	rr := httptest.NewRecorder()
	hh.h.Upload(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d", rr.Code)
	}
}

func TestAttachmentUpload_InvalidOwnerIDUUID(t *testing.T) {
	hh := newHarness()
	req := multipartUpload(t, "/?owner_type="+testOwnerType+"&owner_id=not-a-uuid", realJPEG(t), "image/jpeg")
	rr := httptest.NewRecorder()
	hh.h.Upload(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d", rr.Code)
	}
}

func TestAttachmentUpload_OwnerNotOwned(t *testing.T) {
	hh := newHarness()
	hh.verifier.owns = false
	req := multipartUpload(t, "/?owner_type="+testOwnerType+"&owner_id="+testOwnerID, realJPEG(t), "image/jpeg")
	rr := httptest.NewRecorder()
	hh.h.Upload(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d", rr.Code)
	}
}

func TestAttachmentUpload_HEICRejected(t *testing.T) {
	hh := newHarness()
	req := multipartUpload(t, "/?owner_type="+testOwnerType+"&owner_id="+testOwnerID, []byte("not heic"), "image/heic")
	rr := httptest.NewRecorder()
	hh.h.Upload(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "HEIC") {
		t.Errorf("expected HEIC mention in body: %s", rr.Body.String())
	}
}

func TestAttachmentUpload_CorruptImage(t *testing.T) {
	hh := newHarness()
	req := multipartUpload(t, "/?owner_type="+testOwnerType+"&owner_id="+testOwnerID, []byte("not actually a jpeg"), "image/jpeg")
	rr := httptest.NewRecorder()
	hh.h.Upload(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d", rr.Code)
	}
}

func TestAttachmentUpload_PerOwnerCapReached(t *testing.T) {
	hh := newHarness()
	// Pre-populate the owner with MaxPerOwner attachments so the next upload trips the cap.
	for i := 0; i < 3; i++ {
		_ = hh.repo.Insert(context.Background(), &repository.AttachmentRow{
			ID: "existing-" + string(rune('a'+i)), UserID: testUserID,
			OwnerType: testOwnerType, OwnerID: testOwnerID,
		})
	}
	req := multipartUpload(t, "/?owner_type="+testOwnerType+"&owner_id="+testOwnerID, realJPEG(t), "image/jpeg")
	rr := httptest.NewRecorder()
	hh.h.Upload(rr, req)
	if rr.Code != http.StatusConflict {
		t.Errorf("status = %d", rr.Code)
	}
}

func TestAttachmentUpload_RateLimited(t *testing.T) {
	hh := newHarness()
	// Replace the generous limiter with one that allows a single request.
	hh.h.Owners[testOwnerType] = OwnerConfig{
		Verifier:    hh.verifier,
		RateLimiter: middleware.NewRateLimiterPerHour(1, 1),
		MaxPerOwner: 10,
	}
	// Burn the budget.
	req1 := multipartUpload(t, "/?owner_type="+testOwnerType+"&owner_id="+testOwnerID, realJPEG(t), "image/jpeg")
	hh.h.Upload(httptest.NewRecorder(), req1)

	// Next request is denied.
	req2 := multipartUpload(t, "/?owner_type="+testOwnerType+"&owner_id="+testOwnerID, realJPEG(t), "image/jpeg")
	rr := httptest.NewRecorder()
	hh.h.Upload(rr, req2)
	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, body = %s", rr.Code, rr.Body.String())
	}
	if rr.Header().Get("Retry-After") == "" {
		t.Error("missing Retry-After header on 429")
	}
}

func TestAttachmentUpload_VerifierError(t *testing.T) {
	hh := newHarness()
	hh.verifier.err = errors.New("db down")
	req := multipartUpload(t, "/?owner_type="+testOwnerType+"&owner_id="+testOwnerID, realJPEG(t), "image/jpeg")
	rr := httptest.NewRecorder()
	hh.h.Upload(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d", rr.Code)
	}
}

// ── List tests ─────────────────────────────────────────────────────────

func TestAttachmentList_Success(t *testing.T) {
	hh := newHarness()
	_ = hh.repo.Insert(context.Background(), &repository.AttachmentRow{
		ID: "att-1", UserID: testUserID, OwnerType: testOwnerType, OwnerID: testOwnerID,
		Width: 100, Height: 80, ContentType: "image/jpeg",
	})
	req := authedReq(http.MethodGet, "/?owner_type="+testOwnerType+"&owner_id="+testOwnerID)
	rr := httptest.NewRecorder()
	hh.h.List(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	var out []repository.AttachmentRow
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(out) != 1 || out[0].ID != "att-1" {
		t.Errorf("got %v", out)
	}
}

func TestAttachmentList_OwnerNotOwned(t *testing.T) {
	hh := newHarness()
	hh.verifier.owns = false
	req := authedReq(http.MethodGet, "/?owner_type="+testOwnerType+"&owner_id="+testOwnerID)
	rr := httptest.NewRecorder()
	hh.h.List(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d", rr.Code)
	}
}

func TestAttachmentList_EmptyReturnsArrayNotNull(t *testing.T) {
	hh := newHarness()
	req := authedReq(http.MethodGet, "/?owner_type="+testOwnerType+"&owner_id="+testOwnerID)
	rr := httptest.NewRecorder()
	hh.h.List(rr, req)
	if got := strings.TrimSpace(rr.Body.String()); got != "[]" {
		t.Errorf("body = %q, want []", got)
	}
}

// ── GetFull / GetThumb tests ───────────────────────────────────────────

func TestAttachmentGetFull_Success(t *testing.T) {
	hh := newHarness()
	id := "attach-xyz"
	full := []byte("FULL-BYTES")
	thumb := []byte("THUMB-BYTES")
	ctx := context.Background()
	_ = hh.store.Put(ctx, fullKey(testUserID, id), "image/jpeg", bytes.NewReader(full))
	_ = hh.store.Put(ctx, thumbKey(testUserID, id), "image/jpeg", bytes.NewReader(thumb))
	_ = hh.repo.Insert(ctx, &repository.AttachmentRow{
		ID: id, UserID: testUserID, OwnerType: testOwnerType, OwnerID: testOwnerID,
		StorageKey: fullKey(testUserID, id), ThumbKey: thumbKey(testUserID, id),
		ContentType: "image/jpeg", FullSize: len(full), ThumbSize: len(thumb),
		CreatedAt: time.Now(),
	})

	req := withChiParam(authedReq(http.MethodGet, "/"+id), "id", id)
	rr := httptest.NewRecorder()
	hh.h.GetFull(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	if got := rr.Body.String(); got != string(full) {
		t.Errorf("body = %q", got)
	}
	if rr.Header().Get("Content-Type") != "image/jpeg" {
		t.Errorf("content-type = %q", rr.Header().Get("Content-Type"))
	}
	if rr.Header().Get("ETag") != `"`+id+`"` {
		t.Errorf("etag = %q", rr.Header().Get("ETag"))
	}
	if !strings.Contains(rr.Header().Get("Cache-Control"), "private") {
		t.Errorf("cache-control = %q", rr.Header().Get("Cache-Control"))
	}
}

func TestAttachmentGetThumb_ServesThumbBytes(t *testing.T) {
	hh := newHarness()
	id := "attach-thumb"
	thumb := []byte("THUMB-BYTES")
	ctx := context.Background()
	_ = hh.store.Put(ctx, thumbKey(testUserID, id), "image/jpeg", bytes.NewReader(thumb))
	_ = hh.repo.Insert(ctx, &repository.AttachmentRow{
		ID: id, UserID: testUserID, OwnerType: testOwnerType, OwnerID: testOwnerID,
		StorageKey: fullKey(testUserID, id), ThumbKey: thumbKey(testUserID, id),
	})
	req := withChiParam(authedReq(http.MethodGet, "/"+id+"/thumb"), "id", id)
	rr := httptest.NewRecorder()
	hh.h.GetThumb(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	if rr.Body.String() != string(thumb) {
		t.Errorf("served full bytes instead of thumb: %q", rr.Body.String())
	}
}

func TestAttachmentGetFull_NotFoundInDB(t *testing.T) {
	hh := newHarness()
	req := withChiParam(authedReq(http.MethodGet, "/missing"), "id", "missing")
	rr := httptest.NewRecorder()
	hh.h.GetFull(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d", rr.Code)
	}
}

func TestAttachmentGetFull_OtherUsersAttachment(t *testing.T) {
	hh := newHarness()
	id := "attach-other"
	_ = hh.repo.Insert(context.Background(), &repository.AttachmentRow{
		ID: id, UserID: "other-user-id", OwnerType: testOwnerType, OwnerID: testOwnerID,
	})
	req := withChiParam(authedReq(http.MethodGet, "/"+id), "id", id)
	rr := httptest.NewRecorder()
	hh.h.GetFull(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404 to hide other-user attachments, got %d", rr.Code)
	}
}

func TestAttachmentGetFull_StorageMissing(t *testing.T) {
	hh := newHarness()
	id := "attach-orphan"
	_ = hh.repo.Insert(context.Background(), &repository.AttachmentRow{
		ID: id, UserID: testUserID, OwnerType: testOwnerType, OwnerID: testOwnerID,
		StorageKey: fullKey(testUserID, id), ThumbKey: thumbKey(testUserID, id),
	})
	// Storage intentionally not populated.
	req := withChiParam(authedReq(http.MethodGet, "/"+id), "id", id)
	rr := httptest.NewRecorder()
	hh.h.GetFull(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d", rr.Code)
	}
}

// ── Delete tests ───────────────────────────────────────────────────────

func TestAttachmentDelete_Success(t *testing.T) {
	hh := newHarness()
	id := "to-delete"
	ctx := context.Background()
	_ = hh.store.Put(ctx, fullKey(testUserID, id), "image/jpeg", strings.NewReader("F"))
	_ = hh.store.Put(ctx, thumbKey(testUserID, id), "image/jpeg", strings.NewReader("T"))
	_ = hh.repo.Insert(ctx, &repository.AttachmentRow{
		ID: id, UserID: testUserID, OwnerType: testOwnerType, OwnerID: testOwnerID,
		StorageKey: fullKey(testUserID, id), ThumbKey: thumbKey(testUserID, id),
	})

	req := withChiParam(authedReq(http.MethodDelete, "/"+id), "id", id)
	rr := httptest.NewRecorder()
	hh.h.Delete(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rr.Code)
	}
	if _, _, err := hh.store.Get(ctx, fullKey(testUserID, id)); err == nil {
		t.Error("full still in storage after delete")
	}
	if _, _, err := hh.store.Get(ctx, thumbKey(testUserID, id)); err == nil {
		t.Error("thumb still in storage after delete")
	}
}

func TestAttachmentDelete_NotFound(t *testing.T) {
	hh := newHarness()
	req := withChiParam(authedReq(http.MethodDelete, "/missing"), "id", "missing")
	rr := httptest.NewRecorder()
	hh.h.Delete(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d", rr.Code)
	}
}

// Sanity check that the response from Get/Thumb can be drained without error.
func TestAttachmentServeImage_BodyDrains(t *testing.T) {
	hh := newHarness()
	id := "drain"
	ctx := context.Background()
	_ = hh.store.Put(ctx, fullKey(testUserID, id), "image/jpeg", bytes.NewReader([]byte("payload")))
	_ = hh.repo.Insert(ctx, &repository.AttachmentRow{
		ID: id, UserID: testUserID, OwnerType: testOwnerType, OwnerID: testOwnerID,
		StorageKey: fullKey(testUserID, id), ThumbKey: thumbKey(testUserID, id),
	})
	req := withChiParam(authedReq(http.MethodGet, "/"+id), "id", id)
	rr := httptest.NewRecorder()
	hh.h.GetFull(rr, req)
	got, err := io.ReadAll(rr.Result().Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if string(got) != "payload" {
		t.Errorf("body = %q", got)
	}
}

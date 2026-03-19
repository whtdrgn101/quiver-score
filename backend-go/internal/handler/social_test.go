package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/quiverscore/backend-go/internal/repository"
)

type mockSocialRepo struct {
	followResult    *repository.FollowOut
	followErr       error
	unfollowErr     error
	followersResult []repository.FollowOut
	followersErr    error
	followingResult []repository.FollowOut
	followingErr    error
	feedResult      []repository.FeedItemOut
	feedErr         error
}

func (m *mockSocialRepo) Follow(_ context.Context, _, _ string) (*repository.FollowOut, error) {
	return m.followResult, m.followErr
}

func (m *mockSocialRepo) Unfollow(_ context.Context, _, _ string) error {
	return m.unfollowErr
}

func (m *mockSocialRepo) ListFollowers(_ context.Context, _ string) ([]repository.FollowOut, error) {
	return m.followersResult, m.followersErr
}

func (m *mockSocialRepo) ListFollowing(_ context.Context, _ string) ([]repository.FollowOut, error) {
	return m.followingResult, m.followingErr
}

func (m *mockSocialRepo) GetFeed(_ context.Context, _ string, _, _ int) ([]repository.FeedItemOut, error) {
	return m.feedResult, m.feedErr
}

func socialRequest(method, path, userID, targetID string) *http.Request {
	req := authedRequest(method, path, userID)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("userID", targetID)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestSocial_Follow_Success(t *testing.T) {
	username := "archer1"
	mock := &mockSocialRepo{
		followResult: &repository.FollowOut{ID: "f-1", FollowerID: "user-1", FollowingID: "user-2", FollowingUsername: &username, CreatedAt: time.Now()},
	}
	h := &SocialHandler{Social: mock}

	rr := httptest.NewRecorder()
	h.Follow(rr, socialRequest(http.MethodPost, "/follow/user-2", "user-1", "user-2"))

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr.Code)
	}
}

func TestSocial_Follow_Self(t *testing.T) {
	h := &SocialHandler{Social: &mockSocialRepo{}}

	rr := httptest.NewRecorder()
	h.Follow(rr, socialRequest(http.MethodPost, "/follow/user-1", "user-1", "user-1"))

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rr.Code)
	}
}

func TestSocial_Follow_NotFound(t *testing.T) {
	mock := &mockSocialRepo{followErr: repository.ErrNotFound}
	h := &SocialHandler{Social: mock}

	rr := httptest.NewRecorder()
	h.Follow(rr, socialRequest(http.MethodPost, "/follow/user-2", "user-1", "user-2"))

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestSocial_Follow_AlreadyFollowing(t *testing.T) {
	mock := &mockSocialRepo{followErr: repository.ErrAlreadyMember}
	h := &SocialHandler{Social: mock}

	rr := httptest.NewRecorder()
	h.Follow(rr, socialRequest(http.MethodPost, "/follow/user-2", "user-1", "user-2"))

	if rr.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", rr.Code)
	}
}

func TestSocial_Unfollow_Success(t *testing.T) {
	mock := &mockSocialRepo{}
	h := &SocialHandler{Social: mock}

	rr := httptest.NewRecorder()
	h.Unfollow(rr, socialRequest(http.MethodDelete, "/follow/user-2", "user-1", "user-2"))

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
}

func TestSocial_Unfollow_NotFound(t *testing.T) {
	mock := &mockSocialRepo{unfollowErr: repository.ErrNotFound}
	h := &SocialHandler{Social: mock}

	rr := httptest.NewRecorder()
	h.Unfollow(rr, socialRequest(http.MethodDelete, "/follow/user-2", "user-1", "user-2"))

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestSocial_ListFollowers_Success(t *testing.T) {
	mock := &mockSocialRepo{
		followersResult: []repository.FollowOut{{ID: "f-1"}},
	}
	h := &SocialHandler{Social: mock}

	rr := httptest.NewRecorder()
	h.ListFollowers(rr, authedRequest(http.MethodGet, "/followers", "user-1"))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	var result []repository.FollowOut
	json.NewDecoder(rr.Body).Decode(&result)
	if len(result) != 1 {
		t.Errorf("expected 1, got %d", len(result))
	}
}

func TestSocial_ListFollowing_Success(t *testing.T) {
	mock := &mockSocialRepo{
		followingResult: []repository.FollowOut{{ID: "f-1"}, {ID: "f-2"}},
	}
	h := &SocialHandler{Social: mock}

	rr := httptest.NewRecorder()
	h.ListFollowing(rr, authedRequest(http.MethodGet, "/following", "user-1"))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	var result []repository.FollowOut
	json.NewDecoder(rr.Body).Decode(&result)
	if len(result) != 2 {
		t.Errorf("expected 2, got %d", len(result))
	}
}

func TestSocial_GetFeed_Success(t *testing.T) {
	mock := &mockSocialRepo{
		feedResult: []repository.FeedItemOut{},
	}
	h := &SocialHandler{Social: mock}

	rr := httptest.NewRecorder()
	h.GetFeed(rr, authedRequest(http.MethodGet, "/feed", "user-1"))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestSocial_GetFeed_DBError(t *testing.T) {
	mock := &mockSocialRepo{feedErr: errors.New("db error")}
	h := &SocialHandler{Social: mock}

	rr := httptest.NewRecorder()
	h.GetFeed(rr, authedRequest(http.MethodGet, "/feed", "user-1"))

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestSocial_GetFeed_WithQueryParams(t *testing.T) {
	mock := &mockSocialRepo{feedResult: []repository.FeedItemOut{}}
	h := &SocialHandler{Social: mock}

	rr := httptest.NewRecorder()
	h.GetFeed(rr, authedRequest(http.MethodGet, "/feed?limit=5&offset=10", "user-1"))

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

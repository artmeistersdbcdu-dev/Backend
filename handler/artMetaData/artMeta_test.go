package artmetadata

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Blue-Onion/ArtmeisterBackend/internal/database"
	"github.com/Blue-Onion/ArtmeisterBackend/middleware"
	"github.com/Blue-Onion/ArtmeisterBackend/model"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type mockMetaRepo struct {
	database.ArtMetaDataRepository
	likes        map[string]database.ArtLike // "artID:userID" -> like
	comments     map[uuid.UUID]database.ArtComment
	commentSeq   int
	likeErr      error
	unlikeErr    error
	commentErr   error
	deleteErr    error
	getErr       error
}

func (m *mockMetaRepo) LikeArt(ctx context.Context, arg database.LikeArtParams) (database.ArtLike, error) {
	if m.likeErr != nil {
		return database.ArtLike{}, m.likeErr
	}
	key := arg.ArtID.String() + ":" + arg.UserID.String()
	al := database.ArtLike{
		ID:        uuid.New(),
		ArtID:     arg.ArtID,
		UserID:    arg.UserID,
		CreatedAt: time.Now(),
	}
	m.likes[key] = al
	return al, nil
}

func (m *mockMetaRepo) UnlikeArt(ctx context.Context, arg database.UnlikeArtParams) (uuid.UUID, error) {
	if m.unlikeErr != nil {
		return uuid.UUID{}, m.unlikeErr
	}
	key := arg.ArtID.String() + ":" + arg.UserID.String()
	like, ok := m.likes[key]
	if !ok {
		return uuid.UUID{}, sql.ErrNoRows
	}
	delete(m.likes, key)
	return like.ID, nil
}

func (m *mockMetaRepo) AddArtComment(ctx context.Context, arg database.AddArtCommentParams) (database.ArtComment, error) {
	if m.commentErr != nil {
		return database.ArtComment{}, m.commentErr
	}
	c := database.ArtComment{
		ID:        uuid.New(),
		ArtID:     arg.ArtID,
		UserID:    arg.UserID,
		Comment:   arg.Comment,
		CreatedAt: time.Now(),
	}
	m.comments[c.ID] = c
	return c, nil
}

func (m *mockMetaRepo) GetArtCommentsByArtID(ctx context.Context, artID uuid.UUID) ([]database.GetArtCommentsByArtIDRow, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	var res []database.GetArtCommentsByArtIDRow
	for _, c := range m.comments {
		if c.ArtID == artID {
			res = append(res, database.GetArtCommentsByArtIDRow{
				ID:      c.ID,
				ArtID:   c.ArtID,
				UserID:  c.UserID,
				Comment: c.Comment,
			})
		}
	}
	return res, nil
}

func (m *mockMetaRepo) GetArtCommentsCount(ctx context.Context, artID uuid.UUID) (int32, error) {
	if m.getErr != nil {
		return 0, m.getErr
	}
	var count int32
	for _, c := range m.comments {
		if c.ArtID == artID {
			count++
		}
	}
	return count, nil
}

func (m *mockMetaRepo) GetArtLikesCount(ctx context.Context, artID uuid.UUID) (int32, error) {
	if m.getErr != nil {
		return 0, m.getErr
	}
	var count int32
	for key := range m.likes {
		parts := key[:36]
		parsed, _ := uuid.Parse(parts)
		if parsed == artID {
			count++
		}
	}
	return count, nil
}

func (m *mockMetaRepo) DeleteArtComment(ctx context.Context, arg database.DeleteArtCommentParams) (uuid.UUID, error) {
	if m.deleteErr != nil {
		return uuid.UUID{}, m.deleteErr
	}
	c, ok := m.comments[arg.ID]
	if !ok || c.UserID != arg.UserID {
		return uuid.UUID{}, sql.ErrNoRows
	}
	delete(m.comments, arg.ID)
	return arg.ID, nil
}

func newMockMetaRepo() *mockMetaRepo {
	return &mockMetaRepo{
		likes:    make(map[string]database.ArtLike),
		comments: make(map[uuid.UUID]database.ArtComment),
	}
}

func TestHandleLike(t *testing.T) {
	repo := newMockMetaRepo()
	h := &Handler{Repo: repo}

	artID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name           string
		artIDParam     string
		authCtxUser    *middleware.User
		mockErr        error
		expectedStatus int
	}{
		{
			name:       "successful like",
			artIDParam: artID.String(),
			authCtxUser: &middleware.User{
				ID: userID,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "invalid art id",
			artIDParam: "bad-uuid",
			authCtxUser: &middleware.User{
				ID: userID,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthenticated",
			artIDParam:     artID.String(),
			authCtxUser:    nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "repo error",
			artIDParam: artID.String(),
			authCtxUser: &middleware.User{
				ID: userID,
			},
			mockErr:        fmt.Errorf("db error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo.likeErr = tc.mockErr

			req := httptest.NewRequest(http.MethodPost, "/art/"+tc.artIDParam+"/like", nil)
			if tc.authCtxUser != nil {
				ctx := middleware.WithUser(req.Context(), *tc.authCtxUser)
				req = req.WithContext(ctx)
			}
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("art_id", tc.artIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			rr := httptest.NewRecorder()

			h.HandleLike(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestHandleUnLike(t *testing.T) {
	repo := newMockMetaRepo()
	h := &Handler{Repo: repo}

	artID := uuid.New()
	userID := uuid.New()
	repo.likes[artID.String()+":"+userID.String()] = database.ArtLike{
		ID:     uuid.New(),
		ArtID:  artID,
		UserID: userID,
	}

	tests := []struct {
		name           string
		artIDParam     string
		authCtxUser    *middleware.User
		mockErr        error
		expectedStatus int
	}{
		{
			name:       "successful unlike",
			artIDParam: artID.String(),
			authCtxUser: &middleware.User{
				ID: userID,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "like not found",
			artIDParam: artID.String(),
			authCtxUser: &middleware.User{
				ID: uuid.New(),
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:       "invalid art id",
			artIDParam: "bad-uuid",
			authCtxUser: &middleware.User{
				ID: userID,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthenticated",
			artIDParam:     artID.String(),
			authCtxUser:    nil,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo.likes[artID.String()+":"+userID.String()] = database.ArtLike{
				ID:     uuid.New(),
				ArtID:  artID,
				UserID: userID,
			}
			repo.unlikeErr = tc.mockErr

			req := httptest.NewRequest(http.MethodPost, "/art/"+tc.artIDParam+"/unlike", nil)
			if tc.authCtxUser != nil {
				ctx := middleware.WithUser(req.Context(), *tc.authCtxUser)
				req = req.WithContext(ctx)
			}
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("art_id", tc.artIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			rr := httptest.NewRecorder()

			h.HandleUnLike(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestHandleComment(t *testing.T) {
	repo := newMockMetaRepo()
	h := &Handler{Repo: repo}

	artID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name           string
		artIDParam     string
		authCtxUser    *middleware.User
		body           any
		mockErr        error
		expectedStatus int
	}{
		{
			name:       "successful comment",
			artIDParam: artID.String(),
			authCtxUser: &middleware.User{
				ID: userID,
			},
			body:           model.AddComment{Comment: "Nice art!"},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "invalid art id",
			artIDParam: "bad-uuid",
			authCtxUser: &middleware.User{
				ID: userID,
			},
			body:           model.AddComment{Comment: "Nice!"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthenticated",
			artIDParam:     artID.String(),
			authCtxUser:    nil,
			body:           model.AddComment{Comment: "Nice!"},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid json body",
			artIDParam: artID.String(),
			authCtxUser: &middleware.User{
				ID: userID,
			},
			body:           "{bad-json}",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "repo error",
			artIDParam: artID.String(),
			authCtxUser: &middleware.User{
				ID: userID,
			},
			body:           model.AddComment{Comment: "Nice!"},
			mockErr:        fmt.Errorf("db error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo.commentErr = tc.mockErr

			var buf bytes.Buffer
			if s, ok := tc.body.(string); ok {
				buf.WriteString(s)
			} else {
				json.NewEncoder(&buf).Encode(tc.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/art/"+tc.artIDParam+"/comment", &buf)
			if tc.authCtxUser != nil {
				ctx := middleware.WithUser(req.Context(), *tc.authCtxUser)
				req = req.WithContext(ctx)
			}
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("art_id", tc.artIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			rr := httptest.NewRecorder()

			h.HandleComment(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestHandleGetArtComments(t *testing.T) {
	repo := newMockMetaRepo()
	h := &Handler{Repo: repo}

	artID := uuid.New()
	repo.comments[uuid.New()] = database.ArtComment{
		ID:      uuid.New(),
		ArtID:   artID,
		UserID:  uuid.New(),
		Comment: "Great work!",
	}

	tests := []struct {
		name           string
		artIDParam     string
		mockErr        error
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:           "successful get comments",
			artIDParam:     artID.String(),
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "invalid art id",
			artIDParam:     "bad-uuid",
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
		},
		{
			name:           "no comments",
			artIDParam:     uuid.New().String(),
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "db error",
			artIDParam:     artID.String(),
			mockErr:        fmt.Errorf("db error"),
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo.getErr = tc.mockErr

			req := httptest.NewRequest(http.MethodGet, "/art/"+tc.artIDParam+"/comments", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.artIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			rr := httptest.NewRecorder()

			h.HandleGetArtComments(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
			var env struct{ Success bool }
			json.Unmarshal(rr.Body.Bytes(), &env)
			if env.Success != tc.expectSuccess {
				t.Errorf("expected Success=%v, got %v", tc.expectSuccess, env.Success)
			}
		})
	}
}

func TestHandleGetArtCommentsCount(t *testing.T) {
	repo := newMockMetaRepo()
	h := &Handler{Repo: repo}

	artID := uuid.New()
	repo.comments[uuid.New()] = database.ArtComment{
		ID: uuid.New(), ArtID: artID, Comment: "c1",
	}
	repo.comments[uuid.New()] = database.ArtComment{
		ID: uuid.New(), ArtID: artID, Comment: "c2",
	}

	tests := []struct {
		name           string
		artIDParam     string
		mockErr        error
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:           "successful count",
			artIDParam:     artID.String(),
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "invalid art id",
			artIDParam:     "bad-uuid",
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
		},
		{
			name:           "db error",
			artIDParam:     artID.String(),
			mockErr:        fmt.Errorf("db error"),
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo.getErr = tc.mockErr

			req := httptest.NewRequest(http.MethodGet, "/art/"+tc.artIDParam+"/comments/count", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.artIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			rr := httptest.NewRecorder()

			h.HandleGetArtCommentsCount(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
			var env struct{ Success bool }
			json.Unmarshal(rr.Body.Bytes(), &env)
			if env.Success != tc.expectSuccess {
				t.Errorf("expected Success=%v, got %v", tc.expectSuccess, env.Success)
			}
		})
	}
}

func TestHandleGetArtLikeCount(t *testing.T) {
	repo := newMockMetaRepo()
	h := &Handler{Repo: repo}

	artID := uuid.New()
	repo.likes[artID.String()+":"+uuid.New().String()] = database.ArtLike{ID: uuid.New(), ArtID: artID}

	tests := []struct {
		name           string
		artIDParam     string
		mockErr        error
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:           "successful like count",
			artIDParam:     artID.String(),
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "invalid art id",
			artIDParam:     "bad-uuid",
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
		},
		{
			name:           "db error",
			artIDParam:     artID.String(),
			mockErr:        fmt.Errorf("db error"),
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo.getErr = tc.mockErr

			req := httptest.NewRequest(http.MethodGet, "/art/"+tc.artIDParam+"/likes/count", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.artIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			rr := httptest.NewRecorder()

			h.HandleGetArtLikeCount(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
			var env struct{ Success bool }
			json.Unmarshal(rr.Body.Bytes(), &env)
			if env.Success != tc.expectSuccess {
				t.Errorf("expected Success=%v, got %v", tc.expectSuccess, env.Success)
			}
		})
	}
}

func TestHandleDeleteComment(t *testing.T) {
	repo := newMockMetaRepo()
	h := &Handler{Repo: repo}

	commentID := uuid.New()
	userID := uuid.New()
	repo.comments[commentID] = database.ArtComment{
		ID: commentID, UserID: userID, Comment: "test",
	}

	tests := []struct {
		name           string
		commentIDParam string
		authCtxUser    *middleware.User
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "successful delete own comment",
			commentIDParam: commentID.String(),
			authCtxUser: &middleware.User{
				ID: userID,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "forbidden delete other user comment",
			commentIDParam: commentID.String(),
			authCtxUser: &middleware.User{
				ID: uuid.New(),
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "unauthenticated",
			commentIDParam: commentID.String(),
			authCtxUser:    nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid comment id",
			commentIDParam: "bad-uuid",
			authCtxUser: &middleware.User{
				ID: userID,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "comment not found",
			commentIDParam: uuid.New().String(),
			authCtxUser: &middleware.User{
				ID: userID,
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "repo error",
			commentIDParam: commentID.String(),
			authCtxUser: &middleware.User{
				ID: userID,
			},
			mockErr:        fmt.Errorf("db error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo.comments[commentID] = database.ArtComment{
				ID: commentID, UserID: userID, Comment: "test",
			}
			repo.deleteErr = tc.mockErr

			req := httptest.NewRequest(http.MethodDelete, "/art/comment/"+tc.commentIDParam, nil)
			if tc.authCtxUser != nil {
				ctx := middleware.WithUser(req.Context(), *tc.authCtxUser)
				req = req.WithContext(ctx)
			}
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.commentIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			rr := httptest.NewRecorder()

			h.HandleDeleteComment(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

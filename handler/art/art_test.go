package art

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

type mockArtRepo struct {
	database.ArtRepository
	database.UserRepository
	arts         map[uuid.UUID]database.Art
	users        map[uuid.UUID]database.User
	createErr    error
	getErr       error
	updateErr    error
	deleteErr    error
	listErr      error
}

func (m *mockArtRepo) CreateArt(ctx context.Context, arg database.CreateArtParams) (uuid.UUID, error) {
	if m.createErr != nil {
		return uuid.UUID{}, m.createErr
	}
	a := database.Art{
		ID:          arg.ID,
		Name:        arg.Name,
		Description: arg.Description,
		Image:       arg.Image,
		Tags:        arg.Tags,
		Status:      database.ArtStatusPending,
		UserID:      arg.UserID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	m.arts[arg.ID] = a
	return arg.ID, nil
}

func (m *mockArtRepo) GetArtByID(ctx context.Context, id uuid.UUID) (database.GetArtByIDRow, error) {
	if m.getErr != nil {
		return database.GetArtByIDRow{}, m.getErr
	}
	a, ok := m.arts[id]
	if !ok {
		return database.GetArtByIDRow{}, sql.ErrNoRows
	}
	return database.GetArtByIDRow{
		ID:          a.ID,
		Name:        a.Name,
		Description: a.Description,
		Image:       a.Image,
		Tags:        a.Tags,
	}, nil
}

func (m *mockArtRepo) GetArtByUser(ctx context.Context, userID uuid.UUID) ([]database.GetArtByUserRow, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	var res []database.GetArtByUserRow
	for _, a := range m.arts {
		if a.UserID == userID {
			res = append(res, database.GetArtByUserRow{
				ID:          a.ID,
				Name:        a.Name,
				Description: a.Description,
				Image:       a.Image,
			})
		}
	}
	return res, nil
}

func (m *mockArtRepo) UpdateArt(ctx context.Context, arg database.UpdateArtParams) (uuid.UUID, error) {
	if m.updateErr != nil {
		return uuid.UUID{}, m.updateErr
	}
	a, ok := m.arts[arg.ID]
	if !ok || a.UserID != arg.UserID {
		return uuid.UUID{}, sql.ErrNoRows
	}
	if arg.Name.Valid {
		a.Name = arg.Name.String
	}
	if arg.Description.Valid {
		a.Description = arg.Description
	}
	a.UpdatedAt = time.Now()
	m.arts[arg.ID] = a
	return arg.ID, nil
}

func (m *mockArtRepo) DeleteArt(ctx context.Context, arg database.DeleteArtParams) (uuid.UUID, error) {
	if m.deleteErr != nil {
		return uuid.UUID{}, m.deleteErr
	}
	a, ok := m.arts[arg.ID]
	if !ok || a.UserID != arg.UserID {
		return uuid.UUID{}, sql.ErrNoRows
	}
	delete(m.arts, arg.ID)
	return arg.ID, nil
}

func (m *mockArtRepo) GetArtProfileByID(ctx context.Context, arg database.GetArtProfileByIDParams) (database.GetArtProfileByIDRow, error) {
	if m.getErr != nil {
		return database.GetArtProfileByIDRow{}, m.getErr
	}
	a, ok := m.arts[arg.ID]
	if !ok || a.UserID != arg.UserID {
		return database.GetArtProfileByIDRow{}, sql.ErrNoRows
	}
	return database.GetArtProfileByIDRow{
		ID:     a.ID,
		Name:   a.Name,
		Image:  a.Image,
		Status: a.Status,
		UserID: a.UserID,
	}, nil
}

func (m *mockArtRepo) ListArt(ctx context.Context) ([]database.ListArtRow, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var res []database.ListArtRow
	for _, a := range m.arts {
		if a.Status == database.ArtStatusApproved {
			res = append(res, database.ListArtRow{
				ID:     a.ID,
				Name:   a.Name,
				Image:  a.Image,
				Tags:   a.Tags,
				UserID: a.UserID,
			})
		}
	}
	return res, nil
}

func (m *mockArtRepo) ListPendingArt(ctx context.Context) ([]database.ListPendingArtRow, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var res []database.ListPendingArtRow
	for _, a := range m.arts {
		if a.Status == database.ArtStatusPending {
			res = append(res, database.ListPendingArtRow{
				ID:     a.ID,
				Name:   a.Name,
				Image:  a.Image,
				Status: a.Status,
			})
		}
	}
	return res, nil
}

func (m *mockArtRepo) GetUser(ctx context.Context, id uuid.UUID) (database.GetUserRow, error) {
	u, ok := m.users[id]
	if !ok {
		return database.GetUserRow{}, sql.ErrNoRows
	}
	return database.GetUserRow{
		ID:   u.ID,
		Name: u.Name,
	}, nil
}

func newMockArtRepo() *mockArtRepo {
	return &mockArtRepo{
		arts:  make(map[uuid.UUID]database.Art),
		users: make(map[uuid.UUID]database.User),
	}
}

func seedArt(repo *mockArtRepo, overrides ...func(*database.Art)) (uuid.UUID, uuid.UUID) {
	artID := uuid.New()
	userID := uuid.New()
	a := database.Art{
		ID:          artID,
		Name:        "Default Art",
		Description: sql.NullString{String: "A beautiful piece", Valid: true},
		Image:       "http://example.com/img.jpg",
		Tags:        []string{"painting", "modern"},
		Status:      database.ArtStatusApproved,
		UserID:      userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	for _, o := range overrides {
		o(&a)
	}
	repo.arts[artID] = a
	return artID, userID
}

func TestHandleGetArtById(t *testing.T) {
	repo := newMockArtRepo()
	h := &Handler{Repo: repo}

	artID, _ := seedArt(repo)

	tests := []struct {
		name           string
		artIDParam     string
		mockErr        error
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:           "successful get",
			artIDParam:     artID.String(),
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "invalid uuid",
			artIDParam:     "bad-uuid",
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
		},
		{
			name:           "not found",
			artIDParam:     uuid.New().String(),
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
		},
		{
			name:           "empty id param",
			artIDParam:     "",
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
		},
		{
			name:           "db error",
			artIDParam:     artID.String(),
			mockErr:        fmt.Errorf("connection refused"),
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo.getErr = tc.mockErr

			req := httptest.NewRequest(http.MethodGet, "/art/"+tc.artIDParam, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.artIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			rr := httptest.NewRecorder()

			h.HandleGetArtById(rr, req)

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

func TestHandleGetArts(t *testing.T) {
	repo := newMockArtRepo()
	h := &Handler{Repo: repo}

	_, userID := seedArt(repo)

	tests := []struct {
		name           string
		userIDParam    string
		mockErr        error
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:           "successful list user arts",
			userIDParam:    userID.String(),
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "invalid uuid",
			userIDParam:    "bad-uuid",
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
		},
		{
			name:           "empty list for unknown user",
			userIDParam:    uuid.New().String(),
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "empty param",
			userIDParam:    "",
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
		},
		{
			name:           "db error",
			userIDParam:    userID.String(),
			mockErr:        fmt.Errorf("db error"),
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo.getErr = tc.mockErr

			req := httptest.NewRequest(http.MethodGet, "/art/u/"+tc.userIDParam, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("user_id", tc.userIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			rr := httptest.NewRecorder()

			h.HandleGetArts(rr, req)

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

func TestHandleArtDeletion(t *testing.T) {
	repo := newMockArtRepo()
	h := &Handler{Repo: repo}

	artID, userID := seedArt(repo)

	tests := []struct {
		name           string
		artIDParam     string
		authCtxUser    *middleware.User
		mockErr        error
		expectedStatus int
	}{
		{
			name:       "successful delete own art",
			artIDParam: artID.String(),
			authCtxUser: &middleware.User{
				ID: userID,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "forbidden delete other user art",
			artIDParam: artID.String(),
			authCtxUser: &middleware.User{
				ID: uuid.New(),
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "unauthenticated",
			artIDParam:     artID.String(),
			authCtxUser:    nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid uuid",
			artIDParam: "not-a-uuid",
			authCtxUser: &middleware.User{
				ID: userID,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "art not found",
			artIDParam: uuid.New().String(),
			authCtxUser: &middleware.User{
				ID: userID,
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:       "db internal error",
			artIDParam: artID.String(),
			authCtxUser: &middleware.User{
				ID: userID,
			},
			mockErr:        fmt.Errorf("disk full"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo.arts[artID] = repo.arts[artID]
			repo.deleteErr = tc.mockErr

			req := httptest.NewRequest(http.MethodDelete, "/art/"+tc.artIDParam, nil)
			if tc.authCtxUser != nil {
				ctx := middleware.WithUser(req.Context(), *tc.authCtxUser)
				req = req.WithContext(ctx)
			}
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.artIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			rr := httptest.NewRecorder()

			h.HandleArtDeletion(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

func strPtr(s string) *string { return &s }

func TestHandlerArtUpdation(t *testing.T) {
	repo := newMockArtRepo()
	h := &Handler{Repo: repo}

	artID, userID := seedArt(repo)

	tests := []struct {
		name           string
		artIDParam     string
		authCtxUser    *middleware.User
		body           any
		mockErr        error
		expectedStatus int
	}{
		{
			name:       "successful update",
			artIDParam: artID.String(),
			authCtxUser: &middleware.User{
				ID: userID,
			},
			body:           model.UpdateArtRequest{Name: strPtr("NewName")},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "forbidden update other user art",
			artIDParam: artID.String(),
			authCtxUser: &middleware.User{
				ID: uuid.New(),
			},
			body:           model.UpdateArtRequest{Name: strPtr("Hack")},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "unauthenticated",
			artIDParam:     artID.String(),
			authCtxUser:    nil,
			body:           model.UpdateArtRequest{Name: strPtr("X")},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid uuid",
			artIDParam: "not-a-uuid",
			authCtxUser: &middleware.User{
				ID: userID,
			},
			body:           model.UpdateArtRequest{Name: strPtr("X")},
			expectedStatus: http.StatusBadRequest,
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
			body:           model.UpdateArtRequest{Name: strPtr("Fail")},
			mockErr:        fmt.Errorf("db error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo.arts[artID] = repo.arts[artID]
			repo.updateErr = tc.mockErr

			var buf bytes.Buffer
			if s, ok := tc.body.(string); ok {
				buf.WriteString(s)
			} else {
				json.NewEncoder(&buf).Encode(tc.body)
			}

			req := httptest.NewRequest(http.MethodPatch, "/art/"+tc.artIDParam, &buf)
			if tc.authCtxUser != nil {
				ctx := middleware.WithUser(req.Context(), *tc.authCtxUser)
				req = req.WithContext(ctx)
			}
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.artIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			rr := httptest.NewRecorder()

			h.HandlerArtUpdation(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestHandleGetArtProfileById(t *testing.T) {
	repo := newMockArtRepo()
	h := &Handler{Repo: repo}

	artID, userID := seedArt(repo)

	tests := []struct {
		name           string
		artIDParam     string
		userIDParam    string
		mockErr        error
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:           "successful get profile",
			artIDParam:     artID.String(),
			userIDParam:    userID.String(),
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "invalid art id",
			artIDParam:     "bad-id",
			userIDParam:    userID.String(),
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
		},
		{
			name:           "invalid user id",
			artIDParam:     artID.String(),
			userIDParam:    "bad-id",
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
		},
		{
			name:           "not found",
			artIDParam:     uuid.New().String(),
			userIDParam:    userID.String(),
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo.getErr = tc.mockErr

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/art/p/%s/%s", tc.userIDParam, tc.artIDParam), nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.artIDParam)
			rctx.URLParams.Add("user_id", tc.userIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			rr := httptest.NewRecorder()

			h.HandleGetArtProfileById(rr, req)

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

func TestHandleGetApprovedArt(t *testing.T) {
	repo := newMockArtRepo()
	h := &Handler{Repo: repo}

	seedArt(repo) // approved by default
	seedArt(repo, func(a *database.Art) {
		a.Status = database.ArtStatusPending
	})

	tests := []struct {
		name           string
		mockErr        error
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:           "successful list approved",
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "db error",
			mockErr:        fmt.Errorf("db error"),
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo.listErr = tc.mockErr

			req := httptest.NewRequest(http.MethodGet, "/art/", nil)
			rr := httptest.NewRecorder()
			h.HandleGetApprovedArt(rr, req)

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

func TestHandleGetPendingArt(t *testing.T) {
	repo := newMockArtRepo()
	h := &Handler{Repo: repo}

	seedArt(repo, func(a *database.Art) {
		a.Status = database.ArtStatusPending
	})

	tests := []struct {
		name           string
		mockErr        error
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:           "successful list pending",
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "db error",
			mockErr:        fmt.Errorf("db error"),
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo.listErr = tc.mockErr

			req := httptest.NewRequest(http.MethodGet, "/art/pending-art", nil)
			rr := httptest.NewRecorder()
			h.HandleGetPendingArt(rr, req)

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

func TestHandlerGetArtistProfile(t *testing.T) {
	repo := newMockArtRepo()
	userID := uuid.New()
	repo.users[userID] = database.User{
		ID:   userID,
		Name: "Artist Name",
	}
	seedArt(repo, func(a *database.Art) {
		a.UserID = userID
		a.Status = database.ArtStatusApproved
	})

	h := &ProfileHandler{ArtRepo: repo, UserRepo: repo}

	tests := []struct {
		name           string
		userIDParam    string
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:           "successful get profile",
			userIDParam:    userID.String(),
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "invalid uuid",
			userIDParam:    "bad-uuid",
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
		},
		{
			name:           "user not found",
			userIDParam:    uuid.New().String(),
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/art/u/profile/"+tc.userIDParam, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.userIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			rr := httptest.NewRecorder()

			h.HandlerGetArtistProfile(rr, req)

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

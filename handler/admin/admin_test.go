package admin

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Blue-Onion/ArtmeisterBackend/internal/database"
	"github.com/Blue-Onion/ArtmeisterBackend/middleware"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type mockAdminRepo struct {
	database.UserRepository
	database.ArtRepository
	users     map[uuid.UUID]database.User
	arts      map[uuid.UUID]database.Art
	patchErr  error
	statusErr error
}

func (m *mockAdminRepo) PatchUserAdmin(ctx context.Context, arg database.PatchUserAdminParams) (database.PatchUserAdminRow, error) {
	if m.patchErr != nil {
		return database.PatchUserAdminRow{}, m.patchErr
	}
	u, ok := m.users[arg.ID]
	if !ok {
		return database.PatchUserAdminRow{}, sql.ErrNoRows
	}
	if arg.Role.Valid {
		u.Role = arg.Role.UserRole
	}
	if arg.Status.Valid {
		u.Status = arg.Status.AccountStatus
	}
	m.users[arg.ID] = u
	return database.PatchUserAdminRow{
		ID:     u.ID,
		Status: u.Status,
		Role:   u.Role,
	}, nil
}

func (m *mockAdminRepo) UpdateArtStatus(ctx context.Context, arg database.UpdateArtStatusParams) (database.UpdateArtStatusRow, error) {
	if m.statusErr != nil {
		return database.UpdateArtStatusRow{}, m.statusErr
	}
	a, ok := m.arts[arg.ID]
	if !ok {
		return database.UpdateArtStatusRow{}, sql.ErrNoRows
	}
	a.Status = arg.Status
	m.arts[arg.ID] = a
	return database.UpdateArtStatusRow{
		ID:     a.ID,
		Status: a.Status,
	}, nil
}

func newMockAdminRepo() *mockAdminRepo {
	return &mockAdminRepo{
		users: make(map[uuid.UUID]database.User),
		arts:  make(map[uuid.UUID]database.Art),
	}
}

func TestHandlerUserStatus(t *testing.T) {
	repo := newMockAdminRepo()
	h := &UserHandler{Repo: repo}

	userUUID := uuid.New()
	repo.users[userUUID] = database.User{
		ID:     userUUID,
		Name:   "Alice",
		Role:   database.UserRoleMember,
		Status: database.AccountStatusPending,
	}

	seniorID := uuid.New()
	seniorUser := middleware.User{
		ID:     seniorID,
		Role:   database.UserRolePresident,
		Status: database.AccountStatusApproved,
	}

	tests := []struct {
		name           string
		userIDParam    string
		queryRole      string
		queryStatus    string
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "approve user status",
			userIDParam:    userUUID.String(),
			queryStatus:    "approved",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "change role",
			userIDParam:    userUUID.String(),
			queryRole:      "core_member",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "cannot assign higher role",
			userIDParam:    userUUID.String(),
			queryRole:      "president",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "both empty",
			userIDParam:    userUUID.String(),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "both provided",
			userIDParam:    userUUID.String(),
			queryRole:      "member",
			queryStatus:    "approved",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid role value",
			userIDParam:    userUUID.String(),
			queryRole:      "invalid_role",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "repo error",
			userIDParam:    userUUID.String(),
			queryStatus:    "approved",
			mockErr:        errors.New("db error"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "invalid uuid",
			userIDParam:    "not-a-uuid",
			queryStatus:    "approved",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "user not found",
			userIDParam:    uuid.New().String(),
			queryStatus:    "approved",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo.patchErr = tc.mockErr

			query := fmt.Sprintf("/admin/users/%s/status", tc.userIDParam)
			if tc.queryRole != "" || tc.queryStatus != "" {
				query += "?"
				first := true
				if tc.queryRole != "" {
					query += "role=" + tc.queryRole
					first = false
				}
				if tc.queryStatus != "" {
					if !first {
						query += "&"
					}
					query += "status=" + tc.queryStatus
				}
			}
			req := httptest.NewRequest(http.MethodPatch, query, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("user_id", tc.userIDParam)
			ctx := middleware.WithSenior(req.Context(), seniorUser)
			req = req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))
			rr := httptest.NewRecorder()

			h.HandlerRole(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestHandlerArtStatus(t *testing.T) {
	repo := newMockAdminRepo()
	h := &ArtHandler{Repo: repo}

	artUUID := uuid.New()
	repo.arts[artUUID] = database.Art{
		ID:     artUUID,
		Name:   "Artwork",
		Status: database.ArtStatusPending,
	}

	moderatorID := uuid.New()
	moderatorUser := middleware.User{
		ID:     moderatorID,
		Role:   database.UserRolePresident,
		Status: database.AccountStatusApproved,
	}

	tests := []struct {
		name           string
		artIDParam     string
		queryStatus    string
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "approve art",
			artIDParam:     artUUID.String(),
			queryStatus:    "approved",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "reject art",
			artIDParam:     artUUID.String(),
			queryStatus:    "rejected",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid status",
			artIDParam:     artUUID.String(),
			queryStatus:    "invalid-status",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty status",
			artIDParam:     artUUID.String(),
			queryStatus:    "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "repo failure",
			artIDParam:     artUUID.String(),
			queryStatus:    "approved",
			mockErr:        errors.New("db write failed"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "invalid art uuid",
			artIDParam:     "not-a-uuid",
			queryStatus:    "approved",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "art not found",
			artIDParam:     uuid.New().String(),
			queryStatus:    "approved",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo.statusErr = tc.mockErr

			query := fmt.Sprintf("/admin/arts/%s/status?status=%s", tc.artIDParam, tc.queryStatus)
			req := httptest.NewRequest(http.MethodPatch, query, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("art_id", tc.artIDParam)
			ctx := middleware.WithModerator(req.Context(), moderatorUser)
			req = req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))
			rr := httptest.NewRecorder()

			h.HandlerArtStatus(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

package user

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
	"github.com/Blue-Onion/ArtmeisterBackend/utlis"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type mockUserRepo struct {
	database.UserRepository
	users        map[uuid.UUID]database.User
	emails       map[string]database.User
	createErr    error
	getErr       error
	getAllErr    error
	updateErr    error
	passwordErr  error
}

func (m *mockUserRepo) CreateUser(ctx context.Context, arg database.CreateUserParams) (uuid.UUID, error) {
	if m.createErr != nil {
		return uuid.UUID{}, m.createErr
	}
	id := uuid.New()
	u := database.User{
		ID:        id,
		Name:      arg.Name,
		Email:     arg.Email,
		Password:  arg.Password,
		Batch:     sql.NullString{String: "", Valid: true},
		Status:    database.AccountStatusPending,
		Role:      database.UserRoleMember,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	m.users[id] = u
	m.emails[arg.Email] = u
	return id, nil
}

func (m *mockUserRepo) GetUser(ctx context.Context, id uuid.UUID) (database.GetUserRow, error) {
	if m.getErr != nil {
		return database.GetUserRow{}, m.getErr
	}
	u, ok := m.users[id]
	if !ok {
		return database.GetUserRow{}, sql.ErrNoRows
	}
	return database.GetUserRow{
		ID:          u.ID,
		Name:        u.Name,
		Username:    u.Username,
		Email:       u.Email,
		Batch:       u.Batch,
		Status:      u.Status,
		Role:        u.Role,
		Image:       u.Image,
		BannerImage: u.BannerImage,
		Description: u.Description,
		SocialLinks: u.SocialLinks,
	}, nil
}

func (m *mockUserRepo) GetUserByEmail(ctx context.Context, email string) (database.GetUserByEmailRow, error) {
	u, ok := m.emails[email]
	if !ok {
		return database.GetUserByEmailRow{}, sql.ErrNoRows
	}
	return database.GetUserByEmailRow{
		ID:       u.ID,
		Name:     u.Name,
		Email:    u.Email,
		Password: u.Password,
		Image:    u.Image,
	}, nil
}

func (m *mockUserRepo) GetUserByUsername(ctx context.Context, username sql.NullString) (database.GetUserByUsernameRow, error) {
	if m.getErr != nil {
		return database.GetUserByUsernameRow{}, m.getErr
	}
	for _, u := range m.users {
		if u.Username.Valid && u.Username.String == username.String {
			return database.GetUserByUsernameRow{
				ID:          u.ID,
				Name:        u.Name,
				Username:    u.Username,
				Email:       u.Email,
				Batch:       u.Batch,
				Status:      u.Status,
				Role:        u.Role,
				Image:       u.Image,
				BannerImage: u.BannerImage,
				Description: u.Description,
				SocialLinks: u.SocialLinks,
			}, nil
		}
	}
	return database.GetUserByUsernameRow{}, sql.ErrNoRows
}

func (m *mockUserRepo) PatchUserProfile(ctx context.Context, arg database.PatchUserProfileParams) (database.PatchUserProfileRow, error) {
	if m.updateErr != nil {
		return database.PatchUserProfileRow{}, m.updateErr
	}
	u, ok := m.users[arg.ID]
	if !ok {
		return database.PatchUserProfileRow{}, sql.ErrNoRows
	}
	if arg.Name.Valid {
		u.Name = arg.Name.String
	}
	if arg.Email.Valid {
		u.Email = arg.Email.String
	}
	if arg.Username.Valid {
		u.Username = arg.Username
	}
	m.users[arg.ID] = u
	return database.PatchUserProfileRow{
		ID:          u.ID,
		Name:        u.Name,
		Username:    u.Username,
		Email:       u.Email,
		Batch:       u.Batch,
		Status:      u.Status,
		Role:        u.Role,
		Image:       u.Image,
		BannerImage: u.BannerImage,
		Description: u.Description,
		SocialLinks: u.SocialLinks,
	}, nil
}

func (m *mockUserRepo) PatchUserPassword(ctx context.Context, arg database.PatchUserPasswordParams) (uuid.UUID, error) {
	if m.passwordErr != nil {
		return uuid.UUID{}, m.passwordErr
	}
	u, ok := m.users[arg.ID]
	if !ok {
		return uuid.UUID{}, sql.ErrNoRows
	}
	u.Password = arg.Password
	m.users[arg.ID] = u
	return u.ID, nil
}

func (m *mockUserRepo) GetAllUser(ctx context.Context) ([]database.GetAllUserRow, error) {
	if m.getAllErr != nil {
		return nil, m.getAllErr
	}
	var res []database.GetAllUserRow
	for _, u := range m.users {
		res = append(res, database.GetAllUserRow{
			ID:     u.ID,
			Name:   u.Name,
			Email:  u.Email,
			Status: u.Status,
			Role:   u.Role,
			Image:  u.Image,
		})
	}
	return res, nil
}

func (m *mockUserRepo) GetAllUserApproved(ctx context.Context) ([]database.GetAllUserApprovedRow, error) {
	if m.getAllErr != nil {
		return nil, m.getAllErr
	}
	var res []database.GetAllUserApprovedRow
	for _, u := range m.users {
		if u.Status == database.AccountStatusApproved {
			res = append(res, database.GetAllUserApprovedRow{
				ID:   u.ID,
				Name: u.Name,
				Role: u.Role,
			})
		}
	}
	return res, nil
}

func (m *mockUserRepo) CheckUsrById(ctx context.Context, id uuid.UUID) (database.CheckUsrByIdRow, error) {
	u, ok := m.users[id]
	if !ok {
		return database.CheckUsrByIdRow{}, sql.ErrNoRows
	}
	return database.CheckUsrByIdRow{
		ID:     u.ID,
		Status: u.Status,
		Role:   u.Role,
	}, nil
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:  make(map[uuid.UUID]database.User),
		emails: make(map[string]database.User),
	}
}

func strPtr(s string) *string {
	return &s
}

func seedUser(repo *mockUserRepo, overrides ...func(*database.User)) (uuid.UUID, string) {
	id := uuid.New()
	hashed, _ := utlis.HashPassword("default_password")
	u := database.User{
		ID:        id,
		Name:      "Default User",
		Email:     fmt.Sprintf("user-%s@example.com", id.String()[:8]),
		Password:  hashed,
		Status:    database.AccountStatusApproved,
		Role:      database.UserRoleMember,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	for _, o := range overrides {
		o(&u)
	}
	repo.users[id] = u
	repo.emails[u.Email] = u
	return id, u.Email
}

func TestHandleCreateUser(t *testing.T) {
	repo := newMockUserRepo()
	h := &Handler{Repo: repo}

	tests := []struct {
		name           string
		body           any
		mockErr        error
		expectedStatus int
		expectCookie   bool
	}{
		{
			name: "successful registration",
			body: model.CreateUser{
				Name:     "Alice",
				Email:    "alice@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusCreated,
			expectCookie:   true,
		},
		{
			name:           "invalid json body",
			body:           "{invalid-json}",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty body",
			body:           "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "repo create error",
			body: model.CreateUser{
				Name:     "Bob",
				Email:    "bob@example.com",
				Password: "password123",
			},
			mockErr:        fmt.Errorf("pq: duplicate key value violates unique constraint"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "empty name and email still passes validator",
			body: model.CreateUser{
				Name:     "",
				Email:    "",
				Password: "",
			},
			expectedStatus: http.StatusCreated,
			expectCookie:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo.createErr = tc.mockErr
			var buf bytes.Buffer
			if s, ok := tc.body.(string); ok {
				buf.WriteString(s)
			} else {
				json.NewEncoder(&buf).Encode(tc.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/auth/users", &buf)
			rr := httptest.NewRecorder()

			h.HandleCreateUser(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}

			cookies := rr.Result().Cookies()
			foundCookie := false
			for _, c := range cookies {
				if c.Name == "authToken" {
					foundCookie = true
				}
			}
			if tc.expectCookie && !foundCookie {
				t.Errorf("expected authToken cookie in response")
			}
		})
	}
}

func TestHandleLogin(t *testing.T) {
	repo := newMockUserRepo()
	h := &Handler{Repo: repo}

	pwd := "mypassword"
	hashed, _ := utlis.HashPassword(pwd)
	userID, _ := seedUser(repo, func(u *database.User) {
		u.Password = hashed
		u.Email = "alice@example.com"
		u.Name = "Alice"
	})
	_ = userID

	tests := []struct {
		name           string
		body           any
		expectedStatus int
		expectCookie   bool
	}{
		{
			name: "successful login",
			body: model.AuthenticateUser{
				Email:    "alice@example.com",
				Password: pwd,
			},
			expectedStatus: http.StatusOK,
			expectCookie:   true,
		},
		{
			name: "non-existent email",
			body: model.AuthenticateUser{
				Email:    "unknown@example.com",
				Password: pwd,
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "incorrect password",
			body: model.AuthenticateUser{
				Email:    "alice@example.com",
				Password: "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid json body",
			body:           "{bad-json}",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "empty email",
			body: model.AuthenticateUser{
				Email:    "",
				Password: pwd,
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "empty password",
			body: model.AuthenticateUser{
				Email:    "alice@example.com",
				Password: "",
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			if s, ok := tc.body.(string); ok {
				buf.WriteString(s)
			} else {
				json.NewEncoder(&buf).Encode(tc.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/auth/login", &buf)
			rr := httptest.NewRecorder()

			h.HandleLogin(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}

			cookies := rr.Result().Cookies()
			foundCookie := false
			for _, c := range cookies {
				if c.Name == "authToken" {
					foundCookie = true
				}
			}
			if tc.expectCookie != foundCookie {
				t.Errorf("expected cookie presence %v, got %v", tc.expectCookie, foundCookie)
			}
		})
	}
}

func TestHandleLogOut(t *testing.T) {
	h := &Handler{}
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	rr := httptest.NewRecorder()

	h.HandleLogOut(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	cookies := rr.Result().Cookies()
	foundClearedCookie := false
	for _, c := range cookies {
		if c.Name == "authToken" && c.MaxAge == -1 {
			foundClearedCookie = true
		}
	}
	if !foundClearedCookie {
		t.Errorf("expected authToken cookie to be cleared (MaxAge=-1)")
	}

	var resp struct {
		Success bool
		Data    map[string]string
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Success {
		t.Errorf("expected success=true, got false")
	}
	if resp.Data["message"] != "Logged out successfully" {
		t.Errorf("expected logout message, got %v", resp.Data)
	}
}

func TestHandleUpdateUserProfile(t *testing.T) {
	repo := newMockUserRepo()
	h := &Handler{Repo: repo}

	userID, _ := seedUser(repo, func(u *database.User) {
		u.Name = "Alice"
		u.Email = "alice@example.com"
		u.Role = database.UserRoleMember
		u.Status = database.AccountStatusApproved
	})

	tests := []struct {
		name           string
		userIDParam    string
		authCtxUser    *middleware.User
		body           any
		mockErr        error
		expectedStatus int
	}{
		{
			name:        "successful self update",
			userIDParam: userID.String(),
			authCtxUser: &middleware.User{
				ID:   userID,
				Role: database.UserRoleMember,
			},
			body:           model.PatchUserProfileRequest{UserName: strPtr("Alice Updated")},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "forbidden other user update",
			userIDParam: uuid.New().String(),
			authCtxUser: &middleware.User{
				ID:   userID,
				Role: database.UserRoleMember,
			},
			body:           model.PatchUserProfileRequest{UserName: strPtr("Hacked")},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:        "admin can update any user",
			userIDParam: uuid.New().String(),
			authCtxUser: &middleware.User{
				ID:   userID,
				Role: database.UserRolePresident,
			},
			body:           model.PatchUserProfileRequest{UserName: strPtr("Admin Edit")},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthenticated",
			userIDParam:    userID.String(),
			authCtxUser:    nil,
			body:           model.PatchUserProfileRequest{UserName: strPtr("Sneaky")},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:        "invalid uuid param",
			userIDParam: "not-a-uuid",
			authCtxUser: &middleware.User{
				ID:   userID,
				Role: database.UserRoleMember,
			},
			body:           model.PatchUserProfileRequest{UserName: strPtr("X")},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "invalid json body",
			userIDParam: userID.String(),
			authCtxUser: &middleware.User{
				ID:   userID,
				Role: database.UserRoleMember,
			},
			body:           "{bad-json}",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "repo error on update",
			userIDParam: userID.String(),
			authCtxUser: &middleware.User{
				ID:   userID,
				Role: database.UserRoleMember,
			},
			body:           model.PatchUserProfileRequest{UserName: strPtr("Fail")},
			mockErr:        fmt.Errorf("db connection lost"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			targetUUID, _ := uuid.Parse(tc.userIDParam)
			if targetUUID != userID && targetUUID != uuid.Nil {
				repo.users[targetUUID] = database.User{
					ID:     targetUUID,
					Name:   "Other User",
					Role:   database.UserRoleMember,
					Status: database.AccountStatusApproved,
				}
			}

			repo.updateErr = tc.mockErr

			var buf bytes.Buffer
			if s, ok := tc.body.(string); ok {
				buf.WriteString(s)
			} else {
				json.NewEncoder(&buf).Encode(tc.body)
			}

			req := httptest.NewRequest(http.MethodPatch, "/auth/users/"+tc.userIDParam, &buf)

			if tc.authCtxUser != nil {
				ctx := middleware.WithUser(req.Context(), *tc.authCtxUser)
				req = req.WithContext(ctx)
			}

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.userIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			h.HandleUpdateUserProfile(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestHandlePasswordChange(t *testing.T) {
	repo := newMockUserRepo()
	h := &Handler{Repo: repo}

	pwd := "original_password"
	hashed, _ := utlis.HashPassword(pwd)
	userID, userEmail := seedUser(repo, func(u *database.User) {
		u.Password = hashed
		u.Email = "alice@example.com"
	})

	tests := []struct {
		name            string
		authCtxUser     *middleware.User
		body            any
		mockPasswordErr error
		expectedStatus  int
	}{
		{
			name: "successful password change",
			authCtxUser: &middleware.User{
				ID:    userID,
				Email: userEmail,
			},
			body: model.PatchUserPassword{
				OldPassword: pwd,
				Password:    "new_password_123",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "wrong old password",
			authCtxUser: &middleware.User{
				ID:    userID,
				Email: userEmail,
			},
			body: model.PatchUserPassword{
				OldPassword: "wrong_old_pwd",
				Password:    "new_password_123",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "unauthenticated",
			authCtxUser:    nil,
			body:           model.PatchUserPassword{OldPassword: pwd, Password: "x"},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid json body",
			authCtxUser: &middleware.User{
				ID:    userID,
				Email: userEmail,
			},
			body:           "{not-json}",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "user email not found",
			authCtxUser: &middleware.User{
				ID:    userID,
				Email: "ghost@example.com",
			},
			body: model.PatchUserPassword{
				OldPassword: pwd,
				Password:    "new_password_789",
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "repo save failure",
			authCtxUser: &middleware.User{
				ID:    userID,
				Email: userEmail,
			},
			body: model.PatchUserPassword{
				OldPassword: pwd,
				Password:    "new_password_456",
			},
			mockPasswordErr: fmt.Errorf("db write error"),
			expectedStatus:  http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			user := repo.users[userID]
			user.Password = hashed
			repo.users[userID] = user
			repo.emails[userEmail] = user
			repo.passwordErr = tc.mockPasswordErr

			var buf bytes.Buffer
			if s, ok := tc.body.(string); ok {
				buf.WriteString(s)
			} else {
				json.NewEncoder(&buf).Encode(tc.body)
			}

			req := httptest.NewRequest(http.MethodPatch, "/auth/users/password", &buf)
			if tc.authCtxUser != nil {
				ctx := middleware.WithUser(req.Context(), *tc.authCtxUser)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			h.HandlePasswordChange(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestHandleGetUserById(t *testing.T) {
	repo := newMockUserRepo()
	h := &Handler{Repo: repo}

	userID, _ := seedUser(repo, func(u *database.User) {
		u.Name = "Visible User"
		u.Email = "visible@example.com"
	})

	tests := []struct {
		name           string
		userIDParam    string
		mockErr        error
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:           "successful get user",
			userIDParam:    userID.String(),
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "invalid uuid",
			userIDParam:    "not-a-uuid",
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
		},
		{
			name:           "user not found",
			userIDParam:    uuid.New().String(),
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
		},
		{
			name:           "db error",
			userIDParam:    userID.String(),
			mockErr:        fmt.Errorf("connection refused"),
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo.getErr = tc.mockErr

			req := httptest.NewRequest(http.MethodGet, "/auth/users/"+tc.userIDParam, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.userIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			rr := httptest.NewRecorder()

			h.HandleGetUserById(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}

			var env struct {
				Success bool
			}
			json.Unmarshal(rr.Body.Bytes(), &env)
			if env.Success != tc.expectSuccess {
				t.Errorf("expected Success=%v, got %v", tc.expectSuccess, env.Success)
			}
		})
	}
}

func TestHandleMe(t *testing.T) {
	repo := newMockUserRepo()
	h := &Handler{Repo: repo}

	seedUser(repo)

	tests := []struct {
		name           string
		setCookie      bool
		cookieValue    string
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:           "no cookie present",
			setCookie:      false,
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
			if tc.setCookie {
				req.AddCookie(&http.Cookie{
					Name:  "authToken",
					Value: tc.cookieValue,
				})
			}
			rr := httptest.NewRecorder()
			h.HandleMe(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
			var env struct {
				Success bool
			}
			json.Unmarshal(rr.Body.Bytes(), &env)
			if env.Success != tc.expectSuccess {
				t.Errorf("expected Success=%v, got %v", tc.expectSuccess, env.Success)
			}
		})
	}
}

func TestHandleGetAllUser(t *testing.T) {
	repo := newMockUserRepo()
	h := &Handler{Repo: repo}

	seedUser(repo)
	seedUser(repo)

	tests := []struct {
		name           string
		mockErr        error
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:           "successful list all users",
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
			repo.getAllErr = tc.mockErr

			req := httptest.NewRequest(http.MethodGet, "/auth/users", nil)
			rr := httptest.NewRecorder()
			h.HandleGetAllUser(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}

			var env struct {
				Success bool
			}
			json.Unmarshal(rr.Body.Bytes(), &env)
			if env.Success != tc.expectSuccess {
				t.Errorf("expected Success=%v, got %v", tc.expectSuccess, env.Success)
			}
		})
	}
}

func TestHandleGetApprovedUser(t *testing.T) {
	repo := newMockUserRepo()
	h := &Handler{Repo: repo}

	seedUser(repo, func(u *database.User) {
		u.Status = database.AccountStatusApproved
	})
	seedUser(repo, func(u *database.User) {
		u.Status = database.AccountStatusPending
	})

	tests := []struct {
		name           string
		mockErr        error
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:           "successful list approved users",
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
			repo.getAllErr = tc.mockErr

			req := httptest.NewRequest(http.MethodGet, "/auth/main-users", nil)
			rr := httptest.NewRecorder()
			h.HandleGetApprovedUser(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}

			var env struct {
				Success bool
			}
			json.Unmarshal(rr.Body.Bytes(), &env)
			if env.Success != tc.expectSuccess {
				t.Errorf("expected Success=%v, got %v", tc.expectSuccess, env.Success)
			}
		})
	}
}

func TestHandleGetUserByUserName(t *testing.T) {
	repo := newMockUserRepo()
	h := &Handler{Repo: repo}

	seedUser(repo, func(u *database.User) {
		u.Username = sql.NullString{String: "coolartist", Valid: true}
	})

	tests := []struct {
		name           string
		usernameParam  string
		mockErr        error
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:           "successful get by username",
			usernameParam:  "coolartist",
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "username not found",
			usernameParam:  "unknown",
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
		},
		{
			name:           "empty username param",
			usernameParam:  "",
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
		},
		{
			name:           "db error",
			usernameParam:  "coolartist",
			mockErr:        fmt.Errorf("db error"),
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo.getErr = tc.mockErr

			req := httptest.NewRequest(http.MethodGet, "/auth/users/by-username/"+tc.usernameParam, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("user-name", tc.usernameParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			rr := httptest.NewRecorder()

			h.HandleGetUserByUserName(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}

			var env struct {
				Success bool
			}
			json.Unmarshal(rr.Body.Bytes(), &env)
			if env.Success != tc.expectSuccess {
				t.Errorf("expected Success=%v, got %v", tc.expectSuccess, env.Success)
			}
		})
	}
}

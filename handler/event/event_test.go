package event

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
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

type mockEventRepo struct {
	database.EventRepository
	database.EventAttendeesRepository
	events       map[uuid.UUID]database.Event
	attendees    map[uuid.UUID][]uuid.UUID
	createErr    error
	getErr       error
	updateErr    error
	attendeeErr  error
}

func (m *mockEventRepo) CreateEvent(ctx context.Context, arg database.CreateEventParams) (uuid.UUID, error) {
	if m.createErr != nil {
		return uuid.UUID{}, m.createErr
	}
	e := database.Event{
		ID:          arg.ID,
		Name:        arg.Name,
		Description: arg.Description,
		Venue:       arg.Venue,
		Image:       arg.Image,
		BannerImage: arg.BannerImage,
		EventDate:   arg.EventDate,
		Status:      arg.Status,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	m.events[arg.ID] = e
	return arg.ID, nil
}

func (m *mockEventRepo) GetEventByID(ctx context.Context, id uuid.UUID) (database.GetEventByIDRow, error) {
	if m.getErr != nil {
		return database.GetEventByIDRow{}, m.getErr
	}
	e, ok := m.events[id]
	if !ok {
		return database.GetEventByIDRow{}, sql.ErrNoRows
	}
	return database.GetEventByIDRow{
		ID:          e.ID,
		Name:        e.Name,
		Description: e.Description,
		Venue:       e.Venue,
		Image:       e.Image,
		BannerImage: e.BannerImage,
		EventDate:   e.EventDate,
		Status:      e.Status,
	}, nil
}

func (m *mockEventRepo) ListEvents(ctx context.Context) ([]database.ListEventsRow, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	var res []database.ListEventsRow
	for _, e := range m.events {
		res = append(res, database.ListEventsRow{
			ID:          e.ID,
			Name:        e.Name,
			Description: e.Description,
			Venue:       e.Venue,
			Image:       e.Image,
			EventDate:   e.EventDate,
			Status:      e.Status,
		})
	}
	return res, nil
}

func (m *mockEventRepo) DeleteEvent(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	if m.getErr != nil {
		return uuid.UUID{}, m.getErr
	}
	if _, ok := m.events[id]; !ok {
		return uuid.UUID{}, sql.ErrNoRows
	}
	delete(m.events, id)
	return id, nil
}

func (m *mockEventRepo) UpdateEvent(ctx context.Context, arg database.UpdateEventParams) (uuid.UUID, error) {
	if m.updateErr != nil {
		return uuid.UUID{}, m.updateErr
	}
	e, ok := m.events[arg.ID]
	if !ok {
		return uuid.UUID{}, sql.ErrNoRows
	}
	if arg.Name.Valid {
		e.Name = arg.Name.String
	}
	m.events[arg.ID] = e
	return arg.ID, nil
}

func (m *mockEventRepo) EnrollUserToEvent(ctx context.Context, arg database.EnrollUserToEventParams) (uuid.UUID, error) {
	if m.attendeeErr != nil {
		return uuid.UUID{}, m.attendeeErr
	}
	list := m.attendees[arg.EventID]
	for _, uid := range list {
		if uid == arg.UserID {
			return uuid.UUID{}, errors.New("duplicate key violation (already joined)")
		}
	}
	m.attendees[arg.EventID] = append(list, arg.UserID)
	return arg.ID, nil
}

func (m *mockEventRepo) RemoveUserFromEvent(ctx context.Context, arg database.RemoveUserFromEventParams) (uuid.UUID, error) {
	if m.attendeeErr != nil {
		return uuid.UUID{}, m.attendeeErr
	}
	list, ok := m.attendees[arg.EventID]
	if !ok {
		return uuid.UUID{}, sql.ErrNoRows
	}
	var updated []uuid.UUID
	found := false
	for _, uid := range list {
		if uid == arg.UserID {
			found = true
		} else {
			updated = append(updated, uid)
		}
	}
	if !found {
		return uuid.UUID{}, sql.ErrNoRows
	}
	m.attendees[arg.EventID] = updated
	return arg.EventID, nil
}

func (m *mockEventRepo) ListEventAttendees(ctx context.Context, eventID uuid.UUID) ([]database.ListEventAttendeesRow, error) {
	if m.attendeeErr != nil {
		return nil, m.attendeeErr
	}
	list := m.attendees[eventID]
	var rows []database.ListEventAttendeesRow
	for _, uid := range list {
		rows = append(rows, database.ListEventAttendeesRow{
			ID:   uid,
			Name: "Attendee User",
		})
	}
	return rows, nil
}

func (m *mockEventRepo) GetMyEventById(ctx context.Context, arg database.GetMyEventByIdParams) (uuid.UUID, error) {
	if m.attendeeErr != nil {
		return uuid.UUID{}, m.attendeeErr
	}
	list := m.attendees[arg.EventID]
	for _, uid := range list {
		if uid == arg.UserID {
			return arg.EventID, nil
		}
	}
	return uuid.UUID{}, sql.ErrNoRows
}

func (m *mockEventRepo) ListMyEvents(ctx context.Context, userID uuid.UUID) ([]database.ListMyEventsRow, error) {
	if m.attendeeErr != nil {
		return nil, m.attendeeErr
	}
	var res []database.ListMyEventsRow
	for eid, list := range m.attendees {
		for _, uid := range list {
			if uid == userID {
				e := m.events[eid]
				res = append(res, database.ListMyEventsRow{
					ID:   e.ID,
					Name: e.Name,
				})
			}
		}
	}
	return res, nil
}

func newMockEventRepo() *mockEventRepo {
	return &mockEventRepo{
		events:    make(map[uuid.UUID]database.Event),
		attendees: make(map[uuid.UUID][]uuid.UUID),
	}
}

func createEvent(repo *mockEventRepo, overrides ...func(*database.Event)) uuid.UUID {
	id := uuid.New()
	e := database.Event{
		ID:          id,
		Name:        "Test Event",
		Description: sql.NullString{String: "A test event", Valid: true},
		Venue:       sql.NullString{String: "Hall A", Valid: true},
		Image:       sql.NullString{String: "http://example.com/logo.png", Valid: true},
		BannerImage: sql.NullString{String: "http://example.com/banner.png", Valid: true},
		EventDate:   time.Now().Add(24 * time.Hour),
		Status:      database.ModeOfConductOffline,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	for _, o := range overrides {
		o(&e)
	}
	repo.events[id] = e
	return id
}

func TestHandleGetEventById(t *testing.T) {
	repo := newMockEventRepo()
	h := &EventHandler{Repo: repo}

	eventID := createEvent(repo)

	tests := []struct {
		name           string
		eventIDParam   string
		mockErr        error
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:           "successful get",
			eventIDParam:   eventID.String(),
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "invalid uuid",
			eventIDParam:   "bad",
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
		},
		{
			name:           "not found",
			eventIDParam:   uuid.New().String(),
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
		},
		{
			name:           "db error",
			eventIDParam:   eventID.String(),
			mockErr:        fmt.Errorf("db error"),
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo.getErr = tc.mockErr

			req := httptest.NewRequest(http.MethodGet, "/event/"+tc.eventIDParam, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.eventIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			rr := httptest.NewRecorder()

			h.HandleGetEventById(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
			var env struct{ Success bool }
			json.Unmarshal(rr.Body.Bytes(), &env)
			if env.Success != tc.expectSuccess {
				t.Errorf("expected Success=%v, got %v", tc.expectSuccess, env.Success)
			}
		})
	}
}

func TestHandleGetAllEvent(t *testing.T) {
	repo := newMockEventRepo()
	h := &EventHandler{Repo: repo}

	createEvent(repo)
	createEvent(repo)

	tests := []struct {
		name           string
		mockErr        error
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:           "successful list",
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
			repo.getErr = tc.mockErr

			req := httptest.NewRequest(http.MethodGet, "/event", nil)
			rr := httptest.NewRecorder()
			h.HandleGetAllEvent(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
			var env struct{ Success bool }
			json.Unmarshal(rr.Body.Bytes(), &env)
			if env.Success != tc.expectSuccess {
				t.Errorf("expected Success=%v, got %v", tc.expectSuccess, env.Success)
			}
		})
	}
}

func TestHandleDeleteEvent(t *testing.T) {
	repo := newMockEventRepo()
	h := &EventHandler{Repo: repo}

	eventID := createEvent(repo)

	tests := []struct {
		name           string
		eventIDParam   string
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "successful delete",
			eventIDParam:   eventID.String(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid uuid",
			eventIDParam:   "not-a-uuid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "event not found",
			eventIDParam:   uuid.New().String(),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "db error",
			eventIDParam:   eventID.String(),
			mockErr:        fmt.Errorf("db error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo.events[eventID] = repo.events[eventID]
			repo.getErr = tc.mockErr

			req := httptest.NewRequest(http.MethodDelete, "/event/"+tc.eventIDParam, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.eventIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			rr := httptest.NewRecorder()

			h.HandleDeleteEvent(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestHandleJoinEvent(t *testing.T) {
	repo := newMockEventRepo()
	h := &EventAttendeeHandler{Repo: repo}

	eventID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name           string
		eventIDParam   string
		authCtxUser    *middleware.User
		mockErr        error
		expectedStatus int
	}{
		{
			name:         "successful join",
			eventIDParam: eventID.String(),
			authCtxUser:  &middleware.User{ID: userID},
			expectedStatus: http.StatusOK,
		},
		{
			name:         "invalid event uuid",
			eventIDParam: "bad-uuid",
			authCtxUser:  &middleware.User{ID: userID},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:         "duplicate join",
			eventIDParam: eventID.String(),
			authCtxUser:  &middleware.User{ID: userID},
			mockErr:      errors.New("duplicate key"),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo.attendeeErr = tc.mockErr
			repo.attendees[eventID] = nil

			req := httptest.NewRequest(http.MethodPost, "/event/"+tc.eventIDParam+"/join", nil)
			if tc.authCtxUser != nil {
				ctx := middleware.WithUser(req.Context(), *tc.authCtxUser)
				req = req.WithContext(ctx)
			}
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.eventIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			rr := httptest.NewRecorder()

			h.HandleJoinEvent(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestHandleDeleteEventAttendee(t *testing.T) {
	repo := newMockEventRepo()
	h := &EventAttendeeHandler{Repo: repo}

	eventID := uuid.New()
	userID := uuid.New()
	repo.attendees[eventID] = []uuid.UUID{userID}

	tests := []struct {
		name           string
		eventIDParam   string
		queryUserID    string
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "successful remove",
			eventIDParam:   eventID.String(),
			queryUserID:    userID.String(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid user id",
			eventIDParam:   eventID.String(),
			queryUserID:    "bad-user",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid event id",
			eventIDParam:   "bad-event",
			queryUserID:    userID.String(),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "not found",
			eventIDParam:   eventID.String(),
			queryUserID:    uuid.New().String(),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "repo error",
			eventIDParam:   eventID.String(),
			queryUserID:    userID.String(),
			mockErr:        fmt.Errorf("db error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo.attendees[eventID] = []uuid.UUID{userID}
			repo.attendeeErr = tc.mockErr

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/event/%s/attendee/%s", tc.eventIDParam, tc.queryUserID), nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.eventIDParam)
			rctx.URLParams.Add("user_id", tc.queryUserID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			rr := httptest.NewRecorder()

			h.HandleDeleteEventAttendee(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestHandleAllEventAttendee(t *testing.T) {
	repo := newMockEventRepo()
	h := &EventAttendeeHandler{Repo: repo}

	eventID := uuid.New()
	repo.attendees[eventID] = []uuid.UUID{uuid.New(), uuid.New()}

	tests := []struct {
		name           string
		eventIDParam   string
		mockErr        error
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:           "successful list",
			eventIDParam:   eventID.String(),
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "invalid uuid",
			eventIDParam:   "bad",
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
		},
		{
			name:           "empty list",
			eventIDParam:   uuid.New().String(),
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "db error",
			eventIDParam:   eventID.String(),
			mockErr:        fmt.Errorf("db error"),
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo.attendeeErr = tc.mockErr

			req := httptest.NewRequest(http.MethodGet, "/event/"+tc.eventIDParam+"/attendees", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.eventIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			rr := httptest.NewRecorder()

			h.HandleAllEventAttendee(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
			var env struct{ Success bool }
			json.Unmarshal(rr.Body.Bytes(), &env)
			if env.Success != tc.expectSuccess {
				t.Errorf("expected Success=%v, got %v", tc.expectSuccess, env.Success)
			}
		})
	}
}

func TestHandleUpdateEvent(t *testing.T) {
	repo := newMockEventRepo()
	h := &EventHandler{Repo: repo}

	eventID := createEvent(repo)

	tests := []struct {
		name           string
		eventIDParam   string
		body           any
		mockErr        error
		expectedStatus int
	}{
		{
			name:         "successful update",
			eventIDParam: eventID.String(),
			body: model.UpdateEventRequest{
				Name: strPtr("Updated Event"),
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:         "invalid uuid",
			eventIDParam: "bad-uuid",
			body:         model.UpdateEventRequest{Name: strPtr("X")},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:         "invalid json",
			eventIDParam: eventID.String(),
			body:         "{bad-json}",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:         "not found",
			eventIDParam: uuid.New().String(),
			body:         model.UpdateEventRequest{Name: strPtr("X")},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:         "repo error",
			eventIDParam: eventID.String(),
			body:         model.UpdateEventRequest{Name: strPtr("X")},
			mockErr:      fmt.Errorf("db error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo.events[eventID] = repo.events[eventID]
			repo.updateErr = tc.mockErr

			var buf bytes.Buffer
			if s, ok := tc.body.(string); ok {
				buf.WriteString(s)
			} else {
				json.NewEncoder(&buf).Encode(tc.body)
			}

			req := httptest.NewRequest(http.MethodPatch, "/event/"+tc.eventIDParam, &buf)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.eventIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			rr := httptest.NewRecorder()

			h.HandleUpdateEvent(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestHandleGetMyEvent(t *testing.T) {
	repo := newMockEventRepo()
	h := &EventAttendeeHandler{Repo: repo}

	eventID := uuid.New()
	userID := uuid.New()
	repo.attendees[eventID] = []uuid.UUID{userID}

	tests := []struct {
		name           string
		eventIDParam   string
		authCtxUser    *middleware.User
		mockErr        error
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:         "successful get my event",
			eventIDParam: eventID.String(),
			authCtxUser:  &middleware.User{ID: userID},
			expectedStatus: http.StatusOK,
			expectSuccess: true,
		},
		{
			name:         "not attending",
			eventIDParam: uuid.New().String(),
			authCtxUser:  &middleware.User{ID: userID},
			expectedStatus: http.StatusOK,
			expectSuccess: false,
		},
		{
			name:         "invalid uuid",
			eventIDParam: "bad",
			authCtxUser:  &middleware.User{ID: userID},
			expectedStatus: http.StatusOK,
			expectSuccess: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo.attendeeErr = tc.mockErr

			req := httptest.NewRequest(http.MethodGet, "/event/u/"+tc.eventIDParam, nil)
			if tc.authCtxUser != nil {
				ctx := middleware.WithUser(req.Context(), *tc.authCtxUser)
				req = req.WithContext(ctx)
			}
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.eventIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			rr := httptest.NewRecorder()

			h.HandleGetMyEvent(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
			var env struct{ Success bool }
			json.Unmarshal(rr.Body.Bytes(), &env)
			if env.Success != tc.expectSuccess {
				t.Errorf("expected Success=%v, got %v", tc.expectSuccess, env.Success)
			}
		})
	}
}

func TestHandleGetMyAllEvent(t *testing.T) {
	repo := newMockEventRepo()
	h := &EventAttendeeHandler{Repo: repo}

	eventID := createEvent(repo)
	userID := uuid.New()
	repo.attendees[eventID] = []uuid.UUID{userID}

	tests := []struct {
		name           string
		authCtxUser    *middleware.User
		mockErr        error
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:          "successful list my events",
			authCtxUser:   &middleware.User{ID: userID},
			expectedStatus: http.StatusOK,
			expectSuccess: true,
		},
		{
			name:          "no events",
			authCtxUser:   &middleware.User{ID: uuid.New()},
			expectedStatus: http.StatusOK,
			expectSuccess: true,
		},
		{
			name:          "db error",
			authCtxUser:   &middleware.User{ID: userID},
			mockErr:       fmt.Errorf("db error"),
			expectedStatus: http.StatusOK,
			expectSuccess: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo.attendeeErr = tc.mockErr

			req := httptest.NewRequest(http.MethodGet, "/event/all", nil)
			if tc.authCtxUser != nil {
				ctx := middleware.WithUser(req.Context(), *tc.authCtxUser)
				req = req.WithContext(ctx)
			}
			rr := httptest.NewRecorder()

			h.HandleGetMyAllEvent(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
			var env struct{ Success bool }
			json.Unmarshal(rr.Body.Bytes(), &env)
			if env.Success != tc.expectSuccess {
				t.Errorf("expected Success=%v, got %v", tc.expectSuccess, env.Success)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}



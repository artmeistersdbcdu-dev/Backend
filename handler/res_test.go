package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRespondWithJson(t *testing.T) {
	tests := []struct {
		name           string
		code           int
		payload        any
		expectSuccess  bool
	}{
		{
			name:          "success response (2xx)",
			code:          http.StatusOK,
			payload:       map[string]string{"id": "123"},
			expectSuccess: true,
		},
		{
			name:          "success created",
			code:          http.StatusCreated,
			payload:       map[string]string{"id": "456"},
			expectSuccess: true,
		},
		{
			name:          "error response (4xx)",
			code:          http.StatusBadRequest,
			payload:       "bad request",
			expectSuccess: false,
		},
		{
			name:          "server error (5xx)",
			code:          http.StatusInternalServerError,
			payload:       "server error",
			expectSuccess: false,
		},
		{
			name:          "nil payload",
			code:          http.StatusOK,
			payload:       nil,
			expectSuccess: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			RespondWithJson(w, tc.code, tc.payload)

			resp := w.Result()
			if resp.StatusCode != tc.code {
				t.Errorf("expected status %d, got %d", tc.code, resp.StatusCode)
			}
			if ct := resp.Header.Get("Content-type"); ct != "Application/Json" {
				t.Errorf("expected Content-Type Application/Json, got %s", ct)
			}

			var env struct {
				Success bool
				Data    json.RawMessage
			}
			if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
			if env.Success != tc.expectSuccess {
				t.Errorf("expected Success=%v, got %v", tc.expectSuccess, env.Success)
			}
			if tc.payload == nil && string(env.Data) != "null" {
				t.Errorf("expected Data to be null, got %s", string(env.Data))
			}
		})
	}
}

func TestRespondWithJsonCustom(t *testing.T) {
	w := httptest.NewRecorder()
	RespondWithJsonCustom(w, http.StatusOK, false, map[string]string{"msg": "custom"})

	resp := w.Result()
	var env struct {
		Success bool
		Data    map[string]string
	}
	json.NewDecoder(resp.Body).Decode(&env)

	if env.Success != false {
		t.Errorf("expected Success=false, got true")
	}
	if env.Data["msg"] != "custom" {
		t.Errorf("expected msg=custom, got %v", env.Data)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRespondWithError(t *testing.T) {
	tests := []struct {
		name   string
		code   int
		msg    string
	}{
		{
			name: "bad request error",
			code: http.StatusBadRequest,
			msg:  "Invalid input",
		},
		{
			name: "not found error",
			code: http.StatusNotFound,
			msg:  "Resource not found",
		},
		{
			name: "internal error",
			code: http.StatusInternalServerError,
			msg:  "Something went wrong",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			RespondWithError(w, tc.code, tc.msg)

			resp := w.Result()
			if resp.StatusCode != tc.code {
				t.Errorf("expected %d, got %d", tc.code, resp.StatusCode)
			}

			var env struct {
				Success bool
				Data    struct {
					Msg string `json:"Error"`
				}
			}
			json.NewDecoder(resp.Body).Decode(&env)

			if env.Success {
				t.Errorf("expected Success=false for error response, got true")
			}
			if env.Data.Msg != tc.msg {
				t.Errorf("expected Error=%q, got %q", tc.msg, env.Data.Msg)
			}
		})
	}
}

func TestHealth(t *testing.T) {
	w := httptest.NewRecorder()
	Health(w, nil)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var env struct {
		Success bool
		Data    struct {
			Health string
		}
	}
	json.NewDecoder(resp.Body).Decode(&env)

	if !env.Success {
		t.Errorf("expected Success=true")
	}
	if env.Data.Health != "ok" {
		t.Errorf("expected Health=ok, got %s", env.Data.Health)
	}
}

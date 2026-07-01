package admin

import (
	"fmt"
	"net/http"

	"github.com/Blue-Onion/ArtmeisterBackend/handler"
	"github.com/Blue-Onion/ArtmeisterBackend/handler/logger"
	"github.com/Blue-Onion/ArtmeisterBackend/internal/database"
	"github.com/Blue-Onion/ArtmeisterBackend/middleware"
	"github.com/Blue-Onion/ArtmeisterBackend/utlis"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type UserHandler struct {
	Repo database.UserRepository
}
type ArtHandler struct {
	Repo database.ArtRepository
}

func (h *ArtHandler) HandlerArtStatus(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	_, ok := middleware.GetModerator(r.Context())
	if !ok {
		handler.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	artId := chi.URLParam(r, "art_id")
	id, err := uuid.Parse(artId)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandlerArtStatus: invalid art ID format '%s': %v", artId, err))
		}
		handler.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	status := r.URL.Query().Get("status")
	if status != string(database.ArtStatusApproved) && status != string(database.ArtStatusRejected) && status != string(database.AccountStatusPending) {
		if log != nil {
			log.Error(fmt.Sprintf("HandlerArtStatus: invalid status provided '%s'", status))
		}
		handler.RespondWithError(w, http.StatusBadRequest, "Invalid status provided")
		return
	}
	params := database.UpdateArtStatusParams{
		ID:     id,
		Status: database.ArtStatus(status),
	}
	art, err := h.Repo.UpdateArtStatus(r.Context(), params)
	if err != nil {
		if utlis.IsNotFound(err) {
			handler.RespondWithError(w, http.StatusNotFound, "Art not found")
			return
		}
		if log != nil {
			log.Error(fmt.Sprintf("HandlerArtStatus: failed to update art %s status: %v", id, err))
		}
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to update art status")
		return
	}
	if log != nil {
		log.Info(fmt.Sprintf("HandlerArtStatus: art %s status updated to %s", id, status))
	}
	handler.RespondWithJson(w, http.StatusOK, art)
}
func (h *UserHandler) HandlerRole(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	actor, ok := middleware.GetSenior(r.Context())
	if !ok {
		handler.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID := chi.URLParam(r, "user_id")
	id, err := uuid.Parse(userID)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandlerRole: invalid user ID format '%s': %v", userID, err))
		}
		handler.RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	role := r.URL.Query().Get("role")
	status := r.URL.Query().Get("status")

	if (role == "" && status == "") || (role != "" && status != "") {
		handler.RespondWithError(
			w,
			http.StatusBadRequest,
			"Provide either role or status, but not both",
		)
		return
	}

	if role != "" {
		if !utlis.IsValidUserRole(role) {
			handler.RespondWithError(w, http.StatusBadRequest, "Invalid role value")
			return
		}
		targetRole := database.UserRole(role)
		if !utlis.CanAssignRole(actor.Role, targetRole) {
			handler.RespondWithError(w, http.StatusForbidden, "You cannot assign this role")
			return
		}
	}

	params := database.PatchUserAdminParams{
		ID: id,
	}

	if status != "" {
		params.Status = database.NullAccountStatus{
			AccountStatus: database.AccountStatus(status),
			Valid:         true,
		}
	}

	if role != "" {
		params.Role = database.NullUserRole{
			UserRole: database.UserRole(role),
			Valid:    true,
		}
	}
	user, err := h.Repo.PatchUserAdmin(r.Context(), params)
	if err != nil {
		if utlis.IsNotFound(err) {
			handler.RespondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		if log != nil {
			log.Error(fmt.Sprintf("HandlerRole: failed updating user %s: %v", id, err))
		}
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
		return
	}
	if log != nil {
		log.Info(fmt.Sprintf("HandlerRole: updated user %s", id))
	}

	handler.RespondWithJson(w, http.StatusOK, user)
}

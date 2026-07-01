package artmetadata

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Blue-Onion/ArtmeisterBackend/handler"
	"github.com/Blue-Onion/ArtmeisterBackend/handler/logger"
	"github.com/Blue-Onion/ArtmeisterBackend/internal/database"
	"github.com/Blue-Onion/ArtmeisterBackend/middleware"
	"github.com/Blue-Onion/ArtmeisterBackend/model"
	"github.com/Blue-Onion/ArtmeisterBackend/utlis"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type Handler struct {
	Repo database.ArtMetaDataRepository
}

func (h *Handler) HandleGetArtComments(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	id := chi.URLParam(r, "id")
	artId, err := uuid.Parse(id)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleGetArtComments: invalid art ID format '%s': %v", id, err))
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	comments, err := h.Repo.GetArtCommentsByArtID(r.Context(), artId)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleGetArtComments: failed to get comments for art %s: %v", artId, err))
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	if log != nil {
		log.Info(fmt.Sprintf("HandleGetArtComments: retrieved comments for art %s successfully", artId))
	}
	handler.RespondWithJson(w, http.StatusOK, comments)
}
func (h *Handler) HandleGetArtCommentsCount(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	id := chi.URLParam(r, "id")
	artId, err := uuid.Parse(id)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleGetArtCommentsCount: invalid art ID format '%s': %v", id, err))
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	commentsCount, err := h.Repo.GetArtCommentsCount(r.Context(), artId)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleGetArtCommentsCount: failed for art %s: %v", artId, err))
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	if log != nil {
		log.Info(fmt.Sprintf("HandleGetArtCommentsCount: retrieved comments count for art %s successfully", artId))
	}
	handler.RespondWithJson(w, http.StatusOK, commentsCount)
}
func (h *Handler) HandleGetArtLikeCount(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	id := chi.URLParam(r, "id")
	artId, err := uuid.Parse(id)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleGetArtLikeCount: invalid art ID format '%s': %v", id, err))
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	likeCount, err := h.Repo.GetArtLikesCount(r.Context(), artId)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleGetArtLikeCount: failed for art %s: %v", artId, err))
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	if log != nil {
		log.Info(fmt.Sprintf("HandleGetArtLikeCount: retrieved likes count for art %s successfully", artId))
	}
	handler.RespondWithJson(w, http.StatusOK, likeCount)
}
func (h *Handler) HandleDeleteComment(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		if log != nil {
			log.Error("HandleDeleteComment: unauthenticated request")
		}
		handler.RespondWithError(w, 401, "Not Authorized")
		return
	}
	userId := user.ID
	id := chi.URLParam(r, "id")
	commentId, err := uuid.Parse(id)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleDeleteComment: invalid comment ID format '%s' for user %s: %v", id, userId, err))
		}
		handler.RespondWithError(w, 400, err.Error())
		return
	}
	param := database.DeleteArtCommentParams{
		UserID: userId,
		ID:     commentId,
	}
	_, err = h.Repo.DeleteArtComment(r.Context(), param)
	if err != nil {
		if utlis.IsNotFound(err) {
			handler.RespondWithError(w, http.StatusNotFound, "Comment not found")
			return
		}
		if log != nil {
			log.Error(fmt.Sprintf("HandleDeleteComment: failed to delete comment %s: %v", commentId, err))
		}
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to delete comment")
		return
	}
	if log != nil {
		log.Info(fmt.Sprintf("HandleDeleteComment: comment %s deleted by user %s", commentId, userId))
	}
	handler.RespondWithJson(w, http.StatusOK, "Deleted Successfully")
}
func (h *Handler) HandleComment(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		if log != nil {
			log.Error("HandleComment: unauthenticated request")
		}
		handler.RespondWithError(w, 401, "Not Authorized")
		return
	}
	userId := user.ID
	id := chi.URLParam(r, "art_id")
	artId, err := uuid.Parse(id)

	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleComment: invalid art ID format '%s' for user %s: %v", id, userId, err))
		}
		handler.RespondWithError(w, 400, err.Error())
		return
	}
	req := model.AddComment{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&req)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleComment: failed to decode request body for art %s, user %s: %v", artId, userId, err))
		}
		handler.RespondWithError(w, 400, err.Error())
		return
	}
	param := database.AddArtCommentParams{
		Comment: req.Comment,
		UserID:  userId,
		ArtID:   artId,
	}
	comment, err := h.Repo.AddArtComment(r.Context(), param)

	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleComment: failed to add comment on art %s by user %s: %v", artId, userId, err))
		}
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to add comment")
		return
	}
	if log != nil {
		log.Info(fmt.Sprintf("HandleComment: comment added on art %s by user %s", artId, userId))
	}
	handler.RespondWithJson(w, http.StatusOK, comment)
}
func (h *Handler) HandleLike(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		if log != nil {
			log.Error("HandleLike: unauthenticated request")
		}
		handler.RespondWithError(w, 401, "Not Authorized")
		return
	}
	userId := user.ID
	id := chi.URLParam(r, "art_id")
	artId, err := uuid.Parse(id)

	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleLike: invalid art ID format '%s' for user %s: %v", id, userId, err))
		}
		handler.RespondWithError(w, 400, err.Error())
		return
	}
	param := database.LikeArtParams{
		UserID: userId,
		ArtID:  artId,
	}
	comment, err := h.Repo.LikeArt(r.Context(), param)

	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleLike: failed to like art %s by user %s: %v", artId, userId, err))
		}
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to like art")
		return
	}
	if log != nil {
		log.Info(fmt.Sprintf("HandleLike: art %s liked by user %s", artId, userId))
	}
	handler.RespondWithJson(w, http.StatusOK, comment)
}
func (h *Handler) HandleUnLike(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		if log != nil {
			log.Error("HandleUnLike: unauthenticated request")
		}
		handler.RespondWithError(w, 401, "Not Authorized")
		return
	}
	userId := user.ID
	id := chi.URLParam(r, "art_id")
	artId, err := uuid.Parse(id)

	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleUnLike: invalid art ID format '%s' for user %s: %v", id, userId, err))
		}
		handler.RespondWithError(w, 400, err.Error())
		return
	}
	param := database.UnlikeArtParams{
		UserID: userId,
		ArtID:  artId,
	}
	_, err = h.Repo.UnlikeArt(r.Context(), param)

	if err != nil {
		if utlis.IsNotFound(err) {
			handler.RespondWithError(w, http.StatusNotFound, "Like not found")
			return
		}
		if log != nil {
			log.Error(fmt.Sprintf("HandleUnLike: failed to unlike art %s by user %s: %v", artId, userId, err))
		}
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to unlike art")
		return
	}
	if log != nil {
		log.Info(fmt.Sprintf("HandleUnLike: art %s unliked by user %s", artId, userId))
	}
	handler.RespondWithJson(w, http.StatusOK, "Ok")
}

package art

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Blue-Onion/ArtmeisterBackend/cache"
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
	Repo      database.ArtRepository
	EventRepo database.EventRepository
	UserRepo  database.UserRepository
	Cache     *cache.Cache
}
type ProfileHandler struct {
	ArtRepo  database.ArtRepository
	UserRepo database.UserRepository
	Cache    *cache.Cache
}

type profile struct {
	User database.GetUserRow
	Art  []database.GetArtByUserRow
}

// TODO(Cache): SetArt after creation. Invalidate art list caches (approved, user's gallery).
func (h *Handler) HandleArtCreation(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()

	user, ok := middleware.GetUser(r.Context())
	if !ok {
		if log != nil {
			log.Error("HandleArtCreation: unauthenticated request")
		}
		handler.RespondWithError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	if user.Status != database.AccountStatusApproved {
		if log != nil {
			log.Error(fmt.Sprintf(
				"HandleArtCreation: user %s (%s) is not approved",
				user.ID,
				user.Name,
			))
		}
		handler.RespondWithError(w, http.StatusBadRequest, "User not approved")
		return
	}

	req := model.CreateArtRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf(
				"HandleArtCreation: invalid JSON body from user %s: %v",
				user.ID,
				err,
			))
		}
		handler.RespondWithError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if len(req.Name) < 3 {
		if log != nil {
			log.Error(fmt.Sprintf(
				"HandleArtCreation: art name too short for user %s: '%s'",
				user.ID,
				req.Name,
			))
		}
		handler.RespondWithError(w, http.StatusBadRequest, "Name is too short")
		return
	}

	if len(req.URL) < 3 {
		if log != nil {
			log.Error(fmt.Sprintf(
				"HandleArtCreation: invalid URL for user %s: '%s'",
				user.ID,
				req.URL,
			))
		}
		handler.RespondWithError(w, http.StatusBadRequest, "Invalid URL")
		return
	}

	if req.Tags == nil {
		req.Tags = []string{}
	}

	id := uuid.New()

	params := database.CreateArtParams{
		ID:          id,
		Name:        req.Name,
		Description: utlis.ToNilStr(req.Description),
		Tags:        req.Tags,
		UserID:      user.ID,
		Image:       req.URL,
	}
	if log != nil {
		log.Info(fmt.Sprintf(
			"HandleArtCreation: creating art %s for user %s",
			id,
			user.ID,
		))
	}

	artID, err := h.Repo.CreateArt(r.Context(), params)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf(
				"HandleArtCreation: failed to create art %s for user %s: %v",
				id,
				user.ID,
				err,
			))
		}
		handler.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if log != nil {
		log.Info(fmt.Sprintf(
			"HandleArtCreation: art %s created successfully by user %s",
			id,
			user.ID,
		))
	}

	artCache := cache.Art{
		Name:        params.Name,
		Description: params.Description,
		Tags:        params.Tags,
		UserID:      user.ID,
		Image:       params.Image,
	}
	h.Cache.SetArt(id, artCache)
	h.Cache.DeleteList("approved_arts")
	h.Cache.DeleteList("homepage")
	h.Cache.DeleteList("user_arts:" + user.ID.String())
	handler.RespondWithJson(w, http.StatusOK, map[string]string{"ID": artID.String()})
}
// TODO(Cache): Cache arts list by user_id UUID.
// Cache Miss -> DB GetArtByUser, Set cache.
// Cache Hit -> Return cached list.
// Invalidate when art is created, updated, or deleted for this user.
func (h *Handler) HandleGetArts(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	userId := chi.URLParam(r, "user_id")
	if userId == "" {
		if log != nil {
			log.Error("HandleGetArts: user ID is empty")
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	id, err := uuid.Parse(userId)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("userId=%q err=%v", userId, err))
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	cacheKey := "user_arts:" + id.String()
	cached := h.Cache.GetList(cacheKey)
	if cached != nil {
		fmt.Println("Cache BABY")
		handler.RespondWithJson(w, http.StatusOK, cached)
		return
	}
	arts, err := h.Repo.GetArtByUser(r.Context(), id)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleGetArts: failed to get arts for user %s: %v", id, err))
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	if log != nil {
		log.Info(fmt.Sprintf("HandleGetArts: retrieved arts for user %s successfully", id))
	}
	h.Cache.SetList(cacheKey, arts)
	handler.RespondWithJson(w, http.StatusOK, arts)
}
// TODO(Cache): Cache Art by UUID. cache.Art struct exists but is never read.
// Cache Hit -> Return cached art. Cache Miss -> DB GetArtByID, SetArt.
// Invalidate cache on PATCH or DELETE.
func (h *Handler) HandleGetArtById(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	artid := chi.URLParam(r, "id")
	if artid == "" {
		if log != nil {
			log.Error("HandleGetArts: user ID is empty")
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}

	id, err := uuid.Parse(artid)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("userId=%q err=%v", artid, err))
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	artCache := h.Cache.GetArt(id)
	if artCache != nil {
		fmt.Println("Cache BABY")
		art := database.GetArtByIDRow{
			ID:          id,
			Name:        artCache.Name,
			Description: artCache.Description,
			Image:       artCache.Image,
			Tags:        artCache.Tags,
		}
		handler.RespondWithJson(w, http.StatusOK, art)
		return
	}
	arts, err := h.Repo.GetArtByID(r.Context(), id)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleGetArts: failed to get arts for user %s: %v", id, err))
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	if log != nil {
		log.Info(fmt.Sprintf("HandleGetArts: retrieved arts for user %s successfully", id))
	}
	newCache := cache.Art{
		Name:        arts.Name,
		Description: arts.Description,
		Image:       arts.Image,
		Tags:        arts.Tags,
	}
	h.Cache.SetArt(id, newCache)
	handler.RespondWithJson(w, http.StatusOK, arts)
}

// TODO(Cache): Cache art profile by composite key (art_id + user_id).
// Cache Miss -> DB GetArtProfileByID, Set cache.
// Cache Hit -> Return cached joined result.
// Invalidate when art is updated or deleted.
func (h *Handler) HandleGetArtProfileById(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	Id := chi.URLParam(r, "id")
	usrId := chi.URLParam(r, "user_id")
	if Id == "" {
		if log != nil {
			log.Error("HandleGetArtById: art ID is empty")
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	artId, err := uuid.Parse(Id)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleGetArtById: invalid art ID format '%s': %v", Id, err))
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	userId, err := uuid.Parse(usrId)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleGetArtById: invalid user ID format '%s': %v", usrId, err))
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	cacheKey := "art_profile:" + artId.String() + ":" + userId.String()
	cached := h.Cache.GetList(cacheKey)
	if cached != nil {
		fmt.Println("Cache BABY")
		handler.RespondWithJson(w, http.StatusOK, cached)
		return
	}
	params := database.GetArtProfileByIDParams{
		ID:     artId,
		UserID: userId,
	}
	art, err := h.Repo.GetArtProfileByID(r.Context(), params)

	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleGetArtById: failed to get art %s: %v", artId, err))
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	if log != nil {
		log.Info(fmt.Sprintf("HandleGetArtById: retrieved art %s successfully", artId))
	}
	h.Cache.SetList(cacheKey, art)
	handler.RespondWithJson(w, http.StatusOK, art)
}
// TODO(Cache): DeleteArt from cache. Invalidate art list caches (approved, user's gallery).
func (h *Handler) HandleArtDeletion(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		if log != nil {
			log.Error("HandleArtDeletion: unauthenticated request")
		}
		handler.RespondWithError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	id := chi.URLParam(r, "id")
	if id == "" {
		if log != nil {
			log.Error(fmt.Sprintf("HandleArtDeletion: art ID is empty (user %s)", user.ID))
		}
		handler.RespondWithError(w, http.StatusBadRequest, "Art ID is required")
		return
	}
	artId, err := uuid.Parse(id)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleArtDeletion: invalid art ID format '%s' for user %s: %v", id, user.ID, err))
		}
		handler.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	param := database.DeleteArtParams{
		ID:     artId,
		UserID: user.ID,
	}
	_, err = h.Repo.DeleteArt(r.Context(), param)
	if err != nil {
		if utlis.IsNotFound(err) {
			handler.RespondWithError(w, http.StatusNotFound, "Art not found")
			return
		}
		if log != nil {
			log.Error(fmt.Sprintf("HandleArtDeletion: failed to delete art %s: %v", artId, err))
		}
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to delete art")
		return
	}
	if log != nil {
		log.Info(fmt.Sprintf("HandleArtDeletion: art %s deleted by user %s", artId, user.ID))
	}
	h.Cache.DeleteArt(artId)
	h.Cache.DeleteList("approved_arts")
	h.Cache.DeleteList("homepage")
	h.Cache.DeleteList("user_arts:" + user.ID.String())
	handler.RespondWithJson(w, http.StatusOK, "Art Work Deleted")

}
// TODO(Cache): UpdateArt in cache after successful PATCH. Invalidate art list caches.
func (h *Handler) HandlerArtUpdation(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		if log != nil {
			log.Error("HandlerArtUpdation: unauthenticated request")
		}
		handler.RespondWithError(w, http.StatusUnauthorized, "Not Authorized")
		return
	}
	id := chi.URLParam(r, "id")
	if id == "" {
		if log != nil {
			log.Error(fmt.Sprintf("HandlerArtUpdation: art ID is empty (user %s)", user.ID))
		}
		handler.RespondWithError(w, http.StatusBadRequest, "Art ID is required")
		return
	}
	artId, err := uuid.Parse(id)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandlerArtUpdation: invalid art ID format '%s' for user %s: %v", id, user.ID, err))
		}
		handler.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	params := database.UpdateArtParams{
		ID:     artId,
		UserID: user.ID,
	}
	req := model.UpdateArtRequest{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&req)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandlerArtUpdation: invalid req format '%s' for user %s: %v", id, user.ID, err))
		}
		handler.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	params.Name = utlis.ToNilStr(req.Name)
	params.Description = utlis.ToNilStr(req.Description)
	if req.Tags != nil {
		params.Tags = *req.Tags
	} else {
		params.Tags = nil

	}

	updatedWork, err := h.Repo.UpdateArt(r.Context(), params)
	if err != nil {
		if utlis.IsNotFound(err) {
			handler.RespondWithError(w, http.StatusNotFound, "Art not found")
			return
		}
		if log != nil {
			log.Error(fmt.Sprintf("HandlerArtUpdation: failed to update art %s: %v", artId, err))
		}
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to update art")
		return
	}
	if log != nil {
		log.Info(fmt.Sprintf("HandlerArtUpdation: art %s updated by user %s", artId, user.ID))
	}
	h.Cache.DeleteArt(artId)
	h.Cache.DeleteList("approved_arts")
	h.Cache.DeleteList("homepage")
	h.Cache.DeleteList("user_arts:" + user.ID.String())
	handler.RespondWithJson(w, http.StatusOK, map[string]string{"ID": updatedWork.String()})
}
// TODO(Cache): Cache composite profile (User + Arts) by user_id UUID.
// Cache Miss -> DB GetUser + GetArtByUser (2 calls), Set cache.
// Cache Hit -> Return cached composite.
// Invalidate when user profile or their arts change.
func (h *ProfileHandler) HandlerGetArtistProfile(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	userId := chi.URLParam(r, "id")
	if userId == "" {
		if log != nil {
			log.Error("HandleGetArts: user ID is empty")
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	Id, err := uuid.Parse(userId)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleGetArtById: invalid art ID format '%s': %v", Id, err))
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	cacheKey := "artist_profile:" + Id.String()
	cached := h.Cache.GetList(cacheKey)
	if cached != nil {
		fmt.Println("Cache BABY")
		handler.RespondWithJson(w, 200, cached)
		return
	}
	user, err := h.UserRepo.GetUser(r.Context(), Id)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleGetArtById: invalid art ID format '%s': %v", Id, err))
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	artWork, err := h.ArtRepo.GetArtByUser(r.Context(), Id)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleGetArtById: invalid art ID format '%s': %v", Id, err))
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	res := profile{
		User: user,
		Art:  artWork,
	}
	h.Cache.SetList(cacheKey, res)
	handler.RespondWithJson(w, 200, res)
}
// TODO(Cache): Low value (admin T5). Skip cache or use short TTL.
func (h *Handler) HandleGetPendingArt(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	arts, err := h.Repo.ListPendingArt(r.Context())
	if err != nil {
		if log != nil {
			log.Error("HandleGetPendingArt: err Occurred")
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	if log != nil {
		log.Info("HandleGetPendingArt: retrieved art")
	}
	handler.RespondWithJson(w, http.StatusOK, arts)
}

func (h *Handler) HandleLatestArt(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	cached := h.Cache.GetList("homepage")
	if cached != nil {
		data := cached.(model.HomepageData)
		handler.RespondWithJson(w, http.StatusOK, data.LatestArt)
		return
	}
	arts, err := h.Repo.ListLatestArt(r.Context())
	if err != nil {
		if log != nil {
			log.Error("HandleLatestArt: failed to get latest art")
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	events, err := h.EventRepo.ListEvents(r.Context())
	if err != nil {
		if log != nil {
			log.Error("HandleLatestArt: failed to get events")
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	coreMembers, err := h.UserRepo.GetCoreMembers(r.Context())
	if err != nil {
		if log != nil {
			log.Error("HandleLatestArt: failed to get core members")
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	if log != nil {
		log.Info("HandleLatestArt: retrieved homepage data")
	}
	data := model.HomepageData{
		LatestArt:   arts,
		Events:      events,
		CoreMembers: coreMembers,
	}
	h.Cache.SetList("homepage", data)
	handler.RespondWithJson(w, http.StatusOK, arts)
}

func (h *Handler) HandleGetApprovedArt(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	cached := h.Cache.GetList("approved_arts")
	if cached != nil {
		fmt.Println("Cache BABY")
		handler.RespondWithJson(w, http.StatusOK, cached)
		return
	}
	arts, err := h.Repo.ListArt(r.Context())
	if err != nil {
		if log != nil {
			log.Error("HandleGetPendingArt: err Occurred")
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	if log != nil {
		log.Info("HandleGetPendingArt: retrieved art")
	}
	h.Cache.SetList("approved_arts", arts)
	handler.RespondWithJson(w, http.StatusOK, arts)
}

package user

import (
	"database/sql"
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
	"github.com/sqlc-dev/pqtype"
)

type Handler struct {
	Repo  database.UserRepository
	Cache *cache.Cache
}

// TODO(Cache): Cache User by UUID (already reads cache. SetUser on cache miss).
// Update cache after successful PATCH. Delete cache after role/status change.
func (h *Handler) HandleGetUserById(w http.ResponseWriter, r *http.Request) {

	log, _ := logger.GetLogger()
	id := chi.URLParam(r, "id")
	userId, err := uuid.Parse(id)

	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleGetArts: invalid user ID format '%s': %v", userId, err))
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	userCache := h.Cache.GetUser(userId)
	if userCache != nil {
		fmt.Println("Cache BABY")
		user := database.GetUserRow{
			ID:          userId,
			Name:        userCache.Name,
			Username:    userCache.Username,
			Email:       userCache.Email,
			Batch:       userCache.Batch,
			Status:      database.AccountStatus(userCache.Status),
			Role:        database.UserRole(userCache.Role),
			Image:       userCache.Image,
			BannerImage: userCache.BannerImage,
			Description: userCache.Description,
			SocialLinks: userCache.SocialLinks,
		}

		handler.RespondWithJson(w, http.StatusOK, user)
		return

	}
	user, err := h.Repo.GetUser(r.Context(), userId)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleGetUserById: failed to get art %s: %v", userId, err))
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	if log != nil {
		log.Info(fmt.Sprintf("HandleGetArtById: retrieved art %s successfully", userId))
	}
	newCache := cache.User{
		Name:        user.Name,
		Description: user.Description,
		Image:       user.Image,
		BannerImage: user.BannerImage,
		SocialLinks: user.SocialLinks,
		Username:    user.Username,
		Batch:       user.Batch,
		Status:      string(user.Status),
		Role:        string(user.Role),
	}
	h.Cache.SetUser(userId, newCache)
	handler.RespondWithJson(w, http.StatusOK, user)

}

// TODO(Cache): Bypass cache (auth, includes password). Cache could be seeded after login to warm user cache.
func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	params := model.AuthenticateUser{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		if log != nil {
			log.Error("HandleLogin: invalid request body")
		}
		handler.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	user, err := h.Repo.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		if utlis.IsNotFound(err) {
			if log != nil {
				log.Info(fmt.Sprintf("HandleLogin: no account found for email %s", params.Email))
			}
			handler.RespondWithError(w, http.StatusNotFound, "No account found with this email")
			return
		}
		if log != nil {
			log.Error(fmt.Sprintf("HandleLogin: db error looking up email %s: %v", params.Email, err))
		}
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to look up user")
		return
	}
	isValid := utlis.CheckPassword(user.Password, params.Password)
	if !isValid {
		if log != nil {
			log.Info(fmt.Sprintf("HandleLogin: incorrect password for user %s", user.ID))
		}
		handler.RespondWithError(w, http.StatusUnauthorized, "Incorrect password")
		return
	}
	token, err := utlis.GenerateJwt(user.ID)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleLogin: failed to generate JWT for user %s: %v", user.ID, err))
		}
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to generate authentication token")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "authToken",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		MaxAge:   3600 * 24,
		SameSite: http.SameSiteNoneMode,
	})
	if log != nil {
		log.Info(fmt.Sprintf("HandleLogin: user %s logged in successfully", user.ID))
	}
	// Seed user cache after successful login
	fullUser, err := h.Repo.GetUser(r.Context(), user.ID)
	if err == nil {
		newCache := cache.User{
			Name:        fullUser.Name,
			Description: fullUser.Description,
			Image:       fullUser.Image,
			BannerImage: fullUser.BannerImage,
			SocialLinks: fullUser.SocialLinks,
			Username:    fullUser.Username,
			Batch:       fullUser.Batch,
			Status:      string(fullUser.Status),
			Role:        string(fullUser.Role),
		}
		h.Cache.SetUser(user.ID, newCache)
	}
	handler.RespondWithJson(w, http.StatusOK, map[string]string{
		"id":    user.ID.String(),
		"name":  user.Name,
		"email": user.Email,
		"image": user.Image.String,
	})
}

// TODO(Cache): Cache User by UUID (already reads cache via GetUser). SetUser on cache miss.
func (h *Handler) HandleMe(w http.ResponseWriter, r *http.Request) {
	tokenCookie, err := r.Cookie("authToken")
	if err != nil {
		if err == http.ErrNoCookie {

			handler.RespondWithJson(w, http.StatusOK, nil)
			return
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	userId, err := utlis.GetUserIdJwt(tokenCookie)
	if err != nil {
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}

	id, err := uuid.Parse(userId)
	if err != nil {
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}

	userCache := h.Cache.GetUser(id)
	if userCache != nil {
		fmt.Println("Cache BABY")
		user := database.GetUserRow{
			ID:          id,
			Name:        userCache.Name,
			Username:    userCache.Username,
			Email:       userCache.Email,
			Batch:       userCache.Batch,
			Status:      database.AccountStatus(userCache.Status),
			Role:        database.UserRole(userCache.Role),
			Image:       userCache.Image,
			BannerImage: userCache.BannerImage,
			Description: userCache.Description,
			SocialLinks: userCache.SocialLinks,
		}

		handler.RespondWithJson(w, http.StatusOK, user)
		return

	}
	user, err := h.Repo.GetUser(r.Context(), id)
	if err != nil {
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	newCache := cache.User{
		Name:        user.Name,
		Description: user.Description,
		Image:       user.Image,
		BannerImage: user.BannerImage,
		SocialLinks: user.SocialLinks,
		Username:    user.Username,
		Batch:       user.Batch,
		Status:      string(user.Status),
		Role:        string(user.Role),
	}
	h.Cache.SetUser(id, newCache)
	handler.RespondWithJson(w, 200, user)

}

// TODO(Cache): No DB query. Skip cache.
func (h *Handler) HandleLogOut(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	http.SetCookie(w, &http.Cookie{
		Name:     "authToken",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})
	if log != nil {
		log.Info("HandleLogOut: user logged out")
	}
	handler.RespondWithJson(w, http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}

// TODO(Cache): Invalidate user list caches (approved users, core members) after creation.
func (h *Handler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	param := model.CreateUser{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&param)
	if err != nil {
		if log != nil {
			log.Error("HandleCreateUser: invalid request body")
		}
		handler.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	hashPass, err := utlis.HashPassword(param.Password)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleCreateUser: failed to hash password for %s: %v", param.Email, err))
		}
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to process password")
		return
	}
	userParam := database.CreateUserParams{
		Name:     param.Name,
		Email:    param.Email,
		Password: hashPass,
		Batch:    sql.NullString{String: "", Valid: true},
	}
	fmt.Println(userParam)
	userID, err := h.Repo.CreateUser(r.Context(), userParam)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleCreateUser: failed to create user %s: %v", param.Email, err))
		}
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}
	if log != nil {
		log.Info(fmt.Sprintf("HandleCreateUser: user created with email %s", param.Email))
	}
	token, err := utlis.GenerateJwt(userID)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleLogin: failed to generate JWT for user %s: %v", userID, err))
		}
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to generate authentication token")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "authToken",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   3600 * 24,
	})

	h.Cache.DeleteList("approved_users")
	handler.RespondWithJson(w, http.StatusCreated, map[string]string{"ID": userID.String()})
}

// TODO(Cache): UpdateUser in cache after successful PATCH. Invalidate user list caches (approved, core).
func (h *Handler) HandleUpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		if log != nil {
			log.Error("HandleUpdateUserProfile: unauthenticated request")
		}
		handler.RespondWithError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	id := chi.URLParam(r, "id")
	userId, err := uuid.Parse(id)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleUpdateUserProfile: invalid user ID format '%s': %v", id, err))
		}
		handler.RespondWithError(w, http.StatusBadRequest, "Invalid user ID format")
		return
	}
	if user.Role != database.UserRolePresident && user.ID != userId {
		if log != nil {
			log.Error(fmt.Sprintf("HandleUpdateUserProfile: user %s unauthorized to update profile %s", user.ID, userId))
		}
		handler.RespondWithError(w, http.StatusForbidden, "You are not authorized to update this profile")
		return
	}
	req := model.PatchUserProfileRequest{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&req)
	fmt.Println(&req)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleUpdateUserProfile: failed to decode request body: %v", err))
		}
		handler.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	params := database.PatchUserProfileParams{
		ID:          userId,
		Username:    utlis.ToNilStr(req.UserName),
		Email:       utlis.ToNilStr(req.Email),
		Image:       utlis.ToNilStr(req.Image),
		BannerImage: utlis.ToNilStr(req.Banner_image),
		Batch:       utlis.ToNilStr(req.Batch),

		Description: utlis.ToNilStr(req.Desc),
	}
	if req.Social != nil {
		params.SocialLinks = pqtype.NullRawMessage{
			RawMessage: *req.Social,
			Valid:      true,
		}
	}
	updatedUser, err := h.Repo.PatchUserProfile(r.Context(), params)

	if err != nil {
		if utlis.IsNotFound(err) {
			handler.RespondWithError(w, http.StatusNotFound, "User profile not found")
			return
		}
		fmt.Printf("HandleUpdateUserProfile ERROR: %v\n", err)
		if log != nil {
			log.Error(fmt.Sprintf("HandleUpdateUserProfile: failed to update profile for user %s: %v", userId, err))
		}
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to update user profile")
		return
	}
	if log != nil {
		log.Info(fmt.Sprintf("HandleUpdateUserProfile: profile updated for user %s", userId))
	}
	newCache := cache.User{
		Name:        updatedUser.Name,
		Description: updatedUser.Description,
		Image:       updatedUser.Image,
		BannerImage: updatedUser.BannerImage,
		SocialLinks: updatedUser.SocialLinks,
		Username:    updatedUser.Username,
		Batch:       updatedUser.Batch,
		Status:      string(updatedUser.Status),
		Role:        string(updatedUser.Role),
	}
	h.Cache.SetUser(userId, newCache)
	h.Cache.DeleteList("approved_users")
	handler.RespondWithJson(w, http.StatusOK, updatedUser)
}

// TODO(Cache): Low value. Could invalidate user cache, but password is not cached.
func (h *Handler) HandlePasswordChange(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		if log != nil {
			log.Error("HandlePasswordChange: unauthenticated request")
		}
		handler.RespondWithError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	req := model.PatchUserPassword{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandlePasswordChange: failed to decode request body for user %s: %v", user.ID, err))
		}
		handler.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	dbUser, err := h.Repo.GetUserByEmail(r.Context(), user.Email)
	if err != nil {
		if utlis.IsNotFound(err) {
			if log != nil {
				log.Info(fmt.Sprintf("HandlePasswordChange: user email %s not found in DB: %v", user.Email, err))
			}
			handler.RespondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		if log != nil {
			log.Error(fmt.Sprintf("HandlePasswordChange: DB error looking up email %s: %v", user.Email, err))
		}
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to verify user")
		return
	}
	if !utlis.CheckPassword(dbUser.Password, req.OldPassword) {
		if log != nil {
			log.Info(fmt.Sprintf("HandlePasswordChange: incorrect old password for user %s", user.ID))
		}
		handler.RespondWithError(w, http.StatusUnauthorized, "Current password is incorrect")
		return
	}

	password, err := utlis.HashPassword(req.Password)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandlePasswordChange: failed to hash new password for user %s: %v", user.ID, err))
		}
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to process new password")
		return
	}
	params := database.PatchUserPasswordParams{
		ID:       user.ID,
		Password: password,
	}
	res, err := h.Repo.PatchUserPassword(r.Context(), params)
	if err != nil {
		if utlis.IsNotFound(err) {
			handler.RespondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		if log != nil {
			log.Error(fmt.Sprintf("HandlePasswordChange: failed to update password for user %s: %v", user.ID, err))
		}
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to update password")
		return
	}
	if log != nil {
		log.Info(fmt.Sprintf("HandlePasswordChange: password updated for user %s", user.ID))
	}
	handler.RespondWithJson(w, http.StatusOK, res)
}

// TODO(Cache): Low value (admin T5). Skip cache or use short TTL.
func (h *Handler) HandleGetAllUser(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()

	user, err := h.Repo.GetAllUser(r.Context())
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleGetAllUser: Failed to get All User: %v", err))
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	if log != nil {
		log.Info("HandleGetAllUser: successfully")
	}
	handler.RespondWithJson(w, http.StatusOK, user)
}

// TODO(Cache): Unused route (no router registration). Skip.
func (h *Handler) HandleGetUserByUserName(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	username := chi.URLParam(r, "user-name")
	if username == "" {
		if log != nil {
			log.Error("HandleGetUserByUserName: Failed to get All User:")
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	userName := sql.NullString{
		String: username,
		Valid:  true,
	}
	user, err := h.Repo.GetUserByUsername(r.Context(), userName)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleGetAllUser: Failed to get All User: %v", err))
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	if log != nil {
		log.Info("HandleGetAllUser: successfully")
	}
	handler.RespondWithJson(w, http.StatusOK, user)
}

// TODO(Cache): Cache approved user list (T2 route).
// Cache Miss -> DB GetAllUserApproved, Set cache with list key.
// Cache Hit -> Return cached list.
// Invalidate when user is created, approved, role changes, or is banned.
func (h *Handler) HandleGetApprovedUser(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	cached := h.Cache.GetList("approved_users")
	if cached != nil {
		fmt.Println("Cache BABY")
		handler.RespondWithJson(w, http.StatusOK, cached)
		return
	}
	user, err := h.Repo.GetAllUserApproved(r.Context())
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleGetAllUser: Failed to get All User: %v", err))
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	if log != nil {
		log.Info("HandleGetAllUser: successfully")
	}
	h.Cache.SetList("approved_users", user)
	handler.RespondWithJson(w, http.StatusOK, user)
}

// TODO(Cache): Implement caching for GET /core-member using list cache.
//
// CacheKey: "core_members"
//
// Plan:
//   1. Check h.Cache.GetList("core_members")
//      - If hit, return cached data immediately.
//   2. On miss, call h.Repo.GetCoreMembers(ctx)
//   3. Store result via h.Cache.SetList("core_members", users)
//   4. Return fresh data.
//
// Invalidation triggers (in HandlerRole admin):
//   - User role changes (member <-> core_member/vp/president)
//   - User status changes (e.g., banned)
//
// Invalidation: h.Cache.DeleteList("core_members")
//
// GetCoreMembersRow (sqlc generated in internal/database/user.sql.go):
//   type GetCoreMembersRow struct {
//       ID          uuid.UUID
//       Name        string
//       Email       string
//       Status      AccountStatus
//       Role        UserRole
//       Image       sql.NullString
//       SocialLinks json.RawMessage
//   }
func (h *Handler) HandleGetCoreMember(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	cached := h.Cache.GetList("core_members")
	if cached != nil {
		fmt.Println("Cache baby")
		handler.RespondWithJson(w, http.StatusOK, cached)
		return
	}
	user, err := h.Repo.GetCoreMembers(r.Context())
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleGetAllUser: Failed to get All User: %v", err))
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	if log != nil {
		log.Info("HandleGetAllUser: successfully")
	}
	h.Cache.SetList("core_members", user)
	handler.RespondWithJson(w, http.StatusOK, user)
}

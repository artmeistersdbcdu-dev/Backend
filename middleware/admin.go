package middleware

import (
	"context"
	"net/http"

	"github.com/Blue-Onion/ArtmeisterBackend/handler"
	"github.com/Blue-Onion/ArtmeisterBackend/internal/database"
	"github.com/Blue-Onion/ArtmeisterBackend/utlis"
	"github.com/google/uuid"
)

const senior contextKey = "senior"
const moderator contextKey = "moderator"

func GetSenior(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(senior).(User)
	return user, ok
}

func WithSenior(ctx context.Context, u User) context.Context {
	return context.WithValue(ctx, senior, u)
}

func GetModerator(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(moderator).(User)
	return user, ok
}

func WithModerator(ctx context.Context, u User) context.Context {
	return context.WithValue(ctx, moderator, u)
}

func (h Handler) MiddlewareSeniorAuth(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenCookie, err := r.Cookie("authToken")
		if err != nil {
			handler.RespondWithError(w, http.StatusUnauthorized, "Unauthorized: login required")
			return
		}

		userId, err := utlis.GetUserIdJwt(tokenCookie)
		if err != nil {
			handler.RespondWithError(w, http.StatusUnauthorized, "Unauthorized: invalid or expired token")
			return
		}

		id, err := uuid.Parse(userId)
		if err != nil {
			handler.RespondWithError(w, http.StatusBadRequest, "Invalid user ID format")
			return
		}

		user, err := h.Repo.CheckUsrById(r.Context(), id)
		if err != nil {
			if utlis.IsNotFound(err) {
				handler.RespondWithError(w, http.StatusUnauthorized, "Unauthorized: user not found")
				return
			}
			handler.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}
		if user.Role != database.UserRolePresident {
			handler.RespondWithError(w, http.StatusUnauthorized, "Unauthorized: senior access required")
			return
		}
		if user.Status != database.AccountStatusApproved {
			handler.RespondWithError(w, http.StatusUnauthorized, "You are Banned to do anything")
			return
		}
		ctx := context.WithValue(r.Context(), senior, User{
			ID:     user.ID,
			Status: user.Status,
			Role:   user.Role,
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (h Handler) MiddlewareModeratorAuth(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenCookie, err := r.Cookie("authToken")
		if err != nil {
			handler.RespondWithError(w, http.StatusUnauthorized, "Unauthorized: login required")
			return
		}

		userId, err := utlis.GetUserIdJwt(tokenCookie)
		if err != nil {
			handler.RespondWithError(w, http.StatusUnauthorized, "Unauthorized: invalid or expired token")
			return
		}

		id, err := uuid.Parse(userId)
		if err != nil {
			handler.RespondWithError(w, http.StatusBadRequest, "Invalid user ID format")
			return
		}

		user, err := h.Repo.CheckUsrById(r.Context(), id)
		if err != nil {
			if utlis.IsNotFound(err) {
				handler.RespondWithError(w, http.StatusUnauthorized, "Unauthorized: user not found")
				return
			}
			handler.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}
		if !utlis.CanModerate(user.Role) {
			handler.RespondWithError(w, http.StatusUnauthorized, "Unauthorized: moderator access required")
			return
		}
		if user.Status != database.AccountStatusApproved {
			handler.RespondWithError(w, http.StatusUnauthorized, "You are Banned to do anything")
			return
		}
		ctx := context.WithValue(r.Context(), moderator, User{
			ID:     user.ID,
			Status: user.Status,
			Role:   user.Role,
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

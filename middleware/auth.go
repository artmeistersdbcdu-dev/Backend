package middleware

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/Blue-Onion/ArtmeisterBackend/handler"
	"github.com/Blue-Onion/ArtmeisterBackend/internal/database"
	"github.com/Blue-Onion/ArtmeisterBackend/utlis"
	"github.com/google/uuid"
)

type Handler struct {
	Repo database.UserRepository
}

type contextKey string

const userContextKey contextKey = "user"

type User struct {
	ID          uuid.UUID
	Name        string
	Email       string
	UserName    sql.NullString
	Batch       sql.NullString
	Status      database.AccountStatus
	Role        database.UserRole
	Image       sql.NullString
	BannerImage sql.NullString
}

func GetUser(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(userContextKey).(User)
	return user, ok
}

func WithUser(ctx context.Context, u User) context.Context {
	return context.WithValue(ctx, userContextKey, u)
}

func (h Handler) MiddlewareAuth(next http.Handler) http.HandlerFunc {
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

		usr, err := h.Repo.CheckUsrById(r.Context(), id)
		if err != nil {
			if utlis.IsNotFound(err) {
				handler.RespondWithError(w, http.StatusUnauthorized, "Unauthorized: user not found")
				return
			}
			handler.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, User{
			ID:     usr.ID,
			Status: usr.Status,
			Role:   usr.Role,
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

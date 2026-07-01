package admin

import (
	"github.com/Blue-Onion/ArtmeisterBackend/handler/art"
	"github.com/Blue-Onion/ArtmeisterBackend/handler/user"
	"github.com/Blue-Onion/ArtmeisterBackend/middleware"
	"github.com/go-chi/chi"
	"net/http"
)

func AdminRoute(userHandler *user.Handler, artHandler *art.Handler, middlewareHandler *middleware.Handler) *chi.Mux {
	r := chi.NewRouter()

	senior := middlewareHandler.MiddlewareSeniorAuth
	moderator := middlewareHandler.MiddlewareModeratorAuth

	adminUserHandler := &UserHandler{Repo: userHandler.Repo}
	adminArtHandler := &ArtHandler{Repo: artHandler.Repo}

	r.Patch("/arts/{art_id}/status", moderator(http.HandlerFunc(adminArtHandler.HandlerArtStatus)))
	r.Patch("/users/{user_id}/status", senior(http.HandlerFunc(adminUserHandler.HandlerRole)))

	return r
}

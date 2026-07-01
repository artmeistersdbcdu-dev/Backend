package user

import (
	"net/http"

	"github.com/Blue-Onion/ArtmeisterBackend/middleware"
	"github.com/go-chi/chi"
)

func UserRouter(userHandler *Handler, middlewareHandler *middleware.Handler) *chi.Mux {
	r := chi.NewRouter()

	// Public routes
	r.Post("/users", userHandler.HandleCreateUser)
	r.Post("/login", userHandler.HandleLogin)

	// Protected routes
	auth := middlewareHandler.MiddlewareAuth
	senior := middlewareHandler.MiddlewareSeniorAuth

	r.Get("/main-users", http.HandlerFunc(userHandler.HandleGetApprovedUser))
	r.Get("/users", senior(http.HandlerFunc(userHandler.HandleGetAllUser)))
	r.Patch("/users/{id}", auth(http.HandlerFunc(userHandler.HandleUpdateUserProfile)))
	r.Get("/users/{id}", userHandler.HandleGetUserById)
	r.Patch("/users/password", auth(http.HandlerFunc(userHandler.HandlePasswordChange)))
	r.Post("/logout", auth(http.HandlerFunc(userHandler.HandleLogOut)))
	r.Get("/me", http.HandlerFunc(userHandler.HandleMe))
	r.Get("/core-member", http.HandlerFunc(userHandler.HandleGetCoreMember))

	return r
}

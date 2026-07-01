package art

import (
	"net/http"

	artmetadata "github.com/Blue-Onion/ArtmeisterBackend/handler/artMetaData"
	"github.com/Blue-Onion/ArtmeisterBackend/middleware"
	"github.com/go-chi/chi"
)

// ArtRouter defines routes for the art package.
func ArtRouter(artHandler *Handler, artMetadataHandler *artmetadata.Handler, middlewareHandler *middleware.Handler, profile *ProfileHandler) *chi.Mux {
	r := chi.NewRouter()
	auth := middlewareHandler.MiddlewareAuth
	// Public routes
	r.Get("/u/{user_id}", artHandler.HandleGetArts)
	r.Get("/p/{user_id}/{id}", artHandler.HandleGetArtProfileById)
	r.Get("/{id}", artHandler.HandleGetArtById)
	r.Get("/pending-art", artHandler.HandleGetPendingArt)

	r.Get("/u/profile/{id}", profile.HandlerGetArtistProfile)
	// Art metadata public routes
	r.Get("/{id}/comments", artMetadataHandler.HandleGetArtComments)
	r.Get("/{id}/comments/count", artMetadataHandler.HandleGetArtCommentsCount)
	r.Get("/{id}/likes/count", artMetadataHandler.HandleGetArtLikeCount)
	r.Get("/", artHandler.HandleGetApprovedArt)
	// Protected routes (require user authentication)
	r.Post("/", auth(http.HandlerFunc(artHandler.HandleArtCreation)))
	r.Delete("/{id}", auth(http.HandlerFunc(artHandler.HandleArtDeletion)))
	r.Patch("/{id}", auth(http.HandlerFunc(artHandler.HandlerArtUpdation)))

	// Art metadata protected routes
	r.Post("/{art_id}/comment", auth(http.HandlerFunc(artMetadataHandler.HandleComment)))
	r.Delete("/comment/{id}", auth(http.HandlerFunc(artMetadataHandler.HandleDeleteComment)))
	r.Post("/{art_id}/like", auth(http.HandlerFunc(artMetadataHandler.HandleLike)))
	r.Post("/{art_id}/unlike", auth(http.HandlerFunc(artMetadataHandler.HandleUnLike)))

	return r
}

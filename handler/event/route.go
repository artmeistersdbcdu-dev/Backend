package event

import (
	"net/http"

	"github.com/Blue-Onion/ArtmeisterBackend/middleware"
	"github.com/go-chi/chi"
)

// EventRouter defines routes for the event package.
func EventRouter(eventHandler *EventHandler, attendeeHandler *EventAttendeeHandler, middlewareHandler *middleware.Handler) *chi.Mux {
	r := chi.NewRouter()
	auth := middlewareHandler.MiddlewareAuth
	moderator := middlewareHandler.MiddlewareModeratorAuth

	// Public routes
	r.Get("/", eventHandler.HandleGetAllEvent)
	r.Get("/{id}", eventHandler.HandleGetEventById)

	// Protected routes (require user authentication)
	r.Post("/{id}/join", auth(http.HandlerFunc(attendeeHandler.HandleJoinEvent)))
	r.Get("/u/{id}", auth(http.HandlerFunc(attendeeHandler.HandleGetMyEvent)))
	r.Get("/all", auth(http.HandlerFunc(attendeeHandler.HandleGetMyAllEvent)))
	r.Delete("/{id}/attendee/{user_id}", auth(http.HandlerFunc(attendeeHandler.HandleDeleteEventAttendee)))
	r.Get("/{id}/attendees", auth(http.HandlerFunc(attendeeHandler.HandleAllEventAttendee)))

	// Moderator routes (require president or vp)
	r.Post("/", moderator(http.HandlerFunc(eventHandler.HandleCreateEvent)))
	r.Patch("/{id}", moderator(http.HandlerFunc(eventHandler.HandleUpdateEvent)))
	r.Delete("/{id}", moderator(http.HandlerFunc(eventHandler.HandleDeleteEvent)))

	return r
}

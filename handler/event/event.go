package event

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Blue-Onion/ArtmeisterBackend/handler"
	"github.com/Blue-Onion/ArtmeisterBackend/handler/logger"
	"github.com/Blue-Onion/ArtmeisterBackend/internal/database"
	"github.com/Blue-Onion/ArtmeisterBackend/middleware"
	"github.com/Blue-Onion/ArtmeisterBackend/model"
	"github.com/Blue-Onion/ArtmeisterBackend/utlis"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type EventHandler struct {
	Repo database.EventRepository
}
type EventAttendeeHandler struct {
	Repo database.EventAttendeesRepository
}

func (h *EventHandler) HandleCreateEvent(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	err := r.ParseMultipartForm(20 << 20)

	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleCreateEvent: failed to parse form data: %v", err))
		}
		handler.RespondWithError(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}
	id := uuid.New()
	name := r.FormValue("name")
	if len(name) < 3 {
		if log != nil {
			log.Error(fmt.Sprintf("HandleCreateEvent: name too short: '%s'", name))
		}
		handler.RespondWithError(w, http.StatusBadRequest, "Too Short Name")
		return
	}
	form_date := r.FormValue("date")
	fmt.Println(form_date)
	eventDate, err := time.Parse("2006-01-02", form_date)

	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleCreateEvent: invalid date format '%s': %v", form_date, err))
		}
		handler.RespondWithError(w, 400, "invalid date format (expected YYYY-MM-DD)")
		return
	}

	desc := r.FormValue("description")
	venue := r.FormValue("venue")
	LogoUrl := r.FormValue("LogoUrl")
	bannerUrl := r.FormValue("bannerUrl")

	status := r.FormValue("status")
	mode := database.ModeOfConduct(status)
	if mode == "" {
		if log != nil {
			log.Error(fmt.Sprintf("HandleCreateEvent: unknown mode status '%s'", status))
		}
		handler.RespondWithError(w, http.StatusBadRequest, "Unknown Mode")
		return
	}
	params := database.CreateEventParams{
		ID:          id,
		Name:        name,
		Description: utlis.ToNilStr(&desc),
		Venue:       utlis.ToNilStr(&venue),
		Status:      mode,
		Image:       utlis.ToNilStr(&LogoUrl),
		BannerImage: utlis.ToNilStr(&bannerUrl),
		EventDate:   eventDate,
	}
	res, err := h.Repo.CreateEvent(r.Context(), params)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleCreateEvent: failed to create event %s: %v", id, err))
		}
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to update images")
		return
	}
	if log != nil {
		log.Info(fmt.Sprintf("HandleCreateEvent: event %s created", id))
	}
	handler.RespondWithJson(w, http.StatusOK, map[string]string{"ID": res.String()})
}
func (h *EventHandler) HandleDeleteEvent(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	id := chi.URLParam(r, "id")
	eventId, err := uuid.Parse(id)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleDeleteEvent: invalid ID format '%s': %v", id, err))
		}
		handler.RespondWithError(w, http.StatusBadRequest, "Invalid Id")
		return
	}
	_, err = h.Repo.DeleteEvent(r.Context(), eventId)
	if err != nil {
		if utlis.IsNotFound(err) {
			handler.RespondWithError(w, http.StatusNotFound, "Event not found")
			return
		}
		if log != nil {
			log.Error(fmt.Sprintf("HandleDeleteEvent: failed to delete event %s: %v", eventId, err))
		}
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to delete event")
		return
	}
	handler.RespondWithJson(w, http.StatusOK, "ok")
}
func (h *EventHandler) HandleGetEventById(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	id := chi.URLParam(r, "id")
	eventId, err := uuid.Parse(id)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleGetEventById: invalid ID format '%s': %v", id, err))
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	res, err := h.Repo.GetEventByID(r.Context(), eventId)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleGetEventById: failed to get event %s: %v", eventId, err))
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	if log != nil {
		log.Info(fmt.Sprintf("HandleGetEventById: retrieved event %s successfully", eventId))
	}
	handler.RespondWithJson(w, http.StatusOK, res)
}
func (h *EventHandler) HandleGetAllEvent(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	res, err := h.Repo.ListEvents(r.Context())
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleGetAllEvent: failed to list events: %v", err))
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	if log != nil {
		log.Info("HandleGetAllEvent: retrieved all events successfully")
	}
	handler.RespondWithJson(w, http.StatusOK, res)
}
func (h *EventHandler) HandleUpdateEvent(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()

	eventId := chi.URLParam(r, "id")

	id, err := uuid.Parse(eventId)

	if err != nil {

		handler.RespondWithError(w, http.StatusBadRequest, err.Error())
		return

	}

	req := model.UpdateEventRequest{}
	err = json.NewDecoder(r.Body).Decode(&req)

	if err != nil {

		handler.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return

	}

	params := database.UpdateEventParams{

		ID:          id,
		Name:        utlis.ToNilStr(req.Name),
		Description: utlis.ToNilStr(req.Description),
		Venue:       utlis.ToNilStr(req.Venue),
		Image:       utlis.ToNilStr(req.Image),
		BannerImage: utlis.ToNilStr(req.BannerImage),
	} // optional date
	if req.Date != nil {
		eventDate, err := time.Parse("2006-01-02", *req.Date)
		if err != nil {
			handler.RespondWithError(w, http.StatusBadRequest, "invalid date format")
			return
		}

		params.EventDate = sql.NullTime{
			Valid: true,
			Time:  eventDate,
		}
	}
	if req.Status != nil {
		mode := database.ModeOfConduct(*req.Status)

		if mode == "" {
			handler.RespondWithError(w, http.StatusBadRequest, "Unknown Mode")
			return
		}

		params.Status = database.NullModeOfConduct{
			Valid:         true,
			ModeOfConduct: mode,
		}
	}
	fmt.Println(params)
	res, err := h.Repo.UpdateEvent(r.Context(), params)
	if err != nil {
		if utlis.IsNotFound(err) {
			handler.RespondWithError(w, http.StatusNotFound, "Event not found")
			return
		}

		if log != nil {
			log.Error(fmt.Sprintf("HandleUpdateEvent: failed to update event %s: %v", id, err))
		}

		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to update event")
		return
	}

	if log != nil {
		log.Info(fmt.Sprintf("HandleUpdateEvent: event %s updated", id))
	}

	handler.RespondWithJson(w, http.StatusOK, map[string]string{"ID": res.String()})
}
func (h *EventAttendeeHandler) HandleJoinEvent(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	user, _ := middleware.GetUser(r.Context())
	userId := user.ID
	id := chi.URLParam(r, "id")
	event_id, err := uuid.Parse(id)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleJoinEvent: invalid event ID format '%s' for user %s: %v", id, userId, err))
		}
		handler.RespondWithError(w, 400, err.Error())
		return
	}
	param := database.EnrollUserToEventParams{
		ID:      uuid.New(),
		EventID: event_id,
		UserID:  userId,
	}
	res, err := h.Repo.EnrollUserToEvent(r.Context(), param)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleJoinEvent: failed to enroll user %s to event %s: %v", userId, event_id, err))
		}
		handler.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	if log != nil {
		log.Info(fmt.Sprintf("HandleJoinEvent: user %s joined event %s", userId, event_id))
	}
	handler.RespondWithJson(w, http.StatusOK, map[string]string{"ID": res.String()})
}
func (h *EventAttendeeHandler) HandleDeleteEventAttendee(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()

	user_id := chi.URLParam(r, "user_id")
	userId, err := uuid.Parse(user_id)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleDeleteEventAttendee: invalid user_id format '%s': %v", user_id, err))
		}
		handler.RespondWithError(w, 400, err.Error())
		return
	}
	id := chi.URLParam(r, "id")
	event_id, err := uuid.Parse(id)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleDeleteEventAttendee: invalid event ID format '%s' for user %s: %v", id, userId, err))
		}
		handler.RespondWithError(w, 400, err.Error())
		return
	}
	param := database.RemoveUserFromEventParams{
		EventID: event_id,
		UserID:  userId,
	}
	_, err = h.Repo.RemoveUserFromEvent(r.Context(), param)
	if err != nil {
		if utlis.IsNotFound(err) {
			handler.RespondWithError(w, http.StatusNotFound, "Attendance record not found")
			return
		}
		if log != nil {
			log.Error(fmt.Sprintf("HandleDeleteEventAttendee: failed to remove user %s from event %s: %v", userId, event_id, err))
		}
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to remove user from event")
		return
	}
	if log != nil {
		log.Info(fmt.Sprintf("HandleDeleteEventAttendee: user %s removed from event %s", userId, event_id))
	}
	handler.RespondWithJson(w, http.StatusOK, "ok")
}
func (h *EventAttendeeHandler) HandleAllEventAttendee(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	id := chi.URLParam(r, "id")
	event_id, err := uuid.Parse(id)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleAllEventAttendee: invalid event ID format '%s': %v", id, err))
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	res, err := h.Repo.ListEventAttendees(r.Context(), event_id)

	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleAllEventAttendee: failed to list attendees for event %s: %v", event_id, err))
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	if log != nil {
		log.Info(fmt.Sprintf("HandleAllEventAttendee: retrieved attendees for event %s successfully", event_id))
	}
	handler.RespondWithJson(w, http.StatusOK, res)
}
func (h *EventAttendeeHandler) HandleGetMyEvent(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	user, _ := middleware.GetUser(r.Context())
	userId := user.ID
	id := chi.URLParam(r, "id")
	event_id, err := uuid.Parse(id)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleAllEventAttendee: invalid event ID format '%s': %v", id, err))
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	params := database.GetMyEventByIdParams{
		EventID: event_id,
		UserID:  userId,
	}
	res, err := h.Repo.GetMyEventById(r.Context(), params)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf("HandleAllEventAttendee: failed to list attendees for event %s: %v", event_id, err))
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}
	if log != nil {
		log.Info(fmt.Sprintf("HandleAllEventAttendee: retrieved attendees for event %s successfully", event_id))
	}
	handler.RespondWithJson(w, http.StatusOK, map[string]string{"ID": res.String()})
}

func (h *EventAttendeeHandler) HandleGetMyAllEvent(w http.ResponseWriter, r *http.Request) {
	log, _ := logger.GetLogger()
	user, _ := middleware.GetUser(r.Context())

	res, err := h.Repo.ListMyEvents(r.Context(), user.ID)
	if err != nil {
		if log != nil {
			log.Error(fmt.Sprintf(
				"HandleGetMyAllEvent: failed to list events for user %s: %v",
				user.ID, err,
			))
		}
		handler.RespondWithJsonCustom(w, http.StatusOK, false, nil)
		return
	}

	if log != nil {
		log.Info(fmt.Sprintf(
			"HandleGetMyAllEvent: retrieved %d events for user %s successfully",
			len(res), user.ID,
		))
	}

	handler.RespondWithJson(w, http.StatusOK, res)
}

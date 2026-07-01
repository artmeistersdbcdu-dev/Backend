package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type CreateUser struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
type AuthenticateUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type PatchUserProfileRequest struct {
	UserName     *string          `json:"username"`
	Email        *string          `json:"email"`
	Batch        *string          `json:"batch"`
	Image        *string          `json:"image"`
	Banner_image *string          `json:"banner_image"`
	Desc         *string          `json:"description"`
	Social       *json.RawMessage `json:"social"`
}
type PatchUserPassword struct {
	OldPassword string `json:"old_password"`
	Password    string `json:"password"`
}
type PatchUserStatus struct {
	Role   string `json:"role"`
	Status string `json:"status"`
}
type CreateEvent struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Venue       string `json:"venue"`
	Image       string `json:"image"`
	BannerImage string `json:"banner_image"`
	EventDate   string `json:"event_date"`
	Status      string `json:"status"`
}
type AddComment struct {
	Comment string `json:"comment"`
}
type UpdateArtRequest struct {
	Name        *string   `json:"name"`
	Description *string   `json:"description"`
	Tags        *[]string `json:"tags"`
}
type UpdateEventRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Venue       *string `json:"venue"`
	Image       *string `json:"image"`
	BannerImage *string `json:"banner_image"`
	Date        *string `json:"date"`
	Status      *string `json:"status"`
}
type CreateArtRequest struct {
	Name        string   `json:"name"`
	URL         string   `json:"url"`
	Description *string  `json:"description"`
	Tags        []string `json:"tags"`
}

package cache

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type User struct {
	Name              string
	Username          sql.NullString
	Password          string
	Email             string
	Description       sql.NullString
	BannerImage       sql.NullString
	Image             sql.NullString
	Batch             sql.NullString
	SocialLinks       json.RawMessage
	Status            string
	Role              string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	ResetToken        sql.NullString
	ResetTokenExpires sql.NullTime
}
type Art struct {
	Name        string
	Description sql.NullString
	Image       string
	Tags        []string
	Status      string
	UserID      uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
type Event struct {
	Name        string
	Description sql.NullString
	Venue       sql.NullString
	Image       sql.NullString
	BannerImage sql.NullString
	EventDate   time.Time
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type EventAttendee struct {
	EventID  uuid.UUID
	UserID   uuid.UUID
	JoinedAt time.Time
}
type EventCache struct {
	Event    Event
	LastUsed time.Time
	ExpireAt time.Time
}
type ArtCache struct {
	Art      Art
	LastUsed time.Time
	ExpireAt time.Time
}
type UserCache struct {
	User     User
	LastUsed time.Time
	ExpireAt time.Time
}

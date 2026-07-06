package cache

import (
	"database/sql"
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

type User struct {
	Name        string
	Username    sql.NullString
	Password    string
	Email       string
	Description sql.NullString
	BannerImage sql.NullString
	Image       sql.NullString
	Batch       sql.NullString
	SocialLinks json.RawMessage
	Status      string
	Role        string
}
type Art struct {
	Name        string
	Description sql.NullString
	Image       string
	Tags        []string
	Status      string
	UserID      uuid.UUID
}
type Event struct {
	Name        string
	Description sql.NullString
	Venue       sql.NullString
	Image       sql.NullString
	BannerImage sql.NullString
	EventDate   time.Time
	Status      string
}

type EventAttendee struct {
	EventID  uuid.UUID
	UserID   uuid.UUID
	JoinedAt time.Time
}
type EventCache struct {
	Event    Event
	ExpireAt time.Time
}
type ArtCache struct {
	Art      Art
	ExpireAt time.Time
}
type UserCache struct {
	User     User
	ExpireAt time.Time
}
type ListCache struct {
	Data     any
	ExpireAt time.Time
}

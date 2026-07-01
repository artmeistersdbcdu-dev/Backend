package database

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type UserRepository interface {
	GetUser(ctx context.Context, id uuid.UUID) (GetUserRow, error)
	GetAllUser(ctx context.Context) ([]GetAllUserRow, error)
	GetUserByEmail(ctx context.Context, email string) (GetUserByEmailRow, error)
	PatchUserProfile(ctx context.Context, arg PatchUserProfileParams) (PatchUserProfileRow, error)
	PatchUserAdmin(ctx context.Context, arg PatchUserAdminParams) (PatchUserAdminRow, error)
	GetUserByUsername(ctx context.Context, username sql.NullString) (GetUserByUsernameRow, error)
	GetAllUserApproved(ctx context.Context) ([]GetAllUserApprovedRow, error)
	CreateUser(ctx context.Context, arg CreateUserParams) (uuid.UUID, error)
	PatchUserPassword(ctx context.Context, arg PatchUserPasswordParams) (uuid.UUID, error)
	GetCoreMembers(ctx context.Context) ([]GetCoreMembersRow, error)
	CheckUsrById(ctx context.Context, id uuid.UUID) (CheckUsrByIdRow, error)
}
type ArtRepository interface {
	DeleteArt(ctx context.Context, arg DeleteArtParams) (uuid.UUID, error)
	GetArtByUser(ctx context.Context, userID uuid.UUID) ([]GetArtByUserRow, error)
	GetArtProfileByID(ctx context.Context, arg GetArtProfileByIDParams) (GetArtProfileByIDRow, error)
	GetArtByID(ctx context.Context, id uuid.UUID) (GetArtByIDRow, error)
	ListArt(ctx context.Context) ([]ListArtRow, error)
	ListPendingArt(ctx context.Context) ([]ListPendingArtRow, error)
	ListArtByTag(ctx context.Context, tags []string) ([]ListArtByTagRow, error)
	ListArtByTags(ctx context.Context, dollar_1 []string) ([]ListArtByTagsRow, error)
	UpdateArt(ctx context.Context, arg UpdateArtParams) (uuid.UUID, error)
	UpdateArtStatus(ctx context.Context, arg UpdateArtStatusParams) (UpdateArtStatusRow, error)
	CreateArt(ctx context.Context, arg CreateArtParams) (uuid.UUID, error)
}
type ArtMetaDataRepository interface {
	AddArtComment(ctx context.Context, arg AddArtCommentParams) (ArtComment, error)
	CheckArtLikedByUser(ctx context.Context, arg CheckArtLikedByUserParams) (bool, error)
	DeleteArtComment(ctx context.Context, arg DeleteArtCommentParams) (uuid.UUID, error)
	GetArtCommentsByArtID(ctx context.Context, artID uuid.UUID) ([]GetArtCommentsByArtIDRow, error)
	GetArtCommentsCount(ctx context.Context, artID uuid.UUID) (int32, error)
	GetArtLikesCount(ctx context.Context, artID uuid.UUID) (int32, error)
	LikeArt(ctx context.Context, arg LikeArtParams) (ArtLike, error)
	UnlikeArt(ctx context.Context, arg UnlikeArtParams) (uuid.UUID, error)
}
type EventRepository interface {
	CreateEvent(ctx context.Context, arg CreateEventParams) (uuid.UUID, error)
	DeleteEvent(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
	GetEventByID(ctx context.Context, id uuid.UUID) (GetEventByIDRow, error)
	ListEvents(ctx context.Context) ([]ListEventsRow, error)
	ListEventsByMode(ctx context.Context, status ModeOfConduct) ([]ListEventsByModeRow, error)
	ListUpcomingEvents(ctx context.Context) ([]ListUpcomingEventsRow, error)
	UpdateEvent(ctx context.Context, arg UpdateEventParams) (uuid.UUID, error)
}
type EventAttendeesRepository interface {
	GetMyEventById(ctx context.Context, arg GetMyEventByIdParams) (uuid.UUID, error)
	CountEventAttendees(ctx context.Context, eventID uuid.UUID) (int32, error)
	EnrollUserToEvent(ctx context.Context, arg EnrollUserToEventParams) (uuid.UUID, error)
	ListEventAttendees(ctx context.Context, eventID uuid.UUID) ([]ListEventAttendeesRow, error)
	ListMyEvents(ctx context.Context, userID uuid.UUID) ([]ListMyEventsRow, error)
	RemoveUserFromEvent(ctx context.Context, arg RemoveUserFromEventParams) (uuid.UUID, error)
}

# Struct Usage & Cache Analysis

---

## CLIENT-SIDE ROUTE FREQUENCY TIERS

Routes grouped by expected frequency of client invocation. These tiers are used to weight field-level classification.

| Tier | Frequency | Routes | Examples |
|------|-----------|--------|---------|
| **T1** | Very High (every page load) | 7 routes | `GET /art/`, `GET /event/`, `GET /auth/me`, `GET /art/{id}`, comment/like count endpoints |
| **T2** | High (profile browsing) | 7 routes | `GET /auth/users/{id}`, `GET /art/u/{user_id}`, `GET /art/p/{user_id}/{id}`, `GET /art/u/profile/{id}`, `GET /event/{id}`, `GET /event/{id}/attendees` |
| **T3** | Medium (user actions) | 8 routes | `POST /auth/login`, `POST /auth/users`, likes/comments/joins, `GET /event/all`, `GET /event/u/{id}` |
| **T4** | Low (edits, destructive) | 10 routes | `PATCH /auth/users/{id}`, password change, create/update/delete art/event, logout |
| **T5** | Very Low (admin only) | 4 routes | `GET /auth/users`, `GET /art/pending-art`, status/role patches |

---

## USER

### All Possible User Fields

| Field | Type | Source |
|-------|------|--------|
| ID | `uuid.UUID` | All user queries |
| Name | `string` | All user queries except scalar-ID returns |
| Username | `sql.NullString` | `GetUser`, `GetUserByUsername`, `PatchUserProfile` |
| Email | `string` | `GetAllUser`, `GetUser`, `GetUserByEmail`, `GetUserByUsername`, `GetCoreMembers`, `PatchUserProfile` |
| Password | `string` | `GetUserByEmail` only (never in HTTP response) |
| Batch | `sql.NullString` | `GetUser`, `GetUserByUsername`, `PatchUserProfile` |
| Status | `AccountStatus` | `GetAllUser`, `GetUser`, `GetUserByUsername`, `GetCoreMembers`, `PatchUserAdmin`, `PatchUserProfile` |
| Role | `UserRole` | `GetAllUser`, `GetAllUserApproved`, `GetUser`, `GetUserByUsername`, `GetCoreMembers`, `PatchUserAdmin`, `PatchUserProfile` |
| Image | `sql.NullString` | Every user-data query |
| BannerImage | `sql.NullString` | `GetUser`, `GetUserByUsername`, `PatchUserProfile` |
| Description | `sql.NullString` | `GetAllUserApproved`, `GetUser`, `GetUserByUsername`, `PatchUserProfile` |
| SocialLinks | `json.RawMessage` | `GetAllUserApproved`, `GetUser`, `GetUserByUsername`, `GetCoreMembers`, `PatchUserProfile` |

---

### Routes Using User

| # | Tier | Method | Path | Handler | DB Row Type | Fields Returned | Full/Partial |
|---|------|--------|------|---------|-------------|----------------|-------------|
| 1 | T5 | POST | `/auth/users` | `HandleCreateUser` | `uuid.UUID` | `{ID}` | Scalar |
| 2 | T3 | POST | `/auth/login` | `HandleLogin` | `GetUserByEmailRow` | `id, name, email, image` | Partial |
| 3 | T2 | GET | `/auth/main-users` | `HandleGetApprovedUser` | `[]GetAllUserApprovedRow` | `id, name, role, description, image, social_links` | Partial |
| 4 | T5 | GET | `/auth/users` | `HandleGetAllUser` | `[]GetAllUserRow` | `id, name, email, status, role, image` | Partial |
| 5 | T4 | PATCH | `/auth/users/{id}` | `HandleUpdateUserProfile` | `PatchUserProfileRow` | `id, name, username, email, batch, status, role, image, banner_image, description, social_links` | Full |
| 6 | T2 | GET | `/auth/users/{id}` | `HandleGetUserById` | `GetUserRow` | `id, name, username, email, batch, status, role, image, banner_image, description, social_links` | Full |
| 7 | T4 | PATCH | `/auth/users/password` | `HandlePasswordChange` | `GetUserByEmailRow` + `uuid.UUID` | `{ID}` | Scalar |
| 8 | T4 | POST | `/auth/logout` | `HandleLogOut` | — | `{message}` | None |
| 9 | T1 | GET | `/auth/me` | `HandleMe` | `GetUserRow` | `id, name, username, email, batch, status, role, image, banner_image, description, social_links` | Full |
| 10 | T2 | GET | `/auth/core-member` | `HandleGetCoreMember` | `[]GetCoreMembersRow` | `id, name, email, status, role, image, social_links` | Partial |
| 11 | T5 | PATCH | `/admin/users/{user_id}/status` | `HandlerRole` | `PatchUserAdminRow` | `id, status, role` | Partial |
| 12 | T2 | GET | `/event/{id}/attendees` | `HandleAllEventAttendee` | `[]ListEventAttendeesRow` | `id, name, username, email, image` | Partial |
| 13 | T2 | GET | `/art/u/profile/{id}` | `HandlerGetArtistProfile` | `GetUserRow` (composite) | `id, name, username, email, batch, status, role, image, banner_image, description, social_links` | Full |

---

### User Field Classification (weighted by route frequency)

| Field | Loaded by Routes (Tier) | Estimated Load Frequency | Category |
|-------|------------------------|-------------------------|----------|
| ID | All user routes (T1–T5) | Every user interaction | **Always Used** |
| Name | All user routes (T1–T5) | Every user interaction | **Always Used** |
| Email | T1 (`/me`), T2 (profile, core, attendees), T3 (login), T5 (admin list) | Very frequent | **Frequently Used** |
| Image | T1 (`/me`), T2 (profile, core, approved, attendees), T3 (login), T5 (admin) | Very frequent | **Frequently Used** |
| Role | T1 (`/me`), T2 (profile, core, approved), T5 (admin list, admin patch) | Very frequent | **Frequently Used** |
| Status | T1 (`/me`), T2 (profile, core), T5 (admin list, admin patch) | Frequent | **Frequently Used** |
| Username | T1 (`/me`), T2 (profile, attendees) | Frequent | **Occasionally Used** |
| Description | T1 (`/me`), T2 (profile, approved) | Frequent | **Occasionally Used** |
| SocialLinks | T1 (`/me`), T2 (profile, core, approved) | Frequent | **Occasionally Used** |
| Batch | T1 (`/me`), T2 (profile) | Frequent (but niche) | **Occasionally Used** |
| BannerImage | T1 (`/me`), T2 (profile) | Frequent (but niche) | **Occasionally Used** |
| Password | T3 (login only) | Rare (once per session) | **Rarely Used** |

| Category | Definition | Fields |
|----------|-----------|--------|
| **Always Used** | Loaded by every tier, essential identifier | `ID`, `Name` |
| **Frequently Used** | Loaded by T1 + T2 + at least one lower tier, displayed in most contexts | `Email`, `Image`, `Role`, `Status` |
| **Occasionally Used** | Loaded by T1 + T2 but niche or conditional display | `Username`, `Description`, `SocialLinks`, `Batch`, `BannerImage` |
| **Rarely Used** | Loaded only by low-frequency routes (T3/T4/T5) | `Password` |

---

## EVENT

### All Possible Event Fields

| Field | Type | Source |
|-------|------|--------|
| ID | `uuid.UUID` | All event queries |
| Name | `string` | All event entity queries |
| Description | `sql.NullString` | `GetEventByID`, `ListEvents`, `ListMyEvents` |
| Venue | `sql.NullString` | `GetEventByID`, `ListEvents`, `ListMyEvents` |
| Image | `sql.NullString` | `GetEventByID`, `ListEvents`, `ListMyEvents` |
| BannerImage | `sql.NullString` | `GetEventByID` only |
| EventDate | `time.Time` | `GetEventByID`, `ListEvents`, `ListMyEvents` |
| Status | `ModeOfConduct` | `GetEventByID`, `ListEvents` (not in `ListMyEvents`) |

---

### Routes Using Event

| # | Tier | Method | Path | Handler | DB Row Type | Fields Returned | Full/Partial |
|---|------|--------|------|---------|-------------|----------------|-------------|
| 1 | T1 | GET | `/event/` | `HandleGetAllEvent` | `[]ListEventsRow` | `id, name, description, venue, image, event_date, status` | Partial (no banner_image) |
| 2 | T2 | GET | `/event/{id}` | `HandleGetEventById` | `GetEventByIDRow` | `id, name, description, venue, image, banner_image, event_date, status` | Full |
| 3 | T3 | POST | `/event/{id}/join` | `HandleJoinEvent` | `uuid.UUID` | `{ID}` | Scalar |
| 4 | T3 | GET | `/event/u/{id}` | `HandleGetMyEvent` | `uuid.UUID` | `{ID}` | Scalar |
| 5 | T3 | GET | `/event/all` | `HandleGetMyAllEvent` | `[]ListMyEventsRow` | `id, name, description, venue, image, event_date` | Partial (no status, no banner) |
| 6 | T4 | DELETE | `/event/{id}/attendee/{user_id}` | `HandleDeleteEventAttendee` | `uuid.UUID` | `"ok"` | None |
| 7 | T2 | GET | `/event/{id}/attendees` | `HandleAllEventAttendee` | `[]ListEventAttendeesRow` | user fields | Partial (users, not event) |
| 8 | T4 | POST | `/event/` | `HandleCreateEvent` | `uuid.UUID` | `{ID}` | Scalar |
| 9 | T4 | PATCH | `/event/{id}` | `HandleUpdateEvent` | `uuid.UUID` | `{ID}` | Scalar |
| 10 | T4 | DELETE | `/event/{id}` | `HandleDeleteEvent` | `uuid.UUID` | `"ok"` | None |

---

### Event Field Classification (weighted by route frequency)

| Field | Loaded by Routes (Tier) | Estimated Load Frequency | Category |
|-------|------------------------|-------------------------|----------|
| ID | All event routes | Every event interaction | **Always Used** |
| Name | T1 (list), T2 (detail), T3 (my events) | Very frequent | **Always Used** |
| Description | T1 (list), T2 (detail), T3 (my events) | Very frequent | **Frequently Used** |
| Venue | T1 (list), T2 (detail), T3 (my events) | Very frequent | **Frequently Used** |
| Image | T1 (list), T2 (detail), T3 (my events) | Very frequent | **Frequently Used** |
| EventDate | T1 (list), T2 (detail), T3 (my events) | Very frequent | **Frequently Used** |
| Status | T1 (list), T2 (detail) | Frequent | **Frequently Used** |
| BannerImage | T2 (detail only) | Moderate (detail page) | **Occasionally Used** |

| Category | Definition | Fields |
|----------|-----------|--------|
| **Always Used** | Loaded by every tier | `ID`, `Name` |
| **Frequently Used** | Loaded by T1 + T2 + at least one lower tier | `Description`, `Venue`, `Image`, `EventDate`, `Status` |
| **Occasionally Used** | Loaded only by T2 routes or niche views | `BannerImage` |
| **Rarely Used** | Not loaded by any frequent route | *(none)* |

---

## ART

### All Possible Art Fields

| Field | Type | Source |
|-------|------|--------|
| ID | `uuid.UUID` | All art queries |
| Name | `string` | All art entity queries |
| Description | `sql.NullString` | All art entity queries |
| Image | `string` | All art entity queries |
| Tags | `[]string` | `GetArtByID`, `ListArt`, `ListPendingArt`, `ListArtByTag`, `ListArtByTags` |
| Status | `ArtStatus` | `GetArtProfileByID`, `ListPendingArt`, `UpdateArtStatus` |
| UserID | `uuid.UUID` | `GetArtProfileByID`, `ListArt`, `ListArtByTag`, `ListArtByTags` |
| CreatedAt | `time.Time` | `ListPendingArt` only |

Additionally, `GetArtProfileByIDRow` includes joined user fields:

| Joined Field | Type |
|-------------|------|
| Username | `sql.NullString` |
| UserImage | `sql.NullString` |

---

### Routes Using Art

| # | Tier | Method | Path | Handler | DB Row Type | Fields Returned | Full/Partial |
|---|------|--------|------|---------|-------------|----------------|-------------|
| 1 | T2 | GET | `/art/u/{user_id}` | `HandleGetArts` | `[]GetArtByUserRow` | `id, name, description, image` | Partial |
| 2 | T2 | GET | `/art/p/{user_id}/{id}` | `HandleGetArtProfileById` | `GetArtProfileByIDRow` | `id, name, description, image, status, user_id, username, user_image` | Full (with join) |
| 3 | T1 | GET | `/art/{id}` | `HandleGetArtById` | `GetArtByIDRow` | `id, name, description, image, tags` | Partial |
| 4 | T5 | GET | `/art/pending-art` | `HandleGetPendingArt` | `[]ListPendingArtRow` | `id, name, description, image, tags, status, created_at` | Full |
| 5 | T2 | GET | `/art/u/profile/{id}` | `HandlerGetArtistProfile` | `[]GetArtByUserRow` | `id, name, description, image` | Partial |
| 6 | T1 | GET | `/art/` | `HandleGetApprovedArt` | `[]ListArtRow` | `id, name, description, image, tags, user_id` | Partial |
| 7 | T4 | POST | `/art/` | `HandleArtCreation` | `uuid.UUID` | `{ID}` | Scalar |
| 8 | T4 | DELETE | `/art/{id}` | `HandleArtDeletion` | `uuid.UUID` | `"Art Work Deleted"` | None |
| 9 | T4 | PATCH | `/art/{id}` | `HandlerArtUpdation` | `uuid.UUID` | `{ID}` | Scalar |
| 10 | T5 | PATCH | `/admin/arts/{art_id}/status` | `HandlerArtStatus` | `UpdateArtStatusRow` | `id, status` | Partial |
| 11 | T3 | POST | `/art/{art_id}/comment` | `HandleComment` | `ArtComment` | `id, art_id, user_id, comment, created_at` | Metadata |
| 12 | T4 | DELETE | `/art/comment/{id}` | `HandleDeleteComment` | `uuid.UUID` | `"Deleted Successfully"` | None |
| 13 | T3 | POST | `/art/{art_id}/like` | `HandleLike` | `ArtLike` | `id, art_id, user_id, created_at` | Metadata |
| 14 | T3 | POST | `/art/{art_id}/unlike` | `HandleUnLike` | `uuid.UUID` | `"Ok"` | None |
| 15 | T1 | GET | `/art/{id}/comments` | `HandleGetArtComments` | `[]GetArtCommentsByArtIDRow` | `id, art_id, user_id, user_name, user_image, comment, created_at` | Metadata |
| 16 | T1 | GET | `/art/{id}/comments/count` | `HandleGetArtCommentsCount` | `int32` | scalar | Scalar |
| 17 | T1 | GET | `/art/{id}/likes/count` | `HandleGetArtLikeCount` | `int32` | scalar | Scalar |

---

### Art Field Classification (weighted by route frequency)

| Field | Loaded by Routes (Tier) | Estimated Load Frequency | Category |
|-------|------------------------|-------------------------|----------|
| ID | All art routes | Every art interaction | **Always Used** |
| Name | T1 (home, detail), T2 (user gallery, profile), T4 (updates), T5 (pending, admin) | Very frequent | **Always Used** |
| Description | T1 (home, detail), T2 (user gallery, profile), T4 (updates), T5 (pending) | Very frequent | **Frequently Used** |
| Image | T1 (home, detail), T2 (user gallery, profile), T4 (updates), T5 (pending, admin) | Very frequent | **Frequently Used** |
| Tags | T1 (home, detail), T5 (pending) | Frequent | **Frequently Used** |
| UserID | T2 (art profile), T1 (home list) | Frequent | **Occasionally Used** |
| Status | T2 (art profile), T5 (pending, admin patch) | Moderate | **Occasionally Used** |
| CreatedAt | T5 (pending only) | Rare (admin) | **Rarely Used** |

Joined fields from `GetArtProfileByIDRow` (T2 only):

| Joined Field | Loaded by | Estimated Frequency | Category |
|-------------|-----------|-------------------|----------|
| Username | T2 (art profile) | Moderate | **Occasionally Used** |
| UserImage | T2 (art profile) | Moderate | **Occasionally Used** |

| Category | Definition | Fields |
|----------|-----------|--------|
| **Always Used** | Loaded by every art interaction | `ID`, `Name` |
| **Frequently Used** | Loaded by T1 + at least one other frequent tier | `Description`, `Image`, `Tags` |
| **Occasionally Used** | Loaded by T2 or admin views but not home/detail | `UserID`, `Status`, `Username`, `UserImage` |
| **Rarely Used** | Loaded only by very-low-frequency routes (T5) | `CreatedAt` |

---

## SQL QUERY ANALYSIS

### User Queries

| Query Name | Return Type | Returned Columns | Shape | Can Reuse Cached Entity? | Should Bypass Cache? |
|-----------|-------------|-----------------|-------|-------------------------|----------------------|
| `CreateUser` | `uuid.UUID` | scalar ID | N/A | N/A (write) | **Yes** |
| `GetUser` | `GetUserRow` | id, name, username, email, batch, status, role, image, banner_image, description, social_links (11 cols) | Full | **Yes** — `cache.User` has all these fields | No |
| `GetAllUser` | `[]GetAllUserRow` | id, name, email, status, role, image (6 cols) | Partial | Partially — reconstruct from `cache.User` entries | Could use list cache |
| `GetAllUserApproved` | `[]GetAllUserApprovedRow` | id, name, role, description, image, social_links (6 cols) | Partial | Partially — reconstruct from `cache.User` | Could use list cache |
| `GetUserByEmail` | `GetUserByEmailRow` | id, name, email, password, image (5 cols) | Partial (incl. password hash) | No — `cache.User` has no Password | **Yes** (auth, sensitive) |
| `GetUserByUsername` | `GetUserByUsernameRow` | Same as `GetUserRow` (11 cols) | Full | **Yes** | No *(unused)* |
| `PatchUserProfile` | `PatchUserProfileRow` | Same as `GetUserRow` (11 cols) | Full | **Yes** — should update cache on write | **Yes** (write op) |
| `PatchUserPassword` | `uuid.UUID` | scalar ID | N/A | N/A | **Yes** (write, sensitive) |
| `PatchUserAdmin` | `PatchUserAdminRow` | id, status, role (3 cols) | Partial | Yes — `cache.User` has Status, Role | **Yes** (write op) |
| `CheckUsrById` | `CheckUsrByIdRow` | id, status, role (3 cols) | Partial | **Yes** — `cache.User` has Status, Role | No *(unused)* |
| `GetCoreMembers` | `[]GetCoreMembersRow` | id, name, email, status, role, image, social_links (7 cols) | Partial | Partially — reconstruct from `cache.User` | Could use list cache |

---

### Art Queries

| Query Name | Return Type | Returned Columns | Shape | Can Reuse Cached Entity? | Should Bypass Cache? |
|-----------|-------------|-----------------|-------|-------------------------|----------------------|
| `CreateArt` | `uuid.UUID` | scalar ID | N/A | N/A (write) | **Yes** |
| `GetArtByID` | `GetArtByIDRow` | id, name, description, image, tags (5 cols) | Partial | **Yes** — `cache.Art` has name, description, image, tags | No *(handler doesn't read cache)* |
| `GetArtByUser` | `[]GetArtByUserRow` | id, name, description, image (4 cols) | Partial | Partially — aggregate from `cache.Art` | Could use list cache |
| `GetArtProfileByID` | `GetArtProfileByIDRow` | id, name, description, image, status, user_id, username, user_image (8 cols, JOIN) | Joined | Partially — `cache.Art` + `cache.User` | No |
| `ListArt` | `[]ListArtRow` | id, name, description, image, tags, user_id (6 cols) | Partial | Partially — aggregate from `cache.Art` | **High-value list cache candidate** |
| `ListPendingArt` | `[]ListPendingArtRow` | id, name, description, image, tags, status, created_at (7 cols) | Partial | Partially | Low value (dynamic, admin) |
| `ListArtByTag` | `[]ListArtByTagRow` | Same as `ListArtRow` (6 cols) | Partial | Partially | Could use list cache *(unused)* |
| `ListArtByTags` | `[]ListArtByTagsRow` | Same as `ListArtRow` (6 cols) | Partial | Partially | Could use list cache *(unused)* |
| `UpdateArt` | `uuid.UUID` | scalar ID | N/A | N/A (write) | **Yes** |
| `UpdateArtStatus` | `UpdateArtStatusRow` | id, status (2 cols) | Partial | Yes — `cache.Art` has Status | **Yes** (write op) |
| `DeleteArt` | `uuid.UUID` | scalar ID | N/A | N/A (write) | **Yes** |

### Art Metadata Queries

| Query Name | Return Type | Returned Columns | Shape | Can Reuse Cached Entity? | Should Bypass Cache? |
|-----------|-------------|-----------------|-------|-------------------------|----------------------|
| `LikeArt` | `ArtLike` | id, art_id, user_id, created_at | Full row | N/A (write) | **Yes** |
| `UnlikeArt` | `uuid.UUID` | scalar art_id | N/A | N/A (write) | **Yes** |
| `CheckArtLikedByUser` | `bool` | scalar boolean | Scalar | Yes — short-lived boolean | Could use cache |
| `GetArtLikesCount` | `int32` | scalar count | Scalar | **Yes** — excellent candidate (read-only, computed) | No |
| `AddArtComment` | `ArtComment` | id, art_id, user_id, comment, created_at | Full row | N/A (write) | **Yes** |
| `DeleteArtComment` | `uuid.UUID` | scalar ID | N/A | N/A (write) | **Yes** |
| `GetArtCommentsByArtID` | `[]GetArtCommentsByArtIDRow` | id, art_id, user_id, user_name, user_image, comment, created_at (7 cols, JOIN) | Full (with user join) | **Yes** — good list cache candidate (by art_id) | No |
| `GetArtCommentsCount` | `int32` | scalar count | Scalar | **Yes** — excellent candidate (read-only, computed) | No |

### Event Queries

| Query Name | Return Type | Returned Columns | Shape | Can Reuse Cached Entity? | Should Bypass Cache? |
|-----------|-------------|-----------------|-------|-------------------------|----------------------|
| `CreateEvent` | `uuid.UUID` | scalar ID | N/A | N/A (write) | **Yes** |
| `GetEventByID` | `GetEventByIDRow` | id, name, description, venue, image, banner_image, event_date, status (8 cols) | Full | **Yes** — `cache.Event` has all these fields | No *(already cached)* |
| `ListEvents` | `[]ListEventsRow` | id, name, description, venue, image, event_date, status (7 cols, no banner_image) | Partial | Partially — aggregate from `cache.Event` | **High-value list cache candidate** |
| `ListUpcomingEvents` | `[]ListUpcomingEventsRow` | Same as `ListEventsRow` (7 cols) | Partial | Partially | Could use list cache *(unused)* |
| `ListEventsByMode` | `[]ListEventsByModeRow` | Same as `ListEventsRow` (7 cols) | Partial | Partially | Could use list cache *(unused)* |
| `UpdateEvent` | `uuid.UUID` | scalar ID | N/A | N/A (write) | **Yes** |
| `DeleteEvent` | `uuid.UUID` | scalar ID | N/A | N/A (write) | **Yes** |

### Event Attendees Queries

| Query Name | Return Type | Returned Columns | Shape | Can Reuse Cached Entity? | Should Bypass Cache? |
|-----------|-------------|-----------------|-------|-------------------------|----------------------|
| `EnrollUserToEvent` | `uuid.UUID` | scalar ID | N/A | N/A (write) | **Yes** |
| `RemoveUserFromEvent` | `uuid.UUID` | scalar event_id | N/A | N/A (write) | **Yes** |
| `ListEventAttendees` | `[]ListEventAttendeesRow` | id, name, username, email, image (5 user cols) | Partial (user subset) | Partially — reconstruct from `cache.User` | Could use list cache |
| `CountEventAttendees` | `int32` | scalar count | Scalar | **Yes** — good candidate *(unused)* | No |
| `ListMyEvents` | `[]ListMyEventsRow` | id, name, description, venue, image, event_date (6 cols, no status) | Partial | Partially — from `cache.Event` | Could use per-user cache |
| `GetMyEventById` | `uuid.UUID` | scalar event ID | Scalar | Low value (existence check) | Could use short-lived cache |

---

## SUMMARY: UNUSED QUERIES

The following sqlc queries are defined in repository interfaces but never called by any handler:

| Query | Repository | Tier if Used | Purpose |
|-------|-----------|-------------|---------|
| `CheckUsrById` | UserRepository | Would be T1/T2 | Quick user status/role lookup (cache would serve this) |
| `GetUserByUsername` | UserRepository | T2 | Full user lookup by username |
| `ListArtByTag` | ArtRepository | T2 | List approved art matching a single tag |
| `ListArtByTags` | ArtRepository | T1 | List approved art matching any of tags (browse) |
| `ListEventsByMode` | EventRepository | T2 | List events filtered by online/offline |
| `ListUpcomingEvents` | EventRepository | T1 | List future events |
| `CountEventAttendees` | EventAttendeesRepository | T2 | Count attendees for an event |

---

## SUMMARY: CACHE IMPLEMENTATION GAPS

### Already Cached

| Entity | Cache Type | Cached By | Read By | TTL |
|--------|-----------|-----------|---------|-----|
| User | `map[uuid.UUID]UserCache` | `HandleGetUserById`, `HandleMe` (on miss) | `HandleGetUserById`, `HandleMe` | 7 days (renewed on access) |
| Event | `map[uuid.UUID]EventCache` | `HandleGetEventById` (on miss) | `HandleGetEventById` | 7 days (renewed on access) |
| Art | `map[uuid.UUID]ArtCache` | `HandleArtCreation` only | *(never read)* | 7 days (renewed on access) |

### Not Cached But Excellent Candidates (T1 routes)

| Route | Entity | Estimated Impact | Notes |
|-------|--------|-----------------|-------|
| `GET /art/` | Art list (T1) | **Highest** — homepage, every visit | 6 cols per entry, stable "approved" filter |
| `GET /event/` | Event list (T1) | **High** — events page | 7 cols per entry, infrequent updates |
| `GET /art/{id}` | Single art (T1) | **High** — art detail | `cache.Art` exists but is never read |
| `GET /art/{id}/comments` | Comment list (T1) | **High** — art detail | Short TTL acceptable |
| `GET /art/{id}/likes/count` | Scalar (T1) | **Medium** — art detail | Computed, trivial to cache |
| `GET /art/{id}/comments/count` | Scalar (T1) | **Medium** — art detail | Computed, trivial to cache |

### Not Cached But Good Candidates (T2 routes)

| Route | Entity | Impact | Notes |
|-------|--------|--------|-------|
| `GET /auth/users/{id}` | Single user (T2) | High | Already cached in code! But handler reads cache directly |
| `GET /auth/core-member` | User list (T2) | Medium | Very stable data, rarely changes |
| `GET /auth/main-users` | User list (T2) | Medium | Stable, only changes on new member approval |
| `GET /event/{id}/attendees` | User list (T2) | Medium | Per-event attendee list |
| `GET /art/u/profile/{id}` | Composite (T2) | Medium | Eliminates 2 sequential DB calls |
| `GET /art/p/{user_id}/{id}` | Joined art (T2) | Medium | Composite art+user query |

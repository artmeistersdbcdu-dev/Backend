-- name: CreateEvent :one
INSERT INTO events (
    id, name, description, venue, image, banner_image, event_date, status
)
VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING id;

-- name: GetEventByID :one
SELECT
    id,
    name,
    description,
    venue,
    image,
    banner_image,
    event_date,
    status
FROM events
WHERE id = $1;


-- name: ListEvents :many
SELECT
    id,
    name,
    description,
    venue,
    image,
    event_date,
    status
FROM events
ORDER BY created_at DESC;


-- name: ListUpcomingEvents :many
SELECT
    id,
    name,
    description,
    venue,
    image,
    event_date,
    status
FROM events
WHERE event_date >= CURRENT_DATE
ORDER BY event_date ASC;


-- name: ListEventsByMode :many
SELECT
    id,
    name,
    description,
    venue,
    image,
    event_date,
    status
FROM events
WHERE status = $1
ORDER BY event_date ASC;


-- name: UpdateEvent :one
UPDATE events
SET
    name = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    venue = COALESCE(sqlc.narg('venue'), venue),
    image = COALESCE(sqlc.narg('image'), image),
    banner_image = COALESCE(sqlc.narg('banner_image'), banner_image),
    event_date = COALESCE(sqlc.narg('event_date'), event_date),
    status = COALESCE(sqlc.narg('status'), status),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING id;


-- name: DeleteEvent :one
DELETE FROM events
WHERE id = $1
RETURNING id;

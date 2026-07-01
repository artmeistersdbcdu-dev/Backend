-- name: EnrollUserToEvent :one
INSERT INTO event_attendees (
    id, event_id, user_id
)
VALUES (
    $1, $2, $3
)
RETURNING id;


-- name: RemoveUserFromEvent :one
DELETE FROM event_attendees
WHERE event_id = $1 AND user_id = $2
RETURNING event_id;


-- name: ListEventAttendees :many
SELECT 
    u.id,
    u.name,
    u.username,
    u.email,
    u.image
FROM users u
JOIN event_attendees ea ON ea.user_id = u.id
WHERE ea.event_id = $1
ORDER BY ea.joined_at ASC;

-- name: CountEventAttendees :one
SELECT COUNT(*)::int
FROM event_attendees
WHERE event_id = $1;


-- name: ListMyEvents :many
SELECT
    e.id,
    e.name,
    e.description,
    e.venue,
    e.image,
    e.event_date
FROM events e
JOIN event_attendees ea ON ea.event_id = e.id
WHERE ea.user_id = $1
ORDER BY e.event_date ASC;

-- name: GetMyEventById :one
SELECT e.id
FROM events e
JOIN event_attendees ea ON ea.event_id = e.id
WHERE ea.user_id = $1
AND ea.event_id = $2
LIMIT 1;

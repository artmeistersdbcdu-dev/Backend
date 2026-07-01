-- name: CreateArt :one
INSERT INTO art (id, name, description, image, tags, user_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;

-- name: GetArtProfileByID :one
SELECT
    a.id,
    a.name,
    a.description,
    a.image,
    a.status,
    a.user_id,
    u.username,
    u.image AS user_image
FROM art a
JOIN users u ON a.user_id = u.id
WHERE a.id = $1
AND a.user_id = $2;

-- name: GetArtByID :one
SELECT
  id,
  name,
  description,
  image,
  tags
FROM art
WHERE id = $1;


-- name: GetArtByUser :many
SELECT 
    id,
    name,
    description,
    image
FROM art
WHERE user_id = $1
ORDER BY created_at DESC;


-- name: ListPendingArt :many
SELECT
    id,
    name,
    description,
    image,
    tags,
    status,
    created_at
FROM art
WHERE status = 'pending'
ORDER BY created_at DESC;
-- name: ListArt :many
SELECT
    id,
    name,
    description,
    image,
    tags,
    user_id
FROM art
WHERE status = 'approved'
ORDER BY created_at DESC;


-- name: ListArtByTag :many
SELECT
    id,
    name,
    description,
    image,
    tags,
    user_id
FROM art
WHERE status = 'approved'
  AND $1 = ANY(tags)
ORDER BY created_at DESC;


-- name: ListArtByTags :many
SELECT
    id,
    name,
    description,
    image,
    tags,
    user_id
FROM art
WHERE status = 'approved'
  AND tags && $1::text[]
ORDER BY created_at DESC;


-- name: UpdateArt :one
UPDATE art
SET
    name = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    tags = COALESCE(sqlc.narg('tags'), tags),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
AND user_id = sqlc.arg('user_id')
RETURNING id;

-- name: UpdateArtStatus :one
UPDATE art
SET
    status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING id, status;

-- name: DeleteArt :one
DELETE FROM art
WHERE id = $1 AND user_id = $2
RETURNING id;

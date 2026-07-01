-- name: CreateUser :one
INSERT INTO users (
    name,
    email,
    password,
    batch
)
VALUES (
    sqlc.arg('name'),
    sqlc.arg('email'),
    sqlc.arg('password'),
    COALESCE(sqlc.narg('batch'), '')
)
RETURNING id;


-- name: GetUser :one
SELECT 
    id,
    name,
    username,
    email,
    batch,
    status,
    role,
    image,
    banner_image,
    description,
    social_links
FROM users
WHERE id = $1;


-- name: GetAllUser :many
SELECT 
    id,
    name,
    email,
    status,
    role,
    image
FROM users;
-- name: GetAllUserApproved :many
SELECT 
    id,
    name,
    role,
    description,
    image,
    social_links
FROM users WHERE status='approved';


-- name: GetUserByEmail :one
SELECT 
    id,
    name,
    email,
    password,
    image
FROM users
WHERE email = $1;


-- name: GetUserByUsername :one
SELECT 
    id,
    name,
    username,
    email,
    batch,
    status,
    role,
    image,
    banner_image,
    description,
    social_links
FROM users
WHERE username = $1;


-- name: PatchUserProfile :one
UPDATE users
SET
    name = COALESCE(sqlc.narg('name')::text, name),
    username = COALESCE(sqlc.narg('username')::text, username),
    email = COALESCE(sqlc.narg('email')::text, email),
    batch = COALESCE(sqlc.narg('batch')::text, batch),
    description = COALESCE(sqlc.narg('description')::text, description),
    image = COALESCE(sqlc.narg('image')::text, image),
    banner_image = COALESCE(sqlc.narg('banner_image')::text, banner_image),
    social_links = COALESCE(sqlc.narg('social_links')::jsonb, social_links),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING
    id,
    name,
    username,
    email,
    batch,
    status,
    role,
    image,
    banner_image,
    description,
    social_links;



-- name: PatchUserPassword :one
UPDATE users
SET
    password = sqlc.arg('password'),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING id;




-- name: PatchUserAdmin :one
UPDATE users
SET
    status = COALESCE(sqlc.narg('status')::account_status, status),
    role = COALESCE(sqlc.narg('role')::user_role, role),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING
    id,
    status,
    role;


-- name: CheckUsrById :one
SELECT 
    id,
    status,
    role
FROM users
WHERE id = $1;


-- name: GetCoreMembers :many
SELECT 
    id,
    name,
    email,
    status,
    role,
    image
FROM users
WHERE role != 'member'
  AND status != 'banned';

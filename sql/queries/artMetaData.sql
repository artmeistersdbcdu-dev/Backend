-- name: LikeArt :one
INSERT INTO art_likes (art_id, user_id)
VALUES ($1, $2)
RETURNING *;


-- name: UnlikeArt :one
DELETE FROM art_likes
WHERE art_id = $1 AND user_id = $2
RETURNING art_id;


-- name: CheckArtLikedByUser :one
SELECT EXISTS (
    SELECT 1 FROM art_likes
    WHERE art_id = $1 AND user_id = $2
) AS liked;


-- name: GetArtLikesCount :one
SELECT COUNT(*)::int AS likes_count
FROM art_likes
WHERE art_id = $1;


-- name: AddArtComment :one
INSERT INTO art_comments (art_id, user_id, comment)
VALUES ($1, $2, $3)
RETURNING *;


-- name: DeleteArtComment :one
DELETE FROM art_comments
WHERE id = $1 AND user_id = $2
RETURNING id;


-- name: GetArtCommentsByArtID :many
SELECT 
    ac.id,
    ac.art_id,
    ac.user_id,
    u.name AS user_name,
    u.image AS user_image,
    ac.comment,
    ac.created_at
FROM art_comments ac
JOIN users u ON u.id = ac.user_id
WHERE ac.art_id = $1
ORDER BY ac.created_at DESC;


-- name: GetArtCommentsCount :one
SELECT COUNT(*)::int AS comments_count
FROM art_comments
WHERE art_id = $1;

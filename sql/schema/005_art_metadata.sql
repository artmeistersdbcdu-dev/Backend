-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS art_likes (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    art_id uuid NOT NULL REFERENCES art(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    UNIQUE (art_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_art_likes_art_id ON art_likes(art_id);
CREATE INDEX IF NOT EXISTS idx_art_likes_user_id ON art_likes(user_id);

CREATE TABLE IF NOT EXISTS art_comments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    art_id uuid NOT NULL REFERENCES art(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    comment TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_art_comments_art_id ON art_comments(art_id);
CREATE INDEX IF NOT EXISTS idx_art_comments_user_id ON art_comments(user_id);

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS art_comments;
DROP TABLE IF EXISTS art_likes;

-- +goose StatementEnd

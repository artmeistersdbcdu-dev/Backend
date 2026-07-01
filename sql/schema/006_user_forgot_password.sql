-- +goose Up
ALTER TABLE users ADD COLUMN reset_token TEXT;
ALTER TABLE users ADD COLUMN reset_token_expires TIMESTAMP;

-- +goose Down
ALTER TABLE users DROP COLUMN reset_token;
ALTER TABLE users DROP COLUMN reset_token_expires;

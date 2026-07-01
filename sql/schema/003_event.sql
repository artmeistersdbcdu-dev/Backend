-- +goose Up
CREATE TYPE public.mode_of_conduct AS ENUM ('online', 'offline');

CREATE TABLE events (
    id uuid PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    venue TEXT,
    image TEXT,
    banner_image TEXT,
    event_date DATE NOT NULL,
    status public.mode_of_conduct NOT NULL DEFAULT 'offline',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_events_created_at ON events(created_at);
CREATE INDEX idx_events_status ON events(status);

-- +goose Down
DROP TABLE IF EXISTS events;
DROP TYPE IF EXISTS public.mode_of_conduct;

-- +goose Up
CREATE TYPE public.user_role_new AS ENUM (
    'president',
    'vice_president',
    'general_secretary',
    'logistic',
    'social_media_head',
    'content_head',
    'core_member',
    'member'
);

ALTER TABLE users ALTER COLUMN role DROP DEFAULT;

ALTER TABLE users
    ALTER COLUMN role TYPE public.user_role_new
    USING (CASE role::text
        WHEN 'admin' THEN 'president'::public.user_role_new
        ELSE 'member'::public.user_role_new
    END);

ALTER TABLE users ALTER COLUMN role SET DEFAULT 'member';

DROP TYPE IF EXISTS public.user_role;

ALTER TYPE public.user_role_new RENAME TO user_role;

-- +goose Down
CREATE TYPE public.user_role_old AS ENUM ('admin', 'user');

ALTER TABLE users ALTER COLUMN role DROP DEFAULT;

ALTER TABLE users
    ALTER COLUMN role TYPE public.user_role_old
    USING (CASE role::text
        WHEN 'president' THEN 'admin'::public.user_role_old
        ELSE 'user'::public.user_role_old
    END);

ALTER TABLE users ALTER COLUMN role SET DEFAULT 'user';

DROP TYPE IF EXISTS public.user_role;

ALTER TYPE public.user_role_old RENAME TO user_role;

-- +goose Up
CREATE TYPE user_role AS ENUM (
    'USER',
    'ADMIN'
);

ALTER TABLE users
    ADD COLUMN role user_role NOT NULL DEFAULT 'USER';

UPDATE users
SET role = 'ADMIN'
WHERE is_admin = TRUE;

ALTER TABLE users
    DROP COLUMN is_admin;

CREATE INDEX users_role_admin_idx
    ON users (id)
    WHERE role = 'ADMIN' AND is_deleted = FALSE;

COMMENT ON COLUMN users.role IS 'Platform role: USER or ADMIN. ADMIN is bootstrapped from ADMIN_USERNAMES.';

-- +goose Down
ALTER TABLE users
    ADD COLUMN is_admin BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE users
SET is_admin = TRUE
WHERE role = 'ADMIN';

DROP INDEX IF EXISTS users_role_admin_idx;

ALTER TABLE users
    DROP COLUMN role;

DROP TYPE IF EXISTS user_role;

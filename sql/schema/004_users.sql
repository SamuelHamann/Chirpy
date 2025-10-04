-- +goose up
ALTER TABLE users
ADD COLUMN IF NOT EXISTS is_chirpy_red BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose down
ALTER TABLE users
DROP COLUMN IF EXISTS is_chirpy_red;
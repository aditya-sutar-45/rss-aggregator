-- +goose UP
ALTER TABLE users ADD CONSTRAINT users_username_key UNIQUE (username);
ALTER TABLE users ADD COLUMN password_hash TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE users DROP CONSTRAINT users_username_key;
ALTER TABLE users DROP COLUMN password_hash;

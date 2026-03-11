CREATE TABLE IF NOT EXISTS users (
    id                   UUID PRIMARY KEY,
    prep_user_id         BIGINT NOT NULL UNIQUE,
    email                VARCHAR(255) NOT NULL DEFAULT '',
    name                 VARCHAR(255) NOT NULL DEFAULT '',
    created_at           TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at           TIMESTAMPTZ
);

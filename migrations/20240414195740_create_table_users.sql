-- +goose Up
CREATE TYPE access_level AS ENUM ('user', 'admin');
CREATE TABLE users
(
    id                 BIGSERIAL PRIMARY KEY,
    access_level access_level          NOT NULL
);
-- +goose Down
DROP TABLE users;
DROP TYPE access_level;

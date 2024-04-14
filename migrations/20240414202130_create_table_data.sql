-- +goose Up
CREATE TABLE data
(
    id                 BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users,
    data TEXT
);

-- +goose Down
DROP table data;

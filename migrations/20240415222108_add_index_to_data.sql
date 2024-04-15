-- +goose NO TRANSACTION
-- +goose Up
CREATE INDEX CONCURRENTLY data_user_id_idx ON data (user_id);

-- +goose Down
DROP INDEX CONCURRENTLY data_user_id_idx;

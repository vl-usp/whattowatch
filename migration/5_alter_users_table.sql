-- +goose Up
-- +goose StatementBegin
alter table users alter column id type bigint;
alter table user_favorites alter column user_id type bigint;
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
-- alter table users alter column id type bigint;
-- alter table user_favorites alter column user_id type bigint;
-- +goose StatementEnd
-- +goose Up
-- +goose StatementBegin
alter table genres add column if not exists formatted_name text;
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
alter table genres drop column if exists formatted_name;
-- +goose StatementEnd
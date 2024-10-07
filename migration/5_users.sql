-- +goose Up
-- +goose StatementBegin
create table if not exists public.users (
	id           bigint primary key,
	first_name    text,
	last_name     text,
	username      text,
	language_code text,
	created_at    timestamptz,
	updated_at    timestamptz,
	deleted_at    timestamptz
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists public.users;
-- +goose StatementEnd

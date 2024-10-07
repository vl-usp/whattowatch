-- +goose Up
-- +goose StatementBegin
create table if not exists public.genres (
	id uuid primary key,
	name text not null,
	pretty_name text
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists public.genres;
-- +goose StatementEnd

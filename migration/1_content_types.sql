-- +goose Up
-- +goose StatementBegin
create table if not exists public.content_types (
	id serial primary key,
	name text
);

insert into public.content_types (name) values ('movie'), ('tv');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists public.content_types;
-- +goose StatementEnd

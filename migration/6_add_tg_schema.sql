-- +goose Up
-- +goose StatementBegin
create schema if not exists tg;

create table if not exists tg.users (
	id           int primary key,
	first_name    text,
	last_name     text,
	username      text,
	language_code text,
	created_at    timestamptz,
	updated_at    timestamptz,
	deleted_at    timestamptz
);

create table if not exists tg.favorite_types (
	id serial primary key,
	favorite_type text
);

insert into tg.favorite_types (favorite_type) values ('movie'), ('tv'), ('genre');

create table if not exists tg.user_favorites (
	id serial primary key,
	user_id int not null,
	favorite_type_id int not null,
	favorite_id int not null
);

ALTER TABLE tg.user_favorites ADD CONSTRAINT tg_fk_user_favorites_user_id FOREIGN KEY (user_id) REFERENCES tg.users(id) ON DELETE CASCADE;
ALTER TABLE tg.user_favorites ADD CONSTRAINT tg_fk_user_favorites_favorite_type_id FOREIGN KEY (favorite_type_id) REFERENCES tg.favorite_types(id) ON DELETE CASCADE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop schema if exists tg cascade;
-- +goose StatementEnd

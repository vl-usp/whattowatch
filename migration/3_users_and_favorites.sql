-- +goose Up
-- +goose StatementBegin
create table if not exists public.users (
	id           int primary key,
	first_name    text,
	last_name     text,
	username      text,
	language_code text,
	created_at    timestamptz,
	updated_at    timestamptz,
	deleted_at    timestamptz
);

create table if not exists public.user_favorites (
	id serial primary key,
	user_id int not null,
	favorite_id uuid not null,
	unique(user_id, favorite_id)
);


ALTER TABLE public.user_favorites ADD CONSTRAINT fk_user_favorites_user_id FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists public.user_favorites;
drop table if exists public.users;
-- +goose StatementEnd

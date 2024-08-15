-- +goose Up
-- +goose StatementBegin
create table if not exists public.genres (
	id uuid primary key,
	tmdb_id int unique not null,
	name text not null,
	slug text
);

create table if not exists public.content_genres (
	id serial primary key,
	content_id uuid,
	genre_id uuid
);

ALTER TABLE public.content_genres ADD CONSTRAINT fk_conent_genres_content_id FOREIGN KEY (content_id) REFERENCES public.content(id) ON DELETE CASCADE;
ALTER TABLE public.content_genres ADD CONSTRAINT fk_conent_genres_genre_id FOREIGN KEY (genre_id) REFERENCES public.genres(id) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists content_genres;
drop table if exists genres;
-- +goose StatementEnd

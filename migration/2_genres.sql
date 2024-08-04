-- +goose Up
-- +goose StatementBegin
create table if not exists public.film_genres (
	id uuid primary key,
	tmdb_id int unique not null,
	name text not null,
	slug text
);

create table if not exists public.film_content_genres (
	id serial primary key,
	film_content_id uuid,
	film_genre_id uuid
);

ALTER TABLE public.film_content_genres ADD CONSTRAINT fk_film_conent_genres_film_content_id FOREIGN KEY (film_content_id) REFERENCES public.film_content(id) ON DELETE CASCADE;
ALTER TABLE public.film_content_genres ADD CONSTRAINT fk_film_conent_genres_film_genre_id FOREIGN KEY (film_genre_id) REFERENCES public.film_genres(id) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists film_content_genres;
drop table if exists film_genres;
-- +goose StatementEnd

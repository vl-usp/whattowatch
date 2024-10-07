-- +goose Up
-- +goose StatementBegin
create table if not exists public.link_tmdb_genres (
	id serial primary key,
	genre_id uuid not null,
	tmdb_genre_id int not null,
	unique(genre_id, tmdb_genre_id),
	constraint fk_link_tmdb_genres_genre_id foreign key (genre_id) references public.genres(id) on delete cascade
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists public.tmdb_content;
-- +goose StatementEnd

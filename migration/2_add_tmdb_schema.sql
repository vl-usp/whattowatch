-- +goose Up
-- +goose StatementBegin
create schema if not exists tmdb;

create table if not exists tmdb.movies (
	id int primary key,
	tmdb_id text,
	title text not null,
	overview text,
	popularity numeric,
	poster_path text,
	release_date date,
	budget bigint,
	revenue bigint,
	runtime int,
	vote_average numeric,
	vote_count int
);

create table if not exists tmdb.genres (
	id int primary key,
	name text not null
);

create table if not exists tmdb.movies_genres (
	id serial primary key,
	movie_id int,
	genre_id int
);

ALTER TABLE tmdb.movies_genres DROP CONSTRAINT if exists tmdb_fk_movies_genres_genre_id;
ALTER TABLE tmdb.movies_genres ADD CONSTRAINT tmdb_fk_movies_genres_genre_id FOREIGN KEY (genre_id) REFERENCES tmdb.genres(id) ON DELETE CASCADE;
ALTER TABLE tmdb.movies_genres DROP CONSTRAINT if exists tmdb_fk_movies_genres_movie_id;
ALTER TABLE tmdb.movies_genres ADD CONSTRAINT tmdb_fk_movies_genres_movie_id FOREIGN KEY (movie_id) REFERENCES tmdb.movies(id) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop schema if exists "tmdb";
-- +goose StatementEnd

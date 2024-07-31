-- +goose Up
-- +goose StatementBegin
alter table tmdb.genres add column if not exists ru_name text;

ALTER TABLE tmdb.movies_genres DROP CONSTRAINT if exists tmdb_fk_movies_genres_genre_id;
ALTER TABLE tmdb.movies_genres ADD CONSTRAINT tmdb_fk_movies_genres_genre_id FOREIGN KEY (genre_id) REFERENCES tmdb.genres(id) ON DELETE CASCADE;
ALTER TABLE tmdb.movies_genres DROP CONSTRAINT if exists tmdb_fk_movies_genres_movie_id;
ALTER TABLE tmdb.movies_genres ADD CONSTRAINT tmdb_fk_movies_genres_movie_id FOREIGN KEY (movie_id) REFERENCES tmdb.movies(id) ON DELETE CASCADE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table tmdb.genres drop column if exists ru_name;
-- +goose StatementEnd

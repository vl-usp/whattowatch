-- +goose Up
-- +goose StatementBegin
create table if not exists tmdb.tvs (
	id           int primary key,
	title        text,
	overview     text,
	popularity   numeric,
	poster_path   text,
	first_air_date date,
	vote_average  numeric,
	vote_count    int
);

create table if not exists tmdb.tvs_genres (
	id serial primary key,
	tv_id int,
	genre_id int
);


ALTER TABLE tmdb.tvs_genres ADD CONSTRAINT tmdb_fk_tvs_genres_genre_id FOREIGN KEY (genre_id) REFERENCES tmdb.genres(id) ON DELETE CASCADE;
ALTER TABLE tmdb.tvs_genres ADD CONSTRAINT tmdb_fk_tvs_genres_tv_id FOREIGN KEY (tv_id) REFERENCES tmdb.tvs(id) ON DELETE CASCADE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists tmdb.tvs_genres;
drop table if exists tmdb.tvs;
-- +goose StatementEnd

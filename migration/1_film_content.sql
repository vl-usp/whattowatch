-- +goose Up
-- +goose StatementBegin
create table if not exists public.film_content_type (
	id serial primary key,
	name text
);

insert into public.film_content_type (name) values ('movie'), ('tv');

create table if not exists public.film_content (
	id uuid primary key,
	tmdb_id int unique not null,
	film_content_type_id int not null,
	title text not null,
	overview text,
	popularity numeric,
	poster_path text,
	release_date date,
	vote_average numeric,
	vote_count int
);

alter table public.film_content add constraint public_fk_film_content_film_content_type_id foreign key (film_content_type_id) references public.film_content_type(id) on delete cascade;

create index if not exists film_content_title_idx on public.film_content (title);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS film_content_type;
DROP TABLE IF EXISTS film_content;
-- +goose StatementEnd

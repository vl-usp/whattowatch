-- +goose Up
-- +goose StatementBegin
create table if not exists public.content_type (
	id serial primary key,
	name text
);

insert into public.content_type (name) values ('movie'), ('tv');

create table if not exists public.content (
	id uuid primary key,
	tmdb_id int unique not null,
	content_type_id int not null,
	title text not null,
	overview text,
	popularity numeric,
	poster_path text,
	release_date date,
	vote_average numeric,
	vote_count int
);

alter table public.content add constraint public_fk_content_content_type_id foreign key (content_type_id) references public.content_type(id) on delete cascade;

create index if not exists content_title_idx on public.content (title);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS content_type;
DROP TABLE IF EXISTS content;
-- +goose StatementEnd

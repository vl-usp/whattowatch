-- +goose Up
-- +goose StatementBegin
create table if not exists public.link_tmdb_content (
	id serial primary key,
	content_id uuid not null,
	tmdb_content_id int not null,
	unique(content_id, tmdb_content_id),
	constraint fk_link_tmdb_content_content_id foreign key (content_id) references public.content(id) on delete cascade
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists public.tmdb_content;
-- +goose StatementEnd

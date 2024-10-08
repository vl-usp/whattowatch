-- +goose Up
-- +goose StatementBegin
create table if not exists public.content (
	id int primary key,
	content_type_id int not null,
	title text unique not null,
	overview text,
	popularity numeric,
	poster_path text,
	release_date date,
	vote_average numeric,
	vote_count int,
	constraint fk_content_content_types_id foreign key (content_type_id) references public.content_types(id) on delete cascade
);

create index if not exists content_title_idx on public.content (title);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists public.content;
-- +goose StatementEnd

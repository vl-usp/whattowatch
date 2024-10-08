-- +goose Up
-- +goose StatementBegin
create table if not exists public.link_content_genres (
	id serial primary key,
	content_id int not null,
	genre_id int not null,
	unique(content_id, genre_id),
	constraint fk_link_conent_genres_content_id foreign key (content_id) references public.content(id) on delete cascade,
	constraint fk_link_conent_genres_genre_id foreign key (genre_id) references public.genres(id) on delete cascade
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists public.link_content_genres;
-- +goose StatementEnd

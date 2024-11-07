-- +goose Up
-- +goose StatementBegin
create table if not exists public.content (
	id int not null,
	content_type_id int not null,
	title text not null,
	popularity numeric,
	constraint fk_content_content_types_id foreign key (content_type_id) references public.content_types(id) on delete cascade,
	primary key (id, content_type_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists public.content;
-- +goose StatementEnd

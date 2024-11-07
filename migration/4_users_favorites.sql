-- +goose Up
-- +goose StatementBegin
create table if not exists public.users_favorites (
	id serial primary key,
	user_id bigint not null,
	content_id int not null,
	content_type_id int not null,
	unique(user_id, content_id),
	constraint public_fk_users_favorites_user_id foreign key (user_id) references public.users(id) on delete cascade,
	constraint public_fk_users_favorites_content_id foreign key (content_id, content_type_id) references public.content(id, content_type_id) on delete cascade
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists public.users_favorites;
-- +goose StatementEnd

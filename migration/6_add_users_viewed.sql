-- +goose Up
-- +goose StatementBegin
create table if not exists public.users_viewed (
	id serial primary key,
	user_id uuid not null,
	content_id uuid not null,
	constraint public_fk_users_viewed_user_id foreign key (user_id) references public.users(id) on delete cascade,
	constraint public_fk_users_viewed_content_id foreign key (content_id) references public.content(id) on delete cascade
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists public.users_viewed;
-- +goose StatementEnd

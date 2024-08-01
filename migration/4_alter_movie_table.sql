-- +goose Up
-- +goose StatementBegin
alter table tmdb.movies drop column if exists budget;
alter table tmdb.movies drop column if exists revenue;
alter table tmdb.movies drop column if exists runtime;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd

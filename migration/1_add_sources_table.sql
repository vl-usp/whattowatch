-- +goose Up
-- +goose StatementBegin
CREATE TABLE if not exists "sources" (
  "id" int PRIMARY KEY,
  "name" text NOT NULL,
  "url" text NOT NULL
);

INSERT INTO "sources" ("id", "name", "url") VALUES 
(1, 'Kinopoisk', 'https://kinopoiskapiunofficial.tech/api'),
(2, 'TMDb', 'https://api.themoviedb.org/3/')
on conflict do nothing;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "sources";
-- +goose StatementEnd

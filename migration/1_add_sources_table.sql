-- +goose Up
-- +goose StatementBegin
CREATE TABLE "sources" (
  "id" int PRIMARY KEY,
  "name" text NOT NULL,
  "url" text NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT 'now()',
  "updated_at" timestamp,
  "deleted_at" timestamp
);

INSERT INTO "sources" ("id", "name", "hostname") VALUES 
(1, 'Kinopoisk', 'https://kinopoiskapiunofficial.tech/api'),
(2, 'TMDb', 'https://api.themoviedb.org/3/');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "sources";
-- +goose StatementEnd

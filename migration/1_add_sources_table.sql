-- +goose Up
-- +goose StatementBegin
CREATE TABLE "sources" (
  "id" int PRIMARY KEY,
  "name" text NOT NULL,
  "parse_url" text NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT 'now()',
  "updated_at" timestamp,
  "deleted_at" timestamp
);


INSERT INTO "sources" ("id", "name", "parse_url") VALUES 
(1, 'TMDb', 'https://www.themoviedb.org/discover/movie/items?language=ru-RU');

INSERT INTO "sources" ("name", "hostname") VALUES 
('TMDb', 'https://www.themoviedb.org/discover/movie/items?language=ru-RU');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "sources";
-- +goose StatementEnd

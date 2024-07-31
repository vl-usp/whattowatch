-- +goose Up
-- +goose StatementBegin
CREATE TABLE "genres" (
  "id" serial PRIMARY KEY,
  "name" text NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT 'now()',
  "updated_at" timestamp,
  "deleted_at" timestamp
);

CREATE TABLE "link_movies_genres" (
  "id" serial PRIMARY KEY,
  "movie_id" int NOT NULL,
  "genre_id" int NOT NULL
);

ALTER TABLE "link_movies_genres" ADD FOREIGN KEY ("movie_id") REFERENCES "movies" ("id");

ALTER TABLE "link_movies_genres" ADD FOREIGN KEY ("genre_id") REFERENCES "genres" ("id");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "genres";
DROP TABLE IF EXISTS "link_movies_genres";
-- +goose StatementEnd

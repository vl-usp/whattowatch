-- +goose Up
-- +goose StatementBegin
CREATE TABLE "categories" (
  "id" serial PRIMARY KEY,
  "name" text NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT 'now()',
  "updated_at" timestamp,
  "deleted_at" timestamp
);

CREATE TABLE "link_movies_categories" (
  "id" serial PRIMARY KEY,
  "movie_id" int NOT NULL,
  "category_id" int NOT NULL
);

ALTER TABLE "link_movies_categories" ADD FOREIGN KEY ("movie_id") REFERENCES "movies" ("id");

ALTER TABLE "link_movies_categories" ADD FOREIGN KEY ("category_id") REFERENCES "categories" ("id");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "categories";
DROP TABLE IF EXISTS "link_movies_categories";
-- +goose StatementEnd

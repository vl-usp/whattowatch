-- +goose Up
-- +goose StatementBegin
CREATE TABLE "movies" (
  "id" serial PRIMARY KEY,
  "source_id" int NOT NULL,
  "title" text NOT NULL,
  "description" text NOT NULL,
  "runtime" interval,
  "release_year" date,
  "created_at" timestamp NOT NULL DEFAULT 'now()',
  "updated_at" timestamp,
  "deleted_at" timestamp
);

CREATE INDEX ON "movies" ("title");

CREATE INDEX ON "movies" ("release_year");

ALTER TABLE "movies" ADD FOREIGN KEY ("source_id") REFERENCES "sources" ("id");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "movies";
-- +goose StatementEnd

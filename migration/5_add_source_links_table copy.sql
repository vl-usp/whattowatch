-- +goose Up
-- +goose StatementBegin
CREATE TABLE "source_links" (
	"id" serial PRIMARY KEY,
	"original_id" int NOT NULL UNIQUE,
	"source_id" int NOT NULL,
	"page" int NOT NULL,
	"title" text NOT NULL,
	"movie_url" text NOT NULL,
	"created_at" timestamp NOT NULL DEFAULT 'now()',
	"updated_at" timestamp,
	"deleted_at" timestamp
);

ALTER TABLE "source_links" ADD FOREIGN KEY ("source_id") REFERENCES "sources" ("id");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "source_links";
-- +goose StatementEnd

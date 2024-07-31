-- +goose Up
-- +goose StatementBegin
INSERT INTO "sources" ("name", "hostname") VALUES 
('Kinopoisk', 'kinopoisk.ru');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM "sources"
WHERE "hostname" = 'kinopoisk.ru';
-- +goose StatementEnd

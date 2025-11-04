-- +goose Up
-- +goose StatementBegin
CREATE TABLE "urls" (
    "key" text NOT NULL,
    "url" text NULL,
    "visits" bigint NULL,
    "expire" timestamptz NULL,
    "created_at" timestamptz NULL,
    "updated_at" timestamptz NULL,
    PRIMARY KEY ("key")
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "urls";
-- +goose StatementEnd

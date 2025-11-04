-- +goose Up
-- +goose StatementBegin
ALTER TABLE "urls" ADD CONSTRAINT "chk_urls_expire" CHECK (expire > created_at);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE "urls" ADD CONSTRAINT "chk_urls_visits" CHECK (visits >= 0);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX "idx_urls_url" ON "urls" ("url");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS "idx_urls_url";
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE "urls" DROP CONSTRAINT IF EXISTS "chk_urls_visits";
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE "urls" DROP CONSTRAINT IF EXISTS "chk_urls_expire";
-- +goose StatementEnd

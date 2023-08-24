-- Modify "urls" table
ALTER TABLE "urls" ADD CONSTRAINT "chk_urls_expire" CHECK (expire > created_at), ADD CONSTRAINT "chk_urls_visits" CHECK (visits >= 0);
-- Create index "idx_urls_url" to table: "urls"
CREATE INDEX "idx_urls_url" ON "urls" ("url");

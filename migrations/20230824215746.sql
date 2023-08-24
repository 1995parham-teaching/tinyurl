-- Create "urls" table
CREATE TABLE "urls" ("key" text NOT NULL, "url" text NULL, "visits" bigint NULL, "expire" timestamptz NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, PRIMARY KEY ("key"));

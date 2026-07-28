-- +goose Up
-- +goose StatementBegin
-- Backs the counter and feistel key generators, which take their uniqueness from this counter
-- instead of from a collision check. CACHE hands each session a block of values up front, so
-- handing out an identifier costs neither a round trip nor coordination between instances. The
-- price is gaps in the numbering when a session ends with values unused, which costs nothing
-- but key space.
CREATE SEQUENCE "url_id_seq" AS bigint START WITH 1 INCREMENT BY 1 CACHE 1000;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP SEQUENCE IF EXISTS "url_id_seq";
-- +goose StatementEnd

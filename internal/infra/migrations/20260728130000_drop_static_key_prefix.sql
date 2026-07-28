-- +goose Up
-- +goose StatementBegin
-- Chosen names used to be stored with a "static_" prefix so they could not collide with
-- generated keys. Nothing needed that: the primary key already rejects a name that is taken,
-- and the prefix cost every lookup of a chosen name a second query, because the key the caller
-- used was not the key that was stored. This puts both kinds of key back in one space.
--
-- Rows whose stripped name is already taken are left alone rather than dropped: there is no
-- correct way to merge two urls onto one key, and losing one of them would be worse than
-- leaving it reachable under the name it already has.
UPDATE urls AS prefixed
SET key = substring(prefixed.key FROM 8)
WHERE prefixed.key LIKE 'static\_%'
  AND NOT EXISTS (
    SELECT 1 FROM urls AS existing WHERE existing.key = substring(prefixed.key FROM 8)
  );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Deliberately not reversed. Once the prefix is gone there is no way to tell which keys were
-- chosen by a caller and which were generated, so restoring it would have to guess, and a wrong
-- guess breaks the links it renames. Rolling this back leaves the data as it is.
SELECT 1;
-- +goose StatementEnd

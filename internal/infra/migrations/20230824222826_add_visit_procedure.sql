-- +goose Up
-- +goose StatementBegin
CREATE PROCEDURE visit(key urls.key%TYPE) LANGUAGE SQL BEGIN ATOMIC
UPDATE
  urls
SET
  visits = visits + 1
WHERE
  urls.key = key;
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP PROCEDURE IF EXISTS visit;
-- +goose StatementEnd

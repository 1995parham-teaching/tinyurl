-- atlas:delimiter -- end

CREATE PROCEDURE visit(key urls.key%TYPE) LANGUAGE SQL BEGIN ATOMIC
UPDATE
  urls
set
  visits = visits + 1
where
  urls.key = key;

END;

-- end

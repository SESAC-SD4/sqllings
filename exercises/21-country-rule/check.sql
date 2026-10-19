SELECT count(*) FROM pg_constraint WHERE conrelid = 'titles'::regclass AND conname = 'titles_country_check';
SELECT md5(string_agg(t::text, '' ORDER BY show_id)) FROM titles t;
CREATE TEMP TABLE rule_test (v text, result text);
DO $$
BEGIN
  BEGIN
    INSERT INTO titles (show_id, type, title, country) VALUES ('s9901', 'Movie', 'Test', ', Korea');
    INSERT INTO rule_test VALUES (', Korea'::text, '들어감');
  EXCEPTION WHEN check_violation THEN
    INSERT INTO rule_test VALUES (', Korea'::text, '막힘');
  END;
  BEGIN
    INSERT INTO titles (show_id, type, title, country) VALUES ('s9902', 'Movie', 'Test', 'Korea');
    INSERT INTO rule_test VALUES ('Korea'::text, '들어감');
  EXCEPTION WHEN check_violation THEN
    INSERT INTO rule_test VALUES ('Korea'::text, '막힘');
  END;
END $$;
SELECT v, result FROM rule_test ORDER BY v;

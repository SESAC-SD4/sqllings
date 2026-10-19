SELECT count(*) FROM pg_constraint WHERE conrelid = 'titles'::regclass AND conname = 'titles_minutes_check';
SELECT md5(string_agg(t::text, '' ORDER BY show_id)) FROM titles t;
CREATE TEMP TABLE rule_test (v text, result text);
DO $$
BEGIN
  BEGIN
    INSERT INTO titles (show_id, type, title, minutes) VALUES ('s9901', 'Movie', 'Test', 0);
    INSERT INTO rule_test VALUES (0::text, '들어감');
  EXCEPTION WHEN check_violation THEN
    INSERT INTO rule_test VALUES (0::text, '막힘');
  END;
  BEGIN
    INSERT INTO titles (show_id, type, title, minutes) VALUES ('s9902', 'Movie', 'Test', 1);
    INSERT INTO rule_test VALUES (1::text, '들어감');
  EXCEPTION WHEN check_violation THEN
    INSERT INTO rule_test VALUES (1::text, '막힘');
  END;
END $$;
SELECT v, result FROM rule_test ORDER BY v;

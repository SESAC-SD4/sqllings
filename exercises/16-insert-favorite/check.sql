SELECT count(*), bool_and(type IS NOT NULL AND title IS NOT NULL AND release_year IS NOT NULL) FROM titles WHERE show_id = 's9001';
SELECT md5(string_agg(t::text, '' ORDER BY show_id)) FROM titles t WHERE show_id <> 's9001';

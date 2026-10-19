SELECT description FROM titles WHERE show_id = 's2822';
SELECT md5(string_agg(t::text, '' ORDER BY show_id)) FROM titles t WHERE show_id <> 's2822';

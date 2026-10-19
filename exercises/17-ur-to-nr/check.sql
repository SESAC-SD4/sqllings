SELECT rating, count(*) FROM titles WHERE rating IN ('NR', 'UR') GROUP BY rating ORDER BY rating;
SELECT md5(string_agg(t::text, '' ORDER BY show_id)) FROM titles t;

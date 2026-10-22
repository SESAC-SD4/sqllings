SELECT max(cast_count) FROM (SELECT title_id, count(*) AS cast_count FROM title_cast GROUP BY title_id) AS per_title;

SELECT count(*) FROM (SELECT title_id FROM title_countries GROUP BY title_id HAVING count(*) >= 3) AS many;

SELECT rating, count(*) FROM titles WHERE type = 'TV Show' GROUP BY rating ORDER BY count(*) DESC LIMIT 5;

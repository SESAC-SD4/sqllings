SELECT seasons, count(*) FROM titles WHERE seasons IS NOT NULL GROUP BY seasons ORDER BY count(*) DESC LIMIT 3;

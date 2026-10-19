SELECT rating, count(*) FROM titles WHERE minutes > 120 GROUP BY rating ORDER BY count(*) DESC LIMIT 3;

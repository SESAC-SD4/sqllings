SELECT type, count(*) FROM titles WHERE director IS NULL GROUP BY type;

SELECT type, count(*) FROM titles WHERE date_added BETWEEN '2019-01-01' AND '2019-12-31' GROUP BY type;

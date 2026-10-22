SELECT release_year, count(*) FROM titles WHERE release_year >= 2015 GROUP BY release_year ORDER BY release_year;

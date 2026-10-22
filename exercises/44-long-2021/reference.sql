SELECT count(*) FROM titles WHERE minutes > (SELECT avg(minutes) FROM titles) AND release_year = 2021;

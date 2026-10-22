SELECT title, release_year FROM titles WHERE release_year = (SELECT min(release_year) FROM titles);

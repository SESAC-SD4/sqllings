SELECT count(*) FROM titles LEFT JOIN title_directors ON title_directors.title_id = titles.id WHERE title_directors.title_id IS NULL AND titles.type = 'Movie' AND titles.release_year = 2021;

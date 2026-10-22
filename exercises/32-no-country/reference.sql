SELECT count(*) FROM titles LEFT JOIN title_countries ON title_countries.title_id = titles.id WHERE title_countries.title_id IS NULL;

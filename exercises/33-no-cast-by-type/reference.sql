SELECT titles.type, count(*) FROM titles LEFT JOIN title_cast ON title_cast.title_id = titles.id WHERE title_cast.title_id IS NULL GROUP BY titles.type;

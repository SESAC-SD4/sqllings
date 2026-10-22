SELECT ratings.name_ko, count(*) FROM titles JOIN ratings ON ratings.id = titles.rating_id WHERE titles.type = 'TV Show' GROUP BY ratings.name_ko ORDER BY count(*) DESC;

SELECT titles.title FROM titles JOIN ratings ON ratings.id = titles.rating_id WHERE ratings.name_ko = '전체관람가' AND titles.release_year = 2021;

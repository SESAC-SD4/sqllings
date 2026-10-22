SELECT count(*) FROM title_genres WHERE genre_id = (SELECT id FROM genres WHERE name = 'Korean TV Shows');

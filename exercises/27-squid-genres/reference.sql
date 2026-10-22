SELECT genres.name FROM titles JOIN title_genres ON title_genres.title_id = titles.id JOIN genres ON genres.id = title_genres.genre_id WHERE titles.title = 'Squid Game';

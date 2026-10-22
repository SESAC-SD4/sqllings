SELECT people.name FROM titles JOIN title_directors ON title_directors.title_id = titles.id JOIN people ON people.id = title_directors.person_id WHERE titles.title = 'Kingdom';

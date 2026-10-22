SELECT people.name, count(*) FROM people JOIN title_directors ON title_directors.person_id = people.id GROUP BY people.name ORDER BY count(*) DESC LIMIT 3;

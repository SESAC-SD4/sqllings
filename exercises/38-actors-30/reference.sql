SELECT people.name, count(*) FROM people JOIN title_cast ON title_cast.person_id = people.id GROUP BY people.name HAVING count(*) >= 30;

-- 3장(잇고 세기) 시작 상태 = 교안 fixes.sql 1~10 (강사 레포 SESAC-SD4/DB, 2026-10-21 D84 · D86)
-- ch2와 달리 《The Glory》가 없다(D80). 등급 · 장르 · 나라 · 사람을 나누고 원본 칸을 지운 표 9개
-- 1. 등급 칸에 상영 시간이 들어간 세 편: 상영 시간 칸으로 옮기고 등급은 비운다
UPDATE titles
SET duration = rating, rating = NULL
WHERE rating LIKE '%min';

-- 2. D.P.의 나라 칸 앞에 붙은 쉼표를 지운다
UPDATE titles
SET country = 'South Korea'
WHERE show_id = 's194';

-- 3. date_added를 글자에서 날짜로
ALTER TABLE titles ALTER COLUMN date_added TYPE date USING date_added::date;

-- 4. 상영 시간(minutes)과 시즌 수(seasons)를 숫자 열로
ALTER TABLE titles ADD COLUMN IF NOT EXISTS minutes integer;
UPDATE titles SET minutes = split_part(duration, ' ', 1)::integer WHERE duration LIKE '% min';
ALTER TABLE titles ADD COLUMN IF NOT EXISTS seasons integer;
UPDATE titles SET seasons = split_part(duration, ' ', 1)::integer WHERE duration LIKE '% Season%';


-- 5. 규칙 (데이터를 고친 뒤에 건다).
ALTER TABLE titles ALTER COLUMN show_id SET NOT NULL;
ALTER TABLE titles ALTER COLUMN title SET NOT NULL;
ALTER TABLE titles ALTER COLUMN type SET NOT NULL;
ALTER TABLE titles ADD CONSTRAINT titles_type_check CHECK (type IN ('Movie', 'TV Show'));
ALTER TABLE titles ADD CONSTRAINT titles_rating_check CHECK (rating IN ('G', 'PG', 'PG-13', 'R', 'NC-17', 'NR', 'UR', 'TV-Y', 'TV-Y7', 'TV-Y7-FV', 'TV-G', 'TV-PG', 'TV-14', 'TV-MA'));
ALTER TABLE titles ADD CONSTRAINT titles_release_year_check CHECK (release_year BETWEEN 1900 AND 2100);
ALTER TABLE titles ADD CONSTRAINT titles_show_id_key UNIQUE (show_id);
ALTER TABLE titles ADD COLUMN id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY;


-- 6. 등급 나누기 (rating → ratings · titles.rating_id)
CREATE TABLE ratings (
    id      bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    code    text NOT NULL UNIQUE,
    name_ko text NOT NULL
);
INSERT INTO ratings (code, name_ko) VALUES
    ('G', '전체관람가'),
    ('TV-G', '전체관람가'),
    ('TV-Y', '어린이'),
    ('TV-Y7', '7세 이상'),
    ('TV-Y7-FV', '7세 이상(판타지 폭력)'),
    ('PG', '보호자 지도'),
    ('TV-PG', '보호자 지도'),
    ('PG-13', '13세 미만 주의'),
    ('TV-14', '14세 미만 주의'),
    ('R', '17세 미만 보호자 동반'),
    ('NC-17', '17세 이하 관람불가'),
    ('TV-MA', '성인용'),
    ('NR', '등급 없음'),
    ('UR', '등급 없음');
ALTER TABLE titles ADD COLUMN rating_id bigint REFERENCES ratings (id);
UPDATE titles SET rating_id = ratings.id FROM ratings WHERE titles.rating = ratings.code;


-- 7. 장르 나누기 (listed_in → genres · title_genres)
CREATE TABLE genres (
    id   bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name text NOT NULL UNIQUE
);
INSERT INTO genres (name)
SELECT DISTINCT genre
FROM titles, unnest(string_to_array(listed_in, ', ')) AS genre
ORDER BY genre;

CREATE TABLE title_genres (
    title_id bigint NOT NULL REFERENCES titles (id) ON DELETE CASCADE,
    genre_id bigint NOT NULL REFERENCES genres (id),
    PRIMARY KEY (title_id, genre_id)
);
INSERT INTO title_genres (title_id, genre_id)
SELECT titles.id, genres.id
FROM titles, unnest(string_to_array(titles.listed_in, ', ')) AS genre, genres
WHERE genres.name = genre;

-- 8. 나라 나누기 (country → countries · title_countries)
--    쉼표로 자르고 앞뒤 공백을 지운다. 끝 쉼표가 남긴 빈 조각은 빈 이름 나라 하나로 남는다
CREATE TABLE countries (
    id   bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name text NOT NULL UNIQUE
);
INSERT INTO countries (name)
SELECT DISTINCT trim(part)
FROM titles, unnest(string_to_array(country, ',')) AS part
ORDER BY 1;

CREATE TABLE title_countries (
    title_id   bigint NOT NULL REFERENCES titles (id) ON DELETE CASCADE,
    country_id bigint NOT NULL REFERENCES countries (id),
    PRIMARY KEY (title_id, country_id)
);
INSERT INTO title_countries (title_id, country_id)
SELECT titles.id, countries.id
FROM titles, unnest(string_to_array(titles.country, ',')) AS part, countries
WHERE countries.name = trim(part);


-- 9. 사람 나누기 (director · cast_members → people · title_directors · title_cast)
--    감독이자 배우인 사람이 있어서 사람 표는 하나. 배우를 넣을 때 이미 있는 이름은 넘어간다
--    한 작품에 같은 이름이 두 번 적힌 곳이 있어서 짝은 DISTINCT로 넣는다
CREATE TABLE people (
    id   bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name text NOT NULL UNIQUE
);
INSERT INTO people (name)
SELECT DISTINCT part
FROM titles, unnest(string_to_array(director, ', ')) AS part
ORDER BY part;
INSERT INTO people (name)
SELECT DISTINCT part
FROM titles, unnest(string_to_array(cast_members, ', ')) AS part
ORDER BY part
ON CONFLICT (name) DO NOTHING;

CREATE TABLE title_directors (
    title_id  bigint NOT NULL REFERENCES titles (id) ON DELETE CASCADE,
    person_id bigint NOT NULL REFERENCES people (id),
    PRIMARY KEY (title_id, person_id)
);
INSERT INTO title_directors (title_id, person_id)
SELECT DISTINCT titles.id, people.id
FROM titles, unnest(string_to_array(titles.director, ', ')) AS part, people
WHERE people.name = part;

CREATE TABLE title_cast (
    title_id  bigint NOT NULL REFERENCES titles (id) ON DELETE CASCADE,
    person_id bigint NOT NULL REFERENCES people (id),
    PRIMARY KEY (title_id, person_id)
);
INSERT INTO title_cast (title_id, person_id)
SELECT DISTINCT titles.id, people.id
FROM titles, unnest(string_to_array(titles.cast_members, ', ')) AS part, people
WHERE people.name = part;


-- 10. 나눈 칸의 원본을 지운다 (같은 정보는 한 곳에)
--     등급 · 장르 · 나라 · 감독 · 배우는 이제 나눈 표에만 있다. 보려면 JOIN
ALTER TABLE titles
    DROP COLUMN rating,
    DROP COLUMN listed_in,
    DROP COLUMN country,
    DROP COLUMN director,
    DROP COLUMN cast_members;

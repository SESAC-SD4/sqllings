package exercises

import (
	"regexp"
	"strings"
	"testing"

	"sqllings/internal/grade"
	"sqllings/internal/pg"
)

type tcase struct{ wrong, wrongMsg, alt string }

var cases = map[int]tcase{
	1:  {"SELECT count(*) FROM titles WHERE type = 'Movie';", "값이 다릅니다", "SELECT count(show_id) FROM titles"},
	2:  {"SELECT count(*) FROM titles WHERE type = 'movie';", "대소문자", "SELECT count(*) FROM titles WHERE type <> 'TV Show'"},
	3:  {"SELECT title FROM titles WHERE director = 'David Fincher' ORDER BY title DESC;", "순서", "SELECT title FROM titles WHERE director LIKE 'David Fincher' ORDER BY 1"},
	4:  {"SELECT count(*) FROM titles WHERE type = 'TV Show' AND release_year <= 2000;", "값이 다릅니다", "SELECT count(*) FROM titles WHERE release_year <= 1999 AND type = 'TV Show'"},
	5:  {"SELECT count(*) FROM titles WHERE type = 'TV Show' AND release_year BETWEEN 2000 AND 2010;", "값이 다릅니다", "SELECT count(*) FROM titles WHERE release_year >= 2000 AND release_year <= 2009 AND type = 'TV Show'"},
	6:  {"SELECT count(*) FROM titles WHERE rating IN ('TV-MA', 'R');", "값이 다릅니다", "SELECT count(*) FROM titles WHERE (rating = 'R' OR rating = 'TV-MA') AND type = 'Movie'"},
	7:  {"SELECT type, count(*) FROM titles WHERE director = NULL GROUP BY type;", "= NULL", "SELECT type, count(show_id) FROM titles WHERE NOT (director IS NOT NULL) GROUP BY type"},
	8:  {"SELECT title, release_year FROM titles ORDER BY release_year, title LIMIT 5;", "값이 다릅니다", "SELECT title, release_year FROM titles WHERE type = 'TV Show' ORDER BY 2, 1 LIMIT 5"},
	9:  {"SELECT rating, count(*) FROM titles WHERE type = 'TV Show' GROUP BY rating ORDER BY count(*) LIMIT 5;", "값이 다릅니다", "SELECT rating, count(show_id) FROM titles WHERE type = 'TV Show' GROUP BY 1 ORDER BY 2 DESC LIMIT 5"},
	10: {"SELECT count(*) FROM titles WHERE type = 'Movie' AND country = 'India' OR country = 'United Kingdom';", "값이 다릅니다", "SELECT count(*) FROM titles WHERE country IN ('India', 'United Kingdom') AND type = 'Movie'"},
	11: {"SELECT count(*) FROM titles WHERE type = 'TV Show' AND duration = '1 Seasons';", "대소문자", "SELECT count(*) FROM titles WHERE duration = '1 Season'"},
	12: {"SELECT count(*) FROM titles WHERE title LIKE '%Night%';", "값이 다릅니다", "SELECT count(*) FROM titles WHERE lower(title) LIKE '%night%'"},
	13: {"SELECT count(*) FROM titles WHERE country = 'Japan';", "값이 다릅니다", "SELECT count(*) FROM titles WHERE position('Japan' in country) > 0"},
	14: {"SELECT title, rating FROM titles WHERE rating LIKE 'min%' ORDER BY title;", "0행", "SELECT title, rating FROM titles WHERE rating LIKE '% min' ORDER BY 1"},
	15: {"UPDATE titles SET description = '조선 시대 좀비 사극' WHERE title LIKE 'Kingdom%';", "상태가 다릅니다", "UPDATE titles SET description = '조선 시대 좀비 사극' WHERE title = 'Kingdom'"},
	16: {"INSERT INTO titles (show_id, type, title) VALUES ('s9001', 'Movie', 'X');", "상태가 다릅니다", "INSERT INTO titles (show_id, type, title, release_year, country) VALUES ('s9001', 'TV Show', '나의 아저씨', 2018, 'South Korea')"},
	17: {"UPDATE titles SET rating = 'NR';", "상태가 다릅니다", "UPDATE titles SET rating = 'NR' WHERE rating IN ('UR')"},
	18: {"UPDATE titles SET country = 'France, Algeria' WHERE country LIKE '%,%';", "상태가 다릅니다", "UPDATE titles SET country = ltrim(country, ', ') WHERE country LIKE ',%'"},
	19: {"SELECT title, date_added FROM titles WHERE country = 'South Korea' ORDER BY date_added DESC LIMIT 3;", "값이 다릅니다", "SELECT title, date_added FROM titles WHERE country = 'South Korea' ORDER BY date_added DESC NULLS LAST LIMIT 3"},
	20: {"SELECT count(*) FROM titles WHERE seasons > 5;", "값이 다릅니다", "SELECT count(*) FROM titles WHERE type = 'TV Show' AND seasons >= 5"},
	21: {"ALTER TABLE titles ADD CONSTRAINT titles_country_check CHECK (country NOT LIKE '%,%');", "에러", "ALTER TABLE titles ADD CONSTRAINT titles_country_check CHECK (left(country, 1) <> ',')"},
	22: {"ALTER TABLE titles ADD CONSTRAINT titles_minutes_check CHECK (minutes >= 0);", "상태가 다릅니다", "ALTER TABLE titles ADD CONSTRAINT titles_minutes_check CHECK (minutes >= 1)"},
	23: {"SELECT title, minutes FROM titles WHERE date_added BETWEEN '2020-01-01' AND '2020-12-31' ORDER BY minutes DESC LIMIT 3;", "값이 다릅니다", "SELECT title, minutes FROM titles WHERE type = 'Movie' AND date_added >= '2020-01-01' AND date_added < '2021-01-01' ORDER BY minutes DESC LIMIT 3"},
	24: {"SELECT type, count(*) FROM titles WHERE release_year = 2019 GROUP BY type;", "값이 다릅니다", "SELECT type, count(*) FROM titles WHERE date_added >= '2019-01-01' AND date_added < '2020-01-01' GROUP BY 1"},
	25: {"SELECT rating, count(*) FROM titles WHERE minutes >= 120 GROUP BY rating ORDER BY count(*) DESC LIMIT 3;", "값이 다릅니다", "SELECT rating, count(*) FROM titles WHERE type = 'Movie' AND minutes > 120 GROUP BY rating ORDER BY count(*) DESC LIMIT 3"},
	26: {"SELECT duration, count(*) FROM titles WHERE type = 'TV Show' GROUP BY duration ORDER BY count(*) DESC LIMIT 3;", "값이 다릅니다", "SELECT seasons, count(*) FROM titles WHERE type = 'TV Show' AND seasons IS NOT NULL GROUP BY 1 ORDER BY 2 DESC LIMIT 3"},
	27: {"SELECT genres.name FROM titles JOIN title_genres ON title_genres.title_id = titles.id JOIN genres ON genres.id = title_genres.title_id WHERE titles.title = 'Squid Game';", "0행", "SELECT g.name FROM genres AS g, title_genres AS tg, titles AS t WHERE g.id = tg.genre_id AND tg.title_id = t.id AND t.title = 'Squid Game'"},
	28: {"SELECT people.name FROM titles JOIN title_cast ON title_cast.title_id = titles.id JOIN people ON people.id = title_cast.person_id WHERE titles.title = 'Kingdom';", "행이어야", "SELECT name FROM people WHERE id IN (SELECT person_id FROM title_directors WHERE title_id = (SELECT id FROM titles WHERE title = 'Kingdom'))"},
	29: {"SELECT titles.title, titles.release_year FROM people JOIN title_cast ON title_cast.person_id = people.id JOIN titles ON titles.id = title_cast.title_id WHERE people.name = 'Shah Rukh Khan' ORDER BY titles.release_year, titles.title LIMIT 5;", "값이 다릅니다", "SELECT t.title, t.release_year FROM titles AS t JOIN title_cast AS c ON c.title_id = t.id JOIN people AS p ON p.id = c.person_id WHERE p.name = 'Shah Rukh Khan' ORDER BY 2 DESC, 1 LIMIT 5"},
	30: {"SELECT titles.title FROM titles JOIN ratings ON ratings.id = titles.rating_id WHERE ratings.code = 'G' AND titles.release_year = 2021;", "0행", "SELECT titles.title FROM titles JOIN ratings ON titles.rating_id = ratings.id WHERE ratings.code IN ('G', 'TV-G') AND titles.release_year = 2021"},
	31: {"SELECT genres.name FROM people JOIN title_directors ON title_directors.person_id = people.id JOIN title_genres ON title_genres.title_id = title_directors.title_id JOIN genres ON genres.id = title_genres.genre_id WHERE people.name = 'Bong Joon Ho';", "행이어야", "SELECT genres.name FROM genres JOIN title_genres ON title_genres.genre_id = genres.id JOIN title_directors ON title_directors.title_id = title_genres.title_id JOIN people ON people.id = title_directors.person_id WHERE people.name = 'Bong Joon Ho' GROUP BY genres.name"},
	32: {"SELECT count(*) FROM titles JOIN title_countries ON title_countries.title_id = titles.id;", "값이 다릅니다", "SELECT count(*) FROM titles WHERE id NOT IN (SELECT title_id FROM title_countries)"},
	33: {"SELECT titles.type, count(*) FROM titles LEFT JOIN title_cast ON title_cast.title_id = titles.id GROUP BY titles.type;", "값이 다릅니다", "SELECT type, count(*) FROM titles WHERE NOT EXISTS (SELECT 1 FROM title_cast WHERE title_cast.title_id = titles.id) GROUP BY type"},
	34: {"SELECT count(*) FROM titles LEFT JOIN title_directors ON title_directors.title_id = titles.id WHERE title_directors.title_id IS NULL AND titles.release_year = 2021;", "값이 다릅니다", "SELECT count(*) FROM titles WHERE type = 'Movie' AND release_year = 2021 AND id NOT IN (SELECT title_id FROM title_directors)"},
	35: {"SELECT count(*) FROM titles LEFT JOIN title_directors ON title_directors.title_id = titles.id LEFT JOIN title_cast ON title_cast.title_id = titles.id WHERE title_directors.title_id IS NULL OR title_cast.title_id IS NULL;", "값이 다릅니다", "SELECT count(*) FROM titles WHERE id NOT IN (SELECT title_id FROM title_directors) AND id NOT IN (SELECT title_id FROM title_cast)"},
	36: {"SELECT people.name, count(*) FROM people JOIN title_cast ON title_cast.person_id = people.id GROUP BY people.name ORDER BY count(*) DESC LIMIT 3;", "값이 다릅니다", "SELECT p.name, count(d.title_id) FROM title_directors AS d JOIN people AS p ON p.id = d.person_id GROUP BY p.id, p.name ORDER BY 2 DESC LIMIT 3"},
	37: {"SELECT release_year, count(*) FROM titles WHERE release_year > 2015 GROUP BY release_year ORDER BY release_year;", "행이어야", "SELECT release_year, count(id) FROM titles WHERE release_year BETWEEN 2015 AND 2100 GROUP BY 1 ORDER BY 1"},
	38: {"SELECT people.name, count(*) FROM people JOIN title_cast ON title_cast.person_id = people.id GROUP BY people.name HAVING count(*) > 30;", "행이어야", "SELECT people.name, count(title_cast.title_id) FROM title_cast JOIN people ON people.id = title_cast.person_id GROUP BY people.name HAVING count(title_cast.title_id) >= 30"},
	39: {"SELECT ratings.code, count(*) FROM titles JOIN ratings ON ratings.id = titles.rating_id WHERE titles.type = 'TV Show' GROUP BY ratings.code ORDER BY count(*) DESC;", "값이 다릅니다", "SELECT r.name_ko, count(t.id) FROM ratings AS r JOIN titles AS t ON t.rating_id = r.id WHERE t.type = 'TV Show' GROUP BY r.name_ko ORDER BY 2 DESC"},
	40: {"SELECT genres.name, count(*) FROM countries JOIN title_countries ON title_countries.country_id = countries.id JOIN title_genres ON title_genres.title_id = title_countries.title_id JOIN genres ON genres.id = title_genres.genre_id WHERE countries.name = 'South Korea' GROUP BY genres.name ORDER BY count(*) LIMIT 5;", "값이 다릅니다", "SELECT g.name, count(DISTINCT tg.title_id) FROM genres AS g JOIN title_genres AS tg ON tg.genre_id = g.id JOIN title_countries AS tc ON tc.title_id = tg.title_id JOIN countries AS c ON c.id = tc.country_id WHERE c.name = 'South Korea' GROUP BY 1 ORDER BY 2 DESC LIMIT 5"},
	41: {"SELECT people.name, count(*) FROM people LEFT JOIN title_directors ON title_directors.person_id = people.id WHERE people.name IN ('Bong Joon Ho', 'Steven Yeun', 'Tilda Swinton') GROUP BY people.name;", "값이 다릅니다", "SELECT name, (SELECT count(*) FROM title_directors WHERE title_directors.person_id = people.id) FROM people WHERE name IN ('Bong Joon Ho', 'Steven Yeun', 'Tilda Swinton')"},
	42: {"SELECT count(*) FROM title_countries WHERE country_id = (SELECT id FROM countries WHERE name = 'South Korea');", "값이 다릅니다", "SELECT count(*) FROM title_genres JOIN genres ON genres.id = title_genres.genre_id WHERE genres.name = 'Korean TV Shows'"},
	43: {"SELECT title, release_year FROM titles WHERE release_year = (SELECT max(release_year) FROM titles) LIMIT 1;", "값이 다릅니다", "SELECT title, release_year FROM titles ORDER BY release_year LIMIT 1"},
	44: {"SELECT count(*) FROM titles WHERE minutes > (SELECT avg(minutes) FROM titles WHERE release_year = 2021) AND release_year = 2021;", "값이 다릅니다", "SELECT count(*) FROM titles WHERE type = 'Movie' AND release_year = 2021 AND minutes > (SELECT avg(minutes) FROM titles WHERE type = 'Movie')"},
	45: {"SELECT count(*) FROM (SELECT title_id FROM title_countries GROUP BY title_id HAVING count(*) > 3) AS many;", "값이 다릅니다", "SELECT count(*) FROM titles WHERE (SELECT count(*) FROM title_countries WHERE title_countries.title_id = titles.id) >= 3"},
	46: {"SELECT max(person_id) FROM title_cast;", "값이 다릅니다", "SELECT count(*) FROM title_cast GROUP BY title_id ORDER BY 1 DESC LIMIT 1"},
	47: {"SELECT count(DISTINCT title_genres.title_id) FROM title_genres JOIN genres ON genres.id = title_genres.genre_id JOIN title_countries ON title_countries.title_id = title_genres.title_id JOIN countries ON countries.id = title_countries.country_id WHERE genres.name LIKE '%Dramas%' AND countries.name = 'South Korea';", "값이 다릅니다", "SELECT count(*) FROM title_genres JOIN genres ON genres.id = title_genres.genre_id JOIN title_countries ON title_countries.title_id = title_genres.title_id JOIN countries ON countries.id = title_countries.country_id WHERE genres.name = 'Dramas' AND countries.name = 'South Korea'"},
}

func TestRegistryShape(t *testing.T) {
	slug := regexp.MustCompile(`^\d{2}-[a-z0-9-]+$`)
	all := All()
	// 배포본(publish.sh)은 앞 장만 담으므로 개수는 고정하지 않는다. 번호가 1부터 이어지는지만 본다.
	if len(all) == 0 {
		t.Fatal("문제가 없다")
	}
	for i, ex := range all {
		if ex.ID != i+1 {
			t.Fatalf("번호가 이어지지 않는다: %d", ex.ID)
		}
		if !slug.MatchString(ex.Slug) || !strings.HasPrefix(ex.Title, ex.Slug[:2]) {
			t.Errorf("%s 형식", ex.Slug)
		}
		if ex.Optional != strings.Contains(ex.Title, "★") {
			t.Errorf("%s ★ 표시", ex.Slug)
		}
		if len(ex.Hints) != 3 || ex.Chapter < 1 {
			t.Errorf("%s 힌트·장", ex.Slug)
		}
	}
}

// 힌트·문제 설명에 기준 SQL의 20글자 이상 조각이 없어야 한다(플레이북 §5).
func TestNoAnswerInHints(t *testing.T) {
	norm := func(s string) string { return strings.Join(strings.Fields(strings.ToLower(s)), " ") }
	for _, ex := range All() {
		refText := strings.TrimSuffix(ex.Reference, ";")
		text := norm(ex.Prompt + " " + strings.Join(ex.Hints, " "))
		// 문제가 알려 줘야 하는 이름(예: 규칙 이름 `titles_country_check`)은 정답 조각으로 보지 않는다.
		for _, m := range regexp.MustCompile("`([^`]+)`").FindAllStringSubmatch(ex.Prompt, -1) {
			refText = strings.ReplaceAll(refText, m[1], " ")
			text = strings.ReplaceAll(text, norm(m[1]), " ")
		}
		ref := []rune(norm(refText))
		for i := 0; i+20 <= len(ref); i++ {
			if frag := string(ref[i : i+20]); strings.Contains(text, frag) {
				t.Errorf("%s: 힌트에 정답 조각 %q", ex.Slug, frag)
				break
			}
		}
	}
}

func TestEachExercise(t *testing.T) {
	if err := pg.Ping(); err != nil {
		t.Skip(err)
	}
	for _, ex := range All() {
		c, ok := cases[ex.ID]
		if !ok {
			t.Errorf("%s: 테스트 케이스 없음", ex.Slug)
			continue
		}
		if r := grade.Grade(ex, ""); r.Passed {
			t.Errorf("%s: 빈 답 통과", ex.Slug)
		}
		if ex.Reference != "" {
			if r := grade.Grade(ex, ex.Reference); !r.Passed {
				t.Errorf("%s: 기준 실패 %s\n%s", ex.Slug, r.Message, r.Display)
			}
		}
		if r := grade.Grade(ex, c.alt); !r.Passed {
			t.Errorf("%s: 다른 정답 실패 %s\n%s", ex.Slug, r.Message, r.Display)
		}
		if r := grade.Grade(ex, c.wrong); r.Passed || !strings.Contains(r.Message, c.wrongMsg) {
			t.Errorf("%s: 오답 %q → %q", ex.Slug, c.wrong, r.Message)
		}
	}
}

// Windows 한국어 서버는 에러를 한국어로 낸다(오류: · …을 위반했습니다). 영어 문구만 적힌 안내는 학생 화면과 다르다.
func TestHintsNotEnglishOnly(t *testing.T) {
	pairs := [][2]string{{"violates", "위반"}, {"ERROR:", "오류:"}, {"does not exist", "없습니다"}, {"canceling statement", "취소"}}
	for _, ex := range All() {
		text := ex.Prompt + "\n" + strings.Join(ex.Hints, "\n")
		for _, p := range pairs {
			if strings.Contains(text, p[0]) && !strings.Contains(text, p[1]) {
				t.Errorf("%s: 영어 서버 문구 %q만 있고 한국어(%q)가 없다", ex.Slug, p[0], p[1])
			}
		}
	}
}

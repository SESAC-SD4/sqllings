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
		ref := []rune(norm(strings.TrimSuffix(ex.Reference, ";")))
		text := norm(ex.Prompt + " " + strings.Join(ex.Hints, " "))
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

package grade

import (
	"strings"
	"testing"

	"sqllings/internal/exercise"
	"sqllings/internal/pg"
)

func need(t *testing.T) {
	t.Helper()
	if err := pg.Ping(); err != nil {
		t.Skip(err)
	}
}

var countMovies = &exercise.Exercise{ID: 1, Chapter: 1, Mode: exercise.ModeResult, Reference: "SELECT count(*) FROM titles WHERE type = 'Movie';"}

func TestResultPassDifferentSpelling(t *testing.T) {
	need(t)
	r := Grade(countMovies, "select COUNT(show_id)\r\nfrom titles\r\nwhere type <> 'TV Show'\r\n")
	if !r.Passed {
		t.Fatalf("%+v", r)
	}
	if !strings.Contains(r.Display, "6131") {
		t.Fatalf("화면에 결과가 없다: %q", r.Display)
	}
}

func TestResultZeroHint(t *testing.T) {
	need(t)
	r := Grade(countMovies, "SELECT count(*) FROM titles WHERE type = 'movie';")
	if r.Passed || !strings.Contains(r.Message, "대소문자") {
		t.Fatalf("%+v", r)
	}
}

func TestResultSQLError(t *testing.T) {
	need(t)
	r := Grade(countMovies, `SELECT count(*) FROM titles WHERE type = "Movie";`)
	if r.Passed || !(strings.Contains(r.Display, `column "Movie" does not exist`) || strings.Contains(r.Display, `"Movie" 이름의 칼럼은 없습니다`)) || strings.Contains(r.Display, "psql:<stdin>") {
		t.Fatalf("%+v", r)
	}
	if !strings.Contains(r.Message, "에러") {
		t.Fatal(r.Message)
	}
}

func TestEmptyForbiddenAndMany(t *testing.T) {
	if r := Grade(countMovies, "  \n"); r.Passed || !strings.Contains(r.Message, "비어") {
		t.Fatal(r)
	}
	if r := Grade(countMovies, "\\d titles"); r.Passed || !strings.Contains(r.Message, "SQL만") {
		t.Fatal(r)
	}
	if r := Grade(countMovies, "SELECT 1; SELECT 2;"); r.Passed || !strings.Contains(r.Message, "한 문장") {
		t.Fatal(r)
	}
	// 결과 예시의 값을 그대로 적은 답
	if r := Grade(countMovies, "SELECT 6131;"); r.Passed || !strings.Contains(r.Message, "FROM이 없습니다") {
		t.Fatal(r)
	}
}

func TestOrdered(t *testing.T) {
	need(t)
	ex := &exercise.Exercise{Chapter: 1, Mode: exercise.ModeResult, Ordered: true,
		Reference: "SELECT title, release_year FROM titles ORDER BY release_year, title LIMIT 5;"}
	if r := Grade(ex, "SELECT title, release_year FROM titles ORDER BY release_year, title LIMIT 5"); !r.Passed {
		t.Fatal(r)
	}
	if r := Grade(ex, "SELECT title, release_year FROM titles ORDER BY release_year DESC, title LIMIT 5"); r.Passed {
		t.Fatal("값이 다른데 통과")
	}
	if r := Grade(ex, "SELECT title FROM titles ORDER BY release_year, title LIMIT 5"); r.Passed || !strings.Contains(r.Message, "열이 2개") {
		t.Fatal(r)
	}
	r := Grade(ex, "SELECT * FROM (SELECT title, release_year FROM titles ORDER BY release_year, title LIMIT 5) t ORDER BY title DESC")
	if r.Passed || !strings.Contains(r.Message, "순서") {
		t.Fatal(r)
	}
}

func TestUnordered(t *testing.T) {
	need(t)
	ex := &exercise.Exercise{Chapter: 1, Mode: exercise.ModeResult, Reference: "SELECT title FROM titles WHERE director = 'Bong Joon Ho';"}
	if r := Grade(ex, "SELECT title FROM titles WHERE director = 'Bong Joon Ho' ORDER BY title DESC"); !r.Passed {
		t.Fatal(r)
	}
}

func TestHangulValue(t *testing.T) {
	need(t)
	ex := &exercise.Exercise{Chapter: 1, Mode: exercise.ModeResult, Ordered: true,
		Reference: "SELECT title, COALESCE(director, '(알 수 없음)') FROM titles WHERE country = 'South Korea' AND release_year = 2020 ORDER BY title LIMIT 8;"}
	if r := Grade(ex, "SELECT title, COALESCE(director, '(알 수 없음)') AS d FROM titles WHERE release_year = 2020 AND country = 'South Korea' ORDER BY 1 LIMIT 8;"); !r.Passed {
		t.Fatal(r)
	}
	if r := Grade(ex, "SELECT title, COALESCE(director, '(모름)') FROM titles WHERE country = 'South Korea' AND release_year = 2020 ORDER BY title LIMIT 8;"); r.Passed {
		t.Fatal("다른 한글 값인데 통과")
	}
}

func TestNullIsNotEmpty(t *testing.T) {
	need(t)
	ex := &exercise.Exercise{Chapter: 1, Mode: exercise.ModeResult, Reference: "SELECT director FROM titles WHERE show_id = 's2';"}
	if r := Grade(ex, "SELECT '' FROM titles WHERE show_id = 's2';"); r.Passed {
		t.Fatal("NULL과 빈 글자를 같게 봤다")
	}
}

func TestState(t *testing.T) {
	need(t)
	ex := &exercise.Exercise{Chapter: 2, Mode: exercise.ModeState, FirstWord: "UPDATE",
		Reference: "UPDATE titles SET description = 'x' WHERE show_id = 's2822';",
		Check:     "SELECT description FROM titles WHERE show_id = 's2822'; SELECT md5(string_agg(t::text, '' ORDER BY show_id)) FROM titles t WHERE show_id <> 's2822';"}
	r := Grade(ex, "update titles set description='x' where title = 'Kingdom' and release_year = 2020;")
	if !r.Passed || !strings.Contains(r.Display, "UPDATE 1") {
		t.Fatal(r)
	}
	if r := Grade(ex, "UPDATE titles SET description = 'x';"); r.Passed {
		t.Fatal("전체를 바꿨는데 통과")
	}
	if r := Grade(ex, "SELECT 1;"); r.Passed || !strings.Contains(r.Message, "UPDATE") {
		t.Fatal(r)
	}
}

func TestErrorMode(t *testing.T) {
	need(t)
	ex := &exercise.Exercise{Chapter: 2, Mode: exercise.ModeError, ErrorName: "titles_type_check", FirstWord: "INSERT"}
	if r := Grade(ex, "INSERT INTO titles (show_id, type, title) VALUES ('s9002', 'Series', 'Test');"); !r.Passed {
		t.Fatal(r)
	}
	if r := Grade(ex, "INSERT INTO titles (show_id, type, title) VALUES ('s9002', 'Movie', 'Test');"); r.Passed || !strings.Contains(r.Message, "에러 없이") {
		t.Fatal(r)
	}
	if r := Grade(ex, "INSERT INTO titles (show_id, type) VALUES ('s9002', 'Movie');"); r.Passed || !strings.Contains(r.Message, "다른 규칙") {
		t.Fatal(r)
	}
}

func TestBigResultTruncated(t *testing.T) {
	need(t)
	r := Grade(countMovies, "SELECT * FROM titles;")
	if r.Passed || strings.Count(r.Display, "\n") > 205 || !strings.Contains(r.Display, "줄 더") {
		t.Fatalf("lines=%d", strings.Count(r.Display, "\n"))
	}
}

func TestTimeout(t *testing.T) {
	need(t)
	old := Timeout
	Timeout = "1s"
	defer func() { Timeout = old }()
	r := Grade(countMovies, "SELECT count(*) FROM titles a, titles b, titles c;")
	if r.Passed || !strings.Contains(r.Message, "10초") {
		t.Fatal(r)
	}
}

func TestStudentCannotTouchPublic(t *testing.T) {
	need(t)
	pg.Run("CREATE TABLE IF NOT EXISTS public.grade_probe (a int); INSERT INTO public.grade_probe VALUES (1);")
	defer pg.Run("DROP TABLE IF EXISTS public.grade_probe;")
	Grade(&exercise.Exercise{Chapter: 1, Mode: exercise.ModeState, Reference: "SELECT 1;", Check: "SELECT 1;"}, "DELETE FROM public.grade_probe;")
	out, _ := pg.Run("SELECT count(*) FROM public.grade_probe;")
	if !strings.Contains(out.Text, "1") {
		t.Fatal("채점이 public을 바꿨다")
	}
}

func TestServerDownMessage(t *testing.T) {
	t.Setenv("PGPORT", "5999")
	r := Grade(countMovies, "SELECT count(*) FROM titles;")
	if r.Passed || !strings.Contains(r.Message, "서버가 꺼져") {
		t.Fatal(r)
	}
}

func TestErrorModeOnlyRealConstraint(t *testing.T) {
	need(t)
	ex := &exercise.Exercise{Chapter: 2, Mode: exercise.ModeError, ErrorName: "titles_type_check", FirstWord: "INSERT"}
	if r := Grade(ex, "INSERT INTO titles_type_check VALUES (1);"); r.Passed {
		t.Fatal("없는 표 에러인데 통과")
	}
	ex2 := &exercise.Exercise{Chapter: 2, Mode: exercise.ModeError, ErrorName: "titles_release_year_check", FirstWord: "INSERT"}
	if r := Grade(ex2, "INSERT INTO titles (show_id, type) VALUES ('titles_release_year_check', 'Movie');"); r.Passed {
		t.Fatal("NOT NULL 에러인데 DETAIL에 이름이 있어 통과")
	}
}

func TestCommentAtLineEnd(t *testing.T) {
	need(t)
	ex := &exercise.Exercise{Chapter: 2, Mode: exercise.ModeState, Reference: "ALTER TABLE titles ADD COLUMN x integer; UPDATE titles SET x = 1;",
		Check: "SELECT count(*) FROM titles WHERE x = 1;"}
	if r := Grade(ex, "ALTER TABLE titles ADD COLUMN x integer -- 새 열\n;\nUPDATE titles SET x = 1;"); !r.Passed {
		t.Fatalf("%+v", r)
	}
}

func TestLeftoverGradeSchemaRecovers(t *testing.T) {
	need(t)
	pg.Run("CREATE SCHEMA IF NOT EXISTS sqllings_grade;")
	defer pg.Run("DROP SCHEMA IF EXISTS sqllings_grade CASCADE;")
	if r := Grade(countMovies, "SELECT count(*) FROM titles WHERE type = 'Movie';"); !r.Passed {
		t.Fatalf("남은 채점 스키마 때문에 실패: %+v", r)
	}
}

// 순서 무시 비교에서 값이 다르면 학생 결과의 어느 행·어느 열인지 알려 준다. 정답 쪽 값은 보여 주지 않는다.
func TestUnorderedDiffPointsToStudentRow(t *testing.T) {
	ex := &exercise.Exercise{Mode: exercise.ModeResult}
	ref := [][]string{{"Parasite", "2019"}, {"Mother", "2009"}, {"Okja", "2017"}}
	stu := [][]string{{"Okja", "2017"}, {"Snowpiercer", "2013"}, {"Parasite", "2019"}}
	r := compareResult(ex, ref, stu)
	if r.Passed || !strings.Contains(r.Message, "값이 다릅니다") || !strings.Contains(r.Message, "내 결과 2행") {
		t.Fatal(r.Message)
	}
	for _, leak := range []string{"Mother", "2009"} {
		if strings.Contains(r.Message, leak) {
			t.Fatalf("정답 값 %q가 안내에 보인다: %s", leak, r.Message)
		}
	}
	// 제목은 다 맞고 연도만 틀렸으면 몇 번째 열인지도
	stu2 := [][]string{{"Okja", "2017"}, {"Mother", "2010"}, {"Parasite", "2019"}}
	if r := compareResult(ex, ref, stu2); !strings.Contains(r.Message, "내 결과 2행") || !strings.Contains(r.Message, "2열") {
		t.Fatal(r.Message)
	}
	// 한 칸짜리 결과(count)는 행·열을 말할 필요가 없다 — 지금처럼 0 안내만
	if r := compareResult(ex, [][]string{{"6131"}}, [][]string{{"0"}}); strings.Contains(r.Message, "행:") || strings.Contains(r.Message, "열") || !strings.Contains(r.Message, "대소문자") {
		t.Fatal(r.Message)
	}
}

// 에러 모드 판정은 영어 서버와 Windows 한국어 서버(lc_messages Korean_Korea.949) 문구를 모두 받는다.
func TestViolatesBothLanguages(t *testing.T) {
	en := "psql:<stdin>:3: ERROR:  new row for relation \"titles\" violates check constraint \"titles_type_check\"\nDETAIL:  Failing row contains (s9002, Series, Test)."
	ko := "psql:<stdin>:3: 오류:  새 자료가 \"titles\" 릴레이션의 \"titles_type_check\" 체크 제약 조건을 위반했습니다\n상세정보:  실패한 자료: (s9002, Series, Test)."
	for _, d := range []string{en, ko} {
		if !violates(pg.FirstErrorLine(d), "titles_type_check") {
			t.Errorf("규칙 위반을 못 알아봤다: %s", d)
		}
		if violates(pg.FirstErrorLine(d), "titles_release_year_check") {
			t.Errorf("다른 규칙인데 통과: %s", d)
		}
	}
	// 표 이름 자리에만 규칙 이름이 있으면 위반이 아니다
	if violates(`오류:  "titles_type_check" 이름의 릴레이션(relation)이 없습니다`, "titles_type_check") {
		t.Error("없는 표 에러를 규칙 위반으로 봤다")
	}
}

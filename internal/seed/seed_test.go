package seed

import (
	"strings"
	"testing"
	"time"

	"sqllings/internal/pg"
)

func last(t *testing.T, ch int, q string) string {
	t.Helper()
	out, _ := pg.Run("BEGIN;\nCREATE SCHEMA seed_test; SET LOCAL search_path TO seed_test;\n" + Script(ch) +
		"\n\\pset format unaligned\n\\pset tuples_only on\n" + q + "\nROLLBACK;\n")
	if out.ExitCode != 0 {
		t.Fatalf("장 %d 스크립트 실패:\n%s", ch, out.Text)
	}
	lines := strings.Split(strings.TrimSpace(out.Text), "\n")
	return lines[len(lines)-1]
}

func TestChapters(t *testing.T) {
	if err := pg.Ping(); err != nil {
		t.Skip(err)
	}
	if got := last(t, 1, "SELECT count(*) FROM titles;"); got != "8807" {
		t.Fatalf("ch1 %s", got)
	}
	if got := last(t, 1, "SELECT data_type FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = 'titles' AND column_name = 'date_added';"); got != "text" {
		t.Fatalf("ch1 date_added %s", got)
	}
	if got := last(t, 2, "SELECT count(*) FROM titles;"); got != "8808" {
		t.Fatalf("ch2 %s", got)
	}
	if got := last(t, 2, "SELECT count(DISTINCT rating) FROM titles;"); got != "14" {
		t.Fatalf("ch2 rating %s", got)
	}
	if got := last(t, 2, "SELECT data_type FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = 'titles' AND column_name = 'date_added';"); got != "date" {
		t.Fatal(got)
	}
	// 3장: 교안 fixes.sql 1~10 — 표 9개, 원본 칸(listed_in 등)은 지워졌다
	if got := last(t, 3, "SELECT count(*) FROM titles;"); got != "8807" {
		t.Fatalf("ch3 %s", got)
	}
	if got := last(t, 3, "SELECT count(*) FROM information_schema.tables WHERE table_schema = current_schema();"); got != "9" {
		t.Fatalf("ch3 표 수 %s", got)
	}
	if got := last(t, 3, "SELECT count(*) FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = 'titles' AND column_name IN ('rating', 'listed_in', 'country', 'director', 'cast_members');"); got != "0" {
		t.Fatalf("ch3 원본 칸 %s", got)
	}
	if got := last(t, 3, "SELECT (SELECT count(*) FROM title_genres) || ' ' || (SELECT count(*) FROM title_countries) || ' ' || (SELECT count(*) FROM people) || ' ' || (SELECT count(*) FROM title_cast);"); got != "19323 10018 40948 64124" {
		t.Fatalf("ch3 나눈 표 %s", got)
	}
}

func TestResetPlayKeepsOthers(t *testing.T) {
	if err := pg.Ping(); err != nil {
		t.Skip(err)
	}
	pg.Run("CREATE TABLE IF NOT EXISTS public.my_notes (a int);")
	defer pg.Run("DROP TABLE IF EXISTS public.my_notes;")
	others, err := ResetPlay(1)
	if err != nil {
		t.Fatal(err)
	}
	// 다른 패키지 테스트가 같은 DB에 표를 잠깐 만들 수 있으니 my_notes가 들어 있는지만 본다.
	if !strings.Contains(","+strings.Join(others, ",")+",", ",my_notes,") || strings.Contains(strings.Join(others, ","), "titles") {
		t.Fatalf("others=%v", others)
	}
	out, _ := pg.Run("SELECT count(*) FROM public.titles;")
	if !strings.Contains(out.Text, "8807") {
		t.Fatal(out.Text)
	}
	if _, err := ResetPlay(2); err != nil {
		t.Fatal(err)
	}
	out, _ = pg.Run("SELECT count(*) FROM public.titles;")
	if !strings.Contains(out.Text, "8808") {
		t.Fatal(out.Text)
	}
}

func TestResetPlayDoesNotHangOnLock(t *testing.T) {
	if err := pg.Ping(); err != nil {
		t.Skip(err)
	}
	if _, err := ResetPlay(1); err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() { pg.Run("BEGIN;\nSELECT count(*) FROM public.titles;\nSELECT pg_sleep(6);\nCOMMIT;"); close(done) }()
	time.Sleep(500 * time.Millisecond)
	start := time.Now()
	_, err := ResetPlay(1)
	if err == nil || !strings.Contains(err.Error(), "ROLLBACK") || time.Since(start) > 5*time.Second {
		t.Fatalf("err=%v after %v", err, time.Since(start))
	}
	<-done
	out, _ := pg.Run("SELECT count(*) FROM public.titles;")
	if !strings.Contains(out.Text, "8807") {
		t.Fatal("실패한 초기화가 터미널 2의 표를 반쯤 지웠다:", out.Text)
	}
}

// Ctrl+R가 실패하면 TUI 상태 줄 한 줄에 보인다. 서버의 여러 줄 원문 대신 한국어 한 줄 + 첫 에러 줄만.
func TestResetErrorIsShort(t *testing.T) {
	cases := map[string]string{
		"psql:<stdin>:5: ERROR:  42501: must be owner of table titles\nLOCATION:  aclcheck_error, aclchk.c:2823\n":       "ERROR:  42501: must be owner of table titles",
		"psql:<stdin>:5: 오류:  42501: titles 테이블의 소유주여야 함\n위치:  aclcheck_error, aclchk.c:2823\n":               "오류:  42501: titles 테이블의 소유주여야 함",
		"psql:<stdin>:3: NOTICE:  table \"x\" does not exist, skipping\npsql:<stdin>:9: ERROR:  53100: could not extend\n": "ERROR:  53100: could not extend",
	}
	for raw, want := range cases {
		msg := resetError(raw).Error()
		if strings.Contains(msg, "\n") || !strings.HasPrefix(msg, "연습 데이터를 만들지 못했습니다") || !strings.HasSuffix(msg, want) || strings.Contains(msg, "psql:<stdin>") {
			t.Errorf("%q → %q", raw, msg)
		}
	}
	lock := resetError("psql:<stdin>:5: ERROR:  55P03: canceling statement due to lock timeout\nCONTEXT:  ...\n").Error()
	if !strings.Contains(lock, "ROLLBACK") || strings.Contains(lock, "lock timeout") {
		t.Fatal(lock)
	}
}

func TestResetServerDownIsKorean(t *testing.T) {
	t.Setenv("PGPORT", "5999")
	_, err := ResetPlay(1)
	if err == nil || strings.Contains(err.Error(), "\n") || !strings.Contains(err.Error(), "서버가 꺼져") {
		t.Fatalf("%v", err)
	}
}

// 3장 표 9개를 만든 뒤 1장으로 되돌리면 3장 표가 남지 않아야 한다(다른 테이블 목록에도 안 나온다).
func TestResetPlayChapter3ThenBack(t *testing.T) {
	if err := pg.Ping(); err != nil {
		t.Skip(err)
	}
	if _, err := ResetPlay(3); err != nil {
		t.Fatal(err)
	}
	out, _ := pg.Run("SELECT count(*) FROM public.title_cast;")
	if !strings.Contains(out.Text, "64124") {
		t.Fatal(out.Text)
	}
	others, err := ResetPlay(1)
	if err != nil {
		t.Fatal(err)
	}
	for _, o := range others {
		for _, tb := range Tables {
			if o == tb {
				t.Fatalf("장 표 %s가 다른 테이블로 남았다: %v", tb, others)
			}
		}
	}
	out, _ = pg.Run("SELECT count(*) FROM pg_tables WHERE schemaname = 'public' AND tablename = 'people';")
	if !strings.Contains(out.Text, "0") {
		t.Fatal(out.Text)
	}
}

package exercises

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"sqllings/internal/exercise"
	"sqllings/internal/grade"
	"sqllings/internal/pg"
)

var update = flag.Bool("update", false, "example.txt를 기준 SQL 결과로 다시 쓴다")

// exampleRows는 결과 예시에 보여 줄 앞쪽 행 수다.
const exampleRows = 2

var rowCount = regexp.MustCompile(`^\((\d+) rows?\)$|^\((\d+)개 행\)$`)

// exampleOf는 기준 SQL의 psql 결과에서 머리줄 + 앞 exampleRows행 + 총 행 수를 만든다.
// 결과가 한 행인 문제(숫자 하나 · 한 행 여러 열)는 예시를 두지 않는다(값이 곧 답이다).
func exampleOf(display string) string {
	lines := strings.Split(strings.TrimRight(display, "\n"), "\n")
	if len(lines) < 3 {
		return ""
	}
	var body []string
	n := ""
	for _, l := range lines[2:] {
		if m := rowCount.FindStringSubmatch(strings.TrimSpace(l)); m != nil {
			n = m[1] + m[2]
			break
		}
		body = append(body, l)
	}
	if n == "" || len(body) == 1 {
		return ""
	}
	show := body
	if len(show) > exampleRows {
		show = append(append([]string{}, show[:exampleRows]...), " …")
	}
	return strings.Join(append(append([]string{lines[0], lines[1]}, show...), fmt.Sprintf("(총 %s행)", n)), "\n")
}

func TestExamples(t *testing.T) {
	if err := pg.Ping(); err != nil {
		t.Skip(err)
	}
	for _, ex := range All() {
		want := ""
		if ex.Mode == exercise.ModeResult {
			r := grade.Grade(ex, ex.Reference)
			if !r.Passed {
				t.Fatalf("%s: 기준 실패 %s", ex.Slug, r.Message)
			}
			want = exampleOf(r.Display)
		}
		path := ex.Slug + "/example.txt"
		if *update {
			if want == "" {
				os.Remove(path)
				continue
			}
			if err := os.WriteFile(path, []byte(want+"\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if ex.Example != want {
			t.Errorf("%s: example.txt가 기준 결과와 다르다 — go test ./exercises -run TestExamples -update", ex.Slug)
		}
	}
}

// Package grade는 학생 SQL을 채점 스키마 안 트랜잭션에서 실행해 기준과 비교한다.
// 한 번의 채점 = psql 하나 + 트랜잭션 하나. 학생 놀이터(public)는 건드리지 않는다.
package grade

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"sqllings/internal/exercise"
	"sqllings/internal/pg"
	"sqllings/internal/seed"
	"sqllings/internal/sqltext"
)

type Result struct {
	Passed  bool
	Message string // 한국어 안내
	Display string // 학생 SQL을 psql로 실행한 화면
}

// Timeout은 한 문장이 끝나야 하는 시간이다.
var Timeout = "10s"

const (
	fieldSep  = "<<F>>"
	recordSep = "<<R>>"
	nullMark  = "<<N>>"
	maxLines  = 200
)

func machine() string {
	return "\\set QUIET on\n\\pset format unaligned\n\\pset tuples_only on\n\\pset fieldsep " + fieldSep +
		"\n\\pset recordsep " + recordSep + "\n\\pset null " + nullMark + "\n"
}

const human = "\\pset format aligned\n\\pset tuples_only off\n\\pset null ''\n\\pset fieldsep '|'\n\\pset recordsep '\\n'\n\\set QUIET off\n"

func withSemicolon(sql string) string { return strings.TrimSpace(sql) + "\n;\n" }

func build(ex *exercise.Exercise, student string) string {
	var b strings.Builder
	b.WriteString("\\set QUIET on\nBEGIN;\nSET LOCAL statement_timeout = '" + Timeout + "';\n")
	// 앞선 채점이 비정상으로 끝나 스키마가 남았어도 스스로 복구한다(트랜잭션 안이라 지워도 ROLLBACK 된다).
	b.WriteString("DROP SCHEMA IF EXISTS sqllings_grade CASCADE;\nCREATE SCHEMA sqllings_grade;\nSET LOCAL search_path TO sqllings_grade;\n")
	b.WriteString(seed.Script(ex.Chapter))
	b.WriteString("\nSAVEPOINT s0;\n")
	b.WriteString(machine())
	b.WriteString("\\echo @@R@@\n")
	switch ex.Mode {
	case exercise.ModeResult:
		b.WriteString(withSemicolon(ex.Reference))
	case exercise.ModeState:
		b.WriteString(withSemicolon(ex.Reference))
		b.WriteString("\\echo @@CHECK@@\n")
		b.WriteString(withSemicolon(ex.Check))
	}
	b.WriteString("\\echo @@R-END@@\nROLLBACK TO s0;\n")
	b.WriteString(human)
	b.WriteString("\\echo @@D@@\n")
	b.WriteString(withSemicolon(student))
	b.WriteString("\\set QUIET on\n\\echo @@D-END@@\nROLLBACK TO s0;\n")
	if ex.Mode != exercise.ModeError {
		b.WriteString(machine())
		b.WriteString("\\echo @@S@@\n")
		b.WriteString(withSemicolon(student))
		if ex.Mode == exercise.ModeState {
			b.WriteString("\\echo @@CHECK@@\n")
			b.WriteString(withSemicolon(ex.Check))
		}
		b.WriteString("\\echo @@S-END@@\n")
	}
	b.WriteString("ROLLBACK;\n")
	return b.String()
}

// section은 start 표지 뒤부터 end 표지 앞까지를 돌려준다. end가 없으면 ok=false와 start 뒤 전부.
func section(text, start, end string) (string, bool) {
	_, after, found := strings.Cut(text, start+"\n")
	if !found {
		return "", false
	}
	body, _, ok := strings.Cut(after, end)
	return body, ok
}

var stdinPrefix = regexp.MustCompile(`(?m)^psql:<stdin>:\d+: `)

func clean(display string) string {
	s := stdinPrefix.ReplaceAllString(display, "")
	s = strings.TrimRight(s, "\n")
	lines := strings.Split(s, "\n")
	if len(lines) > maxLines {
		s = strings.Join(lines[:maxLines], "\n") + fmt.Sprintf("\n… (%d줄 더)", len(lines)-maxLines)
	}
	return s
}

func rows(sec string) [][]string {
	sec = strings.TrimRight(sec, "\n")
	if sec == "" {
		return nil
	}
	var out [][]string
	for _, r := range strings.Split(sec, recordSep) {
		out = append(out, strings.Split(strings.TrimRight(r, "\n"), fieldSep))
	}
	return out
}

func join(r []string) string { return strings.Join(r, fieldSep) }

func compareResult(ex *exercise.Exercise, ref, stu [][]string) Result {
	refCols, stuCols := 0, 0
	if len(ref) > 0 {
		refCols = len(ref[0])
	}
	if len(stu) > 0 {
		stuCols = len(stu[0])
	}
	if len(ref) > 0 && len(stu) > 0 && refCols != stuCols {
		return Result{Message: fmt.Sprintf("열이 %d개 나와야 하는데 %d개입니다", refCols, stuCols)}
	}
	if len(ref) != len(stu) {
		msg := fmt.Sprintf("%d행이어야 하는데 %d행입니다", len(ref), len(stu))
		if len(stu) == 0 {
			msg += " — 0행이 나오면 대소문자나 = NULL을 의심해 보세요"
		}
		return Result{Message: msg}
	}
	a, b := make([]string, len(ref)), make([]string, len(stu))
	for i := range ref {
		a[i], b[i] = join(ref[i]), join(stu[i])
	}
	if ex.Ordered {
		for i := range a {
			if a[i] != b[i] {
				sa, sb := append([]string(nil), a...), append([]string(nil), b...)
				sort.Strings(sa)
				sort.Strings(sb)
				if strings.Join(sa, "\n") == strings.Join(sb, "\n") {
					return Result{Message: "값은 다 맞는데 순서가 다릅니다. 같은 값이 여럿이면 두 번째 정렬 기준이 필요합니다"}
				}
				return Result{Message: fmt.Sprintf("행 수는 맞지만 값이 다릅니다. 처음 다른 곳: %d행 %d열", i+1, firstDiff(ref[i], stu[i])+1)}
			}
		}
		return Result{Passed: true, Message: "정답입니다!"}
	}
	sort.Strings(a)
	sort.Strings(b)
	for i := range a {
		if a[i] != b[i] {
			msg := "행 수는 맞지만 값이 다릅니다"
			if len(stu) == 1 && len(stu[0]) == 1 && stu[0][0] == "0" {
				msg += " — 0이 나오면 대소문자나 = NULL을 의심해 보세요"
			}
			// 정답 쪽 행을 보여 주면 답이 드러난다. 학생 결과에서 정답에 없는 행의 번호와 값이 다른 열만 알려 준다.
			if len(stu) > 1 {
				msg += fmt.Sprintf(". 정답에 없는 행: 내 결과 %d행", extraRow(ref, stu)+1)
			}
			if c := diffColumn(ref, stu); c >= 0 && refCols > 1 {
				msg += fmt.Sprintf(" · 값이 다른 열: %d열", c+1)
			}
			return Result{Message: msg}
		}
	}
	return Result{Passed: true, Message: "정답입니다!"}
}

// extraRow는 학생 결과(나온 순서 그대로)에서 기준에 없는 첫 행의 위치다. 같은 행이 여럿이면 개수까지 센다.
func extraRow(ref, stu [][]string) int {
	left := map[string]int{}
	for _, r := range ref {
		left[join(r)]++
	}
	for i, r := range stu {
		if left[join(r)] == 0 {
			return i
		}
		left[join(r)]--
	}
	return 0
}

// diffColumn은 열마다 값 모음(순서 무시)을 비교해 처음 다른 열의 위치를 준다. 모두 같으면(조합만 다르면) -1.
func diffColumn(ref, stu [][]string) int {
	for c := range ref[0] {
		x, y := make([]string, len(ref)), make([]string, len(stu))
		for i := range ref {
			x[i] = ref[i][c]
		}
		for i := range stu {
			y[i] = stu[i][c]
		}
		sort.Strings(x)
		sort.Strings(y)
		if strings.Join(x, recordSep) != strings.Join(y, recordSep) {
			return c
		}
	}
	return -1
}

// violates는 에러 줄이 그 규칙(제약조건) 위반인지 본다. 영어 `constraint "이름"` · 한국어 `"이름" … 제약 조건`.
// 표 이름 같은 다른 자리에 같은 글자가 있어도 통과시키지 않는다.
func violates(line, name string) bool {
	q := `"` + name + `"`
	return strings.Contains(line, "constraint "+q) || (strings.Contains(line, q) && strings.Contains(line, "제약"))
}

func firstDiff(a, b []string) int {
	for i := range a {
		if i >= len(b) || a[i] != b[i] {
			return i
		}
	}
	return 0
}

// Grade는 학생 SQL 하나를 채점한다.
func Grade(ex *exercise.Exercise, student string) Result {
	student = sqltext.Normalize(student)
	stmts := sqltext.Split(student)
	if len(stmts) == 0 {
		return Result{Message: "정답 칸이 비어 있습니다. 터미널 2에서 된 SQL을 붙여넣으세요"}
	}
	if why := sqltext.Forbidden(student); why != "" {
		return Result{Message: why}
	}
	if ex.Mode == exercise.ModeResult && len(stmts) != 1 {
		return Result{Message: fmt.Sprintf("이 문제는 SELECT 한 문장으로 답합니다 (지금 %d문장)", len(stmts))}
	}
	if ex.FirstWord != "" && sqltext.FirstWord(stmts[0]) != ex.FirstWord {
		return Result{Message: "이 문제는 " + ex.FirstWord + " 문장으로 답합니다"}
	}
	// 원문을 그대로 넘긴다. 문장을 다시 이어 붙이면 줄 끝 -- 주석 안으로 세미콜론이 들어간다.
	out, err := pg.Run(build(ex, student))
	if err != nil {
		return Result{Message: "채점기를 실행하지 못했습니다: " + err.Error()}
	}
	refSec, ok := section(out.Text, "@@R@@", "@@R-END@@")
	if !ok {
		// 접속 실패(서버 꺼짐 등)이거나 기준 SQL이 실패했다. 학생 탓이 아니다.
		if err := pg.Ping(); err != nil {
			return Result{Message: "채점기를 실행하지 못했습니다: " + err.Error()}
		}
		return Result{Message: "채점기를 실행하지 못했습니다 (기준 SQL 실패 — 강사에게 알려 주세요)", Display: clean(out.Text)}
	}
	display, finished := section(out.Text, "@@D@@", "@@D-END@@")
	r := Result{Display: clean(display)}
	// 한국어 서버(Windows 기본 로케일)는 "명령실행시간 초과로 작업을 취소합니다"
	if strings.Contains(display, "canceling statement due to statement timeout") || strings.Contains(display, "명령실행시간 초과") {
		r.Message = "10초 안에 끝나지 않았습니다. 조건이 빠진 JOIN이나 너무 큰 결과를 의심해 보세요"
		return r
	}
	if ex.Mode == exercise.ModeError {
		switch {
		case finished:
			r.Message = "에러 없이 들어갔습니다. 규칙 " + ex.ErrorName + "에 걸리는 값을 넣어 보세요"
		// 규칙 이름은 첫 에러 줄에서만 찾는다 — DETAIL의 "Failing row contains (…)"나 표 이름에 들어 있어도 통과시키지 않는다.
		case !violates(pg.FirstErrorLine(display), ex.ErrorName):
			r.Message = "에러는 났지만 다른 규칙(또는 다른 이유)입니다. 에러 문구를 읽어 보세요"
		default:
			r.Passed, r.Message = true, "정답입니다! 에러 문구에 규칙 이름이 보입니다"
		}
		return r
	}
	if !finished {
		r.Message = "SQL에서 에러가 났습니다. 에러를 소리 내어 읽어 보세요 — 터미널 2에서 같은 SQL을 쳐 보면 같은 에러가 납니다"
		return r
	}
	stuSec, _ := section(out.Text, "@@S@@", "@@S-END@@")
	if ex.Mode == exercise.ModeState {
		_, refCheck, _ := strings.Cut(refSec, "@@CHECK@@\n")
		_, stuCheck, _ := strings.Cut(stuSec, "@@CHECK@@\n")
		if strings.TrimSpace(refCheck) == strings.TrimSpace(stuCheck) {
			r.Passed, r.Message = true, "정답입니다!"
		} else {
			r.Message = "실행은 됐지만 결과 상태가 다릅니다. 터미널 2에서 고치기 전과 후를 SELECT로 비교해 보세요"
		}
		return r
	}
	c := compareResult(ex, rows(refSec), rows(stuSec))
	c.Display = r.Display
	return c
}

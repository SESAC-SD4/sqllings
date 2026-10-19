package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"

	"sqllings/exercises"
	"sqllings/internal/console"
	"sqllings/internal/exercise"
	"sqllings/internal/grade"
	"sqllings/internal/pg"
	"sqllings/internal/progress"
	"sqllings/internal/seed"
	"sqllings/internal/tui"
	"sqllings/internal/workspace"
)

func main() {
	console.Init()
	needDB := len(os.Args) < 2 || !(os.Args[1] == "list" || os.Args[1] == "hint" || os.Args[1] == "help" || os.Args[1] == "-h" || os.Args[1] == "--help")
	if needDB {
		if err := preflight(); err != nil {
			fmt.Fprintln(os.Stderr, "❌", err)
			os.Exit(1)
		}
	}
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout))
}

// preflight는 psql 18 이상과 sqllings DB 접속을 확인한다.
func preflight() error {
	v, err := pg.Version()
	if err != nil {
		return err
	}
	if v < 18 {
		return fmt.Errorf("psql 버전이 %d입니다. 18 이상이 필요합니다 (교안 02의 PATH 설정을 확인하세요)", v)
	}
	return pg.Ping()
}

func self() string {
	if runtime.GOOS == "windows" {
		return "./sqllings.exe"
	}
	return "./sqllings"
}

func run(args []string, in io.Reader, out io.Writer) int {
	prog, err := progress.Load()
	if err != nil {
		fmt.Fprintf(out, "❌ 진행 상황을 읽지 못했습니다: %v\n", err)
		return 1
	}
	cmd := "watch"
	if len(args) > 0 {
		cmd = args[0]
	}
	switch cmd {
	case "watch":
		if err := tui.Run(exercises.All(), prog); err != nil {
			fmt.Fprintf(out, "❌ 화면을 띄우지 못했습니다: %v\n대신 `%s list`, `%s check N 파일`을 쓸 수 있습니다\n", err, self(), self())
			return 1
		}
		return 0
	case "list":
		return cmdList(out, prog)
	case "start":
		return cmdStart(args[1:], in, out, prog)
	case "check":
		return cmdCheck(args[1:], out, prog)
	case "hint":
		return cmdHint(args[1:], out, prog)
	case "reset":
		return cmdReset(args[1:], in, out, prog)
	case "help", "-h", "--help":
		cmdHelp(out)
		return 0
	}
	fmt.Fprintf(out, "❌ 모르는 명령입니다: %s\n\n", cmd)
	cmdHelp(out)
	return 1
}

func cmdHelp(out io.Writer) {
	fmt.Fprintf(out, `sqllings — SQL 연습 문제

  %[1]s                 분할 화면으로 시작 (= %[1]s watch)
  %[1]s list            문제 목록
  %[1]s start N         N번 문제를 보여 주고, 터미널 2의 표를 그 문제에 맞춥니다
  %[1]s check N 파일     파일에 적은 SQL로 N번을 채점합니다
  %[1]s hint N          N번 힌트를 한 단계씩
  %[1]s reset N         터미널 2의 표(public)를 N번 문제의 처음 상태로 되돌립니다
  %[1]s help            이 도움말

터미널 2: psql -U sd4 -d sqllings
`, self())
}

func pick(args []string, out io.Writer, usage string) *exercise.Exercise {
	all := exercises.All()
	if len(args) < 1 {
		fmt.Fprintf(out, "사용법: %s %s (N은 1~%d)\n", self(), usage, len(all))
		return nil
	}
	id, err := strconv.Atoi(args[0])
	ex := exercises.ByID(id)
	if err != nil || ex == nil {
		fmt.Fprintf(out, "❌ %s번 문제가 없습니다 (1~%d)\n", args[0], len(all))
		return nil
	}
	return ex
}

func cmdList(out io.Writer, prog *progress.Progress) int {
	all := exercises.All()
	n, last := 0, 0
	for _, ex := range all {
		if ex.Chapter != last {
			last = ex.Chapter
			fmt.Fprintf(out, "\n[%s]\n", exercise.ChapterName(ex.Chapter))
		}
		mark := "⬜"
		if prog.IsPassed(ex.ID) {
			mark = "✅"
			n++
		}
		fmt.Fprintf(out, "%s %s\n", mark, ex.Title)
	}
	fmt.Fprintf(out, "\n%d/%d 통과\n", n, len(all))
	return 0
}

func yes(in io.Reader) bool {
	ans, _ := bufio.NewReader(in).ReadString('\n')
	return strings.TrimSpace(ans) == "y"
}

func resetTo(ch int, out io.Writer, prog *progress.Progress) int {
	others, err := seed.ResetPlay(ch)
	if err != nil {
		fmt.Fprintln(out, "❌", err)
		return 1
	}
	prog.PlayChapter = ch
	if err := prog.Save(); err != nil {
		fmt.Fprintf(out, "❌ 진행 상황을 저장하지 못했습니다: %v\n", err)
		return 1
	}
	fmt.Fprintln(out, "✅ 터미널 2의 표를 되돌렸습니다.")
	if len(others) > 0 {
		fmt.Fprintf(out, "다른 테이블은 그대로 두었습니다: %s\n", strings.Join(others, ", "))
	}
	return 0
}

func cmdStart(args []string, in io.Reader, out io.Writer, prog *progress.Progress) int {
	ex := pick(args, out, "start N")
	if ex == nil {
		return 1
	}
	if prog.PlayChapter != ex.Chapter {
		fmt.Fprint(out, "터미널 2의 표가 이 문제와 다릅니다. 이 문제의 표로 바꿀까요? 터미널 2에서 한 작업이 지워집니다 (y/n) ")
		if yes(in) {
			if code := resetTo(ex.Chapter, out, prog); code != 0 {
				return code
			}
		}
	}
	prog.LastStarted = ex.ID
	_ = prog.Save()
	fmt.Fprintf(out, "\n# %s\n\n%s\n\n터미널 2: psql -U sd4 -d sqllings\n풀었으면: %s check %d 파일.sql   ·   막히면: %s hint %d\n",
		ex.Title, strings.TrimSpace(ex.Prompt), self(), ex.ID, self(), ex.ID)
	return 0
}

func cmdCheck(args []string, out io.Writer, prog *progress.Progress) int {
	ex := pick(args, out, "check N 파일")
	if ex == nil {
		return 1
	}
	if len(args) < 2 {
		fmt.Fprintf(out, "사용법: %s check %d 파일.sql\n", self(), ex.ID)
		return 1
	}
	b, err := os.ReadFile(args[1])
	if err != nil {
		fmt.Fprintf(out, "❌ 파일을 읽지 못했습니다: %v\n", err)
		return 1
	}
	// Windows 편집기가 CRLF로 저장한 파일도 LF와 같게 다룬다(TUI 정답 칸의 붙여넣기와 같음).
	sql := strings.ReplaceAll(string(b), "\r\n", "\n")
	_ = workspace.SaveAnswer(ex.Slug, sql)
	r := grade.Grade(ex, sql)
	if r.Display != "" {
		fmt.Fprintln(out, r.Display)
		fmt.Fprintln(out)
	}
	if !r.Passed {
		fmt.Fprintf(out, "❌ %s — %s\n\n막히면: %s hint %d\n", ex.Title, r.Message, self(), ex.ID)
		return 1
	}
	prog.MarkPassed(ex.ID)
	if err := prog.Save(); err != nil {
		fmt.Fprintf(out, "❌ 진행 상황을 저장하지 못했습니다: %v\n", err)
		return 1
	}
	fmt.Fprintf(out, "✅ %s — %s\n", ex.Title, r.Message)
	if next := exercises.ByID(ex.ID + 1); next != nil {
		fmt.Fprintf(out, "\n다음: %s start %d\n", self(), next.ID)
	}
	return 0
}

func cmdHint(args []string, out io.Writer, prog *progress.Progress) int {
	ex := pick(args, out, "hint N")
	if ex == nil {
		return 1
	}
	level := prog.NextHint(ex.ID, len(ex.Hints))
	if err := prog.Save(); err != nil {
		fmt.Fprintf(out, "❌ 진행 상황을 저장하지 못했습니다: %v\n", err)
		return 1
	}
	for i := 0; i < level; i++ {
		fmt.Fprintf(out, "힌트 %d/%d\n%s\n\n", i+1, len(ex.Hints), ex.Hints[i])
	}
	if level == len(ex.Hints) {
		fmt.Fprintln(out, "마지막 힌트입니다.")
	}
	return 0
}

func cmdReset(args []string, in io.Reader, out io.Writer, prog *progress.Progress) int {
	ex := pick(args, out, "reset N")
	if ex == nil {
		return 1
	}
	fmt.Fprint(out, "터미널 2의 표를 이 문제의 처음 상태로 되돌릴까요? 터미널 2에서 한 작업이 지워집니다 (y/n) ")
	if !yes(in) {
		fmt.Fprintln(out, "그대로 두었습니다.")
		return 0
	}
	return resetTo(ex.Chapter, out, prog)
}

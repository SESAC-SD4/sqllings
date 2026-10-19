package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sqllings/exercises"
	"sqllings/internal/pg"
)

func runCmd(t *testing.T, in string, args ...string) (int, string) {
	t.Helper()
	var out bytes.Buffer
	code := run(args, strings.NewReader(in), &out)
	return code, out.String()
}

func TestList(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	code, out := runCmd(t, "", "list")
	total := fmt.Sprintf("0/%d 통과", len(exercises.All()))
	if code != 0 || !strings.Contains(out, "01 전체 작품 수") || !strings.Contains(out, total) || !strings.Contains(out, "[1장]") {
		t.Fatalf("%d\n%s", code, out)
	}
}

func TestUnknownCommand(t *testing.T) {
	code, out := runCmd(t, "", "nope")
	if code != 1 || !strings.Contains(out, "모르는 명령") {
		t.Fatal(out)
	}
}

func TestCheckFile(t *testing.T) {
	if err := pg.Ping(); err != nil {
		t.Skip(err)
	}
	t.Setenv("HOME", t.TempDir())
	f := filepath.Join(t.TempDir(), "a.sql")
	os.WriteFile(f, []byte("SELECT count(*) FROM titles;"), 0o644)
	code, out := runCmd(t, "", "check", "1", f)
	if code != 0 || !strings.Contains(out, "정답") || !strings.Contains(out, "8807") {
		t.Fatalf("%d\n%s", code, out)
	}
	_, out = runCmd(t, "", "list")
	if !strings.Contains(out, fmt.Sprintf("1/%d 통과", len(exercises.All()))) {
		t.Fatal(out)
	}
}

// Windows 메모장·VS Code가 CRLF로 저장한 파일도 LF와 똑같이 채점한다. 남기는 제출 SQL에도 CR이 없어야
// TUI 정답 칸에 다시 불러올 때 줄 사이가 벌어지지 않는다.
func TestCheckFileCRLF(t *testing.T) {
	if err := pg.Ping(); err != nil {
		t.Skip(err)
	}
	home := t.TempDir()
	t.Setenv("HOME", home)
	lf := filepath.Join(t.TempDir(), "lf.sql")
	crlf := filepath.Join(t.TempDir(), "crlf.sql")
	sql := "-- 2번 영화 수\nSELECT count(*) -- 전부 세기\nFROM titles\nWHERE type = 'Movie'\n;\n"
	os.WriteFile(lf, []byte(sql), 0o644)
	os.WriteFile(crlf, []byte(strings.ReplaceAll(sql, "\n", "\r\n")), 0o644)
	codeLF, outLF := runCmd(t, "", "check", "2", lf)
	codeCRLF, outCRLF := runCmd(t, "", "check", "2", crlf)
	if codeLF != 0 || codeCRLF != 0 || outLF != outCRLF || strings.Contains(outCRLF, "\r") {
		t.Fatalf("LF %d\n%s\nCRLF %d\n%q", codeLF, outLF, codeCRLF, outCRLF)
	}
	saved, err := os.ReadFile(filepath.Join(home, "sqllings-work", "answers", "02-count-movies.sql"))
	if err != nil || strings.Contains(string(saved), "\r") {
		t.Fatalf("남긴 제출 SQL: %q %v", saved, err)
	}
}

func TestCheckBadArgs(t *testing.T) {
	code, out := runCmd(t, "", "check", "99", "x.sql")
	if code != 1 || !strings.Contains(out, "문제가 없습니다") {
		t.Fatal(out)
	}
	code, out = runCmd(t, "", "check", "1", "/no/such/file.sql")
	if code != 1 || !strings.Contains(out, "파일") {
		t.Fatal(out)
	}
}

func TestHint(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	_, out := runCmd(t, "", "hint", "7")
	if !strings.Contains(out, "힌트 1/3") || strings.Contains(out, "힌트 2/3") {
		t.Fatal(out)
	}
	_, out = runCmd(t, "", "hint", "7")
	if !strings.Contains(out, "힌트 2/3") {
		t.Fatal(out)
	}
}

func TestResetAsks(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	code, out := runCmd(t, "n\n", "reset", "1")
	if code != 0 || !strings.Contains(out, "그대로") {
		t.Fatalf("%d %s", code, out)
	}
}

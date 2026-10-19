// Package pg는 학생 컴퓨터의 psql을 하위 프로세스로 불러 SQL 스크립트를 실행한다.
// Go DB 드라이버를 쓰지 않는다: 학생이 쓰는 손님(psql)과 같은 화면·에러가 나와야 한다.
package pg

import (
	"unicode/utf8"

	"bytes"
	"errors"
	"fmt"
	"golang.org/x/text/encoding/korean"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func setDefault(env []string, key, val string, force bool) []string {
	for i, kv := range env {
		if strings.HasPrefix(kv, key+"=") {
			if force || kv == key+"=" {
				env[i] = key + "=" + val
			}
			return env
		}
	}
	return append(env, key+"="+val)
}

// Env는 학생 환경에 접속 기본값을 더한다. 학생이 정한 PGUSER·PGPASSWORD·PGHOST·PGPORT는 그대로 둔다.
func Env() []string {
	env := os.Environ()
	env = setDefault(env, "PGUSER", "sd4", false)
	env = setDefault(env, "PGPASSWORD", "sd4", false)
	env = setDefault(env, "PGDATABASE", "sqllings", true)
	env = setDefault(env, "PGCLIENTENCODING", "UTF8", true)
	return env
}

// Output은 stdout과 stderr를 나온 순서대로 합친 것이다.
type Output struct {
	Text     string
	ExitCode int
}

// Run은 스크립트를 stdin으로 넘겨 psql 하나로 실행한다. error는 psql을 띄우지 못했을 때만 돌려준다.
func Run(script string) (Output, error) {
	cmd := exec.Command("psql", "-X", "-q", "-v", "ON_ERROR_STOP=1", "-P", "pager=off", "-f", "-")
	cmd.Env = Env()
	cmd.Stdin = strings.NewReader(script)
	var buf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &buf, &buf
	err := cmd.Run()
	text := normalizeOutput(buf.String())
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return Output{text, ee.ExitCode()}, nil
	}
	if err != nil {
		return Output{}, &ConnError{Kind: ConnNoPsql, Raw: err.Error()}
	}
	return Output{text, 0}, nil
}

// normalizeOutput은 Windows psql이 낼 수 있는 CRLF를 LF로 바꾼다. 표지(@@R@@ 등)를 \n으로 찾기 때문이다.
func normalizeOutput(s string) string {
	return decodeCP949(strings.ReplaceAll(s, "\r\n", "\n"))
}

// decodeCP949 는 UTF-8 출력 사이에 섞인 CP949 바이트를 UTF-8로 바꾼다.
// Windows 한국어 서버(lc_messages Korean_Korea.949)는 "상세정보" 같은 머리말을 CP949로 보낸다.
func decodeCP949(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r != utf8.RuneError || size > 1 {
			b.WriteString(s[i : i+size])
			i += size
			continue
		}
		// 깨진 구간: 두 바이트씩(앞 바이트 0x81 이상) 모아 한꺼번에 바꾼다
		j := i
		for j+1 < len(s) && s[j] >= 0x81 {
			if r, n := utf8.DecodeRuneInString(s[j:]); r != utf8.RuneError || n > 1 {
				break
			}
			j += 2
		}
		if j == i {
			j = i + 1
		}
		out, err := korean.EUCKR.NewDecoder().String(s[i:j])
		if err != nil {
			out = "\uFFFD"
		}
		b.WriteString(strings.ToValidUTF8(out, "\uFFFD"))
		i = j
	}
	return b.String()
}

var verRe = regexp.MustCompile(`\(PostgreSQL\) (\d+)`)

func Version() (int, error) {
	out, err := exec.Command("psql", "--version").Output()
	if err != nil {
		return 0, &ConnError{Kind: ConnNoPsql, Raw: err.Error()}
	}
	m := verRe.FindSubmatch(out)
	if m == nil {
		return 0, fmt.Errorf("psql 버전을 읽지 못했습니다: %s", out)
	}
	return strconv.Atoi(string(m[1]))
}

type ConnKind int

const (
	ConnUnknown ConnKind = iota
	ConnServerDown
	ConnNoDatabase
	ConnBadPassword
	ConnNoRole
	ConnNoPsql
)

type ConnError struct {
	Kind ConnKind
	Raw  string
}

func classify(raw string) ConnKind {
	switch {
	case strings.Contains(raw, "No such file or directory"), strings.Contains(raw, "Connection refused"):
		return ConnServerDown
	// 한국어 서버(Windows): 치명적오류:  "sqllings" 데이터베이스 없음
	case strings.Contains(raw, `database "sqllings" does not exist`), strings.Contains(raw, `"sqllings" 데이터베이스 없음`):
		return ConnNoDatabase
	// 한국어 서버(Windows): 치명적오류:  사용자 "sd4"의 password 인증을 실패했습니다
	case strings.Contains(raw, "password authentication failed"), strings.Contains(raw, "password 인증을 실패"):
		return ConnBadPassword
	// 한국어 서버(Windows): 치명적오류:  "sd4" 롤(role) 없음
	case strings.Contains(raw, `role "`) && strings.Contains(raw, "does not exist"), strings.Contains(raw, "롤(role) 없음"):
		return ConnNoRole
	}
	return ConnUnknown
}

func (e *ConnError) Error() string {
	switch e.Kind {
	case ConnNoPsql:
		return "psql을 찾지 못했습니다. 터미널에서 psql --version이 되는지 확인하세요 (교안 02 PATH)"
	case ConnServerDown:
		return "PostgreSQL 서버가 꺼져 있습니다. mac: brew services start postgresql@18 · Windows: 서비스(postgresql-x64-18) 확인"
	case ConnNoDatabase:
		return "sqllings DB가 없습니다. 관리자로 접속해서 만드세요:\n" +
			"  1. 관리자로 접속   mac: psql -d postgres   Windows: psql -U postgres (비밀번호 postgres)\n" +
			"  2. 이 한 줄        CREATE DATABASE sqllings OWNER sd4;\n" +
			"  3. 나가기          \\q\n" +
			"  4. 다시 켜기       ./sqllings (Windows: ./sqllings.exe)"
	case ConnBadPassword:
		return "비밀번호가 틀렸습니다. sd4의 비밀번호는 sd4입니다"
	case ConnNoRole:
		return "sd4 사용자가 없습니다. 교안 03의 CREATE ROLE을 먼저 하세요"
	}
	return "PostgreSQL에 접속하지 못했습니다:\n" + e.Raw
}

var stdinPrefix = regexp.MustCompile(`^psql:<stdin>:\d+: `)

// FirstErrorLine은 psql 출력에서 첫 에러 줄(영어 ERROR: · 한국어 오류:)을 앞의 "psql:<stdin>:N: " 없이 돌려준다.
// 없으면 빈 글자.
func FirstErrorLine(text string) string {
	for _, line := range strings.Split(text, "\n") {
		if strings.Contains(line, "ERROR:") || strings.Contains(line, "오류:") {
			return stdinPrefix.ReplaceAllString(line, "")
		}
	}
	return ""
}

// Ping은 sqllings DB에 접속되는지 본다.
func Ping() error {
	out, err := Run("SELECT 1;")
	if err != nil {
		return err
	}
	if out.ExitCode != 0 {
		return &ConnError{Kind: classify(out.Text), Raw: strings.TrimSpace(out.Text)}
	}
	return nil
}

package pg

import (
	"strings"
	"testing"
)

func TestClassify(t *testing.T) {
	cases := map[string]ConnKind{
		`psql: 오류: "/tmp/.s.PGSQL.5432" 소켓으로 서버 접속 할 수 없음: No such file or directory`: ConnServerDown,
		`... 5432 포트로 서버 접속 할 수 없음: Connection refused`:                               ConnServerDown,
		`FATAL:  database "sqllings" does not exist`:                                  ConnNoDatabase,
		`치명적오류:  "sqllings" 데이터베이스 없음`:                                                ConnNoDatabase,
		`치명적오류:  "sd4" 롤(role) 없음`:                                                    ConnNoRole,
		`FATAL:  password authentication failed for user "sd4"`:                       ConnBadPassword,
		`psql: 오류: "localhost" (::1), 5432 포트로 서버 접속 할 수 없음: 치명적오류:  사용자 "sd4"의 password 인증을 실패했습니다`: ConnBadPassword,
		`FATAL:  role "sd4" does not exist`: ConnNoRole,
		`something else`:                    ConnUnknown,
	}
	for raw, want := range cases {
		if got := classify(raw); got != want {
			t.Errorf("%q → %v, want %v", raw, got, want)
		}
	}
}

func TestConnErrorMessages(t *testing.T) {
	if !strings.Contains((&ConnError{Kind: ConnNoDatabase}).Error(), "CREATE DATABASE sqllings OWNER sd4;") {
		t.Fatal("DB 없음 안내에 만드는 한 줄이 없다")
	}
	if !strings.Contains((&ConnError{Kind: ConnServerDown}).Error(), "brew services start postgresql@18") {
		t.Fatal("서버 꺼짐 안내")
	}
}

func TestEnvDefaults(t *testing.T) {
	t.Setenv("PGUSER", "")
	t.Setenv("PGPASSWORD", "")
	t.Setenv("PGDATABASE", "other")
	env := strings.Join(Env(), "\n")
	for _, want := range []string{"PGUSER=sd4", "PGPASSWORD=sd4", "PGDATABASE=sqllings", "PGCLIENTENCODING=UTF8"} {
		if !strings.Contains(env, want) {
			t.Errorf("Env에 %s 없음", want)
		}
	}
}

func TestEnvKeepsStudentUser(t *testing.T) {
	t.Setenv("PGUSER", "hong")
	if !strings.Contains(strings.Join(Env(), "\n"), "PGUSER=hong") {
		t.Fatal("학생이 정한 PGUSER를 덮어썼다")
	}
}

func TestRunAgainstServer(t *testing.T) {
	if err := Ping(); err != nil {
		t.Skip("PostgreSQL 없음: ", err)
	}
	out, err := Run("SELECT 1 + 1 AS n;")
	if err != nil || out.ExitCode != 0 || !strings.Contains(out.Text, "2") {
		t.Fatalf("%v %+v", err, out)
	}
	out, _ = Run("SELECT * FROM no_such_table;")
	if out.ExitCode == 0 || !strings.Contains(out.Text, "no_such_table") {
		t.Fatalf("에러가 안 보인다: %+v", out)
	}
}

func TestVersion(t *testing.T) {
	if err := Ping(); err != nil {
		t.Skip(err)
	}
	if v, err := Version(); err != nil || v < 18 {
		t.Fatalf("%d %v", v, err)
	}
}

func TestPingServerDown(t *testing.T) {
	t.Setenv("PGPORT", "5999") // 아무도 안 듣는 포트
	err := Ping()
	ce, ok := err.(*ConnError)
	if !ok || ce.Kind != ConnServerDown {
		t.Fatalf("%#v", err)
	}
}

func TestNormalizeOutputCRLF(t *testing.T) {
	if got := normalizeOutput("@@R@@\r\n1\r\n"); got != "@@R@@\n1\n" {
		t.Fatalf("%q", got)
	}
}

// Windows 한국어 서버는 DETAIL 같은 머리말을 CP949 바이트로 보낸다(2026-10-18 실기).
func TestNormalizeOutputDecodesCP949Label(t *testing.T) {
	cp949 := "\xbb\xf3\xbc\xbc\xc1\xa4\xba\xb8" // "상세정보"
	got := normalizeOutput("오류:  위반\r\n" + cp949 + ":  실패한 자료: (s9002)\r\n")
	want := "오류:  위반\n상세정보:  실패한 자료: (s9002)\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

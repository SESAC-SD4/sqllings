package sqltext

import (
	"reflect"
	"strings"
	"testing"
)

func TestSplit(t *testing.T) {
	cases := map[string][]string{
		"SELECT 1;\r\n":                     {"SELECT 1"},
		"SELECT 1":                          {"SELECT 1"},
		"SELECT 1;;":                        {"SELECT 1"},
		"SELECT 1; -- 끝":                   {"SELECT 1"},
		"SELECT ';' AS a; SELECT 2;":        {"SELECT ';' AS a", "SELECT 2"},
		"SELECT 'it''s; ok';":               {"SELECT 'it''s; ok'"},
		"/* a; b */ SELECT 1;":              {"/* a; b */ SELECT 1"},
		"UPDATE t\nSET a = 1\nWHERE b = 2;": {"UPDATE t\nSET a = 1\nWHERE b = 2"},
		`SELECT "a;b" FROM t;`:              {`SELECT "a;b" FROM t`},
		"  \n ":                             nil,
		"-- 주석만\n":                         nil,
	}
	for in, want := range cases {
		if got := Split(in); !reflect.DeepEqual(got, want) {
			t.Errorf("Split(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFirstWord(t *testing.T) {
	if w := FirstWord("-- 설명\n  insert into t values (1)"); w != "INSERT" {
		t.Fatal(w)
	}
	if w := FirstWord("/* x */select 1"); w != "SELECT" {
		t.Fatal(w)
	}
	if w := FirstWord("(SELECT 1)"); w != "" && w != "(SELECT" {
		t.Log(w)
	}
}

func TestForbidden(t *testing.T) {
	for _, s := range []string{"\\d titles", "SELECT 1;\n  \\q", "BEGIN; UPDATE t SET a=1;", "SELECT 1;\nCOMMIT;", "SET search_path TO public;", "DROP SCHEMA public CASCADE;", "create database x;"} {
		if Forbidden(s) == "" {
			t.Errorf("막아야 함: %q", s)
		}
	}
	for _, s := range []string{"SELECT '\\d';", "UPDATE titles SET description = 'begin' WHERE id = 1;", "SELECT 1", "SELECT 1 -- \\d"} {
		if r := Forbidden(s); r != "" {
			t.Errorf("막으면 안 됨: %q (%s)", s, r)
		}
	}
}

func TestForbiddenFixes(t *testing.T) {
	for _, s := range []string{
		"SELECT count(*) FROM titles \\! echo PWNED",
		"SELECT 1 \\gexec",
		"ABORT; CREATE TABLE x(a int);",
		"SELECT $$'$$; COMMIT; SELECT $$'$$",
		"SELECT count(*) FROM titles WHERE type = 'Movie;",
		`SELECT title FROM titles WHERE director = "Bong Joon Ho;`,
		"SELECT 1 /* 안 닫힘",
	} {
		if Forbidden(s) == "" {
			t.Errorf("막아야 함: %q", s)
		}
	}
	if r := Forbidden("SELECT count(*) FROM titles WHERE type = 'Movie;"); !strings.Contains(r, "따옴표") {
		t.Errorf("따옴표 안내가 아니다: %s", r)
	}
	for _, s := range []string{"SELECT $$a;b$$;", "SELECT E'it\\'s; x';", "SELECT $tag$ ' $tag$;"} {
		if r := Forbidden(s); r != "" {
			t.Errorf("막으면 안 됨: %q (%s)", s, r)
		}
		if n := len(Split(s)); n != 1 {
			t.Errorf("%q는 한 문장인데 %d", s, n)
		}
	}
}

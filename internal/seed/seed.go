// Package seed는 장마다 정해진 시작 상태(titles 테이블)를 만드는 SQL을 준다.
// 데이터: Netflix Movies and TV Shows — Shivam Bansal, Kaggle, CC0: Public Domain
package seed

import (
	_ "embed"
	"fmt"
	"strings"

	"sqllings/internal/pg"
)

//go:embed ch1.sql
var ch1 string

//go:embed ch2.sql
var ch2 string

//go:embed netflix_titles.csv
var csv string

// Chapters는 v1이 다루는 장 수다.
const Chapters = 2

// Script는 지금 search_path에 그 장의 titles를 만드는 SQL이다(DROP 없이 CREATE부터).
// ch1은 교안 04의 titles.sql과 같고, ch2는 교안 fixes.sql 1~5번과 《The Glory》를 더한다.
func Script(ch int) string {
	var b strings.Builder
	b.WriteString(ch1)
	b.WriteString("\nCOPY titles FROM STDIN WITH (FORMAT csv, HEADER true);\n")
	b.WriteString(strings.ReplaceAll(csv, "\r\n", "\n"))
	if !strings.HasSuffix(csv, "\n") {
		b.WriteString("\n")
	}
	b.WriteString("\\.\n")
	if ch >= 2 {
		b.WriteString(ch2)
	}
	return b.String()
}

// ResetPlay는 터미널 2의 표(public)의 titles를 지우고 그 장의 시작 상태로 다시 만든다.
// 학생이 직접 만든 다른 테이블은 지우지 않고 이름만 돌려준다.
func ResetPlay(ch int) ([]string, error) {
	// 한 트랜잭션으로 묶어 중간에 실패해도 반쯤 지워진 상태가 남지 않게 한다.
	// 터미널 2에 끝나지 않은 트랜잭션이 titles를 잡고 있으면 DROP이 끝없이 기다리므로 3초에서 끊는다.
	out, err := pg.Run("\\set VERBOSITY verbose\nBEGIN;\nSET LOCAL lock_timeout = '3s';\nSET LOCAL search_path TO public;\nDROP TABLE IF EXISTS titles CASCADE;\n" + Script(ch) +
		"\nCOMMIT;\n\\pset format unaligned\n\\pset tuples_only on\n\\echo @@OTHERS@@\n" +
		"SELECT coalesce(string_agg(tablename, ',' ORDER BY tablename), '') FROM pg_tables WHERE schemaname = 'public' AND tablename <> 'titles';\n")
	if err != nil {
		return nil, err
	}
	if out.ExitCode != 0 {
		if pg.FirstErrorLine(out.Text) == "" {
			// 에러 줄이 없으면 SQL까지 가지 못한 것(서버 꺼짐 등)이다. 접속 안내를 한국어로 준다.
			if err := pg.Ping(); err != nil {
				return nil, err
			}
		}
		return nil, resetError(out.Text)
	}
	_, after, _ := strings.Cut(out.Text, "@@OTHERS@@\n")
	after = strings.TrimSpace(after)
	if after == "" {
		return nil, nil
	}
	return strings.Split(after, ","), nil
}

// resetError는 실패한 psql 출력을 TUI 상태 줄 한 줄에 들어가는 한국어 안내로 바꾼다.
// 서버 원문은 첫 에러 줄만 붙인다(에러 줄이 없으면 처음 나온 줄).
func resetError(text string) error {
	if strings.Contains(text, "55P03") {
		return fmt.Errorf("터미널 2에 끝나지 않은 트랜잭션이 있어 데이터를 바꿀 수 없습니다. 터미널 2에서 ROLLBACK; 을 치고 다시 하세요")
	}
	line := pg.FirstErrorLine(text)
	if line == "" {
		line, _, _ = strings.Cut(strings.TrimSpace(text), "\n")
	}
	return fmt.Errorf("연습 데이터를 만들지 못했습니다 (강사에게 알려 주세요): %s", strings.TrimSpace(line))
}

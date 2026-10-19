// Package workspace는 연습 기록을 두는 ~/sqllings-work 폴더를 다룬다.
package workspace

import (
	"os"
	"path/filepath"
)

// Git Bash에서 HOME을 따로 정한 학생도 같은 곳을 보도록 HOME을 먼저 쓴다(gitlings와 같음).
func Base() string {
	home := os.Getenv("HOME")
	if home == "" {
		home, _ = os.UserHomeDir()
	}
	return filepath.Join(home, "sqllings-work")
}

func answerPath(slug string) string { return filepath.Join(Base(), "answers", slug+".sql") }

// SaveAnswer는 문제별 마지막 제출 SQL을 남긴다. 다음 날 해설 때 강사가 본다.
func SaveAnswer(slug, sql string) error {
	if err := os.MkdirAll(filepath.Dir(answerPath(slug)), 0o755); err != nil {
		return err
	}
	return os.WriteFile(answerPath(slug), []byte(sql), 0o644)
}

func LoadAnswer(slug string) string {
	b, _ := os.ReadFile(answerPath(slug))
	return string(b)
}

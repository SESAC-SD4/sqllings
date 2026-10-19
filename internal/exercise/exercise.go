// Package exercise는 문제 하나의 정의를 담는다.
package exercise

import (
	"regexp"
	"strings"
)

// Mode는 채점 방식이다.
type Mode int

const (
	ModeResult Mode = iota // SELECT 결과 표를 기준과 비교
	ModeState              // 실행한 뒤 상태 확인 SQL의 결과를 비교
	ModeError              // 정해진 규칙 이름의 에러가 나야 통과
)

type Exercise struct {
	ID        int
	Slug      string // "01-count-all"
	Title     string // "01 전체 작품 수"
	Optional  bool
	Chapter   int
	Mode      Mode
	Ordered   bool   // ModeResult: 순서까지 비교
	ErrorName string // ModeError: 에러에 들어 있어야 할 규칙 이름
	FirstWord string // 학생 첫 문장의 첫 단어 제한(예: "INSERT"). 빈 칸이면 제한 없음
	Prompt    string
	Hints     []string
	Reference string // 기준 SQL
	Check     string // ModeState: 상태 확인 SQL
}

var hintSep = regexp.MustCompile(`(?m)^---[ \t]*$`)

func ParseHints(raw string) []string {
	var out []string
	for _, part := range hintSep.Split(strings.ReplaceAll(raw, "\r\n", "\n"), -1) {
		if s := strings.TrimSpace(part); s != "" {
			out = append(out, s)
		}
	}
	return out
}

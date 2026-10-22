// Package exercise는 문제 하나의 정의를 담는다.
package exercise

import (
	"fmt"
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
	Module    string // 이 문제를 풀 시점(교안 모듈). 목록에 소제목으로 보인다. 빈 칸이면 없음
	Mode      Mode
	Ordered   bool   // ModeResult: 순서까지 비교
	ErrorName string // ModeError: 에러에 들어 있어야 할 규칙 이름
	FirstWord string // 학생 첫 문장의 첫 단어 제한(예: "INSERT"). 빈 칸이면 제한 없음
	Prompt    string
	Hints     []string
	Reference string // 기준 SQL
	Check     string // ModeState: 상태 확인 SQL
	Example   string // 결과 예시(머리줄 + 앞 몇 행 + 총 행 수, example.txt). 빈 칸이면 없음
	Setup     string // 채점 전에 장 데이터 위에 돌리는 준비 SQL(앞 문제에서 고친 상태 등). 빈 칸이면 없음
}

// chapterNames는 학생 화면에 보이는 파트 이름이다. 장 번호(수업 날짜 구분)는 강사용이라 화면에 내지 않는다.
var chapterNames = map[int]string{1: "꺼내기", 2: "고치기와 규칙", 3: "잇고 세기"}

// ChapterName은 장의 파트 이름을 돌려준다. 이름이 없는 장은 "파트 N".
func ChapterName(ch int) string {
	if n, ok := chapterNames[ch]; ok {
		return n
	}
	return fmt.Sprintf("파트 %d", ch)
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

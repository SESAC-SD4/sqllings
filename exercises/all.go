// Package exercises는 문제 목록이다. 문제마다 폴더 하나(prompt.md · hint.md · reference.sql · check.sql).
package exercises

import (
	"embed"
	"fmt"
	"sort"
	"strings"

	"sqllings/internal/exercise"
)

//go:embed */*.md */*.sql */*.txt
var files embed.FS

var registry []*exercise.Exercise

type opt struct {
	optional, ordered  bool
	mode               exercise.Mode
	errName, firstWord string
	module             string
}

func read(slug, name string, must bool) string {
	b, err := files.ReadFile(slug + "/" + name)
	if err != nil && must {
		panic(fmt.Sprintf("%s/%s를 읽지 못했습니다: %v", slug, name, err))
	}
	return strings.TrimSpace(string(b))
}

// readRaw는 앞 공백을 살린다(psql 표의 머리줄은 공백으로 시작한다). 끝 줄바꿈만 지운다.
func readRaw(slug, name string) string {
	b, _ := files.ReadFile(slug + "/" + name)
	return strings.TrimRight(string(b), "\n")
}

func register(id, chapter int, slug, title string, o opt) {
	hints := exercise.ParseHints(read(slug, "hint.md", true))
	if len(hints) != 3 {
		panic(fmt.Sprintf("%s/hint.md의 힌트는 3단계여야 합니다(지금 %d)", slug, len(hints)))
	}
	registry = append(registry, &exercise.Exercise{
		ID: id, Chapter: chapter, Module: o.module, Slug: slug, Title: title, Optional: o.optional,
		Mode: o.mode, Ordered: o.ordered, ErrorName: o.errName, FirstWord: o.firstWord,
		Prompt: read(slug, "prompt.md", true), Hints: hints,
		Reference: read(slug, "reference.sql", o.mode != exercise.ModeError),
		Check:     read(slug, "check.sql", o.mode == exercise.ModeState),
		Setup:     read(slug, "setup.sql", false),
		Example:   readRaw(slug, "example.txt"),
	})
}

func All() []*exercise.Exercise {
	out := append([]*exercise.Exercise(nil), registry...)
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func ByID(id int) *exercise.Exercise {
	for _, e := range registry {
		if e.ID == id {
			return e
		}
	}
	return nil
}

// 2장부터는 교안 모듈마다 몇 문제씩 푼다. 목록 소제목은 주제 이름만 — 언제 풀지는 강사가 수업에서 번호로 말한다(D71).
const (
	m10 = "바꾸기 (UPDATE·INSERT)"
	m11 = "WHERE 먼저"
	m12 = "날짜와 숫자"
	m13 = "규칙"
	m14 = "섞어 풀기"
)

func init() {
	register(1, 1, "01-count-all", "01 전체 작품 수", opt{})
	register(2, 1, "02-count-movies", "02 영화 수", opt{})
	register(3, 1, "03-fincher", "03 데이비드 핀처 감독 작품", opt{ordered: true})
	register(4, 1, "04-tv-before-2000", "04 2000년 전 TV", opt{})
	register(5, 1, "05-tv-2000s", "05 2000년대 TV", opt{})
	register(6, 1, "06-rating-in", "06 TV-MA 또는 R", opt{})
	register(7, 1, "07-no-director-by-type", "07 감독이 빈 작품, 종류별로", opt{})
	register(8, 1, "08-oldest-tv", "08 가장 오래된 TV 5편", opt{ordered: true})
	register(9, 1, "09-tv-rating-top5", "09 TV 등급별 편수 상위 5", opt{ordered: true})
	register(10, 1, "10-and-or-paren", "10 인도 또는 영국 영화 (괄호)", opt{})
	register(11, 1, "11-one-season", "11 시즌 하나뿐인 TV", opt{})
	register(12, 1, "12-night-ilike", "12 제목에 night", opt{})
	register(13, 1, "13-japan-included", "13★ 일본이 포함된 작품", opt{optional: true})
	register(14, 1, "14-rating-min", "14★ 등급 칸의 이상한 값", opt{optional: true, ordered: true})
	register(15, 2, "15-update-kingdom", "15 Kingdom 설명 고치기", opt{module: m10, mode: exercise.ModeState, firstWord: "UPDATE"})
	register(16, 2, "16-insert-favorite", "16 좋아하는 작품 넣기", opt{module: m10, mode: exercise.ModeState, firstWord: "INSERT"})
	register(17, 2, "17-ur-to-nr", "17 UR을 NR로", opt{module: m11, mode: exercise.ModeState, firstWord: "UPDATE"})
	register(18, 2, "18-comma-country", "18 쉼표로 시작하는 나라", opt{module: m11, mode: exercise.ModeState, firstWord: "UPDATE"})
	register(19, 2, "19-recent-korea", "19 최근 추가된 한국 작품", opt{module: m12, ordered: true})
	register(20, 2, "20-many-seasons", "20 5시즌 이상 TV", opt{module: m12})
	register(21, 2, "21-country-rule", "21 나라 칸 규칙 걸기", opt{module: m13, mode: exercise.ModeState, firstWord: "ALTER"})
	register(22, 2, "22-minutes-rule", "22 상영 시간 규칙 걸기", opt{module: m13, mode: exercise.ModeState, firstWord: "ALTER"})
	register(23, 2, "23-long-2020", "23 2020년 추가 긴 영화", opt{module: m14, ordered: true})
	register(24, 2, "24-added-2019-by-type", "24 2019년 추가, 종류별", opt{module: m14})
	register(25, 2, "25-long-movie-rating", "25 120분 넘는 영화 등급별", opt{module: m14, ordered: true})
	register(26, 2, "26-seasons-count", "26★ 시즌 수별 TV 편수", opt{module: m14, optional: true, ordered: true})
}

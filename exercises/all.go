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
	// 3장(잇고 세기)
	m21 = "잇기 (JOIN)"
	m22 = "빠진 것 찾기 (LEFT JOIN)"
	m23 = "묶어 세기 (GROUP BY)"
	m24 = "서브쿼리"
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
	register(27, 3, "27-squid-genres", "27 Squid Game의 장르", opt{module: m21})
	register(28, 3, "28-kingdom-director", "28 Kingdom의 감독", opt{module: m21})
	register(29, 3, "29-srk-recent", "29 Shah Rukh Khan 최근 5편", opt{module: m21, ordered: true})
	register(30, 3, "30-g-rating-2021", "30 2021년 전체관람가", opt{module: m21})
	register(31, 3, "31-bong-genres", "31★ 봉준호 작품의 장르", opt{module: m21, optional: true})
	register(32, 3, "32-no-country", "32 나라가 없는 작품 수", opt{module: m22})
	register(33, 3, "33-no-cast-by-type", "33 배우 없는 작품, 종류별", opt{module: m22})
	register(34, 3, "34-no-director-2021-movie", "34 감독 없는 2021년 영화", opt{module: m22})
	register(35, 3, "35-no-director-no-cast", "35★ 감독도 배우도 없음", opt{module: m22, optional: true})
	register(36, 3, "36-top-directors", "36 감독 작품 수 상위 3", opt{module: m23, ordered: true})
	register(37, 3, "37-year-count", "37 2015년 이후 연도별", opt{module: m23, ordered: true})
	register(38, 3, "38-actors-30", "38 30편 이상 나온 배우", opt{module: m23})
	register(39, 3, "39-tv-rating-name", "39 TV 등급 이름별 수", opt{module: m23, ordered: true})
	register(40, 3, "40-korea-genres", "40 한국 작품 장르 상위 5", opt{module: m23, ordered: true})
	register(41, 3, "41-director-count-zero", "41★ 감독 작품 수 0편", opt{module: m23, optional: true})
	register(42, 3, "42-korean-tv-count", "42 Korean TV Shows 수", opt{module: m24})
	register(43, 3, "43-oldest", "43 가장 오래된 작품", opt{module: m24})
	register(44, 3, "44-long-2021", "44 평균보다 긴 2021 영화", opt{module: m24})
	register(45, 3, "45-three-countries", "45 나라 셋 이상인 작품", opt{module: m24})
	register(46, 3, "46-max-cast", "46★ 가장 많은 배우 수", opt{module: m24, optional: true})
	register(47, 3, "47-korea-dramas", "47★ 한국 Dramas 작품 수", opt{module: m24, optional: true})
}

// Package exercises는 문제 목록이다. 문제마다 폴더 하나(prompt.md · hint.md · reference.sql · check.sql).
package exercises

import (
	"embed"
	"fmt"
	"sort"
	"strings"

	"sqllings/internal/exercise"
)

//go:embed */*.md */*.sql
var files embed.FS

var registry []*exercise.Exercise

type opt struct {
	optional, ordered  bool
	mode               exercise.Mode
	errName, firstWord string
}

func read(slug, name string, must bool) string {
	b, err := files.ReadFile(slug + "/" + name)
	if err != nil && must {
		panic(fmt.Sprintf("%s/%s를 읽지 못했습니다: %v", slug, name, err))
	}
	return strings.TrimSpace(string(b))
}

func register(id, chapter int, slug, title string, o opt) {
	hints := exercise.ParseHints(read(slug, "hint.md", true))
	if len(hints) != 3 {
		panic(fmt.Sprintf("%s/hint.md의 힌트는 3단계여야 합니다(지금 %d)", slug, len(hints)))
	}
	registry = append(registry, &exercise.Exercise{
		ID: id, Chapter: chapter, Slug: slug, Title: title, Optional: o.optional,
		Mode: o.mode, Ordered: o.ordered, ErrorName: o.errName, FirstWord: o.firstWord,
		Prompt: read(slug, "prompt.md", true), Hints: hints,
		Reference: read(slug, "reference.sql", o.mode != exercise.ModeError),
		Check:     read(slug, "check.sql", o.mode == exercise.ModeState),
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
}

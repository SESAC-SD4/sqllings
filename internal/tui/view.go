package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"sqllings/internal/exercise"
)

const (
	listWidth   = 32 // 가장 긴 줄 "  ⬜ 14★ 등급 칸의 이상한 값" + 여백
	paneExtra   = 4  // 테두리 2 + 좌우 여백 2
	inputHeight = 6
	minDesc     = 6 // 문제 칸 최소 높이
	minOut      = 5 // 결과 칸 최소 높이 (채점 전 빈 칸)
)

var (
	pane     = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63")).Padding(0, 1)
	selected = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("226"))
	passed   = lipgloss.NewStyle().Foreground(lipgloss.Color("78"))
	keysTxt  = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	chapter  = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
)

const keyHelp = "Enter 채점 · Tab 힌트 · ↑↓ 문제 · Ctrl+L 칸 비우기 · Ctrl+R 터미널 2 표 되돌리기 · Ctrl+X 전체 초기화 · PgUp/PgDn 스크롤 · Ctrl+C 종료"

func (m *Model) layout() {
	w := m.width - listWidth - paneExtra*2
	if w < 20 {
		w = 20
	}
	avail := m.height - 2 // 상태 줄 · 도움말 줄
	rest := avail - (inputHeight + 2) - 4
	if rest < 6 {
		rest = 6
	}
	m.rest = rest
	m.desc.Width, m.out.Width = w, w
	m.split(rest/2, 0)
	m.input.SetWidth(w)
	m.input.SetHeight(inputHeight)
}

// split은 결과 칸을 결과 줄 수만큼(최소 minOut) 주되 문제 칸 내용(descLines)을 가리지 않게 하고,
// 남는 높이는 모두 문제 칸에 준다.
func (m *Model) split(descLines, outLines int) {
	keep := descLines
	if keep > m.rest-minOut {
		keep = m.rest - minOut
	}
	if keep < minDesc {
		keep = minDesc
	}
	o := outLines
	if o > m.rest-keep {
		o = m.rest - keep
	}
	if o < minOut {
		o = minOut
	}
	if o > m.rest {
		o = m.rest
	}
	m.desc.Height, m.out.Height = m.rest-o, o
}

func (m Model) View() string {
	if m.width == 0 {
		return "불러오는 중..."
	}
	lines := []string{"문제"}
	curLine := 0
	lastCh, lastMod := 0, ""
	for i, ex := range m.exs {
		if ex.Chapter != lastCh {
			lastCh, lastMod = ex.Chapter, ""
			lines = append(lines, "", chapter.Render(exercise.ChapterName(ex.Chapter)))
		}
		if ex.Module != "" && ex.Module != lastMod {
			lastMod = ex.Module
			lines = append(lines, chapter.Render("· "+ex.Module))
		}
		mark := "⬜"
		if m.prog.IsPassed(ex.ID) {
			mark = "✅"
		}
		line := "  " + mark + " " + ex.Title
		switch {
		case i == m.cursor:
			line = selected.Render("> " + mark + " " + ex.Title)
			curLine = len(lines)
		case m.prog.IsPassed(ex.ID):
			line = passed.Render(line)
		}
		lines = append(lines, line)
	}
	total := m.desc.Height + m.out.Height + inputHeight + 4
	lines = listWindow(lines, curLine, total)
	left := pane.Width(listWidth).Height(total).Render(strings.Join(lines, "\n"))
	// lipgloss의 Width는 좌우 여백을 포함하므로 +2를 해야 줄이 다시 접히지 않는다(gitlings와 같음).
	desc := pane.Width(m.desc.Width + 2).Height(m.desc.Height).Render(m.desc.View())
	input := pane.Width(m.desc.Width + 2).Height(inputHeight).Render(m.input.View())
	out := pane.Width(m.desc.Width + 2).Height(m.out.Height).Render(m.out.View())
	right := lipgloss.JoinVertical(lipgloss.Left, desc, input, out)
	body := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	status := lipgloss.NewStyle().MaxWidth(m.width).Render(m.status)
	help := keysTxt.MaxWidth(m.width).Render(keyHelp)
	return lipgloss.JoinVertical(lipgloss.Left, body, status, help)
}

// listWindow는 목록이 창보다 길 때 고른 문제가 보이도록 잘라 낸다. 맨 위 "문제" 줄은 남긴다.
func listWindow(lines []string, cur, height int) []string {
	if len(lines) <= height || height < 2 {
		return lines
	}
	body, cur, h := lines[1:], cur-1, height-1
	start := cur - h/2
	if start < 0 {
		start = 0
	}
	if start > len(body)-h {
		start = len(body) - h
	}
	return append([]string{lines[0]}, body[start:start+h]...)
}

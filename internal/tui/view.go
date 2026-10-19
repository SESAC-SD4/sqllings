package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	listWidth   = 32 // 가장 긴 줄 "  ⬜ 14★ 등급 칸의 이상한 값" + 여백
	paneExtra   = 4  // 테두리 2 + 좌우 여백 2
	inputHeight = 6
)

var (
	pane     = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63")).Padding(0, 1)
	selected = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("226"))
	passed   = lipgloss.NewStyle().Foreground(lipgloss.Color("78"))
	keysTxt  = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	chapter  = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
)

const keyHelp = "Enter 채점 · Tab 힌트 · ↑↓ 문제 · Ctrl+L 칸 비우기 · Ctrl+R 놀이터 초기화 · Ctrl+X 전체 초기화 · PgUp/PgDn 스크롤 · Ctrl+C 종료"

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
	m.desc.Width, m.desc.Height = w, rest/2
	m.out.Width, m.out.Height = w, rest-rest/2
	m.input.SetWidth(w)
	m.input.SetHeight(inputHeight)
}

func (m Model) View() string {
	if m.width == 0 {
		return "불러오는 중..."
	}
	var list strings.Builder
	list.WriteString("문제\n")
	last := 0
	for i, ex := range m.exs {
		if ex.Chapter != last {
			last = ex.Chapter
			list.WriteString("\n" + chapter.Render(map[int]string{1: "1장 꺼내기", 2: "2장 고치기와 규칙"}[ex.Chapter]) + "\n")
		}
		mark := "⬜"
		if m.prog.IsPassed(ex.ID) {
			mark = "✅"
		}
		line := "  " + mark + " " + ex.Title
		switch {
		case i == m.cursor:
			line = selected.Render("> " + mark + " " + ex.Title)
		case m.prog.IsPassed(ex.ID):
			line = passed.Render(line)
		}
		list.WriteString(line + "\n")
	}
	total := m.desc.Height + m.out.Height + inputHeight + 4
	left := pane.Width(listWidth).Height(total).Render(list.String())
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

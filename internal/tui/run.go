package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"sqllings/internal/exercise"
	"sqllings/internal/grade"
	"sqllings/internal/progress"
	"sqllings/internal/seed"
)

func GlamourRenderer(dark bool) Renderer {
	style := "light"
	if dark {
		style = "dark"
	}
	return func(md string, width int) string {
		if width < 20 {
			width = 20
		}
		r, err := glamour.NewTermRenderer(glamour.WithStandardStyle(style), glamour.WithWordWrap(width-2))
		if err != nil {
			return md
		}
		out, err := r.Render(md)
		if err != nil {
			return md
		}
		return out
	}
}

// Run은 처음 켤 때 터미널 2의 표를 만들고 화면을 띄운다. 접속 점검은 부르는 쪽(main)이 먼저 한다.
func Run(exs []*exercise.Exercise, prog *progress.Progress) error {
	if prog.PlayChapter == 0 {
		ch := exs[0].Chapter
		for _, ex := range exs {
			if !prog.IsPassed(ex.ID) {
				ch = ex.Chapter
				break
			}
		}
		if _, err := seed.ResetPlay(ch); err != nil {
			return err
		}
		prog.PlayChapter = ch
		if err := prog.Save(); err != nil {
			return err
		}
	}
	// 배경 밝기는 프로그램을 띄우기 전에 한 번만 묻는다(gitlings와 같음).
	render := GlamourRenderer(lipgloss.HasDarkBackground())
	_, err := tea.NewProgram(New(exs, prog, render, grade.Grade, seed.ResetPlay), tea.WithAltScreen()).Run()
	return err
}

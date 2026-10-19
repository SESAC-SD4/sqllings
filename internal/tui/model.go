// Package tui는 왼쪽 문제 목록 · 오른쪽 설명/정답 칸/채점 결과로 된 분할 화면이다.
// 시험은 터미널 2의 psql에서 하고, 된 SQL을 정답 칸에 붙여넣어 Enter로 채점한다.
package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"sqllings/internal/exercise"
	"sqllings/internal/grade"
	"sqllings/internal/progress"
	"sqllings/internal/workspace"
)

type Renderer func(markdown string, width int) string

// Grader는 채점 함수다. 진짜는 grade.Grade, 테스트는 가짜를 넣는다.
type Grader func(*exercise.Exercise, string) grade.Result

// Resetter는 놀이터(public)를 그 장의 시작 상태로 다시 만든다. 진짜는 seed.ResetPlay.
type Resetter func(ch int) ([]string, error)

type Mode int

const (
	ModePrompt Mode = iota
	ModeHint
	ModeConfirm
)

const readyStatus = "터미널 2: psql -U sd4 -d sqllings  ·  된 SQL을 정답 칸에 붙여넣고 Enter"

type confirmKind int

const (
	confirmPlay confirmKind = iota
	confirmAll
)

type Model struct {
	exs       []*exercise.Exercise
	prog      *progress.Progress
	render    Renderer
	grader    Grader
	resetter  Resetter
	cursor    int
	mode      Mode
	confirm   confirmKind
	hintLevel int
	status    string
	result    string
	answers   map[string]string
	width     int
	height    int
	desc      viewport.Model
	out       viewport.Model
	input     textarea.Model
	// Windows 콘솔은 붙여넣기를 bracketed paste로 알려 주지 않아 줄바꿈이 Enter 키로 온다.
	// 키가 burst 간격 안에 몰려 오면 붙여넣기로 보고, 혼자 온 Enter는 enterWait 동안 조용할 때만 채점한다.
	now      func() time.Time
	lastKey  time.Time
	enterSeq int
	pending  bool
	preEnter string
	lastText bool // 바로 앞 키가 정답 칸에 글자를 넣었는가(붙여넣기 중의 Tab을 힌트와 가르는 데 쓴다)
	// 채점은 tea.Cmd로 따로 돈다(Windows에서는 몇 초 걸린다). 도는 동안에는 Enter로 다시 채점하지 않는다.
	grading bool
}

const (
	burst     = 30 * time.Millisecond
	enterWait = 120 * time.Millisecond
)

// enterTimeout은 Enter 뒤 enterWait가 지났다는 알림이다. seq가 지금 것과 다르면 지난 알림이다.
type enterTimeout struct{ seq int }

// gradeDone은 채점이 끝났다는 알림이다. 채점하는 동안 문제를 옮겼을 수 있으니 채점한 문제를 함께 싣는다.
type gradeDone struct {
	ex *exercise.Exercise
	r  grade.Result
}

const gradingStatus = "⏳ 채점 중… (몇 초 걸릴 수 있습니다)"

func (m Model) EnterSeq() int { return m.enterSeq }

func New(exs []*exercise.Exercise, prog *progress.Progress, render Renderer, g Grader, r Resetter) Model {
	ta := textarea.New()
	ta.ShowLineNumbers = false
	ta.Prompt = ""
	ta.CharLimit = 0
	ta.Placeholder = "터미널 2에서 된 SQL을 여기에 붙여넣고 Enter"
	ta.Focus()
	m := Model{exs: exs, prog: prog, render: render, grader: g, resetter: r,
		status: readyStatus, answers: map[string]string{},
		desc: viewport.New(0, 0), out: viewport.New(0, 0), input: ta, now: time.Now}
	for i, ex := range exs {
		if !prog.IsPassed(ex.ID) {
			m.cursor = i
			break
		}
	}
	m.input.SetValue(m.answerOf(m.cur()))
	m.refresh()
	return m
}

func (m Model) Cursor() int                { return m.cursor }
func (m Model) Mode() Mode                 { return m.mode }
func (m Model) Status() string             { return m.status }
func (m Model) Result() string             { return m.result }
func (m Model) Answer() string             { return m.input.Value() }
func (m Model) Prog() *progress.Progress   { return m.prog }
func (m Model) Init() tea.Cmd              { return textarea.Blink }
func (m Model) cur() *exercise.Exercise    { return m.exs[m.cursor] }

func (m *Model) answerOf(ex *exercise.Exercise) string {
	if a, ok := m.answers[ex.Slug]; ok {
		return a
	}
	return workspace.LoadAnswer(ex.Slug)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.layout()
		m.refresh()
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	case enterTimeout:
		if msg.seq == m.enterSeq && m.pending {
			m.pending = false
			m.input.SetValue(strings.TrimRight(m.preEnter, "\n"))
			if m.grading {
				return m, nil
			}
			cmd := m.check()
			m.refresh()
			return m, cmd
		}
		return m, nil
	case gradeDone:
		m.grading = false
		m.finish(msg.ex, msg.r)
		m.refresh()
		return m, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func isASCII(s string) bool {
	for _, r := range s {
		if r > 127 {
			return false
		}
	}
	return true
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := msg.String()
	if msg.Paste {
		// Windows에서 붙여넣으면 CRLF가 온다. CR을 따로 줄바꿈으로 받으면 줄 사이에 빈 줄이 생긴다.
		text := strings.ReplaceAll(strings.ReplaceAll(string(msg.Runes), "\r\n", "\n"), "\r", "\n")
		msg.Runes = []rune(strings.TrimRight(text, "\n"))
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	if k == "ctrl+c" {
		return m, tea.Quit
	}
	now := m.now()
	inBurst := !m.lastKey.IsZero() && now.Sub(m.lastKey) < burst
	prevText := m.lastText
	m.lastKey = now
	m.lastText = msg.Type == tea.KeyRunes || msg.Type == tea.KeySpace || k == "enter"
	if m.pending && k != "enter" {
		// Enter 바로 뒤에 다른 키가 왔다 — 붙여넣기 안의 줄바꿈이었다.
		m.pending = false
		m.enterSeq++
		inBurst = true
	}
	if m.mode != ModeConfirm {
		switch {
		case k == "enter" && (inBurst || m.pending):
			m.pending = false
			m.enterSeq++
			m.input.InsertString("\n")
			return m, nil
		case k == "tab" && inBurst && prevText:
			m.input.InsertString("\t")
			m.lastText = true
			return m, nil
		case k == "enter":
			m.preEnter = m.input.Value()
			m.input.InsertString("\n")
			m.pending = true
			m.enterSeq++
			seq := m.enterSeq
			return m, tea.Tick(enterWait, func(time.Time) tea.Msg { return enterTimeout{seq: seq} })
		}
	}
	// 확인 질문의 '예'는 영문 y 하나뿐이다. Enter나 다른 키는 취소, 한글은 질문을 연 채 안내만 한다.
	if m.mode == ModeConfirm {
		switch {
		case k == "y":
			m.mode = ModePrompt
			if m.confirm == confirmAll {
				m.resetAll()
			} else {
				m.resetPlay()
			}
		case !isASCII(k):
			m.status = "한글로 들어오고 있습니다. 한/영 키로 영문으로 바꾼 뒤 y를 누르세요 (취소는 Esc)"
		default:
			m.mode = ModePrompt
			m.status = "취소했습니다"
		}
		m.refresh()
		return m, nil
	}
	switch k {
	case "esc":
		if m.mode == ModeHint {
			m.mode, m.status = ModePrompt, readyStatus
			m.refresh()
		}
		return m, nil
	case "up":
		m.move(-1)
	case "down":
		m.move(1)
	case "tab":
		m.hint()
	case "ctrl+l":
		m.input.Reset()
		m.answers[m.cur().Slug] = ""
		return m, nil
	case "ctrl+r":
		m.mode, m.confirm = ModeConfirm, confirmPlay
		m.status = fmt.Sprintf("놀이터(public)를 %d장 시작 상태로 다시 만들까요? 터미널 2에서 한 작업이 지워집니다 (y/n)", m.cur().Chapter)
	case "ctrl+x":
		m.mode, m.confirm = ModeConfirm, confirmAll
		m.status = fmt.Sprintf("전체 초기화: %d문제의 ✅ 기록과 힌트를 모두 지울까요? (y/n)", len(m.exs))
	case "pgup":
		m.desc.HalfPageUp()
		return m, nil
	case "pgdown":
		m.desc.HalfPageDown()
		return m, nil
	default:
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	m.refresh()
	return m, nil
}

func (m *Model) move(d int) {
	n := m.cursor + d
	if n < 0 || n >= len(m.exs) {
		return
	}
	m.answers[m.cur().Slug] = m.input.Value()
	m.cursor, m.hintLevel, m.mode, m.status, m.result = n, 0, ModePrompt, readyStatus, ""
	m.input.SetValue(m.answerOf(m.cur()))
}

// check는 채점을 시작한다. 채점 자체는 돌려준 명령 안에서 돌고, 끝나면 gradeDone이 온다.
func (m *Model) check() tea.Cmd {
	ex := m.cur()
	sql := m.input.Value()
	m.answers[ex.Slug] = sql
	if strings.TrimSpace(sql) != "" {
		_ = workspace.SaveAnswer(ex.Slug, sql)
	}
	m.grading = true
	m.status, m.result = gradingStatus, gradingStatus
	g := m.grader
	return func() tea.Msg { return gradeDone{ex: ex, r: g(ex, sql)} }
}

// finish는 채점 결과를 기록하고 보여 준다. 그 사이 다른 문제로 옮겼으면 기록만 한다.
func (m *Model) finish(ex *exercise.Exercise, r grade.Result) {
	here := ex == m.cur()
	if !here && r.Passed {
		m.prog.MarkPassed(ex.ID)
		_ = m.prog.Save()
	}
	if !here {
		return
	}
	mark := "❌ " + r.Message
	if r.Passed {
		mark = "✅ " + r.Message
		m.prog.MarkPassed(ex.ID)
		if err := m.prog.Save(); err != nil {
			mark += "\n(진행 상황을 저장하지 못했습니다: " + err.Error() + ")"
		}
		m.status = "통과! ↓로 다음 문제"
	} else {
		m.status = "아직입니다. Tab으로 힌트를 볼 수 있습니다"
	}
	m.result = mark
	if r.Display != "" {
		m.result += "\n\n" + r.Display
	}
}

func (m *Model) hint() {
	ex := m.cur()
	if m.hintLevel < len(ex.Hints) {
		m.hintLevel++
	}
	m.mode = ModeHint
	m.status = fmt.Sprintf("힌트 %d/%d — Tab을 다시 누르면 다음 단계, Esc로 닫기", m.hintLevel, len(ex.Hints))
}

func (m *Model) resetPlay() {
	ch := m.cur().Chapter
	others, err := m.resetter(ch)
	if err != nil {
		m.status = "❌ " + err.Error()
		return
	}
	m.prog.PlayChapter = ch
	if err := m.prog.Save(); err != nil {
		m.status = "❌ 진행 상황을 저장하지 못했습니다: " + err.Error()
		return
	}
	m.status = fmt.Sprintf("✅ 놀이터를 %d장 데이터로 만들었습니다. 터미널 2에서 다시 SELECT 해 보세요", ch)
	if len(others) > 0 {
		m.status += " · 다른 테이블은 그대로 두었습니다: " + strings.Join(others, ", ")
	}
}

func (m *Model) resetAll() {
	m.prog.Passed, m.prog.HintLevel = nil, nil
	m.cursor, m.hintLevel, m.result = 0, 0, ""
	m.input.SetValue(m.answerOf(m.cur()))
	if err := m.prog.Save(); err != nil {
		m.status = "❌ 진행 상황을 저장하지 못했습니다: " + err.Error()
		return
	}
	m.status = fmt.Sprintf("✅ %d문제의 기록을 모두 지웠습니다", len(m.exs))
}

func (m Model) Content() string {
	ex := m.cur()
	var b strings.Builder
	if m.prog.PlayChapter != 0 && m.prog.PlayChapter != ex.Chapter {
		fmt.Fprintf(&b, "> ⚠️ 터미널 2의 놀이터는 지금 %d장 데이터입니다. 이 문제는 %d장 — Ctrl+R로 바꾸세요 (채점은 늘 %d장 데이터로 합니다)\n\n",
			m.prog.PlayChapter, ex.Chapter, ex.Chapter)
	}
	b.WriteString("# " + ex.Title + "\n\n")
	if m.mode == ModeHint {
		for i := 0; i < m.hintLevel; i++ {
			fmt.Fprintf(&b, "**힌트 %d/%d**\n\n%s\n\n", i+1, len(ex.Hints), ex.Hints[i])
		}
	} else {
		b.WriteString(strings.TrimSpace(ex.Prompt) + "\n")
	}
	return b.String()
}

func (m *Model) refresh() {
	m.desc.SetContent(m.render(m.Content(), m.desc.Width))
	m.desc.GotoTop()
	m.out.SetContent(m.result)
	m.out.GotoTop()
}

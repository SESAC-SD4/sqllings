package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"sqllings/exercises"
	"sqllings/internal/exercise"
	"sqllings/internal/grade"
	"sqllings/internal/progress"
)

func plain(md string, width int) string { return md }

func noReset(int) ([]string, error) { return nil, nil }

func newModelFull(t *testing.T, g Grader, r Resetter) Model {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	m := New(exercises.All(), &progress.Progress{PlayChapter: 1}, plain, g, r)
	tm, _ := m.Update(tea.WindowSizeMsg{Width: 140, Height: 45})
	return tm.(Model)
}

func newModelWith(t *testing.T, g Grader) Model { return newModelFull(t, g, noReset) }

func fakeGrade(pass bool) Grader {
	return func(ex *exercise.Exercise, s string) grade.Result {
		if pass {
			return grade.Result{Passed: true, Message: "정답입니다!", Display: " count \n 8807"}
		}
		return grade.Result{Message: "값이 다릅니다", Display: " count \n 0"}
	}
}

func key(k string) tea.KeyMsg {
	switch k {
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	case "ctrl+r":
		return tea.KeyMsg{Type: tea.KeyCtrlR}
	case "ctrl+x":
		return tea.KeyMsg{Type: tea.KeyCtrlX}
	case "ctrl+l":
		return tea.KeyMsg{Type: tea.KeyCtrlL}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "pgdown":
		return tea.KeyMsg{Type: tea.KeyPgDown}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
}

func update(m Model, msg tea.Msg) Model { tm, _ := m.Update(msg); return tm.(Model) }

func press(m Model, keys ...string) Model {
	for _, k := range keys {
		m = update(m, key(k))
	}
	return m
}

// deliver는 msg를 보내고, 돌아온 명령(채점)을 실행해 그 결과까지 보낸다.
func deliver(m Model, msg tea.Msg) Model {
	tm, cmd := m.Update(msg)
	m = tm.(Model)
	if cmd != nil {
		m = update(m, cmd())
	}
	return m
}

// submit은 Enter를 누르고 붙여넣기 판정 시간이 지난 것까지 보낸 뒤 채점 결과를 받는다.
func submit(m Model) Model {
	m = press(m, "enter")
	return deliver(m, enterTimeout{seq: m.EnterSeq()})
}

func paste(s string) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s), Paste: true} }

func TestPasteWithNewlineDoesNotGrade(t *testing.T) {
	called := 0
	m := newModelWith(t, func(*exercise.Exercise, string) grade.Result { called++; return grade.Result{} })
	m = update(m, paste("SELECT count(*)\nFROM titles;\n"))
	if called != 0 {
		t.Fatal("붙여넣기 안의 줄바꿈이 채점을 불렀다")
	}
	if !strings.Contains(m.Answer(), "FROM titles;") || !strings.Contains(m.Answer(), "\n") {
		t.Fatalf("%q", m.Answer())
	}
	m = submit(m)
	if called != 1 {
		t.Fatal("Enter가 채점을 안 불렀다")
	}
}

func TestPassMarksAndStays(t *testing.T) {
	m := newModelWith(t, fakeGrade(true))
	m = update(m, paste("SELECT 1;"))
	m = submit(m)
	if !m.Prog().IsPassed(1) || m.Cursor() != 0 {
		t.Fatalf("passed=%v cursor=%d", m.Prog().IsPassed(1), m.Cursor())
	}
	if !strings.Contains(m.Result(), "8807") || !strings.Contains(m.Result(), "✅") {
		t.Fatal(m.Result())
	}
}

func TestFailShowsMessage(t *testing.T) {
	m := newModelWith(t, fakeGrade(false))
	m = submit(update(m, paste("SELECT 1;")))
	if m.Prog().IsPassed(1) || !strings.Contains(m.Result(), "❌") || !strings.Contains(m.Result(), "값이 다릅니다") {
		t.Fatal(m.Result())
	}
}

func TestAnswerSavedPerExercise(t *testing.T) {
	m := newModelWith(t, fakeGrade(false))
	m = update(m, paste("SELECT 1;"))
	m = press(m, "down")
	if m.Answer() != "" {
		t.Fatalf("다음 문제 칸이 비어 있지 않다: %q", m.Answer())
	}
	m = press(m, "up")
	if m.Answer() != "SELECT 1;" {
		t.Fatalf("되돌아왔는데 %q", m.Answer())
	}
}

func TestHangulTyping(t *testing.T) {
	m := newModelWith(t, fakeGrade(false))
	m = press(m, "'", "알", " ", "수", "'")
	if m.Answer() != "'알 수'" {
		t.Fatalf("%q", m.Answer())
	}
}

func TestClearAnswer(t *testing.T) {
	m := press(update(newModelWith(t, fakeGrade(false)), paste("SELECT 1;")), "ctrl+l")
	if m.Answer() != "" {
		t.Fatal(m.Answer())
	}
}

func TestResetPlayAsksY(t *testing.T) {
	resets := 0
	m := newModelFull(t, fakeGrade(false), func(ch int) ([]string, error) { resets++; return []string{"my_notes"}, nil })
	m = press(m, "ctrl+r", "enter")
	if resets != 0 {
		t.Fatal("Enter로 실행됐다")
	}
	m = press(m, "ctrl+r")
	if !strings.Contains(m.Status(), "(y/n)") {
		t.Fatal(m.Status())
	}
	m = press(m, "y")
	if resets != 1 || !strings.Contains(m.Status(), "my_notes") {
		t.Fatal(m.Status())
	}
}

func TestResetPlayUsesExerciseChapter(t *testing.T) {
	if all := exercises.All(); all[len(all)-1].Chapter < 2 {
		t.Skip("1장만 담은 배포본(publish.sh 1)에는 2장이 없다")
	}
	got := 0
	m := newModelFull(t, fakeGrade(false), func(ch int) ([]string, error) { got = ch; return nil, nil })
	for i := 0; i < 14; i++ {
		m = press(m, "down")
	}
	if !strings.Contains(m.Content(), "1장 데이터") {
		t.Fatalf("장이 다르다는 안내가 없다:\n%s", m.Content())
	}
	m = press(m, "ctrl+r", "y")
	if got != 2 || m.Prog().PlayChapter != 2 {
		t.Fatalf("ch=%d play=%d", got, m.Prog().PlayChapter)
	}
	if strings.Contains(m.Content(), "1장 데이터입니다") {
		t.Fatal("바꾼 뒤에도 안내가 남았다")
	}
}

func TestHintAndEsc(t *testing.T) {
	m := press(newModelWith(t, fakeGrade(false)), "tab")
	if !strings.Contains(m.Content(), "힌트 1/3") {
		t.Fatal(m.Content())
	}
	m = press(m, "tab", "tab", "tab")
	if !strings.Contains(m.Content(), "힌트 3/3") {
		t.Fatal(m.Content())
	}
	m = press(m, "esc")
	if strings.Contains(m.Content(), "힌트 1/3") {
		t.Fatal("Esc로 닫히지 않았다")
	}
}

func TestGraderErrorDoesNotCrash(t *testing.T) {
	m := newModelWith(t, func(*exercise.Exercise, string) grade.Result {
		return grade.Result{Message: "채점기를 실행하지 못했습니다: PostgreSQL 서버가 꺼져 있습니다"}
	})
	m = submit(update(m, paste("SELECT 1;")))
	if !strings.Contains(m.Result(), "서버가 꺼져") {
		t.Fatal(m.Result())
	}
}

func TestQuit(t *testing.T) {
	_, cmd := newModelWith(t, fakeGrade(false)).Update(key("ctrl+c"))
	if cmd == nil {
		t.Fatal("Ctrl+C가 끝내지 않는다")
	}
}

func TestViewRenders(t *testing.T) {
	v := newModelWith(t, fakeGrade(false)).View()
	for _, want := range []string{"01 전체 작품 수", "psql -U sd4 -d sqllings", "Enter 채점"} {
		if !strings.Contains(v, want) {
			t.Errorf("화면에 %q가 없다", want)
		}
	}
}

func TestPasteCRLFHasNoBlankLines(t *testing.T) {
	m := update(newModelWith(t, fakeGrade(false)), paste("SELECT count(*)\r\nFROM titles;\r\n"))
	if strings.Contains(m.Answer(), "\n\n") || strings.Contains(m.Answer(), "\r") {
		t.Fatalf("%q", m.Answer())
	}
}

// Windows 콘솔은 붙여넣기를 bracketed paste로 알려 주지 않아 줄바꿈이 Enter 키로 온다(리뷰 C1).
func TestUnbracketedPasteDoesNotGradePerLine(t *testing.T) {
	called := 0
	m := newModelWith(t, func(*exercise.Exercise, string) grade.Result { called++; return grade.Result{} })
	m = press(m, "S", "E", "L", "E", "C", "T", " ", "1", "enter", "tab", "F", "R", "O", "M", " ", "t", "enter")
	if called != 0 {
		t.Fatalf("붙여넣기 중에 %d번 채점했다", called)
	}
	if m.Mode() == ModeHint {
		t.Fatal("붙여넣은 Tab이 힌트를 열었다")
	}
	if !strings.Contains(m.Answer(), "SELECT 1\n") || !strings.Contains(m.Answer(), "FROM t") {
		t.Fatalf("%q", m.Answer())
	}
	time.Sleep(2 * burst) // 붙여넣기가 끝나고 학생이 Enter를 누른다
	m = submit(m)
	if called != 1 {
		t.Fatalf("조용해진 뒤 채점이 한 번이어야 한다: %d", called)
	}
	if strings.HasSuffix(m.Answer(), "\n") {
		t.Fatalf("끝 줄바꿈이 남았다: %q", m.Answer())
	}
}

func TestEnterAloneGradesAfterTimeout(t *testing.T) {
	called := 0
	m := newModelWith(t, func(*exercise.Exercise, string) grade.Result { called++; return grade.Result{} })
	m = update(m, paste("SELECT 1;"))
	m = press(m, "enter")
	stale := m.EnterSeq() - 1
	m = deliver(m, enterTimeout{seq: stale})
	if called != 0 {
		t.Fatal("지난 타이머가 채점했다")
	}
	m = deliver(m, enterTimeout{seq: m.EnterSeq()})
	if called != 1 || m.Answer() != "SELECT 1;" {
		t.Fatalf("called=%d answer=%q", called, m.Answer())
	}
}

// 긴 결과 줄(DETAIL의 null 나열)이 화면 높이를 넘겨 틀이 밀리면 안 된다(2026-10-18 Windows 실기).
func TestLongResultLineKeepsScreenHeight(t *testing.T) {
	long := "상세정보:  실패한 자료: (s9002, Series, Test" + strings.Repeat(", null", 60) + ")."
	g := func(ex *exercise.Exercise, s string) grade.Result {
		return grade.Result{Passed: true, Message: "정답입니다!", Display: "오류:  위반\n" + long}
	}
	m := submit(update(newModelWith(t, g), paste("SELECT 1;")))
	if n := strings.Count(m.View(), "\n") + 1; n > 45 {
		t.Fatalf("화면 높이 %d줄 > 45", n)
	}
	if !strings.Contains(m.Result(), "null)") {
		t.Fatal("긴 줄의 끝이 결과에서 사라졌다")
	}
}

// 채점은 화면을 멈추지 않는다: Windows에서는 몇 초 걸린다. 그동안 "채점 중"을 보이고 Enter는 다시 채점하지 않는다.
func TestGradingIsAsyncAndIgnoresEnter(t *testing.T) {
	called := 0
	m := newModelWith(t, func(ex *exercise.Exercise, s string) grade.Result {
		called++
		return grade.Result{Passed: true, Message: "정답입니다!", Display: " count \n 8807"}
	})
	m = press(update(m, paste("SELECT 1;")), "enter")
	tm, cmd := m.Update(enterTimeout{seq: m.EnterSeq()})
	m = tm.(Model)
	if called != 0 || cmd == nil {
		t.Fatalf("채점이 Update 안에서 돌았다 (called=%d)", called)
	}
	if !strings.Contains(m.Status(), "채점 중") || m.Answer() != "SELECT 1;" {
		t.Fatalf("status=%q answer=%q", m.Status(), m.Answer())
	}
	// 채점 중에 Enter를 다시 눌러도 두 번째 채점은 없다
	time.Sleep(2 * burst) // 붙여넣기의 줄바꿈이 아니라 혼자 누른 Enter
	m = press(m, "enter")
	tm, again := m.Update(enterTimeout{seq: m.EnterSeq()})
	m = tm.(Model)
	if again != nil || m.Answer() != "SELECT 1;" {
		t.Fatalf("채점 중에 다시 채점하려 했다 (answer=%q)", m.Answer())
	}
	m = update(m, cmd())
	if called != 1 || !m.Prog().IsPassed(1) || !strings.Contains(m.Result(), "✅") || strings.Contains(m.Status(), "채점 중") {
		t.Fatalf("called=%d status=%q result=%q", called, m.Status(), m.Result())
	}
	// 끝난 뒤에는 다시 채점된다
	time.Sleep(2 * burst)
	m = submit(m)
	if called != 2 {
		t.Fatalf("채점이 끝난 뒤 Enter가 채점하지 않았다: %d", called)
	}
}

// 채점 중에 다른 문제로 옮겨 가도 결과는 채점한 문제에만 붙는다.
func TestGradeResultAfterMoveGoesToGradedExercise(t *testing.T) {
	m := newModelWith(t, fakeGrade(true))
	m = press(update(m, paste("SELECT 1;")), "enter")
	tm, cmd := m.Update(enterTimeout{seq: m.EnterSeq()})
	m = press(tm.(Model), "down")
	m = update(m, cmd())
	if !m.Prog().IsPassed(1) || m.Prog().IsPassed(2) || m.Cursor() != 1 {
		t.Fatalf("passed1=%v passed2=%v cursor=%d", m.Prog().IsPassed(1), m.Prog().IsPassed(2), m.Cursor())
	}
	if strings.Contains(m.Result(), "8807") || strings.Contains(m.Status(), "채점 중") {
		t.Fatalf("옮겨 간 문제에 앞 문제 결과가 보인다: status=%q result=%q", m.Status(), m.Result())
	}
}

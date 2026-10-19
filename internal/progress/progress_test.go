package progress

import "testing"

func TestPlayChapterSaved(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	p, _ := Load()
	p.PlayChapter = 2
	p.MarkPassed(3)
	if err := p.Save(); err != nil {
		t.Fatal(err)
	}
	q, _ := Load()
	if q.PlayChapter != 2 || !q.IsPassed(3) {
		t.Fatalf("%+v", q)
	}
}

func TestHintCapped(t *testing.T) {
	p := &Progress{}
	for i := 0; i < 5; i++ {
		p.NextHint(1, 3)
	}
	if p.HintLevel[1] != 3 {
		t.Fatal(p.HintLevel)
	}
}

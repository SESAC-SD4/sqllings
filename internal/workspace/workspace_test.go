package workspace

import (
	"path/filepath"
	"testing"
)

func TestAnswerRoundTrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if got := LoadAnswer("01-count-all"); got != "" {
		t.Fatalf("없을 때 빈 문자열이어야 한다, got %q", got)
	}
	if err := SaveAnswer("01-count-all", "SELECT 1;"); err != nil {
		t.Fatal(err)
	}
	if got := LoadAnswer("01-count-all"); got != "SELECT 1;" {
		t.Fatalf("got %q", got)
	}
	if filepath.Base(Base()) != "sqllings-work" {
		t.Fatal(Base())
	}
}

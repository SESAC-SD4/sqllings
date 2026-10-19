package exercise

import "testing"

func TestParseHints(t *testing.T) {
	h := ParseHints("하나\r\n---\n둘\n---  \n셋\n")
	if len(h) != 3 || h[0] != "하나" || h[2] != "셋" {
		t.Fatalf("%q", h)
	}
}

package progress

import (
	"encoding/json"
	"os"
	"path/filepath"

	"sqllings/internal/workspace"
)

type Progress struct {
	Passed      []int       `json:"passed"`
	LastStarted int         `json:"last_started"`
	HintLevel   map[int]int `json:"hint_level,omitempty"`
	PlayChapter int         `json:"play_chapter"` // 터미널 2의 표(public)가 지금 몇 장 상태인가. 0 = 모름
}

func file() string { return filepath.Join(workspace.Base(), ".progress.json") }

func Load() (*Progress, error) {
	data, err := os.ReadFile(file())
	if os.IsNotExist(err) {
		return &Progress{}, nil
	}
	if err != nil {
		return nil, err
	}
	var p Progress
	if err := json.Unmarshal(data, &p); err != nil {
		return &Progress{}, nil
	}
	return &p, nil
}

func (p *Progress) Save() error {
	if err := os.MkdirAll(workspace.Base(), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(file(), data, 0o644)
}

func (p *Progress) MarkPassed(id int) {
	if !p.IsPassed(id) {
		p.Passed = append(p.Passed, id)
	}
}

func (p *Progress) Unpass(id int) {
	kept := p.Passed[:0]
	for _, v := range p.Passed {
		if v != id {
			kept = append(kept, v)
		}
	}
	p.Passed = kept
}

func (p *Progress) IsPassed(id int) bool {
	for _, v := range p.Passed {
		if v == id {
			return true
		}
	}
	return false
}

func (p *Progress) NextHint(id, max int) int {
	if p.HintLevel == nil {
		p.HintLevel = map[int]int{}
	}
	l := p.HintLevel[id] + 1
	if l > max {
		l = max
	}
	p.HintLevel[id] = l
	return l
}

func (p *Progress) ResetHint(id int) { delete(p.HintLevel, id) }

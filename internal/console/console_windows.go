//go:build windows

package console

import (
	"os/exec"
)

func initPlatform() {
	// PowerShell 5.1에서 한글이 깨지지 않도록 UTF-8 코드페이지로 설정
	exec.Command("cmd", "/c", "chcp", "65001").Run()
}

package utils

import (
	"fmt"
	"os"
	"os/exec"
)

func TerminateProcess(pid int) error {
	var cmd *exec.Cmd
	if os := os.Getenv("OS"); os == "windows" {
		// Windows 使用 taskkill
		cmd = exec.Command("taskkill", "/F", "/PID", fmt.Sprintf("%d", pid))
	} else {
		// Unix-like 系统使用 kill -9
		cmd = exec.Command("kill", "-9", fmt.Sprintf("%d", pid))
	}
	return cmd.Run()
}

// RestartServiceProcess 重启 ollama 服务
func RestartServiceProcess() error {
	cmd := exec.Command("systemctl", "restart", "ollama")
	return cmd.Run()
}

func RebootSystem() error {
	cmd := exec.Command("reboot")
	return cmd.Run()
}

func TerminateOllamaProcess(modelName string) error {
	cmd := exec.Command("ollama", "stop", modelName)
	return cmd.Run()
}

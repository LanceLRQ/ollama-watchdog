package utils

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/LanceLRQ/ollama-watchdog/configs"
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
func RestartServiceProcess(typeName string, serviceName string) error {
	if typeName == "" {
		typeName = "restart"
	}
	if serviceName == "" {
		serviceName = "ollama"
	}
	fmt.Printf("执行shell指令> %s %s %s", "systemctl", typeName, serviceName)
	cmd := exec.Command("systemctl", typeName, serviceName)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("重启Ollama服务失败：%s", err.Error())
	}
	return nil
}

func RebootSystem() error {
	cmd := exec.Command("reboot")
	return cmd.Run()
}

func TerminateOllamaProcess(cfg *configs.ServerConfigStruct, modelName string, server string) error {
	cmd := exec.Command("ollama", "stop", modelName)
	// 1. 获取当前环境变量（可选）
	env := os.Environ()
	// 2. 添加自定义环境变量（例如 OLLAMA_HOST）
	if server == "" {
		env = append(env, fmt.Sprintf("OLLAMA_HOST=%s", cfg.OllamaListen))
	} else {
		env = append(env, fmt.Sprintf("OLLAMA_HOST=%s", server))
	}
	// 3. 设置 cmd 的环境变量
	cmd.Env = env
	// 4. 运行命令
	return cmd.Run()
}

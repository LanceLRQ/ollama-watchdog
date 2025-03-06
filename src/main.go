package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// GPUInfo结构体
type GPUInfo struct {
	Name        string    `json:"name"`
	MemoryTotal uint64    `json:"mem_total"`
	MemoryUsed  uint64    `json:"mem_used"`
	MemoryUtil  float64   `json:"mem_util"`
	Temperature uint64    `json:"temp"`
	GPUUtil     float64   `json:"gpu_util"`
	PowerUsage  float64   `json:"power_usage"` // 新增功耗
	Processes   []Process `json:"processes"`   // 新增进程信息
	Timestamp   int64     `json:"timestamp"`
}

type Process struct {
	PID        int    `json:"pid"`
	Name       string `json:"name"`
	MemoryUsed uint64 `json:"mem_used"`
}

func main() {
	app := fiber.New()

	// WebSocket端点
	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			gpuInfo, err := getGPUInfo()
			if err != nil {
				fmt.Println("Error getting GPU info:", err)
				continue
			}

			jsonData, err := json.Marshal(gpuInfo)
			if err != nil {
				fmt.Println("JSON marshal error:", err)
				continue
			}

			if err := c.WriteMessage(websocket.TextMessage, jsonData); err != nil {
				fmt.Println("Write error:", err)
				break
			}
		}
	}))

	// 静态文件服务
	app.Static("/", "./public")

	app.Listen(":23333")
}

func getGPUInfo() ([]GPUInfo, error) {
	cmd := exec.Command(
		"nvidia-smi",
		"--query-gpu=name,memory.total,memory.used,utilization.memory,temperature.gpu,utilization.gpu,power.draw",
		"--query-compute-apps=pid,process_name,used_memory",
		"--format=csv,noheader,nounits",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("command failed: %v\nOutput: %s", err, output)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var gpuInfos []GPUInfo

	for _, line := range lines {
		fields := strings.Split(line, ", ")
		if len(fields) != 7 {
			return nil, fmt.Errorf("invalid output format")
		}

		info := GPUInfo{
			Name:        strings.TrimSpace(fields[0]),
			MemoryTotal: parseUint(fields[1]),
			MemoryUsed:  parseUint(fields[2]),
			MemoryUtil:  parseFloat(fields[3]),
			Temperature: parseUint(fields[4]),
			GPUUtil:     parseFloat(fields[5]),
			PowerUsage:  parseFloat(fields[6]),
			Timestamp:   time.Now().Unix(),
		}

		gpuInfos = append(gpuInfos, info)
	}

	return gpuInfos, nil
}

// 辅助函数：字符串转uint64
func parseUint(s string) uint64 {
	var n uint64
	fmt.Sscanf(strings.TrimSpace(s), "%d", &n)
	return n
}

// 辅助函数：字符串转float64
func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(strings.TrimSpace(s), "%f", &f)
	return f
}

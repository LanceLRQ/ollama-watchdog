package services

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/LanceLRQ/ollama-watchdog/models"
	"github.com/LanceLRQ/ollama-watchdog/utils"
)

func NvidiaSMIWatcher(callback func(models.NvidiaSMIResponse)) {
	ticker := time.NewTicker(time.Duration(0.5 * float64(time.Second)))
	defer ticker.Stop()

	for range ticker.C {
		gpuInfo, err := getGPUInfo()
		if err != nil {
			fmt.Println("Error getting GPU info:", err)
			continue
		}

		gpuProcessesInfo, err := getGPUProcesses()
		if err != nil {
			fmt.Println("Error getting GPU processes info:", err)
			continue
		}
		callback(models.NvidiaSMIResponse{
			GPUInfo:      gpuInfo,
			GPUProcesses: gpuProcessesInfo,
			Timestamp:    time.Now().Unix(),
		})
	}
}

func getGPUInfo() ([]models.GPUInfo, error) {
	cmd := exec.Command(
		"nvidia-smi",
		"--query-gpu=pci.device_id,pci.bus_id,name,memory.total,memory.used,utilization.gpu,temperature.gpu,power.draw,power.limit",
		"--format=csv,noheader,nounits",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("command failed: %v\nOutput: %s", err, output)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var gpuInfos []models.GPUInfo

	for _, line := range lines {
		fields := strings.Split(line, ", ")
		if len(fields) != 9 {
			fmt.Println(line)
			return nil, fmt.Errorf("invalid output format")
		}

		info := models.GPUInfo{
			DeviceId:    strings.TrimSpace(fields[0]),
			BusId:       strings.TrimSpace(fields[1]),
			Name:        strings.TrimSpace(fields[2]),
			MemoryTotal: utils.ParseUint(fields[3]),
			MemoryUsed:  utils.ParseUint(fields[4]),
			GPUUsed:     utils.ParseUint(fields[5]),
			Temperature: utils.ParseUint(fields[6]),
			PowerUsage:  utils.ParseFloat(fields[7]),
			PowerLimit:  utils.ParseFloat(fields[8]),
		}

		gpuInfos = append(gpuInfos, info)
	}

	return gpuInfos, nil
}
func getGPUProcesses() ([]models.GPUProcess, error) {
	cmd := exec.Command(
		"nvidia-smi",
		"--query-compute-apps=gpu_bus_id,pid,process_name,used_memory",
		"--format=csv,noheader,nounits",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("command failed: %v\nOutput: %s", err, output)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var gpuProcesses []models.GPUProcess

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Split(line, ", ")
		if len(fields) != 4 {
			fmt.Println(line)
			return nil, fmt.Errorf("invalid output format")
		}

		// 解析每一行数据并创建GPUProcess对象
		info := models.GPUProcess{
			BusId:      fields[0],
			PID:        utils.ParseUint(fields[1]),
			Name:       fields[2],
			MemoryUsed: utils.ParseUint(fields[3]),
		}

		gpuProcesses = append(gpuProcesses, info)
	}

	return gpuProcesses, nil
}

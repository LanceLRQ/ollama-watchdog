package models

// GPUInfo结构体
type GPUInfo struct {
	DeviceId    string  `json:"device_id"`   // PCI设备ID，用于唯一标识GPU设备
	BusId       string  `json:"bus_id"`      // PCI总线ID，用于标识GPU在PCI总线上的位置
	Name        string  `json:"name"`        // GPU的名称或型号
	MemoryTotal uint64  `json:"mem_total"`   // GPU的总内存大小，以字节为单位
	MemoryUsed  uint64  `json:"mem_used"`    // GPU当前使用的内存大小，以字节为单位
	GPUUsed     uint64  `json:"gpu_used"`    // GPU的利用率，表示当前GPU的使用百分比
	Temperature uint64  `json:"temperature"` // GPU当前的温度
	PowerUsage  float64 `json:"power_usage"` // GPU当前的当前功耗
	PowerLimit  float64 `json:"power_limit"`   // GPU当前的功耗限制
}

type GPUProcess struct {
	BusId      string `json:"bus_id"` // PCI总线ID，用于标识GPU在PCI总线上的位置
	PID        uint64 `json:"pid"`
	Name       string `json:"name"`
	MemoryUsed uint64 `json:"mem_used"`
}

type NvidiaSMIResponse struct {
	GPUInfo      []GPUInfo    `json:"gpu_info"`
	GPUProcesses []GPUProcess `json:"gpu_processes"`
	Timestamp    int64        `json:"timestamp"`
}

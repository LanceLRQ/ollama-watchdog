package utils

import (
	"fmt"
	"strings"
)

// 辅助函数：字符串转uint64
func ParseUint(s string) uint64 {
	var n uint64
	fmt.Sscanf(strings.TrimSpace(s), "%d", &n)
	return n
}

// 辅助函数：字符串转float64
func ParseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(strings.TrimSpace(s), "%f", &f)
	return f
}

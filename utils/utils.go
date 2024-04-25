package utils

import (
	"strconv"
	"strings"
)

// 判断是否满足 ip:port 的格式
func ValidPeerAddr(addr string) bool {
	parts1 := strings.Split(addr, ":")
	// 是否 : 分割两段
	if len(parts1) != 2 {
		return false
	}
	// 是否 端口前缀 0
	if parts1[1][0] == '0' && len(parts1[1]) > 1 {
		return false
	}
	if parts1[0] == "localhost" {
		return true
	}
	if port, _ := strconv.Atoi(parts1[1]); port > 65535 {
		return false
	}
	// 是否 4段ip地址
	parts2 := strings.Split(parts1[0], ".")
	if len(parts2) != 4 {
		return false
	}
	v, _ := strconv.Atoi(parts2[0])
	if v > 255 || v < 0 {
		return false
	}
	// 是否 2 3 4 段 前缀0
	for i := 1; i < len(parts2); i++ {
		parts3 := parts2[i]
		if parts3[0] == '0' && len(parts3) > 1 {
			v, _ := strconv.Atoi(parts3)
			if v > 255 || v < 0 {
				return false
			}
		}
	}
	return true
}

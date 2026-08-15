package utils

import (
	gopsutil_net "github.com/shirou/gopsutil/v3/net"
	gopsutil_process "github.com/shirou/gopsutil/v3/process"
)

// GetPidByTCPPort 根据端口号获取进程ID
func GetPidByTCPPort(port int) (ok bool, pid int, name string) {

	connections, err := gopsutil_net.Connections("tcp")
	if err != nil {
		return false, 0, ""
	}
	for _, conn := range connections {
		if conn.Laddr.Port == uint32(port) && conn.Status == "LISTEN" {
			pid = int(conn.Pid)
			if pid == 0 {
				continue // 有些连接可能拿不到 PID
			}
			proc, err := gopsutil_process.NewProcess(int32(pid))
			if err != nil {
				name = "unknown"
				return true, pid, name
			}
			name, err = proc.Name()
			if err != nil {
				name = "unknown"
			}
			return true, pid, name
		}
	}
	return false, 0, ""
}

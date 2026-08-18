package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

// OsInfo 系统信息响应
// @ref database.User
type OsInfo struct {
	Hostname   string `json:"hostname"`     // 主机名
	OsName     string `json:"os_name"`      // 操作系统名称
	Platform   string `json:"platform"`     // 系统平台
	Arch       string `json:"arch"`         // 系统架构
	StartTime  string `json:"start_time"`   // 启动时间
	TotalMemMB uint64 `json:"total_mem_mb"` // 总内存
	CPUCount   int    `json:"cpu_count"`    // CPU 核心数
	CPUModel   string `json:"cpu_model"`    // CPU 型号
}

type OsState struct {
	Uptime     string `json:"uptime"`      // 运行时长
	FreeMemMB  uint64 `json:"free_mem_mb"` // 空闲内存
	CPUPercent uint8  `json:"cpu_percent"` // CPU 使用率
	CPUFreq    string `json:"cpu_freq"`    // CPU 频率
}

// http_os_info_api_handler 系统信息接口处理
// @Summary      获取系统信息
// @Description  获取服务器的启动时间、总内存和空闲内存等信息。
// @Tags         系统信息
// @Accept       json
// @Produce      json
// @Success      200  {object}  OsInfo
// @Router       /os/info [get]
// @Security     Bearer
func http_os_info_api_handler(s *Server, session *Session, pattern string, w http.ResponseWriter, r *http.Request) {

	v, _ := mem.VirtualMemory()
	info, _ := host.Info()
	bootTime, _ := host.BootTime()

	updateTime := time.Unix(int64(bootTime), 0)

	cpuCount, _ := cpu.Counts(true)
	cpuInfo, _ := cpu.Info()

	s.WriteJsonSuccessResponse(w, OsInfo{
		Hostname:   info.Hostname,
		OsName:     info.OS,
		Platform:   info.Platform,
		Arch:       info.KernelArch,
		StartTime:  updateTime.Format(time.RFC3339),
		TotalMemMB: v.Total / 1024 / 1024,
		CPUCount:   cpuCount,
		CPUModel:   cpuInfo[0].ModelName,
	})
}

// http_os_state_api_handler 系统状态接口处理
// @Summary      获取系统状态
// @Description  获取服务器的运行时长和空闲内存等信息。
// @Tags         系统信息
// @Accept       json
// @Produce      json
// @Success      200  {object}  OsState
// @Router       /os/state [get]
// @Security     Bearer
func http_os_state_api_handler(s *Server, session *Session, pattern string, w http.ResponseWriter, r *http.Request) {

	v, _ := mem.VirtualMemory()
	bootTime, _ := host.BootTime()
	cpuPercent, _ := cpu.Percent(0, false) // 获取整体CPU使用率
	cpuInfo, _ := cpu.Info()

	var (
		updateTime = time.Unix(int64(bootTime), 0)
		since      = time.Since(updateTime)
	)

	s.WriteJsonSuccessResponse(w, OsState{
		Uptime:     formatDuration(since),
		FreeMemMB:  v.Free / 1024 / 1024,
		CPUPercent: uint8(cpuPercent[0]),
		CPUFreq:    fmt.Sprintf("%.2f MHz", cpuInfo[0].Mhz),
	})
}

package server

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/ma6254/bookcocoon-server/build"
)

// SysInfo 系统信息响应
// @ref database.User
type SysInfo struct {
	StartTime    string `json:"start_time"`    // 启动时间
	BuildTime    string `json:"build_time"`    // 编译时间
	GoVersion    string `json:"go_version"`    // Go 版本
	BuildVersion string `json:"build_version"` // 构建版本
}

type SysState struct {
	Uptime       string `json:"uptime"`         // 运行时长
	NumGoroutine int    `json:"num_goroutine"`  // 当前协程数
	HeapUsedMB   int    `json:"heap_used_mb"`   // 堆内存使用量（MB）
	GcTotalCount int    `json:"gc_total_count"` // 垃圾回收总次数
}

// http_sys_info_api_handler 系统信息接口处理
// @Summary      获取系统信息
// @Description  获取服务器的启动时间和编译时间等信息。
// @Tags         系统信息
// @Accept       json
// @Produce      json
// @Success      200  {object}  SysInfo
// @Router       /sys/info [get]
// @Security     Bearer
func http_sys_info_api_handler(s *Server, session *Session, pattern string, w http.ResponseWriter, r *http.Request) {

	s.WriteJsonSuccessResponse(w, SysInfo{
		StartTime:    s.start_time.Format(time.RFC3339),
		BuildTime:    build.BuildTime,
		GoVersion:    runtime.Version(),
		BuildVersion: build.BuildVersion,
	})

}

// http_sys_state_api_handler 系统状态接口处理
// @Summary      获取系统状态
// @Description  获取服务器的当前会话数。
// @Tags         系统信息
// @Accept       json
// @Produce      json
// @Success      200  {object}  SysState
// @Router       /sys/state [get]
// @Security     Bearer
func http_sys_state_api_handler(s *Server, session *Session, pattern string, w http.ResponseWriter, r *http.Request) {

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	s.WriteJsonSuccessResponse(w, SysState{
		Uptime:       formatDuration(time.Since(s.start_time)),
		NumGoroutine: runtime.NumGoroutine(),
		HeapUsedMB:   int(memStats.HeapAlloc / 1024 / 1024),
		GcTotalCount: int(memStats.NumGC),
	})

}

func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %02dh %02dm %02ds", days, hours, minutes, seconds)
	} else if hours > 0 {
		return fmt.Sprintf("%02dh %02dm %02ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%02dm %02ds", minutes, seconds)
	} else {
		return fmt.Sprintf("%02ds", seconds)
	}
}

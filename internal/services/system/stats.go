package system

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// SystemStats holds host system information.
type SystemStats struct {
	Hostname    string      `json:"hostname"`
	OS          string      `json:"os"`
	Arch        string      `json:"arch"`
	CPUPercent  float64     `json:"cpu_percent"`
	CPUCores    int         `json:"cpu_cores"`
	MemTotal    uint64      `json:"mem_total"`
	MemUsed     uint64      `json:"mem_used"`
	MemPercent  float64     `json:"mem_percent"`
	Disks       []DiskStats `json:"disks"`
	Uptime      int64       `json:"uptime"`
	Temperature float64     `json:"temperature,omitempty"`
}

// DiskStats holds filesystem usage for a mount point.
type DiskStats struct {
	Mount   string  `json:"mount"`
	Total   uint64  `json:"total"`
	Used    uint64  `json:"used"`
	Percent float64 `json:"percent"`
}

// GetStats collects system statistics from /proc and OS.
func GetStats() (*SystemStats, error) {
	hostname, _ := os.Hostname()

	stats := &SystemStats{
		Hostname: hostname,
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		CPUCores: runtime.NumCPU(),
	}

	stats.CPUPercent = readCPUPercent()
	readMemInfo(stats)
	stats.Disks = readDiskStats()
	stats.Uptime = readUptime()
	stats.Temperature = readTemperature()

	return stats, nil
}

// readCPUPercent reads CPU usage from /proc/stat.
// Takes two samples 200ms apart to calculate usage.
func readCPUPercent() float64 {
	idle1, total1 := readCPUSample()
	if total1 == 0 {
		return 0
	}
	time.Sleep(200 * time.Millisecond)
	idle2, total2 := readCPUSample()
	if total2 == total1 {
		return 0
	}

	idleDelta := float64(idle2 - idle1)
	totalDelta := float64(total2 - total1)
	return (1.0 - idleDelta/totalDelta) * 100
}

func readCPUSample() (idle, total uint64) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, 0
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)
			if len(fields) < 5 {
				return 0, 0
			}
			var values []uint64
			for _, f := range fields[1:] {
				v, _ := strconv.ParseUint(f, 10, 64)
				values = append(values, v)
			}
			for _, v := range values {
				total += v
			}
			if len(values) > 3 {
				idle = values[3] // idle is 4th field
			}
			return idle, total
		}
	}
	return 0, 0
}

// readMemInfo reads memory stats from /proc/meminfo.
func readMemInfo(stats *SystemStats) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return
	}

	var memTotal, memAvailable uint64
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		val, _ := strconv.ParseUint(fields[1], 10, 64)
		val *= 1024 // Convert from kB to bytes
		switch fields[0] {
		case "MemTotal:":
			memTotal = val
		case "MemAvailable:":
			memAvailable = val
		}
	}

	stats.MemTotal = memTotal
	stats.MemUsed = memTotal - memAvailable
	if memTotal > 0 {
		stats.MemPercent = float64(stats.MemUsed) / float64(memTotal) * 100
	}
}

// readDiskStats reads filesystem usage from /proc/mounts + syscall.
func readDiskStats() []DiskStats {
	data, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return nil
	}

	seen := make(map[string]bool)
	var disks []DiskStats

	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		device := fields[0]
		mount := fields[1]

		// Only include real filesystems
		if !strings.HasPrefix(device, "/dev/") {
			continue
		}
		// Skip duplicate devices
		if seen[device] {
			continue
		}
		seen[device] = true

		total, used, err := getDiskUsage(mount)
		if err != nil || total == 0 {
			continue
		}

		disks = append(disks, DiskStats{
			Mount:   mount,
			Total:   total,
			Used:    used,
			Percent: float64(used) / float64(total) * 100,
		})
	}

	return disks
}

// readUptime reads system uptime from /proc/uptime in seconds.
func readUptime() int64 {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		return 0
	}
	uptime, _ := strconv.ParseFloat(fields[0], 64)
	return int64(uptime)
}

// readTemperature tries to read CPU temperature from thermal zone.
func readTemperature() float64 {
	// Try common thermal zone paths
	paths := []string{
		"/sys/class/thermal/thermal_zone0/temp",
		"/sys/class/thermal/thermal_zone1/temp",
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		temp, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
		if err != nil {
			continue
		}
		// Temperature is in millidegrees
		return temp / 1000.0
	}
	return 0
}

// FormatBytes formats bytes into human-readable string.
func FormatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

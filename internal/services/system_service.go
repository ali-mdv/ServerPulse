package services

import (
	"server-monitoring/internal/models"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
)

type SystemService interface {
	MemUsage() models.Usage
	CPUUsage() float64
	DiskUsage() models.Usage
	NetUsage() models.NetworkUsage
	SystemUsage() models.SystemUsage
}

type systemService struct {
	lastSent     uint64
	lastReceived uint64
	lastTime     time.Time
}

func NewSystemService() SystemService {
	return &systemService{}
}

func (s *systemService) MemUsage() models.Usage {
	vmStat, _ := mem.VirtualMemory()

	return models.Usage{
		UsedPercent: vmStat.UsedPercent,
		Used:        vmStat.Used,
		Total:       vmStat.Total,
	}
}

func (s *systemService) CPUUsage() float64 {
	// CPU usage (average over 1 second)
	cpuPercent, _ := cpu.Percent(time.Second, false)
	return cpuPercent[0]
}

func (s *systemService) DiskUsage() models.Usage {
	diskStat, _ := disk.Usage("/")

	return models.Usage{
		UsedPercent: diskStat.UsedPercent,
		Used:        diskStat.Used,
		Total:       diskStat.Total,
	}
}

func (s *systemService) NetUsage() models.NetworkUsage {
	netIO, _ := net.IOCounters(false)
	if len(netIO) == 0 {
		return models.NetworkUsage{}
	}

	now := time.Now()
	interval := now.Sub(s.lastTime).Seconds()

	currentSent := netIO[0].BytesSent
	currentRecv := netIO[0].BytesRecv

	usage := models.NetworkUsage{}

	// Only compute if we have previous values
	if !s.lastTime.IsZero() {
		usage.Send = uint64(float64(currentSent-s.lastSent) / interval)
		usage.Received = uint64(float64(currentRecv-s.lastReceived) / interval)
	}

	// Update cache
	s.lastSent = currentSent
	s.lastReceived = currentRecv
	s.lastTime = now

	return usage
}

func (s *systemService) SystemUsage() models.SystemUsage {
	return models.SystemUsage{
		MemUsage:  s.MemUsage(),
		CpuUsage:  s.CPUUsage(),
		DiskUsage: s.DiskUsage(),
		NetIO:     s.NetUsage(),
	}
}

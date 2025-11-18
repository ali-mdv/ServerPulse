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

type systemService struct{}

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

	networkUsage := models.NetworkUsage{}
	if len(netIO) > 0 {
		networkUsage.Send = netIO[0].BytesSent
		networkUsage.Received = netIO[0].BytesRecv
	}

	return networkUsage
}

func (s *systemService) SystemUsage() models.SystemUsage {
	return models.SystemUsage{
		MemUsage:  s.MemUsage(),
		CpuUsage:  s.CPUUsage(),
		DiskUsage: s.DiskUsage(),
		NetIO:     s.NetUsage(),
	}
}

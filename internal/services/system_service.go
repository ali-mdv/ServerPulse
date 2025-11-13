package services

import (
	"fmt"
	"server-monitoring/internal/models"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
)

type SystemService interface {
	MemUsage() models.Usage
	CPUUsage() string
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
		UsedPercent: fmt.Sprintf("%.2f%%", vmStat.UsedPercent),
		Used:        fmt.Sprintf("%.2f GB", float64(vmStat.Used)/1024/1024/1024),
		Total:       fmt.Sprintf("%.2f GB", float64(vmStat.Total)/1024/1024/1024),
	}
}

func (s *systemService) CPUUsage() string {
	// CPU usage (average over 1 second)
	cpuPercent, _ := cpu.Percent(time.Second, false)
	return fmt.Sprintf("%.2f%%", cpuPercent[0])
}

func (s *systemService) DiskUsage() models.Usage {
	diskStat, _ := disk.Usage("/")
	return models.Usage{
		UsedPercent: fmt.Sprintf("%.2f%%", diskStat.UsedPercent),
		Used:        fmt.Sprintf("%.2f GB", float64(diskStat.Used)/1024/1024/1024),
		Total:       fmt.Sprintf("%.2f GB", float64(diskStat.Total)/1024/1024/1024),
	}
}

func (s *systemService) NetUsage() models.NetworkUsage {
	netIO, _ := net.IOCounters(false)

	networkUsage := models.NetworkUsage{}
	if len(netIO) > 0 {
		networkUsage.Send = fmt.Sprintf("%.2f MB", float64(netIO[0].BytesSent)/1024/1024)
		networkUsage.Send = fmt.Sprintf("%.2f MB", float64(netIO[0].BytesRecv)/1024/1024)
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

package models

type Usage struct {
	UsedPercent float64 `json:"percent"`
	Used        uint64  `json:"used"`
	Total       uint64  `json:"total"`
}

type NetworkUsage struct {
	Send     uint64 `json:"sent"`
	Received uint64 `json:"received"`
}

type SystemUsage struct {
	MemUsage  Usage        `json:"memUsage"`
	CpuUsage  float64      `json:"cpuUsage"`
	DiskUsage Usage        `json:"diskUsage"`
	NetIO     NetworkUsage `json:"netIO"`
}

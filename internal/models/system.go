package models

type Usage struct {
	UsedPercent string `json:"percent"`
	Used        string `json:"used"`
	Total       string `json:"total"`
}

type NetworkUsage struct {
	Send     string `json:"sent"`
	Received string `json:"received"`
}

type SystemUsage struct {
	MemUsage  Usage        `json:"memUsage"`
	CpuUsage  string       `json:"cpuUsage"`
	DiskUsage Usage        `json:"diskUsage"`
	NetIO     NetworkUsage `json:"netIO"`
}

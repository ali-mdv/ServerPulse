package models

type PM2Env struct {
	Status           string `json:"status"`
	PMUptime         int64  `json:"pm_uptime"`
	RestartTime      int    `json:"restart_time"`
	UnstableRestarts int    `json:"unstable_restarts"`
	ExecPath         string `json:"pm_exec_path"`
	Cwd              string `json:"pm_cwd"`
	PmxMonitoring    bool   `json:"pmx_monitoring"`
}

type PM2Monit struct {
	CPU    float64 `json:"cpu"`
	Memory float64 `json:"memory"`
}

type PM2Process struct {
	Name   string   `json:"name"`
	PID    int      `json:"pid"`
	PMID   int      `json:"pm_id"`
	PM2Env PM2Env   `json:"pm2_env"`
	Monit  PM2Monit `json:"monit"`
}

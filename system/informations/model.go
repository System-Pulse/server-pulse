package informations

type SystemInfo struct {
	Hostname string
	OS       string
	Kernel   string
	Uptime   uint64
}

type SystemMsg SystemInfo

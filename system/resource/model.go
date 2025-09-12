package resource

type Core struct {
	Usage float64
}

type CPUInfo struct {
	Core      []Core
	Usage     float64
	PerCore   []float64
	LoadAvg1  float64
	LoadAvg5  float64
	LoadAvg15 float64
}

type MemoryInfo struct {
	Total     uint64
	Used      uint64
	Free      uint64
	Usage     float64
	SwapTotal uint64
	SwapUsed  uint64
	SwapFree  uint64
	SwapUsage float64
}

type DiskInfo struct {
	Mountpoint string
	Total      uint64
	Used       uint64
	Free       uint64
	Usage      float64
}

type NetworkInterface struct {
	Name      string
    IPs       []string
    RxBytes   uint64
    TxBytes   uint64
    Status    string // "up" or "down"
}

type NetworkInfo struct {
	Connected  bool
	PrivateIPs []string
	PublicIPv4 string
	PublicIPv6 string
	Interfaces []NetworkInterface
}

type CpuMsg CPUInfo
type MemoryMsg MemoryInfo
type DiskMsg []DiskInfo
type NetworkMsg NetworkInfo

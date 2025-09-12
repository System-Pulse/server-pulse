package resource

import (
	"net"
	"strings"
	"time"

	"github.com/System-Pulse/server-pulse/utils"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	network "github.com/shirou/gopsutil/v4/net"
)

func UpdateCPUInfo() tea.Cmd {
	return func() tea.Msg {
		percent, err := cpu.Percent(time.Second, true)
		if err != nil {
			return utils.ErrMsg(err)
		}

		avg, err := cpu.Percent(time.Second, false)
		if err != nil {
			return utils.ErrMsg(err)
		}

		loadAvg, err := utils.LoadAvg()
		if err != nil {
			return utils.ErrMsg(err)
		}

		return CpuMsg{
			Usage:     avg[0],
			PerCore:   percent,
			LoadAvg1:  loadAvg[0],
			LoadAvg5:  loadAvg[1],
			LoadAvg15: loadAvg[2],
		}
	}
}

func UpdateMemoryInfo() tea.Cmd {
	return func() tea.Msg {
		vmem, err := mem.VirtualMemory()
		if err != nil {
			return utils.ErrMsg(err)
		}

		swap, err := mem.SwapMemory()
		if err != nil {
			return utils.ErrMsg(err)
		}

		return MemoryMsg{
			Total:     vmem.Total,
			Used:      vmem.Used,
			Free:      vmem.Free,
			Usage:     vmem.UsedPercent,
			SwapTotal: swap.Total,
			SwapUsed:  swap.Used,
			SwapFree:  swap.Free,
			SwapUsage: swap.UsedPercent,
		}
	}
}

func UpdateDiskInfo() tea.Cmd {
	return func() tea.Msg {
		partitions, err := disk.Partitions(false)
		if err != nil {
			return utils.ErrMsg(err)
		}

		var disks []DiskInfo
		for _, p := range partitions {
			usage, err := disk.Usage(p.Mountpoint)
			if err != nil {
				continue
			}
			if p.Mountpoint == "/" || p.Mountpoint == "/home" || p.Mountpoint == "/efi" {
				disks = append(disks, DiskInfo{
					Mountpoint: p.Mountpoint,
					Total:      usage.Total,
					Used:       usage.Used,
					Free:       usage.Free,
					Usage:      usage.UsedPercent,
				})
			}
		}

		return DiskMsg(disks)
	}
}

func UpdateNetworkInfo() tea.Cmd {
    return func() tea.Msg {
        connected := true

        interfaces, err := net.Interfaces()
        if err != nil {
            return utils.ErrMsg(err)
        }

        var privateIPs []string
        var networkInterfaces []NetworkInterface

        for _, iface := range interfaces {
            var interfaceIPs []string
            status := "down"
            if iface.Flags&net.FlagUp != 0 {
                status = "up"
            }

            addrs, err := iface.Addrs()
            if err == nil {
                for _, addr := range addrs {
                    ipNet, ok := addr.(*net.IPNet)
                    if ok /*&& !ipNet.IP.IsLoopback()*/ {
                        if ipNet.IP.To4() != nil {
                            ipStr := ipNet.IP.String()
                            interfaceIPs = append(interfaceIPs, ipStr)
                            
                            if !strings.HasPrefix(ipStr, "127.") {
                                privateIPs = append(privateIPs, ipStr)
                            }
                        }
                    }
                }
            }

            networkInterfaces = append(networkInterfaces, NetworkInterface{
                Name:   iface.Name,
                IPs:    interfaceIPs,
                Status: status,
            })
        }

        netIO, err := network.IOCounters(true)
        if err == nil {
            for i := range networkInterfaces {
                for _, io := range netIO {
                    if networkInterfaces[i].Name == io.Name {
                        networkInterfaces[i].RxBytes = io.BytesRecv
                        networkInterfaces[i].TxBytes = io.BytesSent
                        break
                    }
                }
            }
        }

        publicIPv4 := "N/A"
        publicIPv6 := "N/A"

        return NetworkMsg{
            Connected:  connected,
            PrivateIPs: privateIPs,
            PublicIPv4: publicIPv4,
            PublicIPv6: publicIPv6,
            Interfaces: networkInterfaces,
        }
    }
}

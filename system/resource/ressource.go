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

		addrs, err := net.Interfaces()
		if err != nil {
			return utils.ErrMsg(err)
		}

		var privateIPs []string
		for _, addr := range addrs {
			addresses, err := addr.Addrs()
			if err != nil {
				continue
			}
			for _, a := range addresses {
				if strings.Contains(a.String(), ".") && !strings.HasPrefix(a.String(), "127.") {
					privateIPs = append(privateIPs, a.String())
				}
			}
		}

		publicIPv4 := "N/A"
		publicIPv6 := "N/A"

		var networkInterfaces []NetworkInterface
		netIO, err := network.IOCounters(true) 
		if err == nil {
			for _, io := range netIO {
				networkInterfaces = append(networkInterfaces, NetworkInterface{
					Name:    io.Name,
					RxBytes: io.BytesRecv,
					TxBytes: io.BytesSent,
				})
			}
		}

		return NetworkMsg{
			Connected:  connected,
			PrivateIPs: privateIPs,
			PublicIPv4: publicIPv4,
			PublicIPv6: publicIPv6,
		}
	}
}

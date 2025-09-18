package widgets

import (
	"time"

	model "github.com/System-Pulse/server-pulse/widgets/model"
)

const (
	ChartCPU model.ChartType = iota
	ChartMemory
	ChartNetworkRX
	ChartNetworkTX
	ChartDiskRead
	ChartDiskWrite
)

var lastNetworkTotals = make(map[string]struct {
	RxBytes uint64
	TxBytes uint64
	Time    time.Time
})

func (m Model) updateCharts() {
	now := time.Now()

	m.Monitor.CpuHistory.Points = append(m.Monitor.CpuHistory.Points, model.DataPoint{
		Timestamp: now,
		Value:     m.Monitor.Cpu.Usage,
	})
	if len(m.Monitor.CpuHistory.Points) > m.Monitor.CpuHistory.MaxPoints {
		m.Monitor.CpuHistory.Points = m.Monitor.CpuHistory.Points[1:]
	}

	m.Monitor.MemoryHistory.Points = append(m.Monitor.MemoryHistory.Points, model.DataPoint{
		Timestamp: now,
		Value:     m.Monitor.Memory.Usage,
	})
	if len(m.Monitor.MemoryHistory.Points) > m.Monitor.MemoryHistory.MaxPoints {
		m.Monitor.MemoryHistory.Points = m.Monitor.MemoryHistory.Points[1:]
	}

	if len(m.Network.NetworkResource.Interfaces) > 0 {
		iface := m.Network.NetworkResource.Interfaces[0]

		key := iface.Name
		last, exists := lastNetworkTotals[key]

		var rxRate, txRate float64
		if exists {
			timeDiff := now.Sub(last.Time).Seconds()
			if timeDiff > 0 {
				rxRate = float64(iface.RxBytes-last.RxBytes) / timeDiff / 1024 / 1024
				txRate = float64(iface.TxBytes-last.TxBytes) / timeDiff / 1024 / 1024
			}
		}

		lastNetworkTotals[key] = struct {
			RxBytes uint64
			TxBytes uint64
			Time    time.Time
		}{
			RxBytes: iface.RxBytes,
			TxBytes: iface.TxBytes,
			Time:    now,
		}

		m.Monitor.NetworkRxHistory.Points = append(m.Monitor.NetworkRxHistory.Points, model.DataPoint{
			Timestamp: now,
			Value:     rxRate,
		})
		m.Monitor.NetworkTxHistory.Points = append(m.Monitor.NetworkTxHistory.Points, model.DataPoint{
			Timestamp: now,
			Value:     txRate,
		})

		if len(m.Monitor.NetworkRxHistory.Points) > m.Monitor.NetworkRxHistory.MaxPoints {
			m.Monitor.NetworkRxHistory.Points = m.Monitor.NetworkRxHistory.Points[1:]
		}
		if len(m.Monitor.NetworkTxHistory.Points) > m.Monitor.NetworkTxHistory.MaxPoints {
			m.Monitor.NetworkTxHistory.Points = m.Monitor.NetworkTxHistory.Points[1:]
		}
	}
}

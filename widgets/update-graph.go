package widgets

import "time"

// Point de données pour les graphiques
type DataPoint struct {
	Timestamp time.Time
	Value     float64
}

// Historique des données
type DataHistory struct {
	Points    []DataPoint
	MaxPoints int
}

// Types de graphiques
type ChartType int

const (
	ChartCPU ChartType = iota
	ChartMemory
	ChartNetworkRX
	ChartNetworkTX
	ChartDiskRead
	ChartDiskWrite
)

// Configuration des graphiques
type ChartConfig struct {
	Title      string
	MaxValue   float64
	Height     int
	Width      int
	ShowLabels bool
}

func (m *Model) updateCharts() {
	now := time.Now()

	// Mettre à jour l'historique CPU
	m.cpuHistory.Points = append(m.cpuHistory.Points, DataPoint{
		Timestamp: now,
		Value:     m.cpu.Usage,
	})
	if len(m.cpuHistory.Points) > m.cpuHistory.MaxPoints {
		m.cpuHistory.Points = m.cpuHistory.Points[1:]
	}

	// Mettre à jour l'historique mémoire
	m.memoryHistory.Points = append(m.memoryHistory.Points, DataPoint{
		Timestamp: now,
		Value:     m.memory.Usage,
	})
	if len(m.memoryHistory.Points) > m.memoryHistory.MaxPoints {
		m.memoryHistory.Points = m.memoryHistory.Points[1:]
	}

	// Mettre à jour l'historique réseau
	if len(m.network.Interfaces) > 0 {
		// Utiliser la première interface (généralement eth0 ou enp0s3)
		iface := m.network.Interfaces[0]

		// Calculer le débit en MB/s depuis la dernière mise à jour
		var rxRate, txRate float64
		if len(m.networkRxHistory.Points) > 0 {
			lastPoint := m.networkRxHistory.Points[len(m.networkRxHistory.Points)-1]
			timeDiff := now.Sub(lastPoint.Timestamp).Seconds()
			if timeDiff > 0 {
				rxRate = (float64(iface.RxBytes) - lastPoint.Value*1024*1024) / timeDiff / 1024 / 1024
			}
		}

		if len(m.networkTxHistory.Points) > 0 {
			lastPoint := m.networkTxHistory.Points[len(m.networkTxHistory.Points)-1]
			timeDiff := now.Sub(lastPoint.Timestamp).Seconds()
			if timeDiff > 0 {
				txRate = (float64(iface.TxBytes) - lastPoint.Value*1024*1024) / timeDiff / 1024 / 1024
			}
		}

		m.networkRxHistory.Points = append(m.networkRxHistory.Points, DataPoint{
			Timestamp: now,
			Value:     rxRate,
		})
		m.networkTxHistory.Points = append(m.networkTxHistory.Points, DataPoint{
			Timestamp: now,
			Value:     txRate,
		})

		if len(m.networkRxHistory.Points) > m.networkRxHistory.MaxPoints {
			m.networkRxHistory.Points = m.networkRxHistory.Points[1:]
		}
		if len(m.networkTxHistory.Points) > m.networkTxHistory.MaxPoints {
			m.networkTxHistory.Points = m.networkTxHistory.Points[1:]
		}
	}

	m.lastChartUpdate = now
}

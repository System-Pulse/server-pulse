package network

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/System-Pulse/server-pulse/utils"
	tea "github.com/charmbracelet/bubbletea"
)

func GetRoutes() tea.Cmd {
	return func() tea.Msg {
		// Ouvrir le fichier de routage du noyau
		file, err := os.Open("/proc/net/route")
		if err != nil {
			return utils.ErrMsg(fmt.Errorf("failed to open /proc/net/route: %w", err))
		}
		defer file.Close()

		var routes []RouteInfo
		scanner := bufio.NewScanner(file)

		// Ignorer l'en-tête
		if scanner.Scan() {
			// Première ligne est l'en-tête, on la saute
		}

		for scanner.Scan() {
			line := scanner.Text()
			fields := strings.Fields(line)

			if len(fields) < 8 {
				continue
			}

			// Le format de /proc/net/route est:
			// Iface Destination Gateway Flags RefCnt Use Metric Mask MTU Window IRTT
			iface := fields[0]
			destHex := fields[1]
			gatewayHex := fields[2]
			flags := fields[3]
			refCnt := fields[4]
			use := fields[5]
			metric := fields[6]
			maskHex := fields[7]

			// Convertir les adresses hexadécimales en IP
			destination := utils.HexToIP(destHex)
			gateway := utils.HexToIP(gatewayHex)
			genmask := utils.HexToIP(maskHex)

			// Convertir les flags numériques en lettres
			flagsStr := utils.ParseRouteFlags(flags)

			routes = append(routes, RouteInfo{
				Destination: destination,
				Gateway:     gateway,
				Genmask:     genmask,
				Flags:       flagsStr,
				Metric:      metric,
				Ref:         refCnt,
				Use:         use,
				Iface:       iface,
			})
		}

		if err := scanner.Err(); err != nil {
			return utils.ErrMsg(fmt.Errorf("error reading /proc/net/route: %w", err))
		}

		return RoutesMsg(routes)
	}
}

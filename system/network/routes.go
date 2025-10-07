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
		var routes []RouteInfo

		ipv4Routes, err := readIPv4Routes()
		if err != nil {
			return utils.ErrMsg(fmt.Errorf("failed to read IPv4 routes: %w", err))
		}
		routes = append(routes, ipv4Routes...)

		ipv6Routes, err := readIPv6Routes()
		if err != nil {
			return utils.ErrMsg(fmt.Errorf("failed to read IPv6 routes: %w", err))
		}
		routes = append(routes, ipv6Routes...)

		return RoutesMsg(routes)
	}
}

func readIPv4Routes() ([]RouteInfo, error) {
	file, err := os.Open("/proc/net/route")
	if err != nil {
		return nil, fmt.Errorf("failed to open /proc/net/route: %w", err)
	}
	defer file.Close()

	var routes []RouteInfo
	scanner := bufio.NewScanner(file)

	if scanner.Scan() {
	}

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		if len(fields) < 8 {
			continue
		}

		iface := fields[0]
		destHex := fields[1]
		gatewayHex := fields[2]
		flags := fields[3]
		refCnt := fields[4]
		use := fields[5]
		metric := fields[6]
		maskHex := fields[7]

		destination := utils.HexToIPv4(destHex)
		gateway := utils.HexToIPv4(gatewayHex)
		genmask := utils.HexToIPv4(maskHex)

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
		return nil, fmt.Errorf("error reading /proc/net/route: %w", err)
	}

	return routes, nil
}

func readIPv6Routes() ([]RouteInfo, error) {
	file, err := os.Open("/proc/net/ipv6_route")
		if err != nil {
			return nil, err
		}
		defer file.Close()

		var routes []RouteInfo
		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			line := scanner.Text()
			fields := strings.Fields(line)
			
			if len(fields) < 10 {
				continue
			}

			destHex := fields[0]
			destPrefix := fields[1]
			gatewayHex := fields[4]
			metric := fields[5]
			refCnt := fields[6]
			use := fields[7]
			flags := fields[8]
			iface := fields[9]

			destination, err := utils.HexToIPv6(destHex)
			if err != nil {
				continue
			}

			destination = fmt.Sprintf("%s/%s", destination, destPrefix)

			gateway, err := utils.HexToIPv6(gatewayHex)
			if err != nil {
				gateway = ""
			}

			flagsStr := utils.ParseIPv6RouteFlags(flags)

			routes = append(routes, RouteInfo{
				Destination: destination,
				Gateway:     gateway,
				Genmask:     destPrefix,
				Flags:       flagsStr,
				Metric:      metric,
				Ref:         refCnt,
				Use:         use,
				Iface:       iface,
			})
		}

		if err := scanner.Err(); err != nil {
			return nil, err
		}

		return routes, nil
}

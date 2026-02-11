package network

import "time"

type ConnectionInfo struct {
	Proto       string
	RecvQ       string
	SendQ       string
	LocalAddr   string
	ForeignAddr string
	State       string
	PID         string
}

type RouteInfo struct {
	Destination string
	Gateway     string
	Genmask     string
	Flags       string
	Metric      string
	Ref         string
	Use         string
	Iface       string
}

type DNSInfo struct {
	Server string
}

type ConnectionsMsg []ConnectionInfo
type RoutesMsg []RouteInfo
type DNSMsg []DNSInfo

type PingResult struct {
	Target     string
	Success    bool
	Latency    time.Duration
	PacketLoss float64
	Error      string
}

type TracerouteResult struct {
	Target string
	Hops   []TracerouteHop
	Error  string
}

type TracerouteHop struct {
	HopNumber int
	IP        string
	Hostname  string
	Latency1  time.Duration
	Latency2  time.Duration
	Latency3  time.Duration
}

type PingMsg PingResult
type TracerouteMsg TracerouteResult
type TracerouteInstallPromptMsg struct {
	Target string
}

type TracerouteInstallResultMsg struct {
	Success         bool
	Target          string
	Error           string
	PasswordInvalid bool
}

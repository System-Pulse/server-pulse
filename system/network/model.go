package network

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

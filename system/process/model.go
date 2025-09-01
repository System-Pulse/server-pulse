package process

type ProcessInfo struct {
	PID     int32
	User    string
	CPU     float64
	Mem     float64
	Command string
}

type ProcessMsg []ProcessInfo

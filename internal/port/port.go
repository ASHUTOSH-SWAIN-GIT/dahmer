package port

// Process describes a TCP listener.
type Process struct {
	PID     int
	Command string
	User    string
	Port    int
}

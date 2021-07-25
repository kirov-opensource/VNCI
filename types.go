package main

type ConnectionType int

const (
	ConnectionType_IN ConnectionType = 1 << iota
	ConnectionType_OUT
)

type ConnectionConfig struct {
	localPort      int
	remotePort     int
	remoteAddress  string
	connectionType ConnectionType
}

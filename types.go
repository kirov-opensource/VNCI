package main

type StatusType string

const (
	StatusTypeActive  StatusType = "Active"
	StatusTypeDisable StatusType = "Disable"
)

type ConnectionType string

const (
	ConnectionTypeClient ConnectionType = "Client"
	ConnectionTypeServer ConnectionType = "Server"
)

type ConnectionConfig struct {
	LocalAddress   string         `json:"localAddress"`
	LocalPort      int            `json:"localPort"`
	RemoteAddress  string         `json:"remoteAddress"`
	RemotePort     int            `json:"remotePort"`
	RawMode        string         `json:"rawMode"`
	CipherMode     string         `json:"cipherMode"`
	AuthMode       string         `json:"authMode"`
	ConnectionType ConnectionType `json:"connectionType"`
	Status         StatusType     `json:"status"`
}

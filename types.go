package main

type StatusType string

const (
	StatusType_ACTIVE  StatusType = "Active"
	StatusType_DISABLE StatusType = "Disable"
)

type ConnectionType string

const (
	ConnectionType_CLIENT ConnectionType = "Client"
	ConnectionType_SERVER ConnectionType = "Server"
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

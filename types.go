package main

type StatusType string

const (
	StatusType_ACTIVE  StatusType = "Active"
	StatusType_DISABLE StatusType = "Disbale"
)

type ConnectionType string

const (
	ConnectionType_IN  ConnectionType = "In"
	ConnectionType_OUT                = "Out"
)

type ConnectionConfig struct {
	LocalPort      int            `json:"localPort"`
	RemotePort     int            `json:"remotePort"`
	RemoteAddress  string         `json:"remoteAddress"`
	ConnectionType ConnectionType `json:"connectionType"`
	Status         StatusType     `json:"status"`
}

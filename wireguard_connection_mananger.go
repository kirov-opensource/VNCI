package main

type WireGuardInterfacePeer struct {
	PublicKey       string `json:"publicKey"`
	Endpoint        string `json:"endpoint"`
	AllowedIps      string `json:"allowedIps"`
	TransferIn      string `json:"transferIn"`
	TransferOut     string `json:"transferOut"`
	KeepAlive       string `json:"keepAlive"`
	LatestHandshake string `json:"latestHandshake"`
}

type WireguardInterface struct {
	Address    string                            `json:"address"`
	PublicKey  string                            `json:"publicKey"`
	PrivateKey string                            `json:"privateKey"`
	ListenPort string                            `json:"listenPort"`
	Peers      map[string]WireGuardInterfacePeer `json:"peers"`
}

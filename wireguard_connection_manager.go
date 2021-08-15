package main

import "sync"

type IWireguardConnectionManager interface {
	// 初始化
	Initial(configPath, templatePath, serviceDestPath, configDestPath, executionPath, localLibraryPath string)

	// 加载数据
	LoadData() (map[int]UDP2RawConnection, error)

	// 打印列表
	List()

	// 保存数据
	SaveConfig() error

	// 添加链接
	Add(item UDP2RawConnection) error

	// 删除链接
	Remove(key int) bool

	// 是否存在
	ContainsKey(key int) bool

	// 获取
	Get(key int) (*UDP2RawConnection, error)

	// 同步物理文件
	SyncPhysicalFiles(item UDP2RawConnection) (string, string)

	// 切换状态
	ToggleStatus(conn UDP2RawConnection) UDP2RawConnectionStatusType
}

type WireguardConnectionManager struct {
	mu sync.Mutex

	data map[string]*WireGuardInterfacePeer

	ConfigPath, TemplatePath, ServiceDestPath, ConfigDestPath, ExecutionPath, LocalLibraryPath string
}

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

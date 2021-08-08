package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"vnci/utils"

	"github.com/google/logger"
	"github.com/olekukonko/tablewriter"
)

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

type IConnectionManager interface {
	// 初始化
	Initial(configPath, templatePath, serviceDestPath, configDestPath, executionPath, localLibraryPath string)

	// 加载数据
	LoadData() (map[int]ConnectionItem, error)

	// 打印列表
	List()

	// 保存数据
	SaveConfig() error

	// 添加链接
	Add(item ConnectionItem) error

	// 删除链接
	Remove(key int) bool

	// 是否存在
	ContainsKey(key int) bool

	// 获取
	Get(key int) (*ConnectionItem, error)

	// 同步物理文件
	SyncPhysicalFiles(item ConnectionItem) (string, string)

	// 切换状态
	ToggleStatus(conn ConnectionItem) StatusType
}

type ConnectionManager struct {
	mu sync.Mutex

	data map[int]*ConnectionItem

	ConfigPath, TemplatePath, ServiceDestPath, ConfigDestPath, ExecutionPath, LocalLibraryPath string
}

func (_self *ConnectionManager) Initial(configPath, templatePath, serviceDestPath, configDestPath, executionPath, localLibraryPath string) {
	_self.TemplatePath = templatePath
	_self.ServiceDestPath = serviceDestPath
	_self.ConfigDestPath = configDestPath
	_self.ExecutionPath = executionPath
	_self.ConfigPath = configPath
	_self.LocalLibraryPath = localLibraryPath

	_self.data = make(map[int]*ConnectionItem)
	_self.mu = sync.Mutex{}

	os.MkdirAll(_self.ConfigDestPath+"/udp2raw", 0777)
	os.MkdirAll(_self.ExecutionPath, 0777)
	utils.Copy(_self.LocalLibraryPath+"/udp2raw/udp2raw", _self.ExecutionPath+"/udp2raw")

	utils.RunCmd("chmod", "+x", "/usr/local/bin/vnci/udp2raw")

	fileExist := utils.Exists(configPath)

	if !fileExist {
		logger.Infoln("未找到配置文件，正在初始化...")
		err := utils.CreateFile(configPath, []byte("{}"))
		if err != nil {
			logger.Fatalln("初始化失败,请检查日志文件", err)
		}
	}

	_self.LoadConfig()
}

func (_self *ConnectionManager) Get(key int) (*ConnectionItem, error) {
	if _self.ContainsKey(key) {
		var data = _self.data[key]
		var a = _self.data[key]

		// &_self.data
		// a := (*sData)[key]

		fmt.Printf("%p\n", &a)
		fmt.Printf("%p\n", &_self.data)
		fmt.Printf("%p\n", &data)
		return data, nil
	}
	return nil, errors.New("not found")
}

func (_self *ConnectionManager) List() {
	tunnelSize := len(_self.data)

	if tunnelSize == 0 {
		fmt.Println("暂无数据")
		return
	}

	activeColor := tablewriter.Colors{tablewriter.FgGreenColor, tablewriter.Bold}
	disableColor := tablewriter.Colors{tablewriter.FgRedColor, tablewriter.Bold}

	inColor := tablewriter.Colors{tablewriter.FgGreenColor, tablewriter.Bold}
	outColor := tablewriter.Colors{tablewriter.FgRedColor, tablewriter.Bold}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"No", "Local", "Remote", "Type", "Status"})

	for key, tunnel := range _self.data {

		rowData := []string{
			strconv.Itoa(key), tunnel.LocalAddress + ":" + strconv.Itoa(tunnel.LocalPort), tunnel.RemoteAddress + ":" + strconv.Itoa(tunnel.RemotePort), string(tunnel.ConnectionType), string(tunnel.Status),
		}
		typeColor := inColor
		if tunnel.ConnectionType == ConnectionTypeServer {
			typeColor = outColor
		}

		statusColor := activeColor
		if tunnel.Status == StatusTypeDisable {
			statusColor = disableColor
		}

		table.Rich(rowData, []tablewriter.Colors{{}, {}, {}, {}, typeColor, statusColor})
	}
	table.Render()
}

func (_self *ConnectionManager) ContainsKey(key int) bool {
	if _, ok := _self.data[key]; ok {
		return true
	} else {
		return false
	}
}

func (_self *ConnectionManager) Add(item *ConnectionItem) bool {
	if !_self.ContainsKey(item.LocalPort) {
		_self.data[item.LocalPort] = item
		_self.SaveConfig()
		return true
	} else {
		logger.Errorln("本地端口已被占用")
		return false
	}
}

func (_self *ConnectionManager) Remove(key int) bool {
	if _self.ContainsKey(key) {
		delete(_self.data, key)
		logger.Infoln("删除成功")
		_self.SaveConfig()
		return true
	} else {
		return false
	}
}

func (_self *ConnectionManager) ToggleStatus(conn *ConnectionItem) StatusType {
	_, serviceFileName := connectionManager.SyncPhysicalFiles(*conn)
	serviceName := serviceFileName[0:strings.LastIndex(serviceFileName, ".")]

	utils.RunCmd("systemctl", "daemon-reload")

	newStatusType := conn.Status

	if conn.Status == StatusTypeActive {
		if ok := utils.RunCmd("systemctl", "stop", serviceName); ok {
			newStatusType = StatusTypeDisable
		}
	} else {
		if ok := utils.RunCmd("systemctl", "stop", serviceName); ok {
			if ok = utils.RunCmd("systemctl", "start", serviceName); ok {
				newStatusType = StatusTypeActive
			}
		}
	}
	conn.Status = newStatusType
	// _self.data[conn.LocalPort] = conn
	_self.SaveConfig()
	return newStatusType
}

func (_self *ConnectionManager) SyncPhysicalFiles(conn ConnectionItem) (string, string) {
	actualRemoteAddress := conn.RemoteAddress
	if !utils.CheckIPAddress(actualRemoteAddress) {
		actualRemoteAddress = utils.GetActualIP(actualRemoteAddress)
	}
	confRenderModel := struct {
		ConnectionType string
		Local          string
		Remote         string
		Password       string
		RawMode        string
		CipherMode     string
		AuthMode       string
		ExtraOptions   string
	}{
		ConnectionType: "",
		Local:          conn.LocalAddress + ":" + strconv.Itoa(conn.LocalPort),
		Remote:         actualRemoteAddress + ":" + strconv.Itoa(conn.RemotePort),
		RawMode:        conn.RawMode,
		CipherMode:     conn.CipherMode,
		AuthMode:       conn.AuthMode,
		Password:       conn.Password,
		ExtraOptions:   conn.ExtraOptions,
	}
	if conn.ConnectionType == ConnectionTypeClient {
		(&confRenderModel).ConnectionType = "c"
	} else {
		(&confRenderModel).ConnectionType = "s"
	}

	udp2rawConfTemplate, _ := template.New("test").Parse(string(utils.ReadFile(_self.TemplatePath + "/udp2raw.config.template")))
	confFileName := strconv.Itoa(conn.LocalPort) + ".conf"
	confFilePath := _self.ConfigDestPath + "/udp2raw/" + confFileName
	fileInfo, err := os.Create(confFilePath)
	if err != nil {
		fmt.Println("创建文件出错:", err)
	}
	udp2rawConfTemplate.Execute(fileInfo, confRenderModel)

	serviceRenderModel := struct {
		Port          string
		ExceutionPath string
	}{
		Port:          strconv.Itoa(conn.LocalPort),
		ExceutionPath: _self.ExecutionPath,
	}

	udp2rawServiceTemplate, _ := template.New("test").Parse(string(utils.ReadFile(_self.TemplatePath + "/udp2raw.service.template")))
	serviceFileName := "vnci@udp2raw@" + strconv.Itoa(conn.LocalPort) + "@" + (&confRenderModel).ConnectionType + ".service"
	serviceFilePath := _self.ServiceDestPath + "/" + serviceFileName
	fileInfo, err = os.Create(serviceFilePath)
	if err != nil {
		fmt.Println("创建文件出错:", err)
	}
	udp2rawServiceTemplate.Execute(fileInfo, serviceRenderModel)

	return confFileName, serviceFileName
}

func (_self *ConnectionManager) LoadConfig() (map[int]*ConnectionItem, error) {

	logger.Infoln("正在加载配置...")

	data := utils.ReadFile(_self.ConfigPath)

	err := json.Unmarshal(data, &_self.data)

	if err != nil {
		logger.Errorln("反序列化失败:", err)
		return nil, err
	} else {
		logger.Infoln("加载配置成功")
		return _self.data, nil
	}
}

func (_self *ConnectionManager) SaveConfig() error {
	logger.Infoln("正在保存...")
	text, err := json.MarshalIndent(_self.data, "", "    ")
	if err != nil {
		logger.Errorln("保存失败(序列化):", err)
		return err
	}
	err = ioutil.WriteFile(_self.ConfigPath, text, 0777)
	if err != nil {
		logger.Errorln("保存失败:", err)
		return err
	} else {
		logger.Info("保存成功")
	}
	return nil
}

type ConnectionItem struct {
	LocalAddress   string         `json:"localAddress"`
	LocalPort      int            `json:"localPort"`
	RemoteAddress  string         `json:"remoteAddress"`
	RemotePort     int            `json:"remotePort"`
	RawMode        string         `json:"rawMode"`
	CipherMode     string         `json:"cipherMode"`
	AuthMode       string         `json:"authMode"`
	Password       string         `json:"password"`
	ConnectionType ConnectionType `json:"connectionType"`
	Status         StatusType     `json:"status"`
	MD5            string         `json:"md5"`
	ExtraOptions   string         `json:"extraOptions"`
}

func NewConnectionItem(item ConnectionItem) *ConnectionItem {
	var message = fmt.Sprintf("%s%s%d%s%s%s%s%s", item.LocalAddress, item.RemoteAddress, item.RemotePort, item.RawMode, item.CipherMode, item.AuthMode, item.Password, item.ExtraOptions)
	item.MD5 = utils.MD5(message)
	return &item
}

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
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

// var templatePath string = "./templates"
// var serviceDestPath string = "/etc/systemd/system"
// var configDestPath string = "/usr/local/etc/vnci"
// var executionPath string = "/usr/local/bin/vnci"

type IConnectionManager interface {
	Initial(configPath, templatePath, serviceDestPath, configDestPath, executionPath string)
	//加载数据
	LoadData() (map[int]ConnectionItem, error)

	//打印列表
	List()

	//保存数据
	SaveConfig() error

	//添加链接
	Add(item ConnectionItem) error

	//删除链接
	Remove(key int) bool

	//是否存在
	ContainsKey(key int) bool

	//获取
	Get(key int) (*ConnectionItem, error)
}

type ConnectionManager struct {
	mu sync.Mutex

	data map[int]ConnectionItem

	ConfigPath, TemplatePath, ServiceDestPath, ConfigDestPath, ExecutionPath string
}

func (_self *ConnectionManager) Get(key int) (*ConnectionItem, error) {
	if _self.ContainsKey(key) {
		var data = _self.data[key]
		return &data, nil
	}
	return nil, errors.New("Not found")
}
func (_self *ConnectionManager) Add(item ConnectionItem) {

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

func (_self *ConnectionManager) Initial(configPath, templatePath, serviceDestPath, configDestPath, executionPath string) {
	_self.TemplatePath = templatePath
	_self.ServiceDestPath = serviceDestPath
	_self.ConfigDestPath = configDestPath
	_self.ExecutionPath = executionPath
	_self.ConfigPath = configPath
	_self.data = make(map[int]ConnectionItem)
	_self.mu = sync.Mutex{}
	_self.LoadConfig()
}

func (_self *ConnectionManager) LoadConfig() (map[int]ConnectionItem, error) {

	logger.Infoln("正在加载配置...")

	initialConfig()

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

func (_self *ConnectionManager) ContainsKey(key int) bool {
	if _, ok := _self.data[key]; ok {
		return true
	} else {
		return false
	}
}

func (_self *ConnectionManager) Remove(key int) bool {

	if _self.ContainsKey(key) {
		delete(_self.data, key)
		logger.Infoln("删除成功")
		return true
	} else {
		return false
	}

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
}

func initialConfig() {
	fileExist := utils.Exists(configPath)

	if !fileExist {
		logger.Infoln("未找到配置文件，正在初始化...")
		err := utils.CreateFile(configPath, []byte("{}"))
		if err != nil {
			logger.Fatalln("初始化失败,请检查日志文件", err)
		}
	}
}

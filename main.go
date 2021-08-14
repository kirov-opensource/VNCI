package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"vnci/utils"
	// "github.com/google/logger"
)

var templatePath string = "./templates"
var serviceDestPath string = "/etc/systemd/system"
var configDestPath string = "/usr/local/etc/vnci"
var executionPath string = "/usr/local/bin/vnci"
var localLibraryPath string = "./library"

var connectionManager UDP2RawConnectionManager

var configPath = "./configs/config.json"

var runAs = *(flag.String("r", "t", "s:Service,t:Tools"))

func GetValueFromStdin(parameterName string, optional bool, hasDefaultValue bool, defaultValue string, validFunc func(string) bool) string {
	f := bufio.NewReader(os.Stdin)
	var result string
	promptText := "请输入" + parameterName
	if optional {
		promptText = "[可选]" + promptText
	} else {
		promptText = "[必填]" + promptText
	}
	if hasDefaultValue {
		promptText += "(默认值:" + defaultValue + ")"
	}
	promptText += ":"
	for {
		fmt.Print(promptText)
		Input, _ := f.ReadString('\n')
		Input = Input[:len(Input)-1]
		if len(Input) == 0 {
			if hasDefaultValue {
				result = defaultValue
				break
			}
			if optional {
				result = ""
				break
			}
		}
		if validFunc != nil {
			if validFunc(Input) {
				result = Input
				break
			}
		}
		if optional || Input != "" {
			result = Input
			break
		}
		fmt.Println(parameterName + "填写有误,请重新输入")
	}
	return result
}

func OperateTunnel(callback func(int)) {
	ListTunnel()

	f := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("请输入编号[NO],输入e退出>")
		Input, _ := f.ReadString('\n')
		Input = Input[:len(Input)-1]
		if len(Input) == 0 {
			continue
		}
		if Input == "e" {
			break
		}
		dataId, err := strconv.Atoi(Input)
		if err != nil {
			fmt.Println("输入正确的编号")
		}
		if connectionManager.ContainsKey(dataId) {
			callback(dataId)
			break
		}
	}
}

func CreateTunnel() {
	var actualType UDP2RawConnectionType
	tunnelType := GetValueFromStdin("通道类型(Server/Client)", false, true, "Client", nil)
	if tunnelType == string(ConnectionTypeClient) {
		actualType = ConnectionTypeClient
	} else {
		actualType = ConnectionTypeServer
	}
	localAddress := GetValueFromStdin("本地监听地址", false, true, "0.0.0.0", utils.CheckIPAddress)
	localPort, _ := strconv.Atoi(GetValueFromStdin("本地端口", false, false, "", nil))
	remoteAddress := GetValueFromStdin("远端地址", false, false, "", nil)
	remotePort, _ := strconv.Atoi(GetValueFromStdin("远端端口", false, false, "", nil))

	rawMode := GetValueFromStdin("模拟方式", false, true, "faketcp", nil)
	cipherMode := GetValueFromStdin("加密方式", false, true, "aes128cbc", nil)
	authMode := GetValueFromStdin("认证方式", false, true, "md5", nil)
	password := GetValueFromStdin("密码", false, false, "", nil)

	connectionManager.Add(NewUDP2RawConnection(UDP2RawConnection{localAddress,
		localPort, remoteAddress, remotePort,
		rawMode, cipherMode, authMode, password, actualType, StatusTypeDisable, "", ""}))
}

func DeleteTunnel() {
	OperateTunnel(func(dataId int) {
		connectionManager.Remove(dataId)
		fmt.Println("删除成功")
	})
}

func ToggleTunnelStatus() {
	OperateTunnel(func(dataId int) {
		data, _ := connectionManager.Get(dataId)
		newStatus := connectionManager.ToggleStatus(data)
		if newStatus == StatusTypeActive {
			fmt.Println("切换成功,已启用")
		} else {
			fmt.Println("切换成功,已停止")
		}
	})
}

func ListTunnel() {
	connectionManager.List()
}

func main() {
	connectionManager.Initial(configPath, templatePath, serviceDestPath, configDestPath, executionPath, localLibraryPath)

	// 把用户传递的命令行参数解析为对应变量的值
	flag.Parse()

	if runAs == "t" {
		f := bufio.NewReader(os.Stdin) //读取输入的内容
		for {
			fmt.Println("l:\t列表")
			fmt.Println("n:\t新增")
			fmt.Println("d:\t删除")
			fmt.Println("s:\t同步")
			fmt.Println("t:\t切换状态")
			fmt.Print("请输入一些字符串,输入e退出>")
			Input, _ := f.ReadString('\n') //定义一行输入的内容分隔符。
			Input = Input[:len(Input)-1]
			if len(Input) == 0 {
				continue //如果用户输入的是一个空行就让用户继续输入。
			}
			if Input == "e" {
				break
			}
			switch Input {
			case "l":
				ListTunnel()
			case "n":
				CreateTunnel()
			case "d":
				DeleteTunnel()
			case "t":
				ToggleTunnelStatus()
			}
		}
	}

	fmt.Println("正在退出...")

	connectionManager.SaveConfig()

	os.Exit(0)
}

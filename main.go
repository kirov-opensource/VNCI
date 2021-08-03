package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/template"
	"vnci/utils"
	// "github.com/google/logger"
)

var templatePath string = "./templates"
var serviceDestPath string = "/etc/systemd/system"
var configDestPath string = "/usr/local/etc/vnci"
var executionPath string = "/usr/local/bin/vnci"

var connectionManager ConnectionManager

// var connections = make(map[int]ConnectionItem)
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
			if validFunc(Input) == true {
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

func CreateTunnel() {
	var actualType ConnectionType
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

	connectionManager.Add(ConnectionItem{localAddress,
		localPort, remoteAddress, remotePort,
		rawMode, cipherMode, authMode, password, actualType, StatusTypeDisable})
}

func DeleteTunnel() {
	connectionManager.List()

	f := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("请输入编号[NO],输入exit退出>")
		Input, _ := f.ReadString('\n')
		Input = Input[:len(Input)-1]
		if len(Input) == 0 {
			continue
		}
		if Input == "exit" {
			break
		}
		dataId, err := strconv.Atoi(Input)
		if err != nil {
			fmt.Println("输入正确的编号")
		}
		if connectionManager.ContainsKey(dataId) {
			connectionManager.Remove(dataId)
			fmt.Println("删除成功")
			break
		}
	}
}

func CreateUDP2RAWConfig(conn ConnectionItem) (string, string) {
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
	}{
		ConnectionType: "",
		Local:          conn.LocalAddress + ":" + strconv.Itoa(conn.LocalPort),
		Remote:         actualRemoteAddress + ":" + strconv.Itoa(conn.RemotePort),
		RawMode:        conn.RawMode,
		CipherMode:     conn.CipherMode,
		AuthMode:       conn.AuthMode,
		Password:       conn.Password,
	}
	if conn.ConnectionType == ConnectionTypeClient {
		(&confRenderModel).ConnectionType = "c"
	} else {
		(&confRenderModel).ConnectionType = "s"
	}

	udp2rawConfTemplate, _ := template.New("test").Parse(string(utils.ReadFile(templatePath + "/udp2raw.config.template")))
	confFileName := strconv.Itoa(conn.LocalPort) + ".conf"
	confFilePath := configDestPath + "/udp2raw/" + confFileName
	fileInfo, err := os.Create(confFilePath)
	if err != nil {
		fmt.Println("创建文件出错:", err)
	}
	udp2rawConfTemplate.Execute(fileInfo, confRenderModel)

	serviceRenderModel := struct {
		Port string
	}{
		Port: strconv.Itoa(conn.LocalPort),
	}

	udp2rawServiceTemplate, _ := template.New("test").Parse(string(utils.ReadFile(templatePath + "/udp2raw.service.template")))
	serviceFileName := "vnci@udp2raw@" + strconv.Itoa(conn.LocalPort) + "@" + (&confRenderModel).ConnectionType + ".service"
	serviceFilePath := serviceDestPath + serviceFileName
	fileInfo, err = os.Create(serviceFilePath)
	if err != nil {
		fmt.Println("创建文件出错:", err)
	}
	udp2rawServiceTemplate.Execute(fileInfo, serviceRenderModel)

	return confFileName, serviceFileName
}

func SyncTunnel() {
	connectionManager.List()

	f := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("请输入编号[NO],输入exit退出>")
		Input, _ := f.ReadString('\n')
		Input = Input[:len(Input)-1]
		if len(Input) == 0 {
			continue
		}
		if Input == "exit" {
			break
		}
		dataId, err := strconv.Atoi(Input)
		if err != nil {
			fmt.Println("输入正确的编号")
		}
		if connectionManager.ContainsKey(dataId) {
			data, _ := connectionManager.Get(dataId)
			CreateUDP2RAWConfig(*data)
			fmt.Println("同步成功")
			break
		}
	}
}

func ToggleTunnelStatus() {
	connectionManager.List()

	f := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("请输入编号[NO],输入exit退出>")
		Input, _ := f.ReadString('\n')
		Input = Input[:len(Input)-1]
		if len(Input) == 0 {
			continue
		}
		if Input == "exit" {
			break
		}
		dataId, err := strconv.Atoi(Input)
		if err != nil {
			fmt.Println("输入正确的编号")
		}
		if connectionManager.ContainsKey(dataId) {
			data, _ := connectionManager.Get(dataId)
			newStatus := ToggleStatus(*data)
			data.Status = newStatus
			if newStatus == StatusTypeActive {
				fmt.Println("切换成功,已启用")
			} else {
				fmt.Println("切换成功,已停止")
			}
			break
		}
	}
}

func ToggleStatus(conn ConnectionItem) StatusType {
	_, serviceFileName := CreateUDP2RAWConfig(conn)
	serviceName := serviceFileName[0:strings.LastIndex(serviceFileName, ".")]

	cmd := exec.Command("systemctl", "daemon-reload")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	fmt.Printf("combined out:\n%s\n", string(out))

	if conn.Status == StatusTypeActive {
		cmd = exec.Command("systemctl", "stop", serviceName)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatalf("cmd.Run() failed with %s\n", err)
		}
		fmt.Printf("combined out:\n%s\n", string(out))
		return StatusTypeDisable
	} else {
		cmd = exec.Command("systemctl", "stop", serviceName)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatalf("cmd.Run() failed with %s\n", err)
		}
		fmt.Printf("combined out:\n%s\n", string(out))
		cmd = exec.Command("systemctl", "start", serviceName)
		out, err = cmd.CombinedOutput()
		if err != nil {
			log.Fatalf("cmd.Run() failed with %s\n", err)
		}
		fmt.Printf("combined out:\n%s\n", string(out))
		return StatusTypeActive
	}
}

// Copy the src file to dst. Any existing file will be overwritten and will not
// copy file attributes.
func Copy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

func deployUDP2Raw() {
	os.MkdirAll("/usr/local/etc/vnci/udp2raw", 0777)
	os.MkdirAll("/usr/local/bin/vnci", 0777)
	// exec.Cmd
	Copy("/usr/local/bin/vnci/udp2raw", "./libraries/udp2raw/udp2raw")

	cmd := exec.Command("chmod", "+x", "/usr/local/bin/vnci/udp2raw")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	fmt.Printf("combined out:\n%s\n", string(out))

	// os.Chmod("/usr/local/bin/vnci/udp2raw", st.Mode()|)
}

func main() {
	// connectionManager = ConnectionManager{}
	connectionManager.Initial(configPath, templatePath, serviceDestPath, configDestPath, executionPath)

	// deployUDP2Raw()

	// 把用户传递的命令行参数解析为对应变量的值
	flag.Parse()

	// loadConfig()

	// createUDP2RAWConfig(connections[1])

	if runAs == "t" {
		f := bufio.NewReader(os.Stdin) //读取输入的内容
		for {
			fmt.Println("l(list):\t列表")
			fmt.Println("n(new):\t新增")
			fmt.Println("d(delete):\t删除")
			fmt.Println("s(sync):\t同步")
			fmt.Println("t(toggle):\t切换状态")
			fmt.Print("请输入一些字符串,输入e(exit)退出>")
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
				connectionManager.List()
			case "n":
				CreateTunnel()
			case "d":
				DeleteTunnel()
			case "s":
				SyncTunnel()
			case "t":
				ToggleTunnelStatus()
			}
		}
	}

	fmt.Println("正在退出...")
	connectionManager.SaveConfig()
	// saveConfig()
	os.Exit(0)
}

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"strconv"

	"github.com/olekukonko/tablewriter"
)

var connections = make(map[int]ConnectionConfig)
var configPath = "./configs/config.json"

var runAs = *(flag.String("r", "t", "s:Service,t:Tools"))

func RunCommand(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	// 命令的错误输出和标准输出都连接到同一个管道
	stdout, err := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout

	if err != nil {
		return err
	}

	if err = cmd.Start(); err != nil {
		return err
	}
	// 从管道中实时获取输出并打印到终端
	for {
		tmp := make([]byte, 1024)
		_, err := stdout.Read(tmp)
		fmt.Print(string(tmp))
		if err != nil {
			break
		}
	}

	if err = cmd.Wait(); err != nil {
		return err
	}
	return nil
}

// Exists reports whether the named file or directory exists.
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func SaveConfig() {
	fmt.Println("正在保存...")
	text, err := json.MarshalIndent(connections, "", "    ")
	if err != nil {
		fmt.Println("保存失败(序列化):", err)

	}
	err = ioutil.WriteFile(configPath, text, 0777)
	if err != nil {
		fmt.Println("保存失败:", err)
	} else {
		fmt.Println("保存成功")
	}
}

func InitialConfig() {
	fileExist := Exists(configPath)

	if !fileExist {
		fmt.Println("未找到配置文件，正在初始化...")
		fileInfo, err := os.Create(configPath)

		if err != nil {
			fmt.Println("初始化失败:", err)
		} else {
			fileInfo.Write([]byte("{}"))
			fmt.Println("初始化成功")
		}
		fileInfo.Close()
	}
}

func LoadConnections() {

	fmt.Println("正在加载配置...")

	InitialConfig()

	fileInfo, err := os.OpenFile(configPath, os.O_RDONLY, 0600)

	if err != nil {
		fmt.Println("加载配置失败(打开配置文件):", err)
	}

	data, err := io.ReadAll(fileInfo)

	if err != nil {
		fmt.Println("加载配置失败(读取配置文件):", err)
	}

	err = json.Unmarshal(data, &connections)

	if err != nil {
		fmt.Println("加载配置失败(反序列化):", err)
	} else {
		fmt.Println("加载配置成功")
	}
}

func listTunnel() {
	tunnelSize := len(connections)

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

	// table.SetColumnColor(tablewriter.Colors{},
	// 	tablewriter.Colors{},
	// 	tablewriter.Colors{},
	// 	tablewriter.Colors{},
	// 	tablewriter.Colors{})

	for key, tunnel := range connections {

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

func checkIPAddress(ip string) bool {
	if net.ParseIP(ip) == nil {
		return false
	} else {
		return true
	}
}

func getValueFromStdin(parameterName string, optional bool, hasDefaultValue bool, defaultValue string, validFunc func(string) bool) string {
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

func createTunnel() {
	var actualType ConnectionType
	tunnelType := getValueFromStdin("通道类型(Server/Client)", false, true, "Client", nil)
	if tunnelType == string(ConnectionTypeClient) {
		actualType = ConnectionTypeClient
	} else {
		actualType = ConnectionTypeServer
	}
	localAddress := getValueFromStdin("本地监听地址", false, true, "0.0.0.0", checkIPAddress)
	localPort, _ := strconv.Atoi(getValueFromStdin("本地端口", false, false, "", nil))
	remoteAddress := getValueFromStdin("远端地址", false, false, "", nil)
	remotePort, _ := strconv.Atoi(getValueFromStdin("远端端口", false, false, "", nil))

	rawMode := getValueFromStdin("模拟方式", false, true, "faketcp", nil)
	cipherMode := getValueFromStdin("加密方式", false, true, "aes128cbc", nil)
	authMode := getValueFromStdin("认证方式", false, true, "md5", nil)

	connections[localPort] = ConnectionConfig{localAddress,
		localPort, remoteAddress, remotePort,
		rawMode, cipherMode, authMode, actualType, StatusTypeActive}
}
func deleteTunnel() {
	listTunnel()

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
		if _, ok := connections[dataId]; ok {
			delete(connections, dataId)
			fmt.Println("删除成功")
			break
		}
	}
}
func main() {

	// 把用户传递的命令行参数解析为对应变量的值
	flag.Parse()

	LoadConnections()

	if runAs == "t" {
		f := bufio.NewReader(os.Stdin) //读取输入的内容
		for {
			fmt.Println("l:\t列表")
			fmt.Println("n:\t新增")
			fmt.Println("d:\t删除")
			fmt.Print("请输入一些字符串,输入exit退出>")
			Input, _ := f.ReadString('\n') //定义一行输入的内容分隔符。
			Input = Input[:len(Input)-1]
			if len(Input) == 0 {
				continue //如果用户输入的是一个空行就让用户继续输入。
			}
			if Input == "exit" {
				break
			}
			switch Input {
			case "l":
				listTunnel()
			case "n":
				createTunnel()
			case "d":
				deleteTunnel()
			}
		}
	}

	fmt.Println("正在退出...")
	SaveConfig()
	os.Exit(0)
}

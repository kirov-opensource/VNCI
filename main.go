package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/template"

	"github.com/miekg/dns"
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

func saveConfig() {
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

func initialConfig() {
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

func readFile(path string) []byte {

	fileInfo, err := os.OpenFile(path, os.O_RDONLY, 0600)

	if err != nil {
		fmt.Println("加载配置失败(打开配置文件):", err)
	}

	data, err := io.ReadAll(fileInfo)

	if err != nil {
		fmt.Println("加载配置失败(读取配置文件):", err)
	}
	return data
}

func loadConfig() {

	fmt.Println("正在加载配置...")

	initialConfig()

	data := readFile(configPath)

	err := json.Unmarshal(data, &connections)

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

func createUDP2RAWConfig(conn ConnectionConfig) (string, string) {
	actualRemoteAddress := conn.RemoteAddress
	if !checkIPAddress(actualRemoteAddress) {
		actualRemoteAddress = getActualIP(actualRemoteAddress)
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
	}
	if conn.ConnectionType == ConnectionTypeClient {
		(&confRenderModel).ConnectionType = "c"
	} else {
		(&confRenderModel).ConnectionType = "s"
	}

	udp2rawConfTemplate, _ := template.New("test").Parse(string(readFile("./templates/udp2raw.config.template")))
	confFileName := strconv.Itoa(conn.LocalPort) + ".conf"
	confFilePath := "/usr/local/etc/vnci/udp2raw/" + confFileName
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

	udp2rawServiceTemplate, _ := template.New("test").Parse(string(readFile("./templates/udp2raw.service.template")))
	serviceFileName := "vnci@udp2raw@" + strconv.Itoa(conn.LocalPort) + "@" + (&confRenderModel).ConnectionType + ".service"
	serviceFilePath := "/etc/systemd/system/" + serviceFileName
	fileInfo, err = os.Create(serviceFilePath)
	if err != nil {
		fmt.Println("创建文件出错:", err)
	}
	udp2rawServiceTemplate.Execute(fileInfo, serviceRenderModel)

	return confFileName, serviceFileName
}

func getActualIP(address string) string {
	config, _ := dns.ClientConfigFromFile("/etc/resolv.conf")
	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(address), dns.TypeA)
	m.RecursionDesired = true
	// client 发起 DNS 请求，其中 c 为上文创建的 client，m 为构造的 DNS 报文
	// config 为从 /etc/resolv.conf 构造出来的配置
	r, _, err := c.Exchange(m, net.JoinHostPort(config.Servers[0], config.Port))
	if r == nil {
		log.Fatalf("*** error: %s\n", err.Error())
	}

	if r.Rcode != dns.RcodeSuccess {
		return "1.1.1.1"
		log.Fatalf("*** invalid answer name %s after MX query for %s\n", os.Args[1], os.Args[1])
	}

	// 如果 DNS 查询成功
	for _, a := range r.Answer {
		if dnsA, ok := a.(*dns.A); ok {
			return dnsA.A.String()
		}
		// fmt.Printf("%v\n", a)
	}
	return "1.1.1.1"
}

func syncTunnel() {
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
			createUDP2RAWConfig(connections[dataId])
			fmt.Println("同步成功")
			break
		}
	}
}

func toggleTunnelStatus() {
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
		if data, ok := connections[dataId]; ok {
			newStatus := toggleStatus(data)
			data.Status = newStatus
			connections[dataId] = data
			if newStatus == StatusTypeActive {
				fmt.Println("切换成功,已启用")
			} else {
				fmt.Println("切换成功,已停止")
			}
			break
		}
	}
}

func toggleStatus(conn ConnectionConfig) StatusType {
	_, serviceFileName := createUDP2RAWConfig(conn)
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

	deployUDP2Raw()

	// 把用户传递的命令行参数解析为对应变量的值
	flag.Parse()

	loadConfig()

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
				listTunnel()
			case "n":
				createTunnel()
			case "d":
				deleteTunnel()
			case "s":
				syncTunnel()
			case "t":
				toggleTunnelStatus()
			}
		}
	}

	fmt.Println("正在退出...")
	saveConfig()
	os.Exit(0)
}

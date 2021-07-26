package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
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
	text, err := json.Marshal(connections)
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

func ListTunnel() {
	tunnelSize := len(connections)

	if tunnelSize == 0 {
		fmt.Println("暂无数据")
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"LocalPort", "RemotePort", "RemoteAddress", "Type", "Status"})

	table.SetColumnColor(tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiBlackColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiRedColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiBlackColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgBlackColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgBlackColor})

	for _, tunnel := range connections {
		rowData := []string{
			strconv.Itoa(tunnel.LocalPort), strconv.Itoa(tunnel.RemotePort), tunnel.RemoteAddress, string(tunnel.ConnectionType), string(tunnel.Status),
		}

		table.Rich(rowData, []tablewriter.Colors{tablewriter.Colors{tablewriter.Bold, tablewriter.FgBlackColor, tablewriter.BgGreenColor}})

		//table.Append(rowData);
	}
	table.Render()
}

func main() {

	// 把用户传递的命令行参数解析为对应变量的值
	flag.Parse()

	LoadConnections()

	if runAs == "t" {
		f := bufio.NewReader(os.Stdin) //读取输入的内容
		for {
			fmt.Print("请输入一些字符串>")
			Input, _ := f.ReadString('\n') //定义一行输入的内容分隔符。
			Input = Input[:len(Input)-1]
			if len(Input) == 0 {
				continue //如果用户输入的是一个空行就让用户继续输入。
			}
			//Input = Input[:len(Input)-1]
			if Input == "exit" {
				break
			}
			switch Input {
			case "l":
				ListTunnel()
				break
			}
		}
	}

	fmt.Println("正在退出...")
	SaveConfig()
}

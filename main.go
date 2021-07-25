package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

var connections map[int]ConnectionConfig
var configPath = "./configs/config.json"

var runAs = flag.String("r", "t", "s:Service,t:Tools")

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

func Exists(name string) (bool, error) {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err != nil, err
}

func SaveConfig() {
	text, err := json.Marshal(connections)
	if err != nil {
		fmt.Println("serialize err:", err)

	}
	err = ioutil.WriteFile(configPath, text, 0777)
	if err != nil {
		fmt.Println("write file err:", err)
	}
}

func InitialConfig() {
	fileExist, err := Exists(configPath)

	if err != nil {
		fmt.Println("check file exist err:", err)
		return
	}

	if !fileExist {
		fileInfo, err := os.Create(configPath)

		if err != nil {
			fmt.Println("create file err:", err)
		}
		fileInfo.Write([]byte("{}"))
		fileInfo.Close()
	}
}

func LoadConnections() {
	InitialConfig()

	fileInfo, err := os.OpenFile(configPath, os.O_RDONLY, 0600)

	if err != nil {
		fmt.Println("read file err:", err)
	}

	data, err := io.ReadAll(fileInfo)

	if err != nil {
		fmt.Println("data read err:", err)
	}

	err = json.Unmarshal(data, &connections)

	if err != nil {
		fmt.Println("data deserialized err:", err)
	}
	//connections = deserializedData
}

func main() {

	// 把用户传递的命令行参数解析为对应变量的值
	flag.Parse()

	fmt.Println(*runAs)

	//go RunCommand("ping", "nps.futa.xyz")
	// cmd := exec.Command("libraries/udp2raw/udp2raw", "-s")
	// stdout, err := cmd.StdoutPipe()

	// if err != nil {
	// 	log.Fatal(err)
	// }

	// cmd.Start()

	// buf := bufio.NewReader(stdout)
	// // num := 0

	// for {
	// 	line, _, _ := buf.ReadLine()
	// 	// if num > 3 {
	// 	// 	os.Exit(0)
	// 	// }
	// 	// num += 1
	// 	fmt.Println(string(line))
	// }
	// go executeCommand()
	// // 监听8080/UDP
	// listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: 8080})
	// if err != nil {
	// 	fmt.Println("server listen error", err)
	// }
	// // 输出本地端的Ip
	// fmt.Printf("Local: <%s> \n", listener.LocalAddr().String())

	// // 创建一个缓冲区
	// data := make([]byte, 1024)
	// for {
	// 	n, remoteAddr, err := listener.ReadFromUDP(data)

	// 	if err != nil {
	// 		fmt.Printf("error during read: %s", err)
	// 	}
	// 	fmt.Printf("<%s> %s\n", remoteAddr, data[:n])

	// 	ipInfo := remoteAddr.IP.String() + ":" + strconv.Itoa(remoteAddr.Port)
	// 	_, err = listener.WriteToUDP([]byte(ipInfo), remoteAddr)
	// 	if err != nil {
	// 		fmt.Printf(err.Error())
	// 	}
	// }

	// fmt.Println("Hello world.")
}

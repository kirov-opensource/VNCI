package main

import (
	"fmt"
	"net"
	"strconv"
)

func main() {
	// 监听8080/UDP
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: 8080})
	if err != nil {
		fmt.Println("server listen error", err)
	}
	// 输出本地端的Ip
	fmt.Printf("Local: <%s> \n", listener.LocalAddr().String())

	// 创建一个缓冲区
	data := make([]byte, 1024)
	for {
		n, remoteAddr, err := listener.ReadFromUDP(data)

		if err != nil {
			fmt.Printf("error during read: %s", err)
		}
		fmt.Printf("<%s> %s\n", remoteAddr, data[:n])

		ipInfo := remoteAddr.IP.String() + ":" + strconv.Itoa(remoteAddr.Port)
		_, err = listener.WriteToUDP([]byte(ipInfo), remoteAddr)
		if err != nil {
			fmt.Printf(err.Error())
		}
	}
	// for {
	// 	conn, err := listener.SyscallConn()
	// 	if err != nil {
	// 		continue
	// 	}
	// 	go func(conn net.UDPConn) {
	// 		fmt.Println(conn.RemoteAddr().String())
	// 		// io.Copy(os.Stdout, conn)
	// 	}(conn)
	// }
	fmt.Println("Hello world.")
}

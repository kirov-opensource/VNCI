// package main

// import (
// 	"fmt"
// 	"net"
// )

// func main() {
// 	// 监听8080/UDP
// 	dialer := &net.Dialer{
// 		LocalAddr: &net.UDPAddr{
// 			IP:   net.ParseIP("127.0.0.1"),
// 			Port: 54444,
// 		},
// 	}
// 	conn, err := dialer.Dial("udp", "127.0.0.1:2525")
// 	if err != nil {
// 		fmt.Println("server listen error", err)
// 	}
// 	conn.Write([]byte("abc"))
// 	// for {
// 	// 	conn, err := listener.SyscallConn()
// 	// 	if err != nil {
// 	// 		continue
// 	// 	}
// 	// 	go func(conn net.UDPConn) {
// 	// 		fmt.Println(conn.RemoteAddr().String())
// 	// 		// io.Copy(os.Stdout, conn)
// 	// 	}(conn)
// 	// }
// 	// fmt.Println("Hello world.")
// }

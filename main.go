package main

import (
	"myproxy/pkg/myhttp"
	"net"
)

func main() {
	l, err := net.Listen("tcp4", ":8080")
	if err != nil {
		print(err)
		return
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			print(err.Error())
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	req, err := myhttp.NewMyReq(conn)
	if err != nil {
		print(err.Error())
		return
	}
	print("==================\n")
	print("|     REQUEST    |\n")
	print("==================\n")
	print(req.Head + "\n" + req.Body)

	proxyReq := req.BuildProxy()
	print("==================\n")
	print("|   CHANGED TO   |\n")
	print("==================\n")
	print(proxyReq.Head + "\n" + proxyReq.Body)

	newConn, err := net.Dial("tcp4", proxyReq.Addr)
	if err != nil {
		print(err.Error())
		return
	}
	_, err = newConn.Write(proxyReq.Raw)
	if err != nil {
		print(err.Error())
		return
	}
	resp, err := myhttp.NewMyResp(newConn)
	if err != nil {
		print(err.Error())
		return
	}
	err = newConn.Close()
	if err != nil {
		print(err.Error())
		return
	}
	print("==================\n")
	print("|    RESPONSE    |\n")
	print("==================\n")
	print(resp.Head + "\n" + resp.Body)
	_, err = conn.Write(resp.Raw)
	if err != nil {
		print(err.Error())
		return
	}
	err = conn.Close()
	if err != nil {
		print(err.Error())
		return
	}
}

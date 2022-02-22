package main

import (
	"crypto/rand"
	"crypto/tls"
	"math/big"
	"myproxy/pkg/myhttp"
	"net"
	"os/exec"
)

const (
	sslResponse = "HTTP/1.0 200 Connection established\r\n\r\n"
)

type Proxy struct {
	serverCfg *tls.Config
	clientCfg *tls.Config
}

var p = Proxy{}

func main() {
	cert, err := tls.LoadX509KeyPair("myproxy-ca.crt", "myproxy-ca.key")
	if err != nil {
		print(err.Error(), "\n")
		return
	}
	p.serverCfg = &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"http/1.1", "http2"},
	}

	l, err := net.Listen("tcp4", ":8080")
	if err != nil {
		print(err)
		return
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			print(err.Error(), "\n")
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	req, err := myhttp.NewMyReq(conn)
	if err != nil {
		print(err.Error(), "\n")
		//break
		return
	}
	print("==================\n")
	print("|     REQUEST    |\n")
	print("==================\n")
	print(req.Head + "\n" + req.Body)
	if req.Method == "CONNECT" {
		handleSSL(req, conn)
	} else {
		handleNoSSL(req, conn)
	}

	err = conn.Close()
	if err != nil {
		print(err.Error(), "\n")
		return
	}
}

func handleSSL(req *myhttp.MyReq, conn net.Conn) {
	newConn := &tls.Conn{}
	_, err := conn.Write([]byte(sslResponse))
	if err != nil {
		print(err.Error(), "\n")
		return
	}
	if err != nil {
		print(err.Error(), "\n")
		return
	}

	sConfig := new(tls.Config)
	*sConfig = *p.serverCfg
	sConfig.GetCertificate = func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		cConfig := new(tls.Config)
		if p.clientCfg != nil {
			*cConfig = *p.clientCfg
		}
		cConfig.ServerName = hello.ServerName
		newConn, err = tls.Dial("tcp", req.Addr, cConfig)
		if err != nil {
			//log.Println("dial", req.Addr, err)
			return nil, err
		}
		return getCertificate(hello.ServerName)
	}

	newwConn := tls.Server(conn, sConfig)
	err = newwConn.Handshake()
	if err != nil {
		print(err.Error(), "\n")
		return
	}
	defer newConn.Close()

	myhttp.PipeConns(newConn, newwConn)
}

func handleNoSSL(req *myhttp.MyReq, conn net.Conn) {
	proxyReq := req.BuildProxy()
	//print("==================\n")
	//print("|   CHANGED TO   |\n")
	//print("==================\n")
	//print(proxyReq.Head + "\n" + proxyReq.Body)

	newConn, err := net.Dial("tcp4", proxyReq.Addr)
	if err != nil {
		print(err.Error(), "\n")
		return
	}
	_, err = newConn.Write(proxyReq.Raw)
	if err != nil {
		print(err.Error(), "\n")
		return
	}
	resp, err := myhttp.NewMyResp(newConn)
	if err != nil {
		print(err.Error(), "\n")
		return
	}
	err = newConn.Close()
	if err != nil {
		print(err.Error(), "\n")
		return
	}
	//print("==================\n")
	//print("|    RESPONSE    |\n")
	//print("==================\n")
	//print(resp.Head + "\n" + resp.Body)
	_, err = conn.Write(resp.Raw)
	if err != nil {
		print(err.Error(), "\n")
		return
	}
}

func getCertificate(name string) (*tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair("./certs/"+name+".crt", "./cert.key")
	if err == nil {
		return &cert, nil
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		//print("lala")
		return nil, err
	}
	command := "touch ./certs/" + name + ".crt; ./gen_cert.sh " +
		name + " " + serialNumber.String() + " > ./certs/" + name + ".crt"
	cmd := exec.Command("bash", "-c", command)
	err = cmd.Run()
	if err != nil {
		return &tls.Certificate{}, err
	}
	cert, err = tls.LoadX509KeyPair("./certs/"+name+".crt", "cert.key")
	if err != nil {
		return nil, err
	}
	return &cert, err
}

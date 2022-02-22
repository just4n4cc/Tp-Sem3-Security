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
		//MaxVersion:   tls.VersionTLS11,
		//MaxVersion: tls.VersionTLS12,
		//MaxVersion: tls.VersionTLS13,
		NextProtos: []string{"http/1.1", "http2"},
	}
	//files, err := ioutil.ReadDir("./certs")
	//if err != nil {
	//	print(err.Error(), "\n")
	//	return
	//}
	//for _, file := range files {
	//	cert, err = tls.LoadX509KeyPair("./certs/"+file.Name(), "./cert.key")
	//	if err == nil {
	//		cfg.Certificates = append(cfg.Certificates, cert)
	//	}
	//}
	//print("certs num = ", len(cfg.Certificates), "\n")

	l, err := net.Listen("tcp4", ":8080")
	//l, err := tls.Listen("tcp4", ":8080", cfg)
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
		//break
		return
	}
	if err != nil {
		print(err.Error(), "\n")
		return
	}

	sConfig := new(tls.Config)
	*sConfig = *p.serverCfg
	sConfig.GetCertificate = func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		//cConfig := sConfig
		cConfig := new(tls.Config)
		if p.clientCfg != nil {
			*cConfig = *p.clientCfg
		}
		//if p.TLSClientConfig != nil {
		//	*cConfig = *p.TLSClientConfig
		//}
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
	//defer newwConn.Close()
	defer newConn.Close()
	//_, err = tls.Dial("tcp", req.Addr, cfg)
	//if err != nil {
	//	print(err.Error(), "\n")
	//	return
	//}
	//time.Sleep(time.Millisecond * 1000)

	//b, err := conn.Read
	//go Pipe(conn, newwConn)
	//go Pipe(conn, newConn)
	Pipe(newConn, newwConn)
	//Pipe(newwConn, conn)
	print("end of pipe")
}

func handleNoSSL(req *myhttp.MyReq, conn net.Conn) {
	//req, err := myhttp.NewMyReq(conn)
	//if err != nil {
	//	print(err.Error(), "\n")
	//	//break
	//	return
	//}
	//print("==================\n")
	//print("|     REQUEST    |\n")
	//print("==================\n")
	//print(req.Head + "\n" + req.Body)
	//if req.Method == "CONNECT" {
	//	_, err = conn.Write([]byte(sslResponse))
	//	if err != nil {
	//		print(err.Error(), "\n")
	//		//break
	//		return
	//	}
	//	req, err = myhttp.NewMyReq(conn)
	//	if err != nil {
	//		print(err.Error(), "\n")
	//		//break
	//		return
	//	}
	//}

	proxyReq := req.BuildProxy()
	//print("==================\n")
	//print("|   CHANGED TO   |\n")
	//print("==================\n")
	//print(proxyReq.Head + "\n" + proxyReq.Body)

	newConn, err := net.Dial("tcp4", proxyReq.Addr)
	if err != nil {
		print(err.Error(), "\n")
		//break
		return
	}
	_, err = newConn.Write(proxyReq.Raw)
	if err != nil {
		print(err.Error(), "\n")
		//break
		return
	}
	resp, err := myhttp.NewMyResp(newConn)
	if err != nil {
		print(err.Error(), "\n")
		//break
		return
	}
	err = newConn.Close()
	if err != nil {
		print(err.Error(), "\n")
		//break
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

	//err = conn.Close()
	//if err != nil {
	//	print(err.Error(), "\n")
	//	return
	//}
}

func getCertificate(name string) (*tls.Certificate, error) {
	//print("lala")
	cert, err := tls.LoadX509KeyPair("./certs/"+name+".crt", "./cert.key")
	if err == nil {
		return &cert, nil
	}

	print("name = ", name, "\n")
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	print("sn: ", serialNumber, "\n")
	print("sn len = ", len(serialNumber.String()), "\n")
	if err != nil {
		//print("lala")
		return nil, err
	}
	command := "touch ./certs/" + name + ".crt; ./gen_cert.sh " + name + " " + serialNumber.String() + " > ./certs/" + name + ".crt"
	//command := "touch ./certs/" + name + ".crt; ./gen_cert.sh " + name + " 343434 > ./certs/" + name + ".crt"
	cmd := exec.Command("bash", "-c", command)
	err = cmd.Run()
	if err != nil {
		print(err.Error(), "\n")
		return &tls.Certificate{}, err
	}
	cert, err = tls.LoadX509KeyPair("./certs/"+name+".crt", "cert.key")
	//cert, err := tls.LoadX509KeyPair("./certs/mail.ru.crt", "cert.key")
	if err != nil {
		print("here!!!")
	}
	print("cert - ", cert.Certificate, "\n")
	return &cert, err
}

func chanFromConn(conn net.Conn) chan []byte {
	c := make(chan []byte)

	go func() {
		b := make([]byte, 1024)

		for {
			n, err := conn.Read(b)
			//print(string(b))
			if n > 0 {
				res := make([]byte, n)
				// Copy the buffer so it doesn't get changed while read by the recipient.
				copy(res, b[:n])
				c <- res
			}
			if err != nil {
				c <- nil
				break
			}
		}
	}()

	return c
}

func Pipe(conn1 net.Conn, conn2 net.Conn) {
	chan1 := chanFromConn(conn1)
	chan2 := chanFromConn(conn2)

	for {
		select {
		case b1 := <-chan1:
			if b1 == nil {
				//print("because b1 is nil\n")
				return
			} else {
				_, err := conn2.Write(b1)
				//fmt.Printf("writing to conn2: %q\n", string(b1))
				if err != nil {
					print("because conn2 write failure\n")
					return
				}
			}
		case b2 := <-chan2:
			if b2 == nil {
				//print("because b2 is nil\n")
				return
			} else {
				_, err := conn1.Write(b2)
				//fmt.Printf("writing to conn1: %q\n", string(b2))
				if err != nil {
					print("because conn1 write failure\n")
					return
				}
			}
		}
	}
}

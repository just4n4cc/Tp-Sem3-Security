package main

import (
	"bufio"
	"io"
	"net"
	"strconv"
	"strings"
)

func readRequest(conn net.Conn) (string, error) {
	r := bufio.NewReader(conn)
	data := ""
	bytesNum := -1
	chunked := false
	for {
		tmp, err := r.ReadString('\n')
		if err != nil {
			return "", err
		}
		if strings.HasPrefix(tmp, "Content-Length:") {
			bytesNum, err = strconv.Atoi(strings.Trim(strings.Split(tmp, " ")[1], "\r\n"))
			if err != nil {
				return "", err
			}
		}
		if strings.HasPrefix(tmp, "Transfer-Encoding: chunked") {
			chunked = true
		}
		data += tmp
		if tmp == "\r\n" {
			break
		}
	}

	if bytesNum < 0 && !chunked {
		return data, nil
	}

	if chunked {
		for {
			tmp, err := r.ReadString('\n')
			if err != nil {
				return "", err
			}
			data += tmp

			bytesNum, err := strconv.ParseInt(tmp[:len(tmp)-2], 16, 64)
			if err != nil {
				return "", err
			}
			buf := make([]byte, bytesNum)
			_, err = io.ReadFull(r, buf)
			if err != nil {
				return "", err
			}
			data += string(buf)

			tmp, err = r.ReadString('\n')
			if err != nil {
				return "", err
			}
			data += tmp

			if bytesNum == 0 {
				return data, nil
			}
		}
	}

	body := make([]byte, bytesNum)
	_, err := io.ReadFull(r, body)
	if err != nil {
		return "", err
	}
	data += string(body)
	return data, nil
}

func parseRequest(data string) (string, string, string) {
	req := ""
	host := ""
	rest := ""
	rows := strings.Split(data, "\r\n")
	req = rows[0] + "\r\n"
	for _, row := range rows[1:] {
		if strings.HasPrefix(row, "Host:") {
			host = row + "\r\n"
			continue
		}
		if strings.HasPrefix(row, "Proxy-Connection:") {
			continue
		}
		rest += row + "\r\n"
	}

	return req, host, rest
}

func rebuildRequest(req string, host string) string {
	hostVal := strings.Trim(strings.Split(host, " ")[1], "\r\n")
	reqArr := strings.Split(req, " ")
	i := strings.Index(reqArr[1], hostVal)
	reqArr[1] = reqArr[1][i+len(hostVal):]
	req = ""
	for _, el := range reqArr {
		req += el + " "
	}
	return req[:len(req)-1]
}

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

func getAddress(host string) string {
	hostVal := strings.Trim(strings.Split(host, " ")[1], "\r\n")
	i := strings.Index(hostVal, ":")
	if i == -1 {
		return hostVal + ":80"
	}
	return hostVal
}

func getResponse(req string, host string, rest string) ([]byte, error) {
	conn, err := net.Dial("tcp4", getAddress(host))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	_, err = conn.Write([]byte(req + host + rest))
	if err != nil {
		return nil, err
	}

	data, err := readRequest(conn)
	if err != nil {
		return nil, err
	}
	return []byte(data), nil
}

func handleConn(conn net.Conn) {
	data, err := readRequest(conn)
	//print("==================")
	//print("|     REQUEST    |")
	//print("==================\n")
	//print(data)

	req, host, rest := parseRequest(data)
	if err != nil {
		print(err.Error())
		return
	}
	req = rebuildRequest(req, host)
	//print("==================")
	//print("|   CHANGED TO   |")
	//print("==================\n")
	//print(req + host + rest)

	resp, err := getResponse(req, host, rest)
	if err != nil {
		print(err.Error())
		return
	}
	//print("==================")
	//print("|    RESPONSE    |")
	//print("==================\n")
	//print(string(resp))
	_, err = conn.Write(resp)
	if err != nil {
		print(err.Error())
		return
	}
	conn.Close()
}

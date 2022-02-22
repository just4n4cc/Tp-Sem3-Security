package myhttp

import (
	"bufio"
	"errors"
	"io"
	"net"
	"strconv"
	"strings"
)

func readFromConn(conn net.Conn) (string, error) {
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

			chunkSize, err := strconv.ParseInt(tmp[:len(tmp)-2], 16, 64)
			if err != nil {
				return "", err
			}
			buf := make([]byte, chunkSize)
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

			if chunkSize == 0 {
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

func headBodySplit(raw string) (string, string, error) {
	bodyStart := strings.Index(raw, "\r\n\r\n") + len("\r\n")
	if bodyStart == -1 {
		return "", "", errors.New("error while parsing raw http")
	}
	head := raw[:bodyStart]
	body := raw[bodyStart+len("\r\n"):]
	return head, body, nil
}

func chanFromConn(conn net.Conn) chan []byte {
	c := make(chan []byte)

	go func() {
		b := make([]byte, 1024)

		for {
			n, err := conn.Read(b)
			if n > 0 {
				res := make([]byte, n)
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

func PipeConns(conn1 net.Conn, conn2 net.Conn) {
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
				if err != nil {
					return
				}
			}
		case b2 := <-chan2:
			if b2 == nil {
				//print("because b2 is nil\n")
				return
			} else {
				_, err := conn1.Write(b2)
				if err != nil {
					return
				}
			}
		}
	}
}

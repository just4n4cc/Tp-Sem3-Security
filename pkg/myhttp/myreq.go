package myhttp

import (
	"net"
	"strings"
)

type MyReq struct {
	Method string
	Url    string
	Host   string
	Addr   string
	Head   string
	Body   string
	Raw    []byte
}

func NewMyReq(conn net.Conn) (*MyReq, error) {
	raw, err := readFromConn(conn)
	if err != nil {
		return nil, err
	}
	rows := strings.Split(raw, "\r\n")
	reqRow := strings.Split(rows[0], " ")

	// Method and Url
	method := reqRow[0]
	url := reqRow[1]

	// Host
	host := ""
	for _, row := range rows {
		if strings.HasPrefix(row, "Host:") {
			host = row[6:]
			break
		}
	}

	// Head and Body
	head, body, err := headBodySplit(raw)
	if err != nil {
		return nil, err
	}

	// Addr
	addr := host
	if strings.Index(host, ":") == -1 {
		addr += ":80"
	}
	return &MyReq{
		Method: method,
		Url:    url,
		Host:   host,
		Addr:   addr,
		Head:   head,
		Body:   body,
		Raw:    []byte(raw),
	}, nil
}

func (mr *MyReq) BuildProxy() *MyReq {
	// - Proxy-Connection
	head := mr.Head
	pcstart := strings.Index(head, "Proxy-Connection:")
	if pcstart != -1 {
		pcend := strings.Index(head[pcstart:], "\r\n") + pcstart + len("\r\n")
		head = head[:pcstart] + head[pcend:]
	}

	// Changing url
	hstart := strings.Index(mr.Url, mr.Host)
	url := mr.Url[hstart+len(mr.Host):]
	uend := strings.Index(head, "\r\n")
	urow := head[:uend]
	uArr := strings.Split(urow, " ")
	uArr[1] = url
	urow = ""
	for _, el := range uArr {
		urow += el + " "
	}
	urow = urow[:len(urow)-1]
	head = urow + head[uend:]
	return &MyReq{
		Method: mr.Method,
		Url:    url,
		Host:   mr.Host,
		Addr:   mr.Addr,
		Head:   head,
		Body:   mr.Body,
		Raw:    []byte(head + "\r\n" + mr.Body),
	}
}

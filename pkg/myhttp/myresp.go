package myhttp

import (
	"net"
	"strconv"
	"strings"
)

type MyResp struct {
	Status int
	Head   string
	Body   string
	Raw    []byte
}

func NewMyResp(conn net.Conn) (*MyResp, error) {
	raw, err := readFromConn(conn)
	if err != nil {
		return nil, err
	}
	rows := strings.Split(raw, "\r\n")
	status, err := strconv.Atoi(strings.Split(rows[0], " ")[1])
	if err != nil {
		return nil, err
	}
	head, body, err := headBodySplit(raw)
	if err != nil {
		return nil, err
	}
	return &MyResp{
		Status: status,
		Head:   head,
		Body:   body,
		Raw:    []byte(raw),
	}, nil
}

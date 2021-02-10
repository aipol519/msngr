package serverapi

import (
	"encoding/json"
	"log"
	"net"
	"time"

	"github.com/gorilla/websocket"
)

const (
	PingPeriod = 10 * time.Second

	Port     = ":80"
	Headport = ":8080"
)

type Client struct {
	Login      string
	Name       string
	Connection *websocket.Conn
}

type Node struct {
	Connection  *websocket.Conn
	ClientsList []string
}

type AddressRequest struct {
	Type    string
	Address string
}

type InfoRequest struct {
	Type string
	Info string
}

type AddressIdRequest struct {
	Type     string
	Address  string
	ClientId string
}

type MessageRequest struct {
	Type       string
	SenderId   string
	ReceiverId string
	Message    string
}

func (inr AddressRequest) ToJSON() []byte {
	req, err := json.Marshal(inr)
	if err != nil {
		return nil
	}
	return req
}

func (inr InfoRequest) ToJSON() []byte {
	req, err := json.Marshal(inr)
	if err != nil {
		return nil
	}
	return req
}

func (inr AddressIdRequest) ToJSON() []byte {
	req, err := json.Marshal(inr)
	if err != nil {
		return nil
	}
	return req
}

func (inr MessageRequest) ToJSON() []byte {
	req, err := json.Marshal(inr)
	if err != nil {
		return nil
	}
	return req
}

func Call(params map[string]interface{}, f func(map[string]interface{}) bool) bool {
	return f(params)
}

func GetLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

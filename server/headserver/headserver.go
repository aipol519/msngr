package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"

	"../serverapi"

	"github.com/gorilla/websocket"
)

var funcs = map[string]func(map[string]interface{}) bool{
	"addNodeAddress":      addNodeAddress,
	"addToAddrTable":      addToAddrTable,
	"removeFromAddrTable": removeFromAddrTable,
	"sendMessage":         sendMessage,
}

var addrtable map[string]*serverapi.Node

var addr = flag.String("addr", "192.168.50.201"+serverapi.Headport, "http service address")

var upgrader = websocket.Upgrader{}

func headserverlisten(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	for {
		mt, message, err := c.ReadMessage()

		if err != nil {
			log.Println("read:", err)
			break
		}

		var result map[string]interface{}
		json.Unmarshal(message, &result)
		var params map[string]interface{}
		params = make(map[string]interface{})
		for k, v := range result {
			params[k] = v
		}
		delete(params, "Type")

		log.Printf("recv: %s", message)

		if funcs[result["Type"].(string)] != nil {
			serverapi.Call(params, funcs[result["Type"].(string)])
		}

		err = c.WriteMessage(mt, message)

		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func connectToNode(headaddr string) (*websocket.Conn, error) {
	u := url.URL{Scheme: "ws", Host: headaddr, Path: "/"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Println("can't connect to node")
		log.Println("dial:", err)
	}

	return c, err
}

func addNodeAddress(params map[string]interface{}) bool {
	addr := params["Address"].(string)

	if nodeconn, err := connectToNode(addr); err == nil {
		addrtable[addr] = &serverapi.Node{nodeconn, make([]string, 0)}
	}

	log.Println("successfully added " + addr + " node to address table")

	return true
}

func addToAddrTable(params map[string]interface{}) bool {
	addr := params["Address"].(string)
	clientid := params["ClientId"].(string)

	addrtable[addr].ClientsList = append(addrtable[addr].ClientsList, clientid)

	log.Println("successfully added client " + addr + " / " + clientid + " to address table")

	return true
}

func removeFromAddrTable(params map[string]interface{}) bool {
	addr := params["Address"].(string)
	clientid := params["ClientId"].(string)

	for i := 0; i < len(addrtable[addr].ClientsList); {
		if addrtable[addr].ClientsList[i] == clientid {
			addrtable[addr].ClientsList = append(addrtable[addr].ClientsList[:i], addrtable[addr].ClientsList[i+1:]...)
		}
	}

	return true
}

func sendMessage(params map[string]interface{}) bool {
	senderid := params["SenderId"].(string)
	receiverid := params["ReceiverId"].(string)
	message := params["Message"].(string)

	for _, v := range addrtable {
		for _, clientid := range v.ClientsList {
			if clientid == receiverid {
				err := v.Connection.WriteMessage(websocket.BinaryMessage, serverapi.MessageRequest{"sendMessage", senderid, receiverid, message}.ToJSON())
				if err != nil {
					log.Println("write:", err)
				}

				return true
			}
		}
	}

	return false
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	addrtable = make(map[string]*serverapi.Node)

	http.HandleFunc("/", headserverlisten)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

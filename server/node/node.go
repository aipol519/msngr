package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"../serverapi"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
)

var funcs = map[string]func(map[string]interface{}) bool{
	"addNewClient":    addNewClient,
	"sendMessage":     sendMessage,
	"registerClient":  registerClient,
	"authorizeClient": authorizeClient,
}

var clients = map[string]serverapi.Client{}

var addr = flag.String("addr", serverapi.GetLocalIP()+serverapi.Port, "http service address")
var headaddr = flag.String("headaddr", "192.168.50.201"+serverapi.Headport, "http service address")

var activeconn = &websocket.Conn{}
var headserverconn = &websocket.Conn{}

var upgrader = websocket.Upgrader{}

func echo(w http.ResponseWriter, r *http.Request) {
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

		activeconn = c

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

func connectToHeadServer(headaddr *string) (*websocket.Conn, error) {
	u := url.URL{Scheme: "ws", Host: *headaddr, Path: "/headserverlisten"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Println("can't connect to head server; working in single mode")
		log.Println("dial:", err)
	}

	return c, err
}

func pingClient(c *websocket.Conn) {
	for {
		if c != nil {
			err := c.WriteMessage(websocket.PingMessage, []byte("ping"))
			if err != nil {
				removeClient(c)
				return
			}
			time.Sleep(serverapi.PingPeriod / 2)
		}
	}
}

func registerClient(params map[string]interface{}) bool {
	clientmap := make(map[string]string)
	for k, v := range params["Client"].(map[string]interface{}) {
		clientmap[k] = v.(string)
	}

	db, err := sql.Open("mysql", "aipol:5210egor@tcp(192.168.50.201:3306)/cn_coursework_db")
	if err != nil {
		log.Println("db:", err)
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	result, _ := db.Query("select * from users where login = ?;", clientmap["Login"])
	rows := make(map[string][]string, 0)
	for result.Next() {
		var id, login, passwordHash, name string
		if err := result.Scan(&id, &login, &passwordHash, &name); err != nil {
			log.Println("db:", err)
		}
		rows[id] = append(rows[id], login, passwordHash, name)
	}

	if len(rows) == 0 {
		db.Exec("insert into users(login, password_hash, name) values (?, ?, ?);", clientmap["Login"], clientmap["PasswordHash"], clientmap["Name"])
		addNewClient(params)
		if err = activeconn.WriteMessage(websocket.BinaryMessage, serverapi.InfoRequest{"infoRequest", "success"}.ToJSON()); err != nil {
			log.Println("write:", err)
		}
		log.Println("user " + clientmap["Login"] + " successfully registered")
	} else {
		if err = activeconn.WriteMessage(websocket.BinaryMessage, serverapi.InfoRequest{"infoRequest", "error"}.ToJSON()); err != nil {
			log.Println("write:", err)
		}
		activeconn.Close()
	}

	db.Close()
	return true
}

func authorizeClient(params map[string]interface{}) bool {
	clientmap := make(map[string]string)
	for k, v := range params["Client"].(map[string]interface{}) {
		clientmap[k] = v.(string)
	}

	db, err := sql.Open("mysql", "aipol:5210egor@tcp(192.168.50.201:3306)/cn_coursework_db")
	if err != nil {
		log.Println("db:", err)
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	result, _ := db.Query("select * from users where login = ?;", clientmap["Login"])
	rows := make(map[string][]string, 0)
	for result.Next() {
		var id, login, passwordHash, name string
		if err := result.Scan(&id, &login, &passwordHash, &name); err != nil {
			log.Println("db:", err)
		}
		rows[id] = append(rows[id], login, passwordHash, name)
	}

	if len(rows) != 0 {
		for _, v := range rows {
			if v[1] == clientmap["PasswordHash"] {
				addNewClient(params)
				if err = activeconn.WriteMessage(websocket.BinaryMessage, serverapi.InfoRequest{"infoRequest", "success"}.ToJSON()); err != nil {
					log.Println("write:", err)
				}
				log.Println("user " + clientmap["Login"] + " successfully authorized")
				break
			}
		}
	} else {
		if err = activeconn.WriteMessage(websocket.BinaryMessage, serverapi.InfoRequest{"infoRequest", "error"}.ToJSON()); err != nil {
			log.Println("write:", err)
		}
		activeconn.Close()
	}

	db.Close()
	return true
}

func addNewClient(params map[string]interface{}) bool {
	clientmap := make(map[string]string)
	for k, v := range params["Client"].(map[string]interface{}) {
		clientmap[k] = v.(string)
	}

	clients[clientmap["Login"]] = serverapi.Client{clientmap["Login"], clientmap["Name"], activeconn}

	if headserverconn != nil {
		err := headserverconn.WriteMessage(websocket.BinaryMessage, serverapi.AddressIdRequest{"addToAddrTable", serverapi.GetLocalIP() + serverapi.Port, clientmap["Login"]}.ToJSON())
		if err != nil {
			log.Println("write:", err)
		}
	}

	//go pingClient(activeconn)

	return true
}

func removeClient(c *websocket.Conn) {
	for k, v := range clients {
		if v.Connection.LocalAddr() == c.LocalAddr() {
			if &headserverconn != nil {
				err := headserverconn.WriteMessage(websocket.BinaryMessage, serverapi.AddressIdRequest{"removeFromAddrTable", serverapi.GetLocalIP() + serverapi.Port, v.Login}.ToJSON())
				if err != nil {
					log.Println("write:", err)
				}
			}

			delete(clients, k)
		}
	}
}

func isClientConnected(clientid string) bool {
	if _, ok := clients[clientid]; ok {
		return true
	} else {
		return false
	}
}

func findClient(clientid string) *serverapi.Client {
	for k, v := range clients {
		if k == clientid {
			return &v
		}
	}

	return nil
}

func sendMessage(params map[string]interface{}) bool {
	senderid := params["SenderId"].(string)
	receiverid := params["ReceiverId"].(string)
	message := params["Message"].(string)

	if isClientConnected(receiverid) {
		if reciever := findClient(receiverid); reciever != nil {
			if err := reciever.Connection.WriteMessage(websocket.BinaryMessage, serverapi.MessageRequest{"receiveMessage", senderid, receiverid, message}.ToJSON()); err != nil {
				log.Println("write:", err)
				return false
			}
		}
	} else {
		if headserverconn != nil {
			if err := headserverconn.WriteMessage(websocket.BinaryMessage, serverapi.MessageRequest{"sendMessage", senderid, receiverid, message}.ToJSON()); err != nil {
				log.Println("write:", err)
				return false
			}
		}
	}

	return true
}

func main() {
	clients = make(map[string]serverapi.Client)

	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	headserverconn, _ = connectToHeadServer(headaddr)
	if headserverconn != nil {
		err := headserverconn.WriteMessage(websocket.BinaryMessage, serverapi.AddressRequest{"addNodeAddress", serverapi.GetLocalIP() + serverapi.Port}.ToJSON())
		if err != nil {
			log.Println("write:", err)
		}
	}

	http.HandleFunc("/", echo)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

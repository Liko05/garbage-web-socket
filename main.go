package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type client struct {
	id int
	ip string
}

var clients = make(map[*websocket.Conn]client)
var blacklist = make(map[string]bool)
var users = 0

func chatEndpoint(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
	}

	fmt.Println("Client connected from ip: " + r.RemoteAddr)

	if blacklist[r.RemoteAddr] {
		println("Blacklisted ip: " + r.RemoteAddr + " tried to connect")
		ws.Close()
		return
	}

	clients[ws] = client{users, r.RemoteAddr}
	users++

	println(clients[ws].id)

	err = ws.WriteMessage(1, []byte("{ \"sender\": \""+"Server"+"\", \"message\": \""+"Connected to chat room"+"\" }"))
	if err != nil {
		log.Println(err)
	}

	reader(ws)
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello world")
}

func setupRoutes() {
	http.HandleFunc("/chat", chatEndpoint)
	http.HandleFunc("/", homePage)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func reader(conn *websocket.Conn) {
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		var user = clients[conn]
		fmt.Println("User: " + user.ip + " Message: " + string(p))
		var username = "User " + strconv.Itoa(user.id)
		if strings.Contains(string(p), "python") {
			blacklist[user.ip] = true
			conn.Close()
			delete(clients, conn)
			println("Blacklisted ip: " + user.ip + " tried to send python")
			broadcast(username+" has been kicked. Bro prolly thought talking about python is something he can do freely. 🙄", "Server")
			return
		}
		broadcast(string(p), username)
	}
}

func broadcast(message string, username string) {
	var jsonMessage = "{ \"sender\": \"" + username + "\", \"message\": \"" + message + "\" }"
	for client := range clients {
		err := client.WriteMessage(1, []byte(jsonMessage))
		if err != nil {
			log.Println(err)
			client.Close()
			delete(clients, client)
		}
	}
}

func main() {
	fmt.Println("Starting websocket server")
	setupRoutes()
	log.Fatal(http.ListenAndServe(":1337", nil))
}

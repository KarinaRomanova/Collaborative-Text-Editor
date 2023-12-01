package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan []byte)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	clients[ws] = true

	for {

		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Println(err)
			delete(clients, ws)
			break
		}

		broadcast <- msg
	}
}

func handleMessages() {
	for {

		msg := <-broadcast

		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Println(err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "ov.html")
}

func main() {

	go handleMessages()

	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/ws", handleConnections)

	port := ":8000"
	fmt.Println("Сервер запущен на порту", port)

	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}

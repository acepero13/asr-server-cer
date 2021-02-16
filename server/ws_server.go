package server

import (
	"bytes"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

// connected clients

type WsConnection struct {
	clients        map[*websocket.Conn]bool
	MessageChannel chan WsClient
}

type WsClient struct {
	Connected       bool
	WebsocketClient *websocket.Conn
	Msg             bytes.Buffer
}

func NewWsMessage(conn *websocket.Conn) *WsClient {
	return &WsClient{
		Connected:       true,
		WebsocketClient: conn,
		Msg:             bytes.Buffer{},
	}
}

func (conn *WsConnection) resets(ws *websocket.Conn) {
	for c := range conn.clients {
		conn.clients[c] = false
	}
	conn.clients[ws] = true
}

func newConnection() *WsConnection {
	return &WsConnection{
		clients:        make(map[*websocket.Conn]bool),
		MessageChannel: make(chan WsClient),
	}
}

var connection = newConnection()

func StartServer(onNewConnection func(*WsConnection, *WsClient)) {
	http.HandleFunc("/ws", handleConnections(onNewConnection))

	go messageHandler()

	log.Println("http server started on :2701")
	err := http.ListenAndServe(":2701", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func messageHandler() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-connection.MessageChannel

		temp := make([]byte, msg.Msg.Len())
		_, errRead := msg.Msg.Read(temp)
		err := msg.WebsocketClient.WriteJSON(temp)
		if errRead != nil {
			fmt.Println("Error reading buffer")
		}
		if err != nil {
			log.Printf("error: %v", err)
			err := msg.WebsocketClient.Close()
			if err != nil {
				fmt.Println("Couldn't close connection with client")
			}
		}

	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleConnections(onNewConnection func(*WsConnection, *WsClient)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		onNewConnection(connection, NewWsMessage(ws))

		if err != nil {
			log.Fatal(err)
		}

		defer func(ws *websocket.Conn) {
			err := ws.Close()
			if err != nil {
				fmt.Println("Error while closing")
			}
		}(ws)
		connection.resets(ws)
	}
}

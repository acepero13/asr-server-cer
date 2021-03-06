package cerence

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"time"
)

var (
	//CRITIC A critical error occurred in the websocket
	CRITIC = Severity{"critic"}
	//UNIMPORTANT A non critical error happened. In general this kind of error should be logged only.
	UNIMPORTANT = Severity{"not important"}
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

//Severity Indicates the error severity
type Severity struct {
	Level string
}

//SError Indicates an error occurred.
//It contains the error itself and the level of severity
type SError struct {
	Err   error
	Level Severity
}

//ClientCallbacks Api for callback functions for ws events
//OnMessage New message arrived from client
//Write sync write to send to client
type ClientCallbacks interface {
	OnMessage(conn *websocket.Conn, msg []byte)
	Write(conn *websocket.Conn, msg []byte)
	OnError(err SError)
	OnClose()
}

//WebSocketApp Higher level APIS for ws connection. Similar to js
func WebSocketApp(port int, tls bool, onNewClient func(conn *websocket.Conn) *Client) {

	http.HandleFunc("/ws", handleConnections(onNewClient))
	log.Println("http server started on :" + strconv.Itoa(port))
	dieIfErr(listenAndServeTo(port, tls), "Cannot serve")
}

func listenAndServeTo(port int, useTLS bool) error {
	if useTLS {
		return http.ListenAndServeTLS(":"+strconv.Itoa(port), "configs/server-certificate.pem", "configs/server-key.pem", nil)
	}
	return http.ListenAndServe(":"+strconv.Itoa(port), nil)
}

//DisconnectClient Closes the connection with the specified client
func DisconnectClient(conn *websocket.Conn) error {
	defer func() { // Deregister from connected clients list
		if _, ok := ConnectedClients.clients[conn]; ok {
			ConnectedClients.clients[conn] = false
		}
	}()
	return conn.Close()

}

func handleConnections(onNewClient func(conn *websocket.Conn) *Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)

		callback := onNewClient(ws)

		go receiveFromWs(ws, callback)

		possibleCritic(err, *callback)

	}
}

func receiveFromWs(ws *websocket.Conn, callbacks *Client) {

	var queue [][]byte
	for {
		_, data, err := ws.ReadMessage()
		possibleNotImportant(err, *callbacks)
		actualQueue := enqueue(queue, data)

		msg, actualQueue := dequeue(actualQueue)
		queue = actualQueue
		time.Sleep(30 * time.Millisecond)
		wsErr := notify(ws, callbacks, err, msg)
		if wsErr != nil {
			break
		}
	}
}

func notify(ws *websocket.Conn, callbacks *Client, err error, msg []byte) error {
	if err == nil {
		callbacks.OnMessage(ws, msg)
		return nil
	}
	callbacks.OnClose()
	logIfErr(DisconnectClient(ws), "Error closing ws")
	return errors.New("error in ws")

}

func possibleNotImportant(err error, callbacks Client) {
	if err != nil {
		callbacks.OnError(SError{err, UNIMPORTANT})
	}
}

func possibleCritic(err error, callbacks Client) {
	if err != nil {
		callbacks.OnError(SError{err, CRITIC})
	}
}

func dieIfErr(err error, msg string) {
	if err != nil {
		fmt.Println(msg+" err: ", err.Error())
	}
}

func enqueue(queue [][]byte, element []byte) [][]byte {
	queue = append(queue, element) // Simply append to enqueue.
	return queue
}

func dequeue(queue [][]byte) ([]byte, [][]byte) {
	element := queue[0]       // The first element is the one to be dequeued.
	return element, queue[1:] // Slice off the element once it is dequeued.
}

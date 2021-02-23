package cerence

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"time"
)

var (
	CRITIC      = Severity{"critic"}
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

type ClientCallbacks interface {
	OnMessage(conn *websocket.Conn, msg []byte)
	Write(conn *websocket.Conn, msg []byte)
	OnError(err SError)
	OnClose()
}

//WebSocketApp Higher level APIS for ws connection. Similar to js
func WebSocketApp(port int, onNewClient func(conn *websocket.Conn) Client) {

	http.HandleFunc("/ws", handleConnections(onNewClient))

	log.Println("http server started on :2701")
	err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
	dieIfErr(err, "Cannot serve")
}

func handleConnections(onNewClient func(conn *websocket.Conn) Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)

		callback := onNewClient(ws)

		go receiveFromWs(ws, callback)

		possibleCritic(err, callback)

		defer func(ws *websocket.Conn) {
			//errClose := ws.Close() // TODO: See why this happens
			callback.OnClose()
			possibleNotImportant(nil, callback)
		}(ws)

	}
}

func receiveFromWs(ws *websocket.Conn, callbacks Client) {
	var queue [][]byte
	for {
		// TODO: Break on close
		_, data, err := ws.ReadMessage()
		possibleNotImportant(err, callbacks)
		actualQueue := enqueue(queue, data)

		msg, actualQueue := dequeue(actualQueue)
		queue = actualQueue
		time.Sleep(30 * time.Millisecond)
		callbacks.OnMessage(ws, msg)
	}
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

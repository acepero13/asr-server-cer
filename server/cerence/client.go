package cerence

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	config3 "github.com/acepero13/asr_server/server/config"
	"github.com/acepero13/asr_server/server/receiver"
	"github.com/acepero13/cloud-client-go/http_v2_client"
	"github.com/gorilla/websocket"
)

//Client Information related to one client
type Client struct {
	wsClient   *websocket.Conn
	sender     *Sender // client that connects to cerence server
	writeMutex *sync.Mutex
}

//ConnectedClients List for connected clients. Not used at the moment
var ConnectedClients = clients{
	clients: make(map[*websocket.Conn]bool),
}

type clients struct {
	clients     map[*websocket.Conn]bool
	clientMutex sync.Mutex
}

//OnConnected When a new ws client connects, it creates and returns a client which is ready to connect to cerence server
func OnConnected(conn *websocket.Conn) *Client {
	ConnectedClients.clientMutex.Lock()
	client := newClient(conn)
	ConnectedClients.clients[conn] = true
	ConnectedClients.clientMutex.Unlock()
	fmt.Printf("New client connected. We have: %d connected client(s)", getNumConnectedClients())
	return client
}

//Write Synchronized method that sends back to the client information. Avoids concurrent socket write
func (c *Client) Write(conn *websocket.Conn, msg []byte) {
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()
	logIfErr(conn.WriteJSON(msg), "Error sending recognition finished")
}

//OnError When an error occurs related with the websocket communication
func (c *Client) OnError(err SError) {
	if err.Level == CRITIC {
		fmt.Printf("Critical error. Disconnecting client %s\n", err.Err.Error())
		logIfErr(DisconnectClient(c.wsClient), "Error closing ws client")
	} else if err.Level == UNIMPORTANT {
		fmt.Printf("Critical error %s\n", err.Err.Error())
	}
}

//OnClose When the websocket connection is closed
func (c *Client) OnClose() {
	c.sender.Close()
}

//OnMessage When a new message arrives from the client. It contains the msg as a byte array
func (c *Client) OnMessage(conn *websocket.Conn, msg []byte) {
	if !c.sender.IsConnected() {
		logIfErr(c.sender.Connect(), "Error connecting to cerence server")
		go startReceiving(c)
	}
	c.sender = receiver.SendWithClient(c.sender, msg).(*Sender)

	if c.sender.GetState().IsFinished {
		c.Write(conn, []byte(`{"recognition_finished": "1"}`))
		c.reconnectToCerence()

	}
}

func startReceiving(cerenceCli *Client) {
	client := cerenceCli.sender.cerenceClient
	go client.Receive()
	for chunk := range client.GetReceivedChunkChannel() {
		if string(chunk.Body.Bytes()) == "Close" {
			fmt.Println("Please close connection")
			time.Sleep(30 * time.Millisecond)
			break
		}
		fmt.Println(chunk.Body.String())
		go process(chunk.Body, cerenceCli)

	}
	fmt.Println("ENDED FOR")
}

func process(msg bytes.Buffer, cli *Client) {
	var singleResult *receiver.AsrResult

	asrResult, errDecod := receiver.NewAsrResultFrom(msg.Bytes())
	if asrResult == nil {
		return
	}
	singleResult = asrResult.GetAtMost(1)

	toSend, errEncod := singleResult.ToBytes()
	if errDecod != nil || errEncod != nil || singleResult == nil {
		return
	}
	cli.Write(cli.wsClient, toSend)

}

func getNumConnectedClients() int {
	var counter = 0
	for _, isConnected := range ConnectedClients.clients {
		if isConnected {
			counter++
		}
	}
	return counter
}

func newClient(conn *websocket.Conn) *Client {
	return &Client{
		wsClient:   conn,
		sender:     newSender(conn),
		writeMutex: &sync.Mutex{},
	}

}

func (c *Client) reconnectToCerence() {
	c.sender.Close()
	c.sender = nil
	time.Sleep(30 * time.Millisecond)
	c.sender = newSender(c.wsClient)
}

func newSender(conn *websocket.Conn) *Sender {
	config, err := config3.GiveMeAConfig()
	disconnectIfErr(err, conn)
	client := http_v2_client.NewHttpV2Client(config.Host,
		config.Port,
		http_v2_client.WithProtocol(config.Protocol),
		http_v2_client.WithPath(config.Path),
		http_v2_client.WithBoundary(config.GetBoundary()),
	)
	return NewSender(client, config)

}

func disconnectIfErr(err error, conn *websocket.Conn) {
	if err != nil {
		fmt.Println("Cannot retrieve a config right now. We will disconnect the client")
		logIfErr(DisconnectClient(conn), "Cannot close connection with client")
	}
}

func logIfErr(err error, msg string) {
	if err != nil {
		fmt.Println(msg + " .Reason: " + err.Error())
	}
}

package cerence

import (
	"bytes"
	"fmt"
	"github.com/acepero13/cloud-client-go/http_v2_client"
	config3 "github.com/alvaro/asr_server/server/config"
	"github.com/alvaro/asr_server/server/receiver"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

//Client Information related to one client
type Client struct {
	wsClient   *websocket.Conn
	sender     *Sender
	writeMutex *sync.Mutex
}

type clients struct {
	clients     []*Client
	clientMutex sync.Mutex
}

//ConnectedClients List for connected clients. Not used at the moment
var ConnectedClients = clients{
	clients: []*Client{},
}

//OnConnected When a new ws client connects, it returns a client which is ready to connect to cerence server
func OnConnected(conn *websocket.Conn) Client {
	ConnectedClients.clientMutex.Lock()
	client := newClient(conn)
	ConnectedClients.clients = append(ConnectedClients.clients, client)
	ConnectedClients.clientMutex.Unlock()
	return *client
}

func newClient(conn *websocket.Conn) *Client {
	config, err := config3.GiveMeAConfig()
	disconnectIfErr(err, conn)

	logIfErr(err, "Problem getting the config")
	client := http_v2_client.NewHttpV2Client(
		config.Host,
		config.Port,
		http_v2_client.WithProtocol(config.Protocol),
		http_v2_client.WithPath(config.Path),
		http_v2_client.WithBoundary(config.GetBoundary()),
	)

	cli := &Client{
		wsClient:   conn,
		sender:     NewSender(client, config),
		writeMutex: &sync.Mutex{},
	}

	return cli
}

func (c *Client) reconnectClient() {
	// TODO: Refactor and consolidate with connect
	// Try to reconnect
	config, err := config3.GiveMeAConfig()
	disconnectIfErr(err, c.wsClient)
	client := http_v2_client.NewHttpV2Client(config.Host,
		config.Port,
		http_v2_client.WithProtocol(config.Protocol),
		http_v2_client.WithPath(config.Path),
		http_v2_client.WithBoundary(config.GetBoundary()),
	)
	c.sender = NewSender(client, config)

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

//OnError When an error occurs related with the websocket communication
func (c *Client) OnError(err SError) {
	if err.Level == CRITIC { // TODO: Die here
		fmt.Printf("Critical error %s\n", err.Err.Error())
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
	currentState := c.sender.GetState()
	if currentState.IsFirstChunk && !currentState.IsFinished {
		logIfErr(c.sender.Connect(), "Error connecting to cerence server")
		go startReceiving(c)
	}
	c.sender = receiver.SendWithClient(c.sender, msg).(*Sender)
	if c.sender.GetState().IsFinished {
		c.Write(conn, []byte(`{"recognition_finished": "1"}`)) // TODO: Too soon to close
		c.sender.Close()

		time.Sleep(30 * time.Millisecond)
		c.reconnectClient()
	}
}

//Write Synchronized method that sends to the client information. Avoids concurrent socket write
func (c *Client) Write(conn *websocket.Conn, msg []byte) {
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()
	logIfErr(conn.WriteJSON(msg), "Error sending recognition finished")
}

func disconnectIfErr(err error, conn *websocket.Conn) {
	if err != nil {
		fmt.Println("Cannot retrieve a config right now. We will disconnect the client")
		logIfErr(conn.Close(), "Cannot close connection with client")
	}
}

func logIfErr(err error, msg string) {
	if err != nil {
		fmt.Println(msg + " .Reason: " + err.Error())
	}
}

package cerence

import (
	"bytes"
	config2 "cloud-client-go/config"
	"fmt"
	. "github.com/acepero13/cloud-client-go/http_v2_client"
	"github.com/alvaro/asr_server/server/receiver"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

type Client struct {
	wsClient       *websocket.Conn
	sender         *Sender
	messageChannel chan bytes.Buffer
	writeMutex     *sync.Mutex
}

type clients struct {
	clients     []*Client
	clientMutex sync.Mutex
}

var ConnectedClients = clients{
	clients: []*Client{},
}

func OnConnected(conn *websocket.Conn) Client {
	ConnectedClients.clientMutex.Lock()
	client := newClient(conn)
	ConnectedClients.clients = append(ConnectedClients.clients, client)
	ConnectedClients.clientMutex.Unlock()
	return *client
}

func newClient(conn *websocket.Conn) *Client {
	config := config2.ReadConfig("config/asr_sem.json")
	client := NewHttpV2Client(config.Host, config.Port, WithProtocol(config.Protocol), WithPath(config.Path), WithBoundary(config.GetBoundary()))

	logIfErr(client.Connect(), "Error connecting to cerence... ")

	cli := &Client{
		wsClient:       conn,
		sender:         NewSender(client, config),
		messageChannel: make(chan bytes.Buffer),
		writeMutex:     &sync.Mutex{},
	}
	//go process(cli)
	go startReceiving(cli)
	return cli
}

func startReceiving(cerenceCli *Client) {
	client := cerenceCli.sender.cerenceClient
	go client.Receive()
	for chunk := range client.GetReceivedChunkChannel() {
		if string(chunk.Body.Bytes()) == "Close" {
			fmt.Println("Please close connection")
			time.Sleep(30 * time.Millisecond)
			logIfErr(client.Connect(), "Error reconnecting") // TODO: Maybe is better to connect on demand
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
	singleResult = asrResult.GetAtMost(1)

	toSend, errEncod := singleResult.ToBytes()
	if errDecod != nil || errEncod != nil || singleResult == nil {
		return
	}
	cli.Write(cli.wsClient, toSend)

}

func (c *Client) OnError(err SError) {
	if err.Level == CRITIC { // TODO: Die here
		fmt.Printf("Critical error %s\n", err.Err.Error())
	} else if err.Level == UNIMPORTANT {
		fmt.Printf("Critical error %s\n", err.Err.Error())
	}
}

func (c *Client) OnClose() {

}

func (c *Client) OnMessage(conn *websocket.Conn, msg []byte) {
	c.sender = receiver.SendWithClient(c.sender, msg).(*Sender)
	if c.sender.GetState().IsFinished {
		c.Write(conn, []byte(`{"recognition_finished": "1"}`)) // TODO: Too soon to close
		c.sender.Close()

		time.Sleep(30 * time.Millisecond)
		c.sender.ReConnect()
	}
}

func (c *Client) Write(conn *websocket.Conn, msg []byte) {
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()
	logIfErr(conn.WriteJSON(msg), "Error sending recognition finished")
}

func logIfErr(err error, msg string) {
	if err != nil {
		fmt.Println(msg + " .Reason: " + err.Error())
	}
}

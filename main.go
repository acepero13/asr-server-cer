package main

import (
	config2 "cloud-client-go/config"
	. "cloud-client-go/http_v2_client"
	. "cloud-client-go/util"
	"fmt"
	"github.com/gorilla/websocket"
	"log"

	"sync"
	"time"

	"encoding/json"
	"github.com/alvaro/asr_server/server"
	"github.com/alvaro/asr_server/server/receiver"
	//"github.com/acepero13/cloud-client-go" // TODO: Use once it becomes stable enough
)

var (
	wg sync.WaitGroup
)

type cerenceClient struct {
	client     *HttpV2Client
	config     *config2.Config
	state      *receiver.RequestState
	connection *server.WsConnection
	wsClient   *websocket.Conn
}

func main() {

	server.StartServer(onNewConnection)

}

func (c *cerenceClient) GetState() receiver.RequestState {
	return *c.state
}

func (c *cerenceClient) SetState(st receiver.RequestState) {
	*c.state = st
}

func (c *cerenceClient) SendHeader() {
	err := c.client.SendHeaders(c.config.Headers)
	ifErrorDie(err, c)
}

func ifErrorDie(err error, c *cerenceClient) {
	if err != nil {
		c.client.Close()
		c.wsClient.Close()
		ConsoleLogger.Println("error %s", err.Error())
	}
}

func (c *cerenceClient) SendRequest() {
	for _, part := range c.config.MultiParts {
		if part.Type == JsonType {
			err := sendJSONMsg(c.client, part)
			ifErrorDie(err, c)
		}
	}
}

func (c *cerenceClient) SendEndRequest() {
	err := c.client.SendMultiPartEnd()
	ifErrorDie(err, c)
}

func (c *cerenceClient) SendAudioChunk(chunk []byte) {

	for _, part := range c.config.MultiParts { // TODO: Not necessary to use a for here
		if part.Type == AudioType {
			sendAudioMsg(c.client, part, chunk)
		}
	}

}

func sendAudioMsg(client *HttpV2Client, part config2.MultiPart, chunk []byte) {

	err := client.SendMultiPart(part.Parameters, chunk)
	ifErrorDie(err, nil) // HERE??
}

func sendJSONMsg(client *HttpV2Client, part config2.MultiPart) error {
	bodyData, _ := json.Marshal(part.Body)
	if err := client.SendMultiPart(part.Parameters, bodyData); err != nil {
		return err
	}
	return nil
}

func enqueue(queue [][]byte, element []byte) [][]byte {
	queue = append(queue, element) // Simply append to enqueue.
	return queue
}

func dequeue(queue [][]byte) ([]byte, [][]byte) {
	element := queue[0]       // The first element is the one to be dequeued.
	return element, queue[1:] // Slice off the element once it is dequeued.
}

func onNewConnection(connection *server.WsConnection, wsClient *server.WsClient) {
	client, cerenceCli := createClient()
	cerenceCli.connection = connection
	cerenceCli.wsClient = wsClient.WebsocketClient
	startReceiverAndSender(client, wsClient, cerenceCli)
}

func startReceiverAndSender(client *HttpV2Client, ws *server.WsClient, cerenceCli *cerenceClient) {
	wg.Add(2)

	go func() {

		defer func() {
			if err := recover(); err != nil {
				ConsoleLogger.Println(err)
			}
			err := client.Close()
			logIfError(err, "Could not close connection")
		}()
		defer wg.Done()

		cerenceCli = sendToCerenceFromWs(ws, cerenceCli)
	}()

	go func() {
		defer func() {
			if err := recover(); err != nil {
				ConsoleLogger.Println(err)
			}
		}()
		defer wg.Done()
		receiveResult(cerenceCli)
		ConsoleLogger.Println("Receive done")
	}()

	wg.Wait()
}

func sendToCerenceFromWs(ws *server.WsClient, cerenceCli *cerenceClient) *cerenceClient {
	var queue [][]byte

	for {
		// Read in a new message as JSON and map it to a Message object
		_, data, err := ws.WebsocketClient.ReadMessage()
		actualQueue := enqueue(queue, data)

		msg, actualQueue := dequeue(actualQueue)
		queue = actualQueue

		cerenceCli = receiver.ReceiveWithClient(cerenceCli, msg).(*cerenceClient)

		if err != nil {
			log.Printf("error: %v", err)
			break
		}
		time.Sleep(30 * time.Millisecond)
		if cerenceCli.state.IsFinished {
			err := ws.WebsocketClient.WriteJSON([]byte(`{"recognition_finished": "1"}`))
			fmt.Println("SENDING END OF RECOGNITION")
			logIfError(err, "Could not send recognition finished")
			time.Sleep(300 * time.Millisecond)
			_ = cerenceCli.client.Close()
			break
		}

	}
	fmt.Println("Finished sending")
	return cerenceCli
}

func logIfError(err error, msg string) {
	if err != nil {
		fmt.Println(msg+" err: %s", err.Error())
	}
}

func createClient() (*HttpV2Client, *cerenceClient) {
	config := config2.ReadConfig("config/asr_sem.json")
	client := NewHttpV2Client(config.Host, config.Port, WithProtocol(config.Protocol), WithPath(config.Path), WithBoundary(config.GetBoundary()))

	if err := client.Connect(); err != nil {
		ConsoleLogger.Fatalln("Can't connect to server")
	}

	return client, newCerenceClient(client, config)
}

func newCerenceClient(client *HttpV2Client, config *config2.Config) *cerenceClient {
	var state *receiver.RequestState
	state = new(receiver.RequestState)
	state.IsFirstChunk = true
	state.IsFinished = false

	var cerenceCli *cerenceClient
	cerenceCli = new(cerenceClient)

	cerenceCli.client = client
	cerenceCli.config = config
	cerenceCli.state = state
	return cerenceCli
}

const receiving = "Receiving:"

func receiveResult(cerenceCli *cerenceClient) {
	client := cerenceCli.client
	go client.Receive()
	for chunk := range client.GetReceivedChunkChannel() {
		if string(chunk.Body.Bytes()) == "Close" {
			fmt.Println("Please close connection")
			break
		}

		formattedJson := PrintPrettyJson(receiving, chunk.Body.Bytes())
		msg := server.NewWsMessage(cerenceCli.wsClient)
		msg.Msg.Write([]byte(formattedJson))
		cerenceCli.connection.MessageChannel <- *msg

	}
	fmt.Println("ENDED FOR")
}

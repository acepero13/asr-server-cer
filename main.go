package main

import (
	"log"
	"net/http"
	config2 "cloud-client-go/config"
	. "cloud-client-go/http_v2_client"
	. "cloud-client-go/util"
	"bytes"
	"github.com/gorilla/websocket"
	"sync"
	"fmt"
	//"time"
	"strings"
	//"github.com/acepero13/cloud-client-go" // TODO: Use once it becomes stable enough
)
import "encoding/json"
import "github.com/alvaro/asr_server/server/receiver"

//TODO: Manage errors

// TODO: Refactor
// TODO: Handle timeouts, so the server does not die

var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan []byte) ;

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}


var (
	wg sync.WaitGroup
)


type cerenceClient struct {
	client *HttpV2Client
	config *config2.Config
	state *receiver.RequestState 
}

func(c* cerenceClient) GetState() receiver.RequestState{
	return *c.state
}

func(c* cerenceClient) SetState(st  receiver.RequestState) {
	*c.state = st
}

func(c * cerenceClient) SendHeder(){
	c.client.SendHeaders(c.config.Headers);
}

func (c* cerenceClient) SendRequest(){
	for _, part := range c.config.MultiParts {
		if part.Type == JsonType {
			sendJsonMsg(c.client, part)
		}
	}
}

func (c* cerenceClient) SendEndRequest(){
	c.client.SendMultiPartEnd()
}

func (c* cerenceClient) SendAudioChunk(chunk []byte) {

	for _, part := range c.config.MultiParts { // TODO: Not necessary to use a for here
		if part.Type == AudioType {
			sendAudioMsg(c.client, part, chunk)
		}
	}


}

func sendAudioMsg(client *HttpV2Client, part config2.MultiPart, chunk []byte){

	if err := client.SendMultiPart(part.Parameters, chunk); err != nil {
		ConsoleLogger.Fatalln(err)
	}
}

func sendJsonMsg(client *HttpV2Client, part config2.MultiPart) error {
	bodyData, _ := json.Marshal(part.Body)
	if err := client.SendMultiPart(part.Parameters, bodyData); err != nil {
		ConsoleLogger.Fatalln(err)
		return err
	}
	return nil
}



func main(){

	http.HandleFunc("/ws", handleConnections)

	// Start listening for incoming chat messages
	go handleMessages()

	// Start the server on localhost port 8000 and log any errors
	log.Println("http server started on :2701")
	err := http.ListenAndServe(":2701", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()

	// Register our new client
	clients[ws] = true

	config := config2.ReadConfig("config/asr_sem.json")
	client := NewHttpV2Client(config.Host, config.Port, WithProtocol(config.Protocol), WithPath(config.Path), WithBoundary(config.GetBoundary()))


	if err := client.Connect(); err != nil {
		ConsoleLogger.Fatalln("Can't connect to server")
	}

	var state *receiver.RequestState
	state = new(receiver.RequestState) 
	state.IsFirstChunk = true;

	var cerenceCli *cerenceClient;
	cerenceCli = new(cerenceClient)

	cerenceCli.client = client;
	cerenceCli.config = config;
	cerenceCli.state = state;

	
	wg.Add(2)

	go func(){

		defer func() {
			if err := recover(); err != nil {
				ConsoleLogger.Println(err)
			}
		}()
		defer wg.Done()

		for {
		
			// Read in a new message as JSON and map it to a Message object
			_, msg, err := ws.ReadMessage()
	
			cerenceCli = receiver.ReceiveWithClient(cerenceCli, msg).(*cerenceClient)
	
			if err != nil {
				log.Printf("error: %v", err)
				break
			}
			//time.Sleep(30)
			// Send the newly received message to the broadcast channel
			//broadcast <- []byte("")
			
		}
	}()


	


	go func() {
		defer func() {
			if err := recover(); err != nil {
				ConsoleLogger.Println(err)
			}
		}()
		defer wg.Done()
		receiveResult(client)
		ConsoleLogger.Println("Receive done")
	}()

	wg.Wait()
}

const Receiving = "Receiving:"

func receiveResult(client *HttpV2Client){
	go client.Receive()
	for chunk := range client.GetReceivedChunkChannel() {
		parameters, _ := handleBoundaryAndParameters(chunk.BoundaryAndParameters)
		if len(parameters) > 0 {
			ConsoleLogger.Println(fmt.Sprintf("%s multiple parts", Receiving))
			for n := range parameters {
				ConsoleLogger.Println(parameters[n])
			
			}
		}

		

		PrintPrettyJson(Receiving, chunk.Body.Bytes())
		
		json := PrintPrettyJson(Receiving, chunk.Body.Bytes())
		

		ConsoleLogger.Println(json + CRLF)

		broadcast <- []byte(json)
		
		

	}
}

func handleBoundaryAndParameters(bytes bytes.Buffer) ([]string, bool) {
	data := strings.Split(bytes.String(), CRLF)
	var parameters []string
	isAudioPart := true
	for n := range data {
		if strings.Trim(data[n], "\r") != "" {
			parameters = append(parameters, data[n])
			if strings.Contains(data[n], "Content-Type: application/JSON;") {
				isAudioPart = false
			}
		}
	}
	return parameters, isAudioPart
}


func handleMessages() {
	for {
		// TODO: Send it only the the current client
		// Grab the next message from the broadcast channel
		msg := <-broadcast
		// Send it out to every client that is currently connected
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
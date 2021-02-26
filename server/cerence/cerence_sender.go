package cerence

import (
	config2 "cloud-client-go/config"
	"cloud-client-go/http_v2_client"
	"cloud-client-go/util"
	"encoding/json"
	"fmt"
	"github.com/acepero13/asr-server-cer/server/config"
	"github.com/acepero13/asr-server-cer/server/receiver"
)

//Sender  Stateful sender. Sends request to cerence client based on the current chunk state
type Sender struct {
	state         *receiver.RequestState
	cerenceClient *http_v2_client.HttpV2Client
	config        *config2.Config
}

//NewSender Creates a new stateful cerence sender. Encapsulate logic for incoming chunks
func NewSender(cerenceClient *http_v2_client.HttpV2Client, config *config2.Config) *Sender {
	var state *receiver.RequestState
	state = new(receiver.RequestState)
	state.IsFirstChunk = true
	state.IsFinished = false

	return &Sender{cerenceClient: cerenceClient, config: config, state: state}
}

//GetState Returns the current state
func (c *Sender) GetState() receiver.RequestState {
	return *c.state
}

//SetState Sets a new state
func (c *Sender) SetState(st receiver.RequestState) {
	*c.state = st
}

//SendHeader Sends Header information to cerence
func (c *Sender) SendHeader() {
	fmt.Println("Sending header")
	logIfErr(c.cerenceClient.SendHeaders(c.config.Headers), "Cannot Send Header")
}

//SendRequest Sends an ASR request to cerence
func (c *Sender) SendRequest() {
	fmt.Println("Sending request")
	for _, part := range c.config.MultiParts {
		if part.Type == util.JsonType {
			bodyData, _ := json.Marshal(part.Body)
			logIfErr(c.cerenceClient.SendMultiPart(part.Parameters, bodyData), "Cannot send Multipart request")
		}
	}
}

//SendEndRequest Sends that the asr request has finished
func (c *Sender) SendEndRequest() {
	fmt.Println("Sending END request")
	logIfErr(c.cerenceClient.SendMultiPartEnd(), "Cannot send end request")
}

//Close Closes the connection with cerence server
func (c *Sender) Close() {
	logIfErr(config.Release(c.config), "Error while releasing the current config")
	logIfErr(c.cerenceClient.Close(), "Error closing connection with cerence client")
}

//SendAudioChunk Sends the audio chunk to cerence server. It accepts a chunk of raw audio data
func (c *Sender) SendAudioChunk(chunk []byte) {
	for _, part := range c.config.MultiParts {
		if part.Type == util.AudioType {
			logIfErr(c.cerenceClient.SendMultiPart(part.Parameters, chunk), "Cannot send audio chunk")
		}
	}

}

//Connect Connects to cerence server
func (c *Sender) Connect() error {
	return c.cerenceClient.Connect()
}

//IsConnected Returns true if the cerence client is connected to the server, false otherwise
func (c *Sender) IsConnected() bool {
	err := c.cerenceClient.CheckConnection()
	if err != nil {
		return false
	}
	return true
}

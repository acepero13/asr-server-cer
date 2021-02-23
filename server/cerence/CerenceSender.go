package cerence

import (
	config2 "cloud-client-go/config"
	"cloud-client-go/util"
	"encoding/json"
	"fmt"
	. "github.com/acepero13/cloud-client-go/http_v2_client"
	"github.com/alvaro/asr_server/server/receiver"
)

type Sender struct {
	state         *receiver.RequestState
	cerenceClient *HttpV2Client
	config        *config2.Config
}

func NewSender(cerenceClient *HttpV2Client, config *config2.Config) *Sender {
	var state *receiver.RequestState
	state = new(receiver.RequestState)
	state.IsFirstChunk = true
	state.IsFinished = false

	return &Sender{cerenceClient: cerenceClient, config: config, state: state}
}

func (c *Sender) GetState() receiver.RequestState {
	return *c.state
}

func (c *Sender) SetState(st receiver.RequestState) {
	*c.state = st
}

func (c *Sender) SendHeader() {
	fmt.Println("Sending header")
	logIfErr(c.cerenceClient.SendHeaders(c.config.Headers), "Cannot Send Header")
}

func (c *Sender) SendRequest() {
	fmt.Println("Sending request")
	for _, part := range c.config.MultiParts {
		if part.Type == util.JsonType {
			bodyData, _ := json.Marshal(part.Body)
			logIfErr(c.cerenceClient.SendMultiPart(part.Parameters, bodyData), "Cannot send Multipart request")
		}
	}
}

func (c *Sender) SendEndRequest() {
	fmt.Println("Sending END request")
	logIfErr(c.cerenceClient.SendMultiPartEnd(), "Cannot send end request")
}

func (c *Sender) Close() {
	logIfErr(c.cerenceClient.Close(), "Error closing connection with cerence client")
}

func (c *Sender) ReConnect() {
	logIfErr(c.cerenceClient.Connect(), "Error reconnecting connection with cerence client")
}

func (c *Sender) SendAudioChunk(chunk []byte) {
	for _, part := range c.config.MultiParts { // TODO: Not necessary to use a for here
		if part.Type == util.AudioType {
			logIfErr(c.cerenceClient.SendMultiPart(part.Parameters, chunk), "Cannot send audio chunk")
		}
	}

}

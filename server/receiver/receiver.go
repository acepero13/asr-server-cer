package receiver

import "strings"
import "encoding/json"

type command interface {
	Execute(data []byte)
}

type RequestState struct {
	IsFirstChunk bool
	IsFinished   bool
}

type Client interface {
	SendHeader()
	SendRequest()
	SendAudioChunk(chunk []byte)
	SendEndRequest()
	GetState() RequestState
	SetState(st RequestState)
}

type asrEvent struct {
	Event string `json:"asr_event"`
}

func Receive(data []byte) (*Client, error) {
	return nil, nil
}

func ReceiveWithClient(c Client, data []byte) Client {
	var s = c.GetState()
	if strings.Contains(string(data), "asr_event") {
		var ev asrEvent
		json.Unmarshal(data, &ev)
		if ev.Event == "stopped" {
			var st = RequestState{}
			st.IsFirstChunk = true
			st.IsFinished = true
			c.SetState(st)
		}
		return c

	}
	if s.IsFirstChunk {
		c.SendHeader()
		c.SendRequest()
		c.SendAudioChunk(data)
	} else if s.IsFinished {
		c.SendAudioChunk(data)
		c.SendEndRequest()
	} else {
		c.SendAudioChunk(data)
	}

	var st = RequestState{}
	st.IsFirstChunk = false
	st.IsFinished = false
	c.SetState(st)

	return c
}

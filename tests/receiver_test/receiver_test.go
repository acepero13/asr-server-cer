package receiver_test

import "testing"
import "github.com/alvaro/asr_server/server/receiver"

type dummyCommand struct {
	received string
}

type dummyClient struct {
	headerCalled      bool
	requestCalled     bool
	endRequestCalled  bool
	audioChunkCounter int
	state             *receiver.RequestState
}

func (m dummyCommand) Execute(data []byte) { // Implements command interface

}

func (c *dummyClient) SendHeader() {
	c.headerCalled = true
}

func (c *dummyClient) SendRequest() {
	c.requestCalled = true
}

func (c *dummyClient) SendAudioChunk(data []byte) {
	c.audioChunkCounter++
}

func (c *dummyClient) SendEndRequest() {
	c.endRequestCalled = true
}

func (c *dummyClient) GetState() receiver.RequestState {
	return *c.state
}

func (c *dummyClient) SetState(st receiver.RequestState) {
	*c.state = st
}

func TestReceivesFirstChunkCallsSendHeaderPlusFirstChunk(t *testing.T) {
	var c *dummyClient
	c = new(dummyClient)

	var s *receiver.RequestState
	s = new(receiver.RequestState)

	s.IsFirstChunk = true

	c.state = s

	var client *dummyClient = receiver.SendWithClient(c, []byte("hello")).(*dummyClient)

	if c == nil {
		t.Errorf("client should not be null")
	}

	if client.headerCalled != true {
		t.Errorf("Should have sent the header")
	}

	if client.requestCalled != true {
		t.Errorf("Should have sent the request")
	}

	if client.audioChunkCounter != 1 {
		t.Errorf("Should have sent one chunk")
	}
}

func TestReceivesSecondChunkCallsSendsOnlyOneChunk(t *testing.T) {
	var c *dummyClient
	c = new(dummyClient)

	var s *receiver.RequestState
	s = new(receiver.RequestState)

	s.IsFirstChunk = false

	c.state = s

	var client *dummyClient = receiver.SendWithClient(c, []byte("hello")).(*dummyClient)

	if c == nil {
		t.Errorf("client should not be null")
	}

	if client.headerCalled != false {
		t.Errorf("Should have sent the header")
	}

	if client.requestCalled != false {
		t.Errorf("Should have sent the request")
	}

	if client.audioChunkCounter != 1 {
		t.Errorf("Should have sent one chunk")
	}
}

func TestReceivesLastChunkCallsSendsChunkAndCloseRequest(t *testing.T) {
	var c *dummyClient
	c = new(dummyClient)

	var s *receiver.RequestState
	s = new(receiver.RequestState)

	s.IsFirstChunk = false
	s.IsFinished = true

	c.state = s

	var client *dummyClient = receiver.SendWithClient(c, []byte("hello")).(*dummyClient)

	if c == nil {
		t.Errorf("client should not be null")
	}

	if client.headerCalled != false {
		t.Errorf("Should have sent the header")
	}

	if client.requestCalled != false {
		t.Errorf("Should have sent the request")
	}

	if client.audioChunkCounter != 1 {
		t.Errorf("Should have sent one chunk")
	}

	if client.endRequestCalled != true {
		t.Errorf("Should have sent the final request")
	}
}

func TestReceivesAsrEndedEventResetsChunkCounter(t *testing.T) {

	var c *dummyClient
	c = new(dummyClient)

	var s *receiver.RequestState
	s = new(receiver.RequestState)

	c.state = s

	var client = receiver.SendWithClient(c, []byte(`{"asr_event": "stopped"}`)).(*dummyClient)

	var st = client.GetState()
	if st.IsFirstChunk != true {
		t.Errorf("Should reset the chunk counter")
	}

}

func TestReceivesAsrEndedEventResetsChunkCounterAndAfterFirstChunkItResetsItToFalse(t *testing.T) {

	var c *dummyClient
	c = new(dummyClient)

	var s *receiver.RequestState
	s = new(receiver.RequestState)

	c.state = s

	var client *dummyClient = receiver.SendWithClient(c, []byte(`{"asr_event": "stopped"}`)).(*dummyClient)

	var st = client.GetState()
	if st.IsFirstChunk != true {
		t.Errorf("Should reset the chunk counter")
	}

	var client2 *dummyClient = receiver.SendWithClient(c, []byte("hello")).(*dummyClient)

	var st2 = client2.GetState()
	if st2.IsFirstChunk != false {
		t.Errorf("Should reset the chunk counter")
	}

}

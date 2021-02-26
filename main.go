package main

import (
	"github.com/acepero13/asr_server/server/cerence"
)

func main() {
	cerence.WebSocketApp(2701, cerence.OnConnected)

}

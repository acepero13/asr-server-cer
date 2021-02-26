package main

import (
	"github.com/acepero13/asr-server-cer/server/cerence"
)

func main() {
	cerence.WebSocketApp(2701, cerence.OnConnected)

}

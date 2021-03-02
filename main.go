package main

import (
	"fmt"
	"github.com/acepero13/asr-server-cer/server/cerence"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

var port = 2701
var noTsl = true

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:        "port",
				Value:       2701,
				Usage:       "port to start listening for raw audio data",
				Destination: &port,
			},
			&cli.BoolFlag{
				Name:        "no-tls",
				Value:       false,
				Usage:       "if present, uses an insecure communication protocol (ws)",
				Destination: &noTsl,
			},
		},
		Action: func(c *cli.Context) error {
			if noTsl {
				fmt.Println("Warning: you are using ws as protocol. Keep in mind that the communication will not be encrypted")
			}
			cerence.WebSocketApp(port, !noTsl, cerence.OnConnected)
			return nil
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}

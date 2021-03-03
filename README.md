# C-Server

This is a small websocket server that expects audio chunks and sends out the Cerence cloud request for
ASR recognition. It uses the library [github.com/acepero13/cloud-client-go](https://github.com/cerence/github.com/acepero13/cloud-client-go).

This app opens the port **2701** and listens to messages sent via **websocket** to that port. It expects raw audio data (ex. taken from a microphone) and connects to **cerence** cloud server and notifies the client with the recognition

## Settings
Make sure that you have the corresponding files inside the **configs** folder. There you
should place the server certificates (if you want to connect using **TLS**), and the configuration files (json) needed
to connect to the **Cerence** server.

## Run manually

1. Set up your local development environment for **Go**. You can follow this tutorial [
How To Install Go and Set Up a Local Programming Environment](https://www.digitalocean.com/community/tutorials/how-to-install-go-and-set-up-a-local-programming-environment-on-ubuntu-18-04)

2. In a terminal, execute the following command: `go get https://github.com/acepero13/asr-server-cer`. This will install the asr-server into your `$GOPATH`

3. Go to `$GOPATH/bin` and copy the **configs** folder that contains the _asr configuration files_ for connecting to **cerence**

4. In `$GOPATH/bin` execute: `./asr-server-cer`  

## Run from Docker

Build image 
```bash
docker build -t asr-server-cer .
```

Run Docker image
```bash
docker run -it -p 2701:2701 --net=host asr-server-cer:latest
```

You can also specify arguments to change the default configuration of the server. For example,
if you want to run the server on a different port other than 2701, or you do not want to use
encrypted communication.

Run Docker image with arguments

```bash
docker run -it -p 2701:2701 --net=host asr-server-cer:latest --port 5005 --no-tls
```

Arguments
> --port **value**  _port to start listening for raw audio data (default: 2701)_
> 
> --no-tls      _if present, uses an insecure communication protocol (ws) (default: false)_
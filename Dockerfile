FROM golang:1.15-alpine3.13

# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/acepero13/asr-server-cer

WORKDIR /go/src/github.com/acepero13/asr-server-cer


# Build the asr-server-cer command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)

RUN go mod download

RUN go install github.com/acepero13/asr-server-cer


# Run the outyet command by default when the container starts.
ENTRYPOINT /go/bin/asr-server-cer

# Document that the service listens on port 2701.
EXPOSE 2701

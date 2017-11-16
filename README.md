# Messenger
A basic messenger app that uses NATS.io and Websockets as the underlying publish/subscribe tech

## Prerequisites
1. NATS
2. Go
3. golang/dep
4. A websocket client

## Getting Started
1. You will need to start a local `gnatsd` server
2. Build the messenger binary
```
dep ensure -v
go build
```
3. Run the binary
```
./messenger
```

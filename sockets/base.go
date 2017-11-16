package sockets

import (
	"net/http"

	"github.com/alecholmez/messenger/users"
	"github.com/gorilla/websocket"
)

// Message ...
type Message struct {
	Timestamp string     `json:"timestamp"`
	Message   string     `json:"message"`
	User      users.User `json:"user"`
	ID        string     `json:"_id"`
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// Insecure but will work for now
			return true
		},
	}
)

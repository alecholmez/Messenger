package sockets

import (
	"log"
	"net/http"
	"time"

	"github.com/alecholmez/messenger/config"
	"github.com/deciphernow/gm-fabric-go/dbutil"
)

// Send ...
func Send(w http.ResponseWriter, r *http.Request) {
	// Upgrade web handler to a websocket connection
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade http connection: " + err.Error())
		return
	}
	defer c.Close()

	ec, err := config.GetNatsConn(r.Context())
	if err != nil {
		log.Println(err)
		return
	}

	// Create a channel and bind the "message" nat tunnel to that channel
	ch := make(chan *Message)
	ec.BindSendChan("message", ch)

	var m Message
	for {
		err := c.ReadJSON(&m)
		if err != nil {
			log.Println(err)
		}
		m.Timestamp = time.Now().Format(time.RFC3339)
		m.ID = dbutil.CreateHash()

		// Send the message to the nat tunnel
		ch <- &m
	}

}

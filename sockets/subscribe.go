package sockets

import (
	"log"
	"net/http"

	"github.com/alecholmez/messenger/config"
)

// Subscribe ...
func Subscribe(w http.ResponseWriter, r *http.Request) {
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

	ch := make(chan *Message)
	ec.BindRecvChan("message", ch)

	var m *Message
	for {
		// Read the channel
		m = <-ch

		// Send the JSON to the client
		err = c.WriteJSON(m)
		if err != nil {
			log.Println(err)
		}
	}
}

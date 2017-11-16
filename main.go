package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/deciphernow/gm-fabric-go/middleware"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	nats "github.com/nats-io/go-nats"
)

// Message ...
type Message struct {
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
	User      User   `json:"user"`
}

// User ...
type User struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type key int

const (
	natsKey key = iota
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			fmt.Println(r.Header.Get("Origin"))
			return true
		},
	}
)

// WithNats ...
func WithNats(conn *nats.EncodedConn) middleware.Middleware {
	return middleware.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			next.ServeHTTP(w, r)
		})
	})
}

// GetNatsConn ...
func GetNatsConn(ctx context.Context) (*nats.EncodedConn, error) {
	conn, ok := ctx.Value(natsKey).(*nats.EncodedConn)
	if !ok {
		return nil, errors.New("Failed to retrieve nats connection")
	}

	return conn, nil
}

func main() {
	nc, _ := nats.Connect(nats.DefaultURL)
	c, _ := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	defer c.Close()

	stack := middleware.Chain(
		WithNats(c),
	)

	mux := mux.NewRouter()

	mux.HandleFunc("/subscribe", Subscribe)
	mux.HandleFunc("/send", Send).Methods("POST")

	s := http.Server{
		Addr:    ":8080",
		Handler: stack.Wrap(mux),
	}

	s.ListenAndServe()
}

// Subscribe ...
func Subscribe(w http.ResponseWriter, r *http.Request) {
	// Upgrade web handler to a websocket connection
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade http connection: " + err.Error())
		return
	}
	defer c.Close()

	ec, err := GetNatsConn(r.Context())
	if err != nil {
		log.Println(err)
		return
	}

	ch := make(chan *Message)
	ec.BindRecvChan("message", ch)

	for {
		m := <-ch
		b, err := json.MarshalIndent(m, "", "    ")
		if err != nil {
			log.Fatal(err)
		}
		err = c.WriteJSON(b)
		if err != nil {
			log.Println(err)
		}
	}
}

// Send ...
func Send(w http.ResponseWriter, r *http.Request) {
	// Upgrade web handler to a websocket connection
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade http connection: " + err.Error())
		return
	}
	defer c.Close()

	ec, err := GetNatsConn(r.Context())
	if err != nil {
		log.Println(err)
		return
	}

	for {
		var m Message
		err := c.ReadJSON(&m)
		if err != nil {
			log.Println(err)
		}

		ch := make(chan *Message)
		ec.BindSendChan("message", ch)

		ch <- &m
	}

}

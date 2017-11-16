package main

import (
	"net/http"
	"os"
	"time"

	mgo "gopkg.in/mgo.v2"

	"github.com/alecholmez/messenger/config"
	"github.com/alecholmez/messenger/sockets"
	"github.com/alecholmez/messenger/users"
	"github.com/deciphernow/gm-fabric-go/dbutil"
	"github.com/deciphernow/gm-fabric-go/middleware"
	"github.com/gorilla/mux"
	nats "github.com/nats-io/go-nats"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

func main() {
	log := zerolog.New(os.Stderr).With().Timestamp().Logger().
		Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Read in config
	sc := config.NewConfig()

	nc, err := nats.Connect(sc.NatsIP)
	if err != nil {
		log.Error().Err(err).Msg("Failed to connect to NATS")
		return
	}

	c, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create encoded NATS conn")
	}
	defer c.Close()

	sess, err := mgo.Dial(sc.MongoIP)
	if err != nil {
		log.Error().Err(err).Msg("Failed to connect to Mongo")
		return
	}
	defer sess.Close()

	stack := middleware.Chain(
		middleware.MiddlewareFunc(hlog.NewHandler(log)),
		middleware.MiddlewareFunc(hlog.AccessHandler(func(r *http.Request, status int, size int, duration time.Duration) {
			hlog.FromRequest(r).Info().
				Str("method", r.Method).
				Str("path", r.URL.String()).
				Int("status", status).
				Int("size", size).
				Dur("duration", duration).
				Msg("Access")
		})),
		middleware.MiddlewareFunc(hlog.UserAgentHandler("user_agent")),
		config.WithNats(c),
		dbutil.WithMongo(sess),
	)

	mux := mux.NewRouter()

	mux.HandleFunc("/subscribe", sockets.Subscribe)
	mux.HandleFunc("/send", sockets.Send)
	mux.HandleFunc("/signup", users.Signup).Methods("POST")
	mux.HandleFunc("/login", users.Login).Methods("POST")

	s := http.Server{
		Addr:    ":8080",
		Handler: stack.Wrap(mux),
	}

	log.Info().Msg("Messenger listening on :8080")
	// Block on this
	s.ListenAndServe()
}

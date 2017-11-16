package config

import (
	"context"
	"errors"
	"net/http"
	"os"

	"github.com/deciphernow/gm-fabric-go/middleware"
	nats "github.com/nats-io/go-nats"
)

type key int

const (
	natsKey key = iota
)

// Service is the basic service configuration object
type Service struct {
	NatsIP  string `json:"natsAddress"`
	MongoIP string `json:"mongoAddress"`
	Host    string `json:"host"`
	Port    string `json:"port"`
}

// NewConfig ...
func NewConfig() Service {
	var s Service

	s.NatsIP = os.Getenv("NATS_URL")
	s.MongoIP = os.Getenv("MONGO_URL")
	s.Host = os.Getenv("HOST")
	s.Port = os.Getenv("PORT")

	return s
}

// WithNats ...
func WithNats(conn *nats.EncodedConn) middleware.Middleware {
	return middleware.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			r = r.WithContext(context.WithValue(r.Context(), natsKey, conn))
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

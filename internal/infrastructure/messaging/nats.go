package messaging

import (
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

func InitNats(url string, ns *server.Server) *nats.Conn {
	clientOpts := []nats.Option{
		nats.InProcessServer(ns),
	}

	conn, err := nats.Connect(url, clientOpts...)
	if err != nil {
		zap.S().Panicw("Cannot connect to nats server")
	}
	return conn
}

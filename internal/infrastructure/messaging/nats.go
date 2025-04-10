package messaging

import (
	"fmt"

	"github.com/guardlight/server/internal/essential/config"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

func InitNatsInProcess(ns *server.Server) *nats.Conn {
	clientOpts := []nats.Option{
		nats.InProcessServer(ns),
		nats.UserInfo(config.Get().Nats.User, config.Get().Nats.Password),
	}

	conn, err := nats.Connect(ns.ClientURL(), clientOpts...)
	if err != nil {
		zap.S().Panicw("Cannot connect to nats server")
	}
	zap.S().Info("Using inprocess NATS")
	return conn
}

func InitNats() *nats.Conn {
	clientOpts := []nats.Option{
		nats.UserInfo(config.Get().Nats.User, config.Get().Nats.Password),
	}

	conString := fmt.Sprintf("%s:%d", config.Get().Nats.Server, config.Get().Nats.Port)
	conn, err := nats.Connect(conString, clientOpts...)
	if err != nil {
		zap.S().Panicw("Cannot connect to nats server")
	}
	zap.S().Info("Using external NATS")
	return conn
}

package natsmessaging

import (
	"errors"
	"time"

	"github.com/guardlight/server/internal/essential/config"
	"github.com/nats-io/nats-server/v2/server"
	"go.uber.org/zap"
)

var (
	nserv *server.Server
)

func NewNatsServer() error {

	opts := &server.Options{
		HTTPPort: 8222,
		Port:     4222,
		Users: []*server.User{
			{
				Username: config.Get().Nats.User,
				Password: config.Get().Nats.Password,
			},
		},
		JetStream:  true,
		MaxPayload: 64 * 1_000_000,
	}

	// Use external nats if Nats.Server is set
	if config.Get().Nats.Server != "" {
		opts.Host = config.Get().Server.Host
		opts.Port = config.Get().Nats.Port
	}

	ns, err := server.NewServer(opts)
	if err != nil {
		zap.S().Errorw("Could not create nats server", "error", err)
		zap.S().Panic("Could not create nats server")
		return err
	}

	// ns.ConfigureLogger()

	go ns.Start()
	if !ns.ReadyForConnections(5 * time.Second) {
		zap.S().Panic("Server not ready after 5 seconds")
		return errors.New("server not ready after 5 seconds")
	}
	zap.S().Infow("nats started", "url", ns.ClientURL())
	nserv = ns
	return nil
}

func GetNatsUrl() string {
	return nserv.ClientURL()
}

func GetServer() *server.Server {
	return nserv
}

func ShutdownNatsServer() {
	nserv.Shutdown()
}

func WaitForShutdown() {
	nserv.WaitForShutdown()
}

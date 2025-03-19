package cmd

import (
	"github.com/guardlight/server/servers/natsmessaging"
	"go.uber.org/zap"
)

func GlExternalServers() {
	// Messaging - Nats server
	err := natsmessaging.NewNatsServer()
	if err != nil {
		zap.S().Panicw("could not start nats server", "error", err)
	}
	zap.S().Info("started servers")
}

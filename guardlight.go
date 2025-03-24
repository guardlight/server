package main

import (
	"os"

	"github.com/guardlight/server/cmd"
	"github.com/guardlight/server/internal/essential/config"
	"github.com/guardlight/server/internal/essential/logging"
)

func init() {
	env := getEnv("environment", "production")
	confFileDir := getEnv("env_file_dir", "./")

	getEnvFile := func() string {
		switch env {
		case "development":
			return confFileDir + "config-development.yaml"
		// case "staging":
		// 	return confFileDir + "env-staging.yaml"
		case "production":
			return confFileDir + "config.yaml"
		default:
			panic("ENVIRONMENT variable not set")
		}
	}

	// Setup the correct logging
	logging.SetupLogging(env)

	config.SetupConfig(getEnvFile())

}

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

func main() {
	cmd.Server()
}

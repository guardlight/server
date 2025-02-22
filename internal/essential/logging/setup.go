package logging

import (
	"log"

	"go.uber.org/zap"
)

var (
	logger *zap.Logger
)

func SetupLogging(e string) {
	if e == "production" {
		l, err := zap.NewProduction()
		if err != nil {
			log.Fatal("production logger failed")
		}
		zap.ReplaceGlobals(l)
		logger = l
	} else {
		l, err := zap.NewDevelopment()
		if err != nil {
			log.Fatal("development logger failed")
		}
		zap.ReplaceGlobals(l)
		logger = l
	}

}

func GetLogger() *zap.Logger {
	return logger
}

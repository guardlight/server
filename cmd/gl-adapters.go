package cmd

import (
	"github.com/guardlight/server/pkg/gladapters/analyzers"
	"github.com/guardlight/server/pkg/gladapters/parsers"
	"github.com/guardlight/server/pkg/gladapters/reporters"
	"github.com/nats-io/nats.go"
)

func GLAdapters(ncon *nats.Conn) {
	parsers.NewFreetextParser(ncon)
	analyzers.NewWordsearchAnalyzer(ncon)
	reporters.NewWordcountReporter(ncon)
}

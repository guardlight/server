package parsers

import (
	"encoding/json"
	"os"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/guardlight/server/internal/essential/logging"
	"github.com/guardlight/server/internal/infrastructure/messaging"
	"github.com/guardlight/server/pkg/parsercontract"
	"github.com/guardlight/server/servers/natsmessaging"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type TestSuiteFreetextParserIntegration struct {
	suite.Suite
	ncon *nats.Conn
}

func (s *TestSuiteFreetextParserIntegration) SetupSuite() {
	logging.SetupLogging("test")

	err := natsmessaging.NewNatsServer()
	s.Assert().NoError(err)

	s.ncon = messaging.InitNats(natsmessaging.GetNatsUrl(), natsmessaging.GetServer())

	zap.S().Info("Setted up")
}

func TestFreetextParserSuiteRun(t *testing.T) {
	suite.Run(t, new(TestSuiteFreetextParserIntegration))
}

func (s *TestSuiteFreetextParserIntegration) TestParser() {
	NewFreetextParser(s.ncon)

	freetext, err := os.ReadFile("../../../testdata/epubs/lion.txt")
	s.Assert().NoError(err)

	jid := uuid.MustParse("1d36a166-2fb9-4028-ad2f-c184980eb33e")
	aid := uuid.MustParse("6a786e6d-e6f9-4ff8-a477-40ba73c6d6d1")
	pr := parsercontract.ParserRequest{
		JobId:      jid,
		AnalysisId: aid,
		Content:    freetext,
	}

	data, err := json.Marshal(pr)
	s.Assert().NoError(err)

	err = s.ncon.Publish("parser.freetext", data)
	s.Assert().NoError(err)

	var wg sync.WaitGroup
	wg.Add(1)
	s.ncon.Subscribe("parser.result", func(m *nats.Msg) {
		var t parsercontract.ParserResponse
		err := json.Unmarshal(m.Data, &t)
		s.Assert().NoError(err)
		s.Assert().Equal(jid, t.JobId)
		zap.S().Infow("parsed text", "text", t.Text)
		// os.WriteFile("./test.txt", []byte(t.Text), os.ModePerm)
		wg.Done()
	})

	wg.Wait()
}

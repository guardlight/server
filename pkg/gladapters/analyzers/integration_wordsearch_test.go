package analyzers

import (
	"encoding/json"
	"os"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/guardlight/server/internal/essential/logging"
	"github.com/guardlight/server/internal/infrastructure/messaging"
	"github.com/guardlight/server/pkg/analyzercontract"
	"github.com/guardlight/server/servers/natsmessaging"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type TestSuiteWordsearchAnalyzerIntegration struct {
	suite.Suite
	ncon *nats.Conn
}

func (s *TestSuiteWordsearchAnalyzerIntegration) SetupSuite() {
	logging.SetupLogging("test")

	err := natsmessaging.NewNatsServer()
	s.Assert().NoError(err)

	s.ncon = messaging.InitNats(natsmessaging.GetNatsUrl(), natsmessaging.GetServer())

	zap.S().Info("Setted up")
}

func TestFreetextParserSuiteRun(t *testing.T) {
	suite.Run(t, new(TestSuiteWordsearchAnalyzerIntegration))
}

func (s *TestSuiteWordsearchAnalyzerIntegration) TestWordsearchAnalyzer() {
	NewWordsearchAnalyzer(s.ncon)

	parsedText, err := os.ReadFile("../../../testdata/epubs/lion-parsed.txt")
	s.Assert().NoError(err)

	jid := uuid.MustParse("1d36a166-2fb9-4028-ad2f-c184980eb33e")
	aid := uuid.MustParse("6a786e6d-e6f9-4ff8-a477-40ba73c6d6d1")

	ar := analyzercontract.AnalyzerRequest{
		JobId:      jid,
		AnalysisId: aid,
		Content:    string(parsedText),
		Inputs: []analyzercontract.AnalysisInput{
			{
				Key:   "strict_words",
				Value: "Magic, Stair, stairs, I do",
			},
		},
	}

	data, err := json.Marshal(ar)
	s.Assert().NoError(err)

	err = s.ncon.Publish("analyzer.word_search", data)
	s.Assert().NoError(err)

	var wg sync.WaitGroup
	wg.Add(1)
	s.ncon.Subscribe("analyzer.result", func(m *nats.Msg) {
		var t analyzercontract.AnalyzerResponse
		err := json.Unmarshal(m.Data, &t)
		s.Assert().NoError(err)
		s.Assert().Equal(jid, t.JobId)
		zap.S().Infow("Analyzer results", "results", t.Results)
		// os.WriteFile("./test.txt", []byte(strings.Join(t.Results, "\n")), os.ModePerm)
		wg.Done()
	})

	wg.Wait()

}

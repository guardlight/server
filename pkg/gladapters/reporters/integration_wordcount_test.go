package reporters

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/guardlight/server/internal/essential/config"
	"github.com/guardlight/server/internal/essential/logging"
	"github.com/guardlight/server/internal/infrastructure/messaging"
	"github.com/guardlight/server/pkg/reportercontract"
	"github.com/guardlight/server/servers/natsmessaging"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type TestSuiteWordcountReporterIntegration struct {
	suite.Suite
	ncon *nats.Conn
}

func (s *TestSuiteWordcountReporterIntegration) SetupSuite() {
	logging.SetupLogging("test")
	config.SetupConfig("testdata/envs/gladapters.yaml")

	err := natsmessaging.NewNatsServer()
	s.Assert().NoError(err)

	s.ncon = messaging.InitNatsInProcess(natsmessaging.GetServer())

	zap.S().Info("Setted up")
}

func TestWordcountReporterSuiteRun(t *testing.T) {
	suite.Run(t, new(TestSuiteWordcountReporterIntegration))
}

func (s *TestSuiteWordcountReporterIntegration) TestWordcountReporter() {
	NewWordcountReporter(s.ncon)

	jid := uuid.MustParse("1d36a166-2fb9-4028-ad2f-c184980eb33e")
	aid := uuid.MustParse("6a786e6d-e6f9-4ff8-a477-40ba73c6d6d1")

	ar := reportercontract.ReporterRequest{
		JobId:      jid,
		AnalysisId: aid,
		Contents:   []string{},
	}

	data, err := json.Marshal(ar)
	s.Assert().NoError(err)

	err = s.ncon.Publish("reporter.word_count", data)
	s.Assert().NoError(err)

	var wg sync.WaitGroup
	wg.Add(1)
	s.ncon.Subscribe("reporter.result", func(m *nats.Msg) {
		var t reportercontract.ReporterResponse
		err := json.Unmarshal(m.Data, &t)
		s.Assert().NoError(err)
		s.Assert().Equal(jid, t.JobId)
		s.Assert().NotZero(t.Score)
		zap.S().Infow("Reporter results", "results", t.Score)
		wg.Done()
	})

	wg.Wait()

}

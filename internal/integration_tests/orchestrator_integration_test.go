package integrationtests

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/google/uuid"
	"github.com/guardlight/server/internal/essential/config"
	"github.com/guardlight/server/internal/essential/logging"
	"github.com/guardlight/server/internal/essential/testcontainers"
	"github.com/guardlight/server/internal/infrastructure/database"
	"github.com/guardlight/server/internal/infrastructure/messaging"
	"github.com/guardlight/server/internal/jobmanager"
	"github.com/guardlight/server/internal/natsclient"
	"github.com/guardlight/server/internal/orchestrator"
	"github.com/guardlight/server/internal/scheduler"
	"github.com/guardlight/server/pkg/analyzercontract"
	"github.com/guardlight/server/pkg/parsercontract"
	"github.com/guardlight/server/servers/natsmessaging"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TestSuiteOrchestratorIntegration struct {
	suite.Suite
	db         *gorm.DB
	jobManager *jobmanager.JobManager
}

func (s *TestSuiteOrchestratorIntegration) SetupSuite() {
	config.SetupConfig("../../testdata/envs/orchestrator.yaml")
	logging.SetupLogging("test")
	ctx, ctxCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer ctxCancel()

	sqlContainer, err := testcontainers.NewPostgresContainer(ctx)
	s.Require().NoError(err)

	conString, err := sqlContainer.ConnectionString(ctx)
	s.Require().NoError(err)

	zap.S().Infow("Connection string", "url", conString)
	s.db = database.InitDatabase(conString)

	jmr := jobmanager.NewJobManagerRepository(s.db)
	s.jobManager = jobmanager.NewJobMananger(jmr)

	sqlDb, _ := s.db.DB()
	fixtures, err := testfixtures.New(
		testfixtures.Database(sqlDb),
		testfixtures.Dialect("postgres"),
		testfixtures.FilesMultiTables(
			"../../testdata/fixtures/orchestrator.yaml",
		),
	)
	s.Assert().NoError(err)

	err = fixtures.Load()
	s.Assert().NoError(err)

	zap.S().Info("Setted up")
}

func TestOrchestratorSuiteRun(t *testing.T) {
	suite.Run(t, new(TestSuiteOrchestratorIntegration))
}

func (s *TestSuiteOrchestratorIntegration) TestOrchestratorSetup() {
	err := natsmessaging.NewNatsServer()
	s.Assert().NoError(err)

	ncon := messaging.InitNatsInProcess(natsmessaging.GetServer())
	loc, err := time.LoadLocation("Europe/Amsterdam")
	s.Assert().NoError(err)
	sch, err := scheduler.NewScheduler(loc)
	s.Assert().NoError(err)
	nc := natsclient.NewNatsClient(ncon)

	_, err = orchestrator.NewOrchestrator(s.jobManager, sch.Gos, nc)
	s.Assert().NoError(err)

	var wg sync.WaitGroup
	wg.Add(2)

	ncon.Subscribe("parser.freetext", func(m *nats.Msg) {
		var t parsercontract.ParserRequest
		err := json.Unmarshal(m.Data, &t)
		s.Assert().NoError(err)
		s.Assert().Equal(uuid.MustParse("a79c05d2-2e07-4431-9269-8f36338142b6"), t.JobId)
		wg.Done()
	})

	ncon.Subscribe("analyzer.word_search", func(m *nats.Msg) {
		var t analyzercontract.AnalyzerRequest
		err := json.Unmarshal(m.Data, &t)
		s.Assert().NoError(err)
		s.Assert().Equal(uuid.MustParse("e36d69cb-c795-4db2-9149-9277b348c3df"), t.JobId)
		wg.Done()
	})

	wg.Wait()
	sch.Gos.Shutdown()

}

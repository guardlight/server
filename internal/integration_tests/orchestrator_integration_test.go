package integrationtests

import (
	"context"
	"testing"
	"time"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/guardlight/server/internal/essential/config"
	"github.com/guardlight/server/internal/essential/logging"
	"github.com/guardlight/server/internal/essential/testcontainers"
	"github.com/guardlight/server/internal/infrastructure/database"
	"github.com/guardlight/server/internal/jobmanager"
	"github.com/guardlight/server/internal/natsclient"
	"github.com/guardlight/server/internal/orchestrator"
	"github.com/guardlight/server/internal/scheduler"
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
	config.SetupConfig("../../env-test.yaml")
	logging.SetupLogging("test")
	ctx, ctxCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer ctxCancel()

	csqlContainer, err := testcontainers.NewCockroachSQLContainer(ctx)
	s.Require().NoError(err)

	s.db = database.InitDatabase(csqlContainer.GetDSN())

	jmr := jobmanager.NewJobManagerRepository(s.db)
	s.jobManager = jobmanager.NewJobMananger(jmr)

	sqlDb, _ := s.db.DB()
	fixtures, err := testfixtures.New(
		testfixtures.Database(sqlDb),
		testfixtures.Dialect("postgres"),
		testfixtures.Files(
			"../../testdata/fixtures/jobs.yaml",
		),
		testfixtures.UseDropConstraint(),
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

	loc, err := time.LoadLocation("Europe/Amsterdam")
	s.Assert().NoError(err)
	sch, err := scheduler.NewScheduler(loc)
	s.Assert().NoError(err)
	nc := natsclient.NewNatsClient()

	_, err = orchestrator.NewOrchestrator(s.jobManager, sch.Gos, nc)
	s.Assert().NoError(err)

	time.Sleep(time.Second * 20)

	s.Assert().Equal("parser.freetext", natsclient.T)

}

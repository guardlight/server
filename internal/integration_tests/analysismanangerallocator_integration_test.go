package integrationtests

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/google/uuid"
	"github.com/guardlight/server/internal/analysismanager"
	"github.com/guardlight/server/internal/essential/config"
	"github.com/guardlight/server/internal/essential/logging"
	"github.com/guardlight/server/internal/essential/testcontainers"
	"github.com/guardlight/server/internal/infrastructure/database"
	"github.com/guardlight/server/internal/infrastructure/messaging"
	"github.com/guardlight/server/internal/jobmanager"
	"github.com/guardlight/server/pkg/parsercontract"
	"github.com/guardlight/server/servers/natsmessaging"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type TestSuiteAnalysisManagerAllocatorIntegration struct {
	suite.Suite
	db                        *gorm.DB
	analysisManagerRepository *analysismanager.AnalysisManagerRepository
	ncon                      *nats.Conn
}

func (sama *TestSuiteAnalysisManagerAllocatorIntegration) SetupSuite() {
	config.SetupConfig("../../testdata/envs/analysismanangerallocator.yaml")
	logging.SetupLogging("test")
	ctx, ctxCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer ctxCancel()
	csqlContainer, err := testcontainers.NewCockroachSQLContainer(ctx)
	sama.Require().NoError(err)

	sama.db = database.InitDatabase(csqlContainer.GetDSN())
	sama.db.Logger = logger.Default.LogMode(logger.Info)
	zap.S().Infow("connection details", "url", csqlContainer.GetDSN())

	jmr := jobmanager.NewJobManagerRepository(sama.db)
	jobManager := jobmanager.NewJobMananger(jmr)

	err = natsmessaging.NewNatsServer()
	sama.Require().NoError(err)
	sama.ncon = messaging.InitNats(natsmessaging.GetNatsUrl(), natsmessaging.GetServer())
	sama.analysisManagerRepository = analysismanager.NewAnalysisManagerRepository(sama.db)
	_ = analysismanager.NewAnalysisManagerAllocator(sama.ncon, sama.analysisManagerRepository, jobManager)

	sqlDb, err := sama.db.DB()
	sama.Require().NoError(err)
	fixtures, err := testfixtures.New(
		testfixtures.Database(sqlDb),
		testfixtures.Dialect("postgres"),
		testfixtures.FilesMultiTables(
			"../../testdata/fixtures/analysismanangerallocator.yaml",
		),
		testfixtures.UseDropConstraint(),
	)
	sama.Assert().NoError(err)

	err = fixtures.Load()
	sama.Assert().NoError(err)

	zap.S().Info("Setted up")

}

func TestAnalysisManangerAllocatorSuiteRun(t *testing.T) {
	suite.Run(t, new(TestSuiteAnalysisManagerAllocatorIntegration))
}

func (sama *TestSuiteAnalysisManagerAllocatorIntegration) TestParserResult() {
	ai := uuid.MustParse("7ffe69cc-7ba2-4500-aee6-1ab36be5ce10")

	pr := parsercontract.ParserResponse{
		JobId:      uuid.MustParse("fed9b891-a38d-41df-b7c5-cc0200726450"),
		AnalysisId: ai,
		Text:       "Running and Walking",
	}

	dat, err := json.Marshal(pr)
	sama.Assert().NoError(err)
	sama.ncon.Publish("parser.result", dat)
	time.Sleep(3 * time.Second)

	var rawDat analysismanager.RawData
	res := sama.db.Model(analysismanager.RawData{AnalysisRequestId: ai}).Find(&rawDat)
	sama.Assert().NoError(res.Error)
	sama.Assert().Equal("Running and Walking", rawDat.ProcessedText)

	var jobs []jobmanager.Job
	res = sama.db.Find(&jobs)
	sama.Assert().NoError(res.Error)
	sama.Assert().Len(jobs, 2)

}

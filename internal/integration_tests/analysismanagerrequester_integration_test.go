package integrationtests

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/google/uuid"
	"github.com/guardlight/server/internal/analysismanager"
	"github.com/guardlight/server/internal/essential/config"
	"github.com/guardlight/server/internal/essential/glsecurity"
	"github.com/guardlight/server/internal/essential/logging"
	"github.com/guardlight/server/internal/essential/testcontainers"
	"github.com/guardlight/server/internal/infrastructure/database"
	"github.com/guardlight/server/internal/jobmanager"
	"github.com/guardlight/server/internal/scheduler"
	"github.com/guardlight/server/internal/ssemanager"
	"github.com/guardlight/server/internal/theme"
	"github.com/guardlight/server/pkg/analysisrequest"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type TestSuiteAnalysisManagerIntegration struct {
	suite.Suite
	db                        *gorm.DB
	router                    *gin.Engine
	analysisManagerRepository *analysismanager.AnalysisManagerRepository
}

func (s *TestSuiteAnalysisManagerIntegration) SetupSuite() {
	config.SetupConfig("../../testdata/envs/analysismanangerequester.yaml")
	logging.SetupLogging("test")
	ctx, ctxCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer ctxCancel()
	sqlContainer, err := testcontainers.NewPostgresContainer(ctx)
	s.Require().NoError(err)

	conString, err := sqlContainer.ConnectionString(ctx)
	s.Require().NoError(err)
	s.db = database.InitDatabase(conString)
	s.db.Logger = logger.Default.LogMode(logger.Info)
	zap.S().Infow("connection details", "url", conString)

	s.router = gin.Default()

	loc, err := time.LoadLocation("Europe/Amsterdam")
	s.Assert().NoError(err)
	sch, err := scheduler.NewScheduler(loc)
	s.Assert().NoError(err)
	jmr := jobmanager.NewJobManagerRepository(s.db)
	jobManager := jobmanager.NewJobMananger(jmr, sch.Gos)
	ssem := ssemanager.NewSseMananger()

	s.analysisManagerRepository = analysismanager.NewAnalysisManagerRepository(s.db)
	tsr := theme.NewThemeRepository(s.db)

	ts := theme.NewThemeService(tsr)
	ars := analysismanager.NewAnalysisResultService(s.analysisManagerRepository, s.analysisManagerRepository, ts)

	analysisManangerRequester := analysismanager.NewAnalysisManangerRequester(jobManager, s.analysisManagerRepository, ssem, ts)

	analysismanager.NewAnalysisRequestController(s.router.Group(""), analysisManangerRequester, ars)

	sqlDb, _ := s.db.DB()
	fixtures, err := testfixtures.New(
		testfixtures.Database(sqlDb),
		testfixtures.Dialect("postgres"),
		testfixtures.Files(),
		testfixtures.UseDropConstraint(),
	)
	s.Assert().NoError(err)

	err = fixtures.Load()
	s.Assert().NoError(err)

	zap.S().Info("Setted up")

}

func TestAnalysisManangerSuiteRun(t *testing.T) {
	suite.Run(t, new(TestSuiteAnalysisManagerIntegration))
}

func (s *TestSuiteAnalysisManagerIntegration) TestSubmitAnalysisRequestUntilParseJob() {
	ui := uuid.MustParse("be7954d2-9c1b-4e96-8605-14a11af397c2")

	tkStr, err := glsecurity.MakeJwtTokenForCompanion(glsecurity.UserTokenCredentials{
		UserId: ui,
		Role:   glsecurity.Admin,
	})
	s.Assert().NoError(err)

	ar := &analysisrequest.AnalysisRequest{
		Title:       "test analysis",
		ContentType: analysisrequest.MOVIE,
		File: analysisrequest.File{
			Content:  base64.StdEncoding.EncodeToString([]byte("Running and walking")),
			Mimetype: "freetext",
		},
		Themes: []analysisrequest.Theme{
			{
				Title: "Test Theme",
				Id:    uuid.MustParse("2864d1b0-411a-4c6c-932a-61acddd67019"),
				Analyzers: []analysisrequest.Analyzer{
					{
						Key: "word_search",
						Inputs: []analysisrequest.AnalyzerInput{
							{
								Key:   "strict_words",
								Value: "Running, Walking",
							},
						},
					},
				},
			},
		},
	}

	jsonValue, err := json.Marshal(ar)
	s.Assert().NoError(err)

	req, err := http.NewRequest("POST", "/analysis/request", bytes.NewBuffer(jsonValue))
	s.Assert().NoError(err)
	req.Header.Set("Authorization", "Bearer "+tkStr)
	req.AddCookie(&http.Cookie{
		Name:     glsecurity.ConsoleApiCookieName,
		Value:    tkStr,
		Path:     "/",
		Domain:   "127.0.0.1",
		MaxAge:   604800,
		Expires:  time.Now().Add(time.Hour * 1),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   true,
	})

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Assert().Equal(http.StatusNoContent, w.Code)

	var allArs []analysismanager.AnalysisRequest
	s.db.Preload("RawData").Preload("AnalysisRequestSteps").Find(&allArs)
	s.Assert().Len(allArs, 1)

	var allJobs []jobmanager.Job
	s.db.Find(&allJobs)
	s.Assert().Len(allJobs, 1)
	zap.S().Infow("jobs", "job", allJobs)

	var allAnalysis []analysismanager.Analysis
	s.db.Find(&allAnalysis)
	s.Assert().Len(allAnalysis, 1)
	zap.S().Infow("analysis", "analysis", allAnalysis)

}

func (s *TestSuiteAnalysisManagerIntegration) TestSubmitAnalysisRequestUntilParseJobEpub() {
	ui := uuid.MustParse("be7954d2-9c1b-4e96-8605-14a11af397c2")

	tkStr, err := glsecurity.MakeJwtTokenForCompanion(glsecurity.UserTokenCredentials{
		UserId: ui,
		Role:   glsecurity.Admin,
	})
	s.Assert().NoError(err)

	epubFile, err := os.ReadFile("alice.epub") //read the content of file
	s.Assert().NoError(err)

	ar := &analysisrequest.AnalysisRequest{
		Title:       "test analysis",
		ContentType: analysisrequest.MOVIE,
		File: analysisrequest.File{
			Content:  base64.StdEncoding.EncodeToString(epubFile),
			Mimetype: "freetext",
		},
		Themes: []analysisrequest.Theme{
			{
				Title: "Test Theme",
				Id:    uuid.MustParse("2864d1b0-411a-4c6c-932a-61acddd67019"),
				Analyzers: []analysisrequest.Analyzer{
					{
						Key: "word_search",
						Inputs: []analysisrequest.AnalyzerInput{
							{
								Key:   "strict_words",
								Value: "Running, Walking",
							},
						},
					},
				},
			},
		},
	}

	jsonValue, err := json.Marshal(ar)
	s.Assert().NoError(err)

	req, err := http.NewRequest("POST", "/analysis/request", bytes.NewBuffer(jsonValue))
	s.Assert().NoError(err)
	req.Header.Set("Authorization", "Bearer "+tkStr)
	req.AddCookie(&http.Cookie{
		Name:     glsecurity.ConsoleApiCookieName,
		Value:    tkStr,
		Path:     "/",
		Domain:   "127.0.0.1",
		MaxAge:   604800,
		Expires:  time.Now().Add(time.Hour * 1),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   true,
	})

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Assert().Equal(http.StatusNoContent, w.Code)

	var allArs []analysismanager.AnalysisRequest
	s.db.Preload("RawData").Preload("AnalysisRequestSteps").Find(&allArs)
	s.Assert().Len(allArs, 1)

	var allJobs []jobmanager.Job
	s.db.Find(&allJobs)
	s.Assert().Len(allJobs, 1)
	zap.S().Infow("jobs", "job", allJobs)

	var allAnalysis []analysismanager.Analysis
	s.db.Find(&allAnalysis)
	s.Assert().Len(allAnalysis, 1)
	zap.S().Infow("analysis", "analysis", allAnalysis)

}

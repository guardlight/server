package integrationtests

import (
	"bytes"
	"context"
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
	"github.com/guardlight/server/internal/api"
	"github.com/guardlight/server/internal/api/analysisapi"
	"github.com/guardlight/server/internal/essential/config"
	"github.com/guardlight/server/internal/essential/glsecurity"
	"github.com/guardlight/server/internal/essential/logging"
	"github.com/guardlight/server/internal/essential/testcontainers"
	"github.com/guardlight/server/internal/infrastructure/database"
	"github.com/guardlight/server/internal/infrastructure/messaging"
	"github.com/guardlight/server/internal/jobmanager"
	"github.com/guardlight/server/internal/natsclient"
	"github.com/guardlight/server/internal/orchestrator"
	"github.com/guardlight/server/internal/scheduler"
	"github.com/guardlight/server/pkg/analysisrequest"
	"github.com/guardlight/server/pkg/gladapters/analyzers"
	"github.com/guardlight/server/pkg/gladapters/parsers"
	"github.com/guardlight/server/servers/natsmessaging"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TestSuiteMainIntegration struct {
	suite.Suite
	db     *gorm.DB
	router *gin.Engine
}

func (s *TestSuiteMainIntegration) SetupSuite() {
	config.SetupConfig("../../testdata/envs/main.yaml")
	logging.SetupLogging("test")
	ctx, ctxCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer ctxCancel()

	csqlContainer, err := testcontainers.NewCockroachSQLContainer(ctx)
	s.Require().NoError(err)

	s.db = database.InitDatabase(csqlContainer.GetDSN())

	sqlDb, _ := s.db.DB()
	fixtures, err := testfixtures.New(
		testfixtures.Database(sqlDb),
		testfixtures.Dialect("postgres"),
		testfixtures.FilesMultiTables(),
		testfixtures.UseDropConstraint(),
	)
	s.Assert().NoError(err)

	err = fixtures.Load()
	s.Assert().NoError(err)

	loc, err := time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		zap.S().Errorw("Could not load timezone", "error", err)
		s.Assert().NoError(err)
	}

	// Repositories
	jmr := jobmanager.NewJobManagerRepository(s.db)
	amr := analysismanager.NewAnalysisManagerRepository(s.db)

	// Controller Groups
	s.router = api.NewRouter(logging.GetLogger())
	baseGroup := s.router.Group("")

	err = natsmessaging.NewNatsServer()
	s.Assert().NoError(err)
	ncon := messaging.InitNats(natsmessaging.GetNatsUrl(), natsmessaging.GetServer())

	parsers.NewFreetextParser(ncon)
	analyzers.NewWordsearchAnalyzer(ncon)

	// Services
	nc := natsclient.NewNatsClient(ncon)
	jm := jobmanager.NewJobMananger(jmr)
	sch, err := scheduler.NewScheduler(loc)
	if err != nil {
		zap.S().Errorw("Could not create scheduler", "error", err)
		s.Assert().NoError(err)
	}
	_, err = orchestrator.NewOrchestrator(jm, sch.Gos, nc)
	if err != nil {
		zap.S().Errorw("Could not create orhestrator", "error", err)
		s.Assert().NoError(err)
	}
	am := analysismanager.NewAnalysisManangerRequester(jm, amr)
	_ = analysismanager.NewAnalysisManagerAllocator(ncon, amr, jm)

	// Controllers
	analysisapi.NewAnalysisRequestController(baseGroup, am)

	// Start the server
	go api.LiveOrLetDie(s.router)

	zap.S().Info("Setted up")
}

func TestMainSuiteRun(t *testing.T) {
	suite.Run(t, new(TestSuiteMainIntegration))
}

func (s *TestSuiteMainIntegration) TestRequestTillResult() {
	uid := uuid.MustParse("be7954d2-9c1b-4e96-8605-14a11af397c2")

	tkStr, err := glsecurity.MakeJwtTokenForCompanion(glsecurity.UserTokenCredentials{
		UserId: uid,
		Role:   glsecurity.Admin,
	})
	s.Assert().NoError(err)

	data, err := os.ReadFile("../../testdata/epubs/lion-parsed.txt")
	s.Assert().NoError(err)

	ar := &analysisrequest.AnalysisRequest{
		Title:       "test analysis",
		ContentType: analysisrequest.BOOK,
		File: analysisrequest.File{
			Content:  data,
			Mimetype: "freetext",
		},
		Themes: []analysisrequest.Theme{
			{
				Title: "Test Theme",
				Id:    uuid.MustParse("2864d1b0-411a-4c6c-932a-61acddd67019"),
				Analyzers: []analysisrequest.Analyzer{
					{
						Key:       "word_search",
						Threshold: 2,
						Inputs: []analysisrequest.AnalyzerInput{
							{
								Key:   "strict_words",
								Value: "Magic, I do",
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

	var arResult analysismanager.AnalysisRequest
	s.Eventually(func() bool {
		err := s.db.First(&arResult).Error
		s.Assert().NoError(err)
		return arResult.Id != uuid.Nil
	}, 30*time.Second, time.Second, "No record found in wait duration")

	reqResult, err := http.NewRequest("GET", "/analysis/analyses", nil)
	s.Assert().NoError(err)
	reqResult.Header.Set("Authorization", "Bearer "+tkStr)
	reqResult.AddCookie(&http.Cookie{
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

	wResult := httptest.NewRecorder()
	s.router.ServeHTTP(wResult, reqResult)
	s.Assert().Equal(http.StatusOK, wResult.Code)

	var arReqs []analysismanager.AnalysisRequest
	err = json.Unmarshal(wResult.Body.Bytes(), &arReqs)
	s.Assert().NoError(err)
	s.Assert().Len(arReqs, 1)
	s.Assert().Equal(arReqs[0].Id, arResult.Id)

	reqResultSpecific, err := http.NewRequest("GET", "/analysis/analyses/"+arResult.Id.String(), nil)
	s.Assert().NoError(err)
	reqResultSpecific.Header.Set("Authorization", "Bearer "+tkStr)
	reqResultSpecific.AddCookie(&http.Cookie{
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

	wResultSpecific := httptest.NewRecorder()
	s.router.ServeHTTP(wResultSpecific, reqResultSpecific)
	s.Assert().Equal(http.StatusOK, wResultSpecific.Code)

	arReqSpecific := &analysismanager.AnalysisRequest{}
	json.Unmarshal(wResultSpecific.Body.Bytes(), arReqSpecific)
	s.Assert().Equal(arReqSpecific.Id, arResult.Id)
}

package analysismanager

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/guardlight/server/internal/essential/glerror"
	"github.com/guardlight/server/internal/essential/glsecurity"
	"github.com/guardlight/server/pkg/analysisrequest"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

type AnalysisRequestController struct {
	manager *AnalysisManagerRequester
	ars     *AnalysisResultService
}

func NewAnalysisRequestController(group *gin.RouterGroup, manager *AnalysisManagerRequester, ars *AnalysisResultService) *AnalysisRequestController {
	arc := &AnalysisRequestController{
		manager: manager,
		ars:     ars,
	}

	analysisGroup := group.Group("analysis")
	analysisGroup.Use(glsecurity.UseGuardlightAuth())
	analysisGroup.POST("", arc.analysisRequest)
	analysisGroup.GET("", arc.analyses)
	analysisGroup.GET("/:arid", arc.analysisById)

	return arc
}

func (arc *AnalysisRequestController) analysisRequest(c *gin.Context) {
	ar := &analysisrequest.AnalysisRequest{}
	err := glsecurity.ReuseBindAndValidate(c, ar)
	if err != nil {
		zap.S().Errorw("error validating analysis request", "error", err)
		c.JSON(glerror.BadRequestError())
		return
	}

	ui := glsecurity.GetUserIdFromContextParsed(c)

	err = arc.manager.RequestAnalysis(ar, ui)
	if err != nil {
		zap.S().Errorw("error creating analysis request", "error", err)
		switch err {
		case ErrInvalidAnalyzer:
		case ErrInvalidParser:
			c.JSON(glerror.BadRequestError())
			return
		default:
			c.JSON(glerror.InternalServerError())
			return
		}
	}

	c.JSON(http.StatusNoContent, gin.H{})
}

func (arc *AnalysisRequestController) analyses(c *gin.Context) {
	uid := glsecurity.GetUserIdFromContextParsed(c)

	pgq := c.Query("page")
	pgNr := lo.If(pgq == "", 0).ElseF(func() int {
		page, err := strconv.Atoi(pgq)
		if err != nil {
			return 0
		}
		return page
	})

	pgl := c.Query("limit")
	pgLim := lo.If(pgl == "", 0).ElseF(func() int {
		lim, err := strconv.Atoi(pgl)
		if err != nil {
			return 0
		}
		return lim
	})

	ars, err := arc.ars.GetAnalysesByUserId(uid, pgLim, pgNr)
	if err != nil {
		zap.S().Errorw("error get analyses", "error", err)
		c.JSON(glerror.InternalServerError())
		return
	}

	c.JSON(http.StatusOK, ars)
}

func (arc *AnalysisRequestController) analysisById(c *gin.Context) {
	uid := glsecurity.GetUserIdFromContextParsed(c)
	arid, err := uuid.Parse(c.Param("arid"))
	if err != nil {
		zap.S().Errorw("Analysis Request id is not uuid", "error", err)
		c.JSON(glerror.BadRequestError())
		return
	}

	ars, err := arc.ars.GetAnalysesByAnalysisIdAndUserId(uid, arid)
	if err != nil {
		zap.S().Errorw("error get analyses", "error", err)
		c.JSON(glerror.InternalServerError())
		return
	}

	c.JSON(http.StatusOK, ars)
}

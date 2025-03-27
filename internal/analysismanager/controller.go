package analysismanager

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/guardlight/server/internal/essential/glerror"
	"github.com/guardlight/server/internal/essential/glsecurity"
	"github.com/guardlight/server/pkg/analysisrequest"
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
	analysisGroup.POST("request", arc.analysisRequest)
	analysisGroup.GET("analyses", arc.analyses)

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

	ars, err := arc.ars.GetAnalysesByUserId(uid)
	if err != nil {
		zap.S().Errorw("error get analyses", "error", err)
		c.JSON(glerror.InternalServerError())
		return
	}

	c.JSON(http.StatusOK, ars)
}

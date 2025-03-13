package analysisapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/guardlight/server/internal/analysismanager"
	"github.com/guardlight/server/internal/essential/glerror"
	"github.com/guardlight/server/internal/essential/glsecurity"
	"github.com/guardlight/server/pkg/analysisrequest"
	"go.uber.org/zap"
)

type AnalysisRequestController struct {
	manager *analysismanager.AnalysisManagerRequester
}

func NewAnalysisRequestController(group *gin.RouterGroup, manager *analysismanager.AnalysisManagerRequester) *AnalysisRequestController {
	arc := &AnalysisRequestController{
		manager: manager,
	}

	analysisGroup := group.Group("analysis")
	analysisGroup.Use(glsecurity.UseGuardlightAuth())
	analysisGroup.POST("request", arc.analysisRequest)

	return arc
}

func (arc AnalysisRequestController) analysisRequest(c *gin.Context) {
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
		case analysismanager.ErrInvalidAnalyzer:
		case analysismanager.ErrInvalidParser:
			c.JSON(glerror.BadRequestError())
			return
		default:
			c.JSON(glerror.InternalServerError())
			return
		}
	}

	c.JSON(http.StatusNoContent, gin.H{})
}

package analysismanager

import (
	"net/http"
	"strconv"
	"strings"

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

	analysisGroupLoom := analysisGroup.Group("dataloom")
	analysisGroupLoom.Use(glsecurity.UseGuardlightAuthApiKey())
	analysisGroupLoom.POST("", arc.analysisRequestDataloom)

	analysisGroup.Use(glsecurity.UseGuardlightAuth())
	analysisGroup.POST("", arc.analysisRequest)
	analysisGroup.GET("", arc.analyses)
	analysisGroup.GET("/:arid", arc.analysisById)
	analysisGroup.DELETE("/:arid", arc.deleteAnalysisRequestById)
	analysisGroup.POST("/update/score", arc.updateAnalysisScore)

	return arc
}

func (arc *AnalysisRequestController) analysisRequestDataloom(c *gin.Context) {
	ard := &analysisrequest.AnalysisRequestDataloom{}
	err := glsecurity.ReuseBindAndValidate(c, ard)
	if err != nil {
		zap.S().Errorw("error validating analysis request", "error", err)
		c.JSON(glerror.BadRequestError())
		return
	}

	ui := glsecurity.GetUserIdFromContextParsed(c)

	aid, err := arc.manager.RequestAnalysisDataloom(ard, ui)
	if err != nil {
		if err == ErrHashAlreadyExist {
			c.JSON(http.StatusOK, analysisrequest.AnalysisRequestResponse{
				Id: aid,
			})
			return
		}
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

	c.JSON(http.StatusOK, analysisrequest.AnalysisRequestResponse{
		Id: aid,
	})

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

	aid, err := arc.manager.RequestAnalysis(ar, ui, string(RequestOriginUser))
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

	c.JSON(http.StatusOK, analysisrequest.AnalysisRequestResponse{
		Id: aid,
	})
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

	cat := c.Query("category")
	pgCatType := lo.If(pgl == "", "").ElseF(func() string {
		var catSpl = strings.Split(cat, ":")
		if len(catSpl) == 2 {
			return catSpl[0]
		} else {
			return ""
		}
	})
	pgCatCat := lo.If(pgl == "", "").ElseF(func() string {
		var catSpl = strings.Split(cat, ":")
		if len(catSpl) == 2 {
			return catSpl[1]
		} else {
			return ""
		}
	})

	pgQuery := c.Query("query")

	ars, err := arc.ars.GetAnalysesByUserId(uid, pgLim, pgNr, pgCatType, pgCatCat, pgQuery)
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

func (arc *AnalysisRequestController) deleteAnalysisRequestById(c *gin.Context) {
	uid := glsecurity.GetUserIdFromContextParsed(c)
	arid, err := uuid.Parse(c.Param("arid"))
	if err != nil {
		zap.S().Errorw("Analysis Request id is not uuid", "error", err)
		c.JSON(glerror.BadRequestError())
		return
	}

	err = arc.ars.DeleteAnalysisRequestById(arid, uid)
	if err != nil {
		zap.S().Errorw("error deleting analysis request", "error", err)
		c.JSON(glerror.InternalServerError())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func (arc *AnalysisRequestController) updateAnalysisScore(c *gin.Context) {
	_ = glsecurity.GetUserIdFromContextParsed(c)

	aus := &analysisrequest.AnalysisUpdateScore{}
	err := glsecurity.ReuseBindAndValidate(c, aus)
	if err != nil {
		zap.S().Errorw("error validating analysis request", "error", err)
		c.JSON(glerror.BadRequestError())
		return
	}

	err = arc.ars.UpdateScore(aus.Id, aus.Score)
	if err != nil {
		zap.S().Errorw("error get analyses", "error", err)
		c.JSON(glerror.InternalServerError())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

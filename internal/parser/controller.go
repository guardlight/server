package parser

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/guardlight/server/internal/essential/config"
	"github.com/guardlight/server/internal/essential/glsecurity"
	"github.com/samber/lo"
)

type ParserController struct {
}

func NewParserController(group *gin.RouterGroup) *ParserController {
	tc := &ParserController{}

	analysisGroup := group.Group("parser")
	analysisGroup.Use(glsecurity.UseGuardlightAuth())
	analysisGroup.GET("", tc.getAllParsers)

	return tc
}

func (pc *ParserController) getAllParsers(c *gin.Context) {
	prs := lo.Map(config.Get().Parsers, func(p config.Parser, _ int) Parser {
		return Parser{
			Name:        p.Name,
			Type:        p.Type,
			Description: p.Description,
		}
	})

	c.JSON(http.StatusOK, prs)
}

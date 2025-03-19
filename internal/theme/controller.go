package theme

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/guardlight/server/internal/essential/glerror"
	"github.com/guardlight/server/internal/essential/glsecurity"
	"go.uber.org/zap"
)

type ThemeController struct {
	s *ThemeService
}

func NewThemeController(group *gin.RouterGroup, s *ThemeService) *ThemeController {
	tc := &ThemeController{
		s: s,
	}

	analysisGroup := group.Group("theme")
	analysisGroup.Use(glsecurity.UseGuardlightAuth())
	analysisGroup.PUT("", tc.updateTheme)
	analysisGroup.GET("", tc.getAllThemes)

	return tc
}

func (tc *ThemeController) updateTheme(c *gin.Context) {
	uid := glsecurity.GetUserIdFromContextParsed(c)

	tDto := ThemeDto{}
	err := glsecurity.ReuseBindAndValidate(c, &tDto)
	if err != nil {
		zap.S().Errorw("error validating theme", "error", err)
		c.JSON(glerror.BadRequestError())
		return
	}

	err = tc.s.updateTheme(tDto, uid)
	if err != nil {
		zap.S().Errorw("error updating theme", "error", err)
		c.JSON(glerror.InternalServerError())
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}

func (tc *ThemeController) getAllThemes(c *gin.Context) {
	uid := glsecurity.GetUserIdFromContextParsed(c)

	t, err := tc.s.getAllThemesByUserId(uid)
	if err != nil {
		zap.S().Errorw("error gettings themes", "error", err)
		c.JSON(glerror.InternalServerError())
		return
	}

	c.JSON(http.StatusOK, t)
}

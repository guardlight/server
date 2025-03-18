package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/guardlight/server/internal/essential/config"
	"github.com/guardlight/server/internal/essential/glerror"
	"github.com/guardlight/server/internal/essential/glsecurity"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

type AuthenticationController struct{}

func NewAuthenticationController(group *gin.RouterGroup) *AuthenticationController {
	ac := &AuthenticationController{}

	authGroup := group.Group("auth")
	authGroup.POST("signin", ac.signIn)

	return ac
}

func (ac *AuthenticationController) signIn(c *gin.Context) {

	ar := &AuthenticationRequest{}
	err := glsecurity.ReuseBindAndValidate(c, ar)
	if err != nil {
		zap.S().Errorw("error validating authentication request", "error", err)
		c.JSON(glerror.BadRequestError())
		return
	}

	u, ok := lo.Find(config.Get().Users, func(u config.User) bool {
		return u.Username == ar.Username
	})
	if !ok {
		zap.S().Errorw("error authorizing user", "error", err)
		c.JSON(glerror.UnauthorizedError())
		return
	}

	tkStr, err := glsecurity.MakeJwtTokenForCompanion(glsecurity.UserTokenCredentials{
		UserId: u.Id,
		Role:   glsecurity.Admin,
	})
	if err != nil {
		zap.S().Errorw("error authorizing user", "error", err)
		c.JSON(glerror.UnauthorizedError())
		return
	}

	c.SetCookie(glsecurity.ConsoleApiCookieName, tkStr, config.Get().Console.Jwt.MaxAge, "/", config.Get().Domain, config.Get().IsProduction(), true)
	c.JSON(http.StatusOK, gin.H{"authenticationStatus": "Authenticated"})
}

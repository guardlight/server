package auth

import "github.com/gin-gonic/gin"

type AuthenticationController struct{}

func NewAuthenticationController(group *gin.RouterGroup) *AuthenticationController {
	ac := &AuthenticationController{}

	authGroup := group.Group("auth")
	authGroup.POST("signin", ac.signIn)

	return ac
}

func (ac *AuthenticationController) signIn(c *gin.Context) {}

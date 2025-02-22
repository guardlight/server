package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthController struct{}

func NewHealthController(group *gin.RouterGroup) {
	controller := &HealthController{}
	group.GET("/health", controller.Status)
}

func (h HealthController) Status(c *gin.Context) {
	c.String(http.StatusOK, "Up!")
}

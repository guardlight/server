package ssemanager

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/guardlight/server/internal/essential/config"
	"github.com/guardlight/server/internal/essential/glsecurity"
	"go.uber.org/zap"
)

type sseController struct {
	ssem *SseManager
}

func NewSseController(group *gin.RouterGroup, ssem *SseManager) {
	ssec := &sseController{
		ssem: ssem,
	}

	sseGroup := group.Group("events")
	sseGroup.Use(glsecurity.UseGuardlightAuth())
	sseGroup.GET("", ssec.events)

}
func (sc sseController) events(c *gin.Context) {
	uid := glsecurity.GetUserIdFromContextParsed(c)

	// Set headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Access-Control-Allow-Origin", config.Get().Cors.Origin)

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		http.Error(c.Writer, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	conn := &SSEConnection{
		writer:  c.Writer,
		flusher: flusher,
	}

	// Save connection
	sc.ssem.Lock()
	sc.ssem.connections[uid] = conn
	zap.S().Infow("User connected", "user_id", uid)
	sc.ssem.Unlock()

	// Keep connection open
	notify := c.Writer.CloseNotify()
	<-notify // wait until client disconnects

	// Cleanup on disconnect
	sc.ssem.Lock()
	delete(sc.ssem.connections, uid)
	sc.ssem.Unlock()
	zap.S().Infow("User disconnected", "user_id", uid)
}

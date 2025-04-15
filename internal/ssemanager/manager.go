package ssemanager

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SSEConnection struct {
	writer  http.ResponseWriter
	flusher http.Flusher
	mutex   sync.Mutex // to avoid concurrent writes to the same connection
}

type SseManager struct {
	sync.RWMutex
	connections map[uuid.UUID]*SSEConnection
}

// Initialize event and Start procnteessing requests
func NewSseMananger() (event *SseManager) {
	ssem := &SseManager{
		connections: make(map[uuid.UUID]*SSEConnection),
	}

	// We are streaming current time to clients in the interval 10 seconds
	go func() {
		for {
			time.Sleep(time.Second * 10)
			now := time.Now().Format("2006-01-02 15:04:05")
			currentTime := fmt.Sprintf("The Current Time Is %v", now)

			for k := range ssem.connections {
				ssem.SendEvent(k, SseEvent{
					Type:   "heartbeat",
					Action: "beat",
					Data:   currentTime,
				})
			}

		}
	}()

	return ssem
}

func (sc *SseManager) SendEvent(userId uuid.UUID, e SseEvent) {
	sc.RLock()
	conn, exists := sc.connections[userId]
	sc.RUnlock()

	if !exists {
		zap.S().Infow("Connection not active", "user_id", userId.String())
		return
	}

	// Send the event
	conn.mutex.Lock()
	defer conn.mutex.Unlock()

	message, err := json.Marshal(e)
	if err != nil {
		zap.S().Errorw("Could not marshal event message", "error", err)
		return
	}

	if _, err = fmt.Fprint(conn.writer, "data: "+string(message)+"\n\n"); err != nil {
		zap.S().Errorw("Error printing to writer", "error", err)
	}
	conn.flusher.Flush()
	zap.S().Infow("Sent sse message", "user_id", userId.String(), "message", string(message))
}

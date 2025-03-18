package natsclient

import (
	"encoding/json"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type Publisher interface {
	Publish(topic string, data interface{}) error
}

var (
	T string = ""
)

type NatsClient struct {
	n *nats.Conn
}

func NewNatsClient(ncon *nats.Conn) *NatsClient {
	return &NatsClient{
		n: ncon,
	}
}

func (nc *NatsClient) Publish(topic string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		zap.S().Errorw("error marshalling request", "error", err)
		return err
	}
	nc.n.Publish(topic, data)
	zap.S().Infow("Published Data", "topic", topic)
	return nil
}

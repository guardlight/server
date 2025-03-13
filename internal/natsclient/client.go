package natsclient

type Sender interface {
	Send(topic string) error
}

var (
	T string = ""
)

type NatsClient struct{}

func NewNatsClient() *NatsClient {
	return &NatsClient{}
}

func (nc NatsClient) Send(topic string) error {
	T = topic

	return nil
}

package eventer

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"
)

type Srv interface {
	Publish(ctx context.Context, sub string, data interface{}) error
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context)
	GetClient() nats.JetStreamContext
}

type srv struct {
	nc *nats.Conn
	js nats.JetStreamContext

	streamUrl string
}

func New(streamUrl string) Srv {
	return &srv{
		streamUrl: streamUrl,
	}
}

func (s *srv) Connect(ctx context.Context) error {
	nc, err := nats.Connect(s.streamUrl)
	if err != nil {
		return err
	}
	js, err := nc.JetStream()
	if err != nil {
		return err
	}
	s.nc = nc
	s.js = js
	return nil
}

func (s *srv) Disconnect(ctx context.Context) {
	s.nc.Close()
}

func (s *srv) Publish(ctx context.Context, sub string, data interface{}) error {
	p, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return s.nc.Publish(sub, p)
}

func (s *srv) GetClient() nats.JetStreamContext {
	return s.js
}

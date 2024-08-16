package eventer

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"
)

type Srv struct {
	nc *nats.Conn
	js nats.JetStreamContext

	streamUrl string
}

func New(streamUrl string) *Srv {
	return &Srv{
		streamUrl: streamUrl,
	}
}

func (s *Srv) Connect(ctx context.Context) error {
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

func (s *Srv) Disconnect(ctx context.Context) error {
	s.nc.Close()
	return nil
}

func (s *Srv) Publish(ctx context.Context, sub string, data interface{}) error {
	p, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return s.nc.Publish(sub, p)
}

func (s *Srv) GetClient() nats.JetStreamContext {
	return s.js
}

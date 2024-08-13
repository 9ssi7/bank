package eventstream

import (
	"context"

	"github.com/9ssi7/bank/pkg/events"
	"github.com/9ssi7/bank/pkg/server"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/trace"
)

type srv struct {
	nc  *nats.Conn
	js  nats.JetStreamContext
	cnf Config

	subs []*nats.Subscription
}

type Config struct {
	StreamUrl string
	Tracer    trace.Tracer

	AuthStartLoginHandler   events.Handler
	UserCreatedHandler      events.Handler
	TransferIncomeHandler   events.Handler
	TransferOutgoingHandler events.Handler
}

func New(cnf Config) (server.Listener, error) {
	nc, err := nats.Connect(cnf.StreamUrl)
	if err != nil {
		return nil, err
	}
	js, err := nc.JetStream()
	if err != nil {
		return nil, err
	}
	return &srv{
		nc:  nc,
		js:  js,
		cnf: cnf,

		subs: make([]*nats.Subscription, 0),
	}, nil
}

func (s *srv) Listen() error {
	ctx := context.Background()
	err := s.addSub(
		ctx,
		eHandler{"Auth.LoginStarted", s.cnf.AuthStartLoginHandler},
		eHandler{"User.Created", s.cnf.UserCreatedHandler},
		eHandler{"Transfer.Incoming", s.cnf.TransferIncomeHandler},
		eHandler{"Transfer.Outgoing", s.cnf.TransferOutgoingHandler},
	)
	if err != nil {
		return err
	}
	return nil
}

type eHandler struct {
	subject string
	handler events.Handler
}

func (s *srv) addSub(ctx context.Context, handlers ...eHandler) error {
	for _, p := range handlers {
		sub, err := events.RegisterHandler(ctx, s.cnf.Tracer, s.js, p.subject, p.handler)
		if err != nil {
			return err
		}
		s.subs = append(s.subs, sub)
	}
	return nil
}

func (s *srv) Shutdown(ctx context.Context) error {
	for _, sub := range s.subs {
		sub.Unsubscribe()
	}
	s.nc.Close()
	return nil
}

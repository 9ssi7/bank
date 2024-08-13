package eventstream

import (
	"context"

	"github.com/9ssi7/bank/internal/domain/account"
	"github.com/9ssi7/bank/internal/domain/auth"
	"github.com/9ssi7/bank/internal/domain/user"
	"github.com/9ssi7/bank/internal/infra/eventer"
	"github.com/9ssi7/bank/pkg/events"
	"github.com/9ssi7/bank/pkg/server"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/trace"
)

type srv struct {
	eventer eventer.Srv
	cnf     Config

	subs []*nats.Subscription
}

type Config struct {
	Eventer eventer.Srv
	Tracer  trace.Tracer

	AuthStartLoginHandler   events.Handler
	UserCreatedHandler      events.Handler
	TransferIncomeHandler   events.Handler
	TransferOutgoingHandler events.Handler
}

func New(cnf Config) server.Listener {
	return &srv{
		eventer: cnf.Eventer,
		cnf:     cnf,
		subs:    make([]*nats.Subscription, 0),
	}
}

func (s *srv) Listen() error {
	ctx := context.Background()
	err := s.addSub(
		ctx,
		eventHandler{auth.SubjectLoginStarted, s.cnf.AuthStartLoginHandler},
		eventHandler{user.SubjectCreated, s.cnf.UserCreatedHandler},
		eventHandler{account.SubjectTransferIncoming, s.cnf.TransferIncomeHandler},
		eventHandler{account.SubjectTransferOutgoing, s.cnf.TransferOutgoingHandler},
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *srv) Shutdown(ctx context.Context) error {
	for _, sub := range s.subs {
		sub.Unsubscribe()
	}
	s.eventer.Disconnect(ctx)
	return nil
}

type eventHandler struct {
	subject string
	handler events.Handler
}

func (s *srv) addSub(ctx context.Context, handlers ...eventHandler) error {
	for _, p := range handlers {
		sub, err := events.RegisterHandler(ctx, s.cnf.Tracer, s.eventer.GetClient(), p.subject, p.handler)
		if err != nil {
			return err
		}
		s.subs = append(s.subs, sub)
	}
	return nil
}

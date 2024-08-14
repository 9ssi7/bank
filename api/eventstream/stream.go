package eventstream

import (
	"context"
	"time"

	"github.com/9ssi7/bank/internal/domain/account"
	"github.com/9ssi7/bank/internal/domain/auth"
	"github.com/9ssi7/bank/internal/domain/user"
	"github.com/9ssi7/bank/internal/eventhandler"
	"github.com/9ssi7/bank/internal/infra/eventer"
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

	AuthHandler    *eventhandler.AuthHandler
	AccountHandler *eventhandler.AccountHandler
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
		eventHandler{auth.SubjectLoginStarted, s.cnf.AuthHandler.OnLoginStart},
		eventHandler{user.SubjectCreated, s.cnf.AuthHandler.OnUserCreated},
		eventHandler{account.SubjectTransferIncoming, s.cnf.AccountHandler.OnTransferIncome},
		eventHandler{account.SubjectTransferOutgoing, s.cnf.AccountHandler.OnTransferOutcome},
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
	handler func(ctx context.Context, msg *nats.Msg) error
}

func (s *srv) addSub(ctx context.Context, handlers ...eventHandler) error {
	for _, p := range handlers {
		sub, err := s.eventer.GetClient().Subscribe(p.subject, func(msg *nats.Msg) {
			ctx, span := s.cnf.Tracer.Start(ctx, msg.Subject, trace.WithTimestamp(time.Now()))
			defer span.End()
			err := p.handler(ctx, msg)
			if err != nil {
				span.RecordError(err)
			}
		})
		if err != nil {
			return err
		}
		s.subs = append(s.subs, sub)
	}
	return nil
}

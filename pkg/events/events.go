package events

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/trace"
)

// Handler is a function that processes a message.
type Handler interface {
	Handle(ctx context.Context, msg *nats.Msg) error
}

// ParseJson parses the JSON data in the message into the given interface.
// If the JSON data is invalid, it returns an error.
func ParseJson(msg *nats.Msg, v interface{}) error {
	if err := json.Unmarshal(msg.Data, v); err != nil {
		return err
	}
	return nil
}

func RegisterHandler(ctx context.Context, tracer trace.Tracer, js nats.JetStreamContext, subject string, h Handler) (*nats.Subscription, error) {
	return js.Subscribe(subject, func(msg *nats.Msg) {
		ctx, span := tracer.Start(ctx, msg.Subject, trace.WithTimestamp(time.Now()))
		defer span.End()
		err := h.Handle(ctx, msg)
		if err != nil {
			span.RecordError(err)
		}
	})
}

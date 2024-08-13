package handlers

import (
	"context"
	"time"

	"github.com/9ssi7/bank/assets"
	"github.com/9ssi7/bank/internal/domain/auth"
	"github.com/9ssi7/bank/internal/infra/mail"
	"github.com/9ssi7/bank/pkg/cancel"
	"github.com/9ssi7/bank/pkg/events"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/trace"
)

type authStartLoginHandler struct {
	mailSrv mail.Srv
}

func NewAuthStartLoginHandler(mailSrv mail.Srv) events.Handler {
	return &authStartLoginHandler{
		mailSrv: mailSrv,
	}
}

func (h *authStartLoginHandler) Handle(ctx context.Context, tracer trace.Tracer, msg *nats.Msg) error {
	var event auth.EventLoginStarted
	if err := events.ParseJson(msg, &event); err != nil {
		return err
	}
	return cancel.RunWithTimeout(ctx, 5*time.Second, func(ctx context.Context) error {
		return h.mailSrv.SendWithTemplate(ctx, mail.SendWithTemplateConfig{
			SendConfig: mail.SendConfig{
				To:      []string{event.Email},
				Subject: "Verify your session",
				Message: event.Code,
			},
			Template: assets.Templates.AuthVerify,
			Data: map[string]interface{}{
				"Code":    event.Code,
				"IP":      mail.GetField(event.Device.IP),
				"Browser": mail.GetField(event.Device.Name),
				"OS":      mail.GetField(event.Device.OS),
			},
		})
	})
}

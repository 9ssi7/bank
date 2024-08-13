package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/9ssi7/bank/assets"
	"github.com/9ssi7/bank/internal/domain/user"
	"github.com/9ssi7/bank/internal/infra/mail"
	"github.com/9ssi7/bank/pkg/cancel"
	"github.com/9ssi7/bank/pkg/events"
	"github.com/nats-io/nats.go"
)

type userCreatedHandler struct {
	mailSrv    mail.Srv
	publicHost string
}

func NewUserCreatedHandler(mailSrv mail.Srv, publicHost string) events.Handler {
	return &userCreatedHandler{
		mailSrv:    mailSrv,
		publicHost: publicHost,
	}
}

func (h *userCreatedHandler) Handle(ctx context.Context, msg *nats.Msg) error {
	var event user.EventCreated
	if err := events.ParseJson(msg, &event); err != nil {
		return err
	}
	return cancel.RunWithTimeout(ctx, 5*time.Second, func(ctx context.Context) error {
		return h.mailSrv.SendWithTemplate(ctx, mail.SendWithTemplateConfig{
			SendConfig: mail.SendConfig{
				To:      []string{event.Email},
				Subject: "Welcome to our service",
			},
			Template: assets.Templates.AuthRegistered,
			Data: map[string]interface{}{
				"Name":            event.Name,
				"VerificationUrl": fmt.Sprintf("%s/auth/verify/%s", h.publicHost, event.TempToken),
			},
		})
	})
}

package eventhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/9ssi7/bank/assets"
	"github.com/9ssi7/bank/internal/domain/auth"
	"github.com/9ssi7/bank/internal/domain/user"
	"github.com/9ssi7/bank/internal/infra/mail"
	"github.com/9ssi7/bank/pkg/cancel"
	"github.com/nats-io/nats.go"
)

type AuthHandler struct {
	mailSrv    *mail.Srv
	publicHost string
}

func NewAuthHandler(mailSrv *mail.Srv, publicHost string) *AuthHandler {
	return &AuthHandler{mailSrv: mailSrv, publicHost: publicHost}
}

func (h *AuthHandler) OnLoginStart(ctx context.Context, msg *nats.Msg) error {
	var event auth.EventLoginStarted
	if err := json.Unmarshal(msg.Data, &event); err != nil {
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

func (h *AuthHandler) OnUserCreated(ctx context.Context, msg *nats.Msg) error {
	var event user.EventCreated
	if err := json.Unmarshal(msg.Data, &event); err != nil {
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

package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/9ssi7/bank/assets"
	"github.com/9ssi7/bank/internal/domain/account"
	"github.com/9ssi7/bank/internal/infra/mail"
	"github.com/9ssi7/bank/pkg/cancel"
	"github.com/9ssi7/bank/pkg/events"
	"github.com/nats-io/nats.go"
)

type transferOutgoingHandler struct {
	mailSrv mail.Srv
}

func NewTransferOutgoingHandler(mailSrv mail.Srv) events.Handler {
	return &transferOutgoingHandler{
		mailSrv: mailSrv,
	}
}

func (h *transferOutgoingHandler) Handle(ctx context.Context, msg *nats.Msg) error {
	var event account.EventTranfserOutgoing
	if err := events.ParseJson(msg, &event); err != nil {
		return err
	}
	return cancel.RunWithTimeout(ctx, 5*time.Second, func(ctx context.Context) error {
		return h.mailSrv.SendWithTemplate(ctx, mail.SendWithTemplateConfig{
			SendConfig: mail.SendConfig{
				To:      []string{event.Email},
				Subject: "Outgoing transaction",
			},
			Template: assets.Templates.TransferOutgoing,
			Data: map[string]interface{}{
				"Name":        event.Name,
				"Amount":      fmt.Sprintf("%s %s", mail.GetField(event.Amount), event.Currency),
				"Account":     mail.GetField(event.Account),
				"Description": mail.GetField(event.Description),
			},
		})
	})
}

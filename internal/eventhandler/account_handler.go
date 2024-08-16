package eventhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/9ssi7/bank/assets"
	"github.com/9ssi7/bank/internal/domain/account"
	"github.com/9ssi7/bank/internal/infra/mail"
	"github.com/9ssi7/bank/pkg/cancel"
	"github.com/nats-io/nats.go"
)

type AccountHandler struct {
	mailSrv *mail.Srv
}

func NewAccountHandler(mailSrv *mail.Srv) *AccountHandler {
	return &AccountHandler{mailSrv: mailSrv}
}

func (h *AccountHandler) OnTransferIncome(ctx context.Context, msg *nats.Msg) error {
	var event account.EventTranfserIncoming
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		return err
	}
	return cancel.NewWithTimeout(ctx, 5*time.Second, func(ctx context.Context) error {
		return h.mailSrv.SendWithTemplate(ctx, mail.SendWithTemplateConfig{
			SendConfig: mail.SendConfig{
				To:      []string{event.Email},
				Subject: "Incoming transaction",
			},
			Template: assets.Templates.TransferIncoming,
			Data: map[string]interface{}{
				"Name":        event.Name,
				"Amount":      fmt.Sprintf("%s %s", mail.GetField(event.Amount), event.Currency),
				"Account":     mail.GetField(event.Account),
				"Description": mail.GetField(event.Description),
			},
		})
	})
}

func (h *AccountHandler) OnTransferOutcome(ctx context.Context, msg *nats.Msg) error {
	var event account.EventTranfserOutgoing
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		return err
	}
	return cancel.NewWithTimeout(ctx, 5*time.Second, func(ctx context.Context) error {
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

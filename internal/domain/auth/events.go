package auth

import "github.com/9ssi7/bank/pkg/agent"

type EventLoginStarted struct {
	Email  string       `json:"email"`
	Code   string       `json:"code"`
	Device agent.Device `json:"device"`
}

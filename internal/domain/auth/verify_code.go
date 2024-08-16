package auth

import (
	"fmt"
	"time"

	"math/rand"

	"github.com/google/uuid"
)

type Verify struct {
	DeviceId  string    `json:"device_id"`
	Locale    string    `json:"locale"`
	UserId    uuid.UUID `json:"user_id"`
	Code      string    `json:"code"`
	TryCount  int       `json:"try_count"`
	ExpiresAt int64     `json:"expires_at"`
}

func (v *Verify) IsExpired() bool {
	return v.ExpiresAt < time.Now().Unix()
}

func (v *Verify) IsExceeded() bool {
	return v.TryCount >= 3
}

func (v *Verify) IncTryCount() {
	v.TryCount++
}

type VerifyConfig struct {
	UserId   uuid.UUID
	DeviceId string
	Locale   string
}

func NewVerify(cnf VerifyConfig) *Verify {
	return &Verify{
		UserId:    cnf.UserId,
		DeviceId:  cnf.DeviceId,
		Locale:    cnf.Locale,
		Code:      fmt.Sprintf("%04d", rand.Intn(9999)),
		TryCount:  0,
		ExpiresAt: time.Now().Add(5 * time.Minute).Unix(),
	}
}

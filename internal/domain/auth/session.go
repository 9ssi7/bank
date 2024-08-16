package auth

import (
	"time"

	"github.com/9ssi7/bank/pkg/agent"
)

type Session struct {
	DeviceId     string    `json:"device_id"`
	DeviceName   string    `json:"device_name"`
	DeviceType   string    `json:"device_type"`
	DeviceOS     string    `json:"device_os"`
	IpAddress    string    `json:"ip_address"`
	FcmToken     string    `json:"fcm_token"`
	RefreshToken string    `json:"refresh_token"`
	AccessToken  string    `json:"access_token"`
	LastLogin    time.Time `json:"last_login"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (s *Session) SetFromDevice(d *agent.Device) {
	s.DeviceName = d.Name
	s.DeviceType = d.Type
	s.DeviceOS = d.OS
	s.IpAddress = d.IP
}

func (s *Session) IsRefreshValid(accessToken, refreshToken string, ipAddress string) bool {
	return s.RefreshToken == refreshToken && s.AccessToken == accessToken && s.IpAddress == ipAddress
}

func (s *Session) IsAccessValid(accessToken, ipAddress string) bool {
	return s.AccessToken == accessToken && s.IpAddress == ipAddress
}

func (s *Session) Refresh(token string) {
	s.AccessToken = token
	s.LastLogin = time.Now()
	s.UpdatedAt = time.Now()
}

func (s *Session) VerifyToken(token string) bool {
	return s.AccessToken == token
}

func (s *Session) VerifyRefreshToken(token string) bool {
	return s.RefreshToken == token
}

func (s *Session) SetFcmToken(token string) {
	s.FcmToken = token
}

type SessionConfig struct {
	Device       agent.Device
	DeviceId     string `example:"550e8400-e29b-41d4-a716-446655440000"`
	AccessToken  string
	RefreshToken string
}

func NewSession(cnf SessionConfig) *Session {
	t := time.Now()
	return &Session{
		DeviceId:     cnf.DeviceId,
		DeviceName:   cnf.Device.Name,
		DeviceType:   cnf.Device.Type,
		DeviceOS:     cnf.Device.OS,
		IpAddress:    cnf.Device.IP,
		LastLogin:    t,
		CreatedAt:    t,
		UpdatedAt:    t,
		RefreshToken: cnf.RefreshToken,
		AccessToken:  cnf.AccessToken,
	}
}

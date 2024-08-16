package user

import (
	"time"

	"github.com/9ssi7/bank/pkg/ptr"
	"github.com/google/uuid"
)

type User struct {
	Id         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	Email      string     `json:"email"`
	IsActive   bool       `json:"is_active"`
	TempToken  *string    `json:"temp_token"`
	VerifiedAt *time.Time `json:"verified_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  time.Time  `json:"deleted_at"`
}

func (u *User) Verify() {
	u.VerifiedAt = ptr.Time(time.Now())
	u.TempToken = nil
}

func (u *User) Enable() {
	u.IsActive = true
}

func (u *User) Disable() {
	u.IsActive = false
}

type Config struct {
	Name  string `example:"John Doe"`
	Email string `example:"john@doe.com"`
}

func New(cnf Config) *User {
	return &User{
		Name:      cnf.Name,
		Email:     cnf.Email,
		TempToken: ptr.String(uuid.New().String()),
	}
}

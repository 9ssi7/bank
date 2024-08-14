package token

import (
	"context"
	"os"
	"time"
)

type Service struct {
	jwt *Jwt
	cnf Config
}

type Config struct {
	PublicKeyFile  string
	PrivateKeyFile string
	Project        string
	SignMethod     string
}

const (
	AccessTokenDuration  time.Duration = time.Hour * 24
	RefreshTokenDuration time.Duration = time.Hour * 24 * 30
)

func New(cnf Config) (*Service, error) {
	j, err := NewJwt(JwtConfig{
		PublicKey:  readFile(cnf.PublicKeyFile),
		PrivateKey: readFile(cnf.PrivateKeyFile),
		SignMethod: cnf.SignMethod,
	})
	if err != nil {
		return nil, err
	}
	return &Service{
		jwt: j,
		cnf: cnf,
	}, nil
}

func readFile(name string) []byte {
	f, err := os.ReadFile(name)
	if err != nil {
		panic(err)
	}
	return f
}

func (t *Service) GenerateAccessToken(ctx context.Context, u User) (string, error) {
	claims := &UserClaim{
		User:      u,
		Project:   t.cnf.Project,
		IsAccess:  true,
		IsRefresh: false,
	}
	claims.SetExpireIn(AccessTokenDuration)
	return t.generate(claims)
}

func (t *Service) GenerateRefreshToken(ctx context.Context, u User) (string, error) {
	claims := &UserClaim{
		User:      u,
		Project:   t.cnf.Project,
		IsAccess:  false,
		IsRefresh: true,
	}
	claims.SetExpireIn(RefreshTokenDuration)
	return t.generate(claims)
}

func (t *Service) generate(u *UserClaim) (string, error) {
	tkn, err := t.jwt.Sign(u)
	if err != nil {
		return "", err
	}
	return tkn, nil
}

func (t *Service) Parse(ctx context.Context, token string) (*UserClaim, error) {
	return t.jwt.GetClaims(ctx, token)
}

func (t *Service) Verify(ctx context.Context, token string) (bool, error) {
	return t.jwt.Verify(ctx, token)
}

func (t *Service) VerifyAndParse(ctx context.Context, token string) (*UserClaim, error) {
	return t.jwt.VerifyAndParse(ctx, token)
}

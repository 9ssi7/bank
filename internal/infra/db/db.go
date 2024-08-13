package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func New(ctx context.Context, cnf Config) (*sql.DB, error) {
	sql, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", cnf.Host, cnf.Port, cnf.User, cnf.Password, cnf.DBName, cnf.SSLMode))
	if err != nil {
		panic(err)
	}
	if err := sql.PingContext(ctx); err != nil {
		panic(err)
	}
	return sql, nil
}

package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/XSAM/otelsql"
	_ "github.com/lib/pq"
	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
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
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", cnf.Host, cnf.Port, cnf.User, cnf.Password, cnf.DBName, cnf.SSLMode)
	db, err := otelsql.Open("postgres", dsn, otelsql.WithAttributes(
		semconv.DBSystemMySQL,
	))
	if err != nil {
		return nil, err
	}
	err = otelsql.RegisterDBStatsMetrics(db, otelsql.WithAttributes(
		semconv.DBSystemMySQL,
	))
	if err != nil {
		return nil, err
	}
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}
	return db, nil
}

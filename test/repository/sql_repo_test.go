package repository_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/lib/pq"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type CancelFunc func()

func createSqlTesting(t *testing.T) (*sql.DB, CancelFunc) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16.4",
		ExposedPorts: []string{"5432"},
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
		},
		WaitingFor: wait.ForListeningPort("5432"),
	}
	sqlC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Could not start sql: %s", err)
	}
	term := func() {
		if err := sqlC.Terminate(ctx); err != nil {
			t.Fatalf("Could not stop sql: %s", err)
		}
	}
	host, err := sqlC.Host(ctx)
	if err != nil {
		t.Fatalf("Could not get sql host: %s", err)
	}
	port, err := sqlC.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("Could not get sql port: %s", err)
	}
	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=postgres password=postgres dbname=postgres sslmode=disable", host, port.Port()))
	if err != nil {
		t.Fatalf("Could not connect to sql: %s", err)
	}
	return db, term
}

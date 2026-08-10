package dbgen

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var testDBPool *pgxpool.Pool

func TestMain(m *testing.M) {
	container, dbpool, err := StartPostgresForTest()
	if err != nil {
		panic("failed to start postgres container: " + err.Error())
	}

	testDBPool = dbpool
	exitCode := m.Run()
	dbpool.Close()

	err = ClosePostgresForTest(container, dbpool)
	if err != nil {
		panic("failed to close postgres container: " + err.Error())
	}

	os.Exit(exitCode)
}

func StartPostgresForTest() (*postgres.PostgresContainer, *pgxpool.Pool, error) {
	ctx := context.Background()
	dbName := "identity"
	dbUser := "user"
	dbPassword := "password"

	postgresContainer, err := postgres.Run(ctx,
		"postgres:18-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		postgres.BasicWaitStrategies(),
	)

	if err != nil {
		log.Printf("failed to start container: %s", err)
		return nil, nil, err
	}

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open sql db for migrations: %w", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return nil, nil, fmt.Errorf("failed to set goose dialect: %w", err)
	}

	migrationsDir := filepath.Join("../../sql", "migrations")

	if err := goose.Up(db, migrationsDir); err != nil {
		return nil, nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	dbpool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect pgxpool: %w", err)
	}

	return postgresContainer, dbpool, nil
}

func ClosePostgresForTest(postgresContainer *postgres.PostgresContainer, dbpool *pgxpool.Pool) error {
	dbpool.Close()

	if err := testcontainers.TerminateContainer(postgresContainer); err != nil {
		log.Printf("failed to terminate container: %s", err)
		return err
	}

	return nil
}

package testutils

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CleanDB(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()

	findTablesQuery := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public'
		  AND table_type = 'BASE TABLE'
		  AND table_name != 'goose_db_version';
	`

	rows, err := pool.Query(ctx, findTablesQuery)
	if err != nil {
		t.Fatalf("failed to fetch table list: %v", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			t.Fatalf("failed to scan table name: %v", err)
		}

		tables = append(tables, fmt.Sprintf(`"%s"`, tableName))
	}

	if err := rows.Err(); err != nil {
		t.Fatalf("error iterating over rows: %v", err)
	}

	if len(tables) == 0 {
		return
	}

	truncateQuery := fmt.Sprintf(
		"TRUNCATE TABLE %s RESTART IDENTITY CASCADE;",
		strings.Join(tables, ", "),
	)

	_, err = pool.Exec(ctx, truncateQuery)
	if err != nil {
		t.Fatalf("failed to truncate tables: %v\nQuery: %s", err, truncateQuery)
	}
}

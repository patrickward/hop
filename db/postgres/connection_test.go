package postgres_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/patrickward/hop/db/postgres"
)

// testTables is a slice of table names used in tests
var testTables = []string{
	"test_table",
	"tx_test",
	"concurrent_test",
}

// setupTestDB initializes a test database connection. It requires an existing PostgreSQL instance.
func setupTestDB(t *testing.T) *postgres.Connection {
	t.Helper()

	uri := os.Getenv("TEST_POSTGRES_URI")
	if uri == "" {
		uri = "postgres://postgres:postgres@localhost:5432/connect_test?sslmode=disable"
	}

	config := postgres.Config{
		URI:             uri,
		Timeout:         5 * time.Second,
		MaxIdleConns:    5,
		MaxIdleTime:     1 * time.Minute,
		MaxConnLifetime: 5 * time.Minute,
		AutoMigrate:     false,
	}

	conn, err := postgres.NewConnection(config)
	require.NoError(t, err)

	return conn
}

// cleanupTestTables drops all test tables
func cleanupTestTables(ctx context.Context, pool *pgxpool.Pool) error {
	for _, table := range testTables {
		//goland:noinspection SqlNoDataSourceInspection
		_, err := pool.Exec(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table))
		if err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}
	return nil
}

//goland:noinspection SqlNoDataSourceInspection
func TestDriver_Connect(t *testing.T) {
	conn := setupTestDB(t)

	ctx := context.Background()
	defer func(conn *postgres.Connection) {
		_ = conn.Close()
	}(conn)

	pool := conn.Pool()

	// Clean up before and after tests
	require.NoError(t, cleanupTestTables(ctx, pool))
	defer func() {
		require.NoError(t, cleanupTestTables(ctx, pool))
	}()

	t.Run("basic operations", func(t *testing.T) {
		// Create test table
		_, err := pool.Exec(ctx, `
            CREATE TABLE test_table (
                id SERIAL PRIMARY KEY,
                name TEXT NOT NULL
            )
        `)
		require.NoError(t, err)

		// Insert data
		_, err = pool.Exec(ctx,
			"INSERT INTO test_table (name) VALUES ($1)",
			"test_name")
		require.NoError(t, err)

		// Read data
		var name string
		err = pool.QueryRow(ctx,
			"SELECT name FROM test_table WHERE id = 1").Scan(&name)
		require.NoError(t, err)
		require.Equal(t, "test_name", name)
	})

	t.Run("connection ping", func(t *testing.T) {
		err := conn.Ping(ctx)
		require.NoError(t, err)
	})
}

//goland:noinspection SqlNoDataSourceInspection
func TestDriver_Transactions(t *testing.T) {
	conn := setupTestDB(t)

	ctx := context.Background()
	defer func(conn *postgres.Connection) {
		_ = conn.Close()
	}(conn)

	pool := conn.Pool()

	// Clean up before and after tests
	require.NoError(t, cleanupTestTables(ctx, pool))
	defer func() {
		require.NoError(t, cleanupTestTables(ctx, pool))
	}()

	// Create test table
	_, err := pool.Exec(ctx, `
        CREATE TABLE tx_test (
            id SERIAL PRIMARY KEY,
            value TEXT NOT NULL
        )
    `)
	require.NoError(t, err)

	t.Run("successful transaction", func(t *testing.T) {
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)
		defer func(tx pgx.Tx, ctx context.Context) {
			_ = tx.Rollback(ctx)
		}(tx, ctx)

		_, err = tx.Exec(ctx,
			"INSERT INTO tx_test (value) VALUES ($1)", "test1")
		require.NoError(t, err)

		err = tx.Commit(ctx)
		require.NoError(t, err)

		// Verify the insert
		var value string
		err = pool.QueryRow(ctx,
			"SELECT value FROM tx_test WHERE id = 1").Scan(&value)
		require.NoError(t, err)
		require.Equal(t, "test1", value)
	})

	t.Run("rollback transaction", func(t *testing.T) {
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)
		defer func(tx pgx.Tx, ctx context.Context) {
			_ = tx.Rollback(ctx)
		}(tx, ctx)

		_, err = tx.Exec(ctx,
			"INSERT INTO tx_test (value) VALUES ($1)", "test2")
		require.NoError(t, err)

		err = tx.Rollback(ctx)
		require.NoError(t, err)

		// Verify the insert was rolled back
		var count int
		err = pool.QueryRow(ctx,
			"SELECT COUNT(*) FROM tx_test WHERE value = $1", "test2").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 0, count)
	})
}

package sqlite_cgo_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/patrickward/hop/db/sqlite_cgo"
)

func setupTestDB(t *testing.T) (*sqlite_cgo.Connection, string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "sqlite_cgo_test_*")
	require.NoError(t, err)

	dbPath := filepath.Join(tmpDir, "test.db")

	config := sqlite_cgo.Config{
		URI:             dbPath,
		Timeout:         5 * time.Second,
		MaxIdleConns:    5,
		MaxIdleTime:     1 * time.Minute,
		MaxConnLifetime: 5 * time.Minute,
		AutoMigrate:     false,
	}

	conn, err := sqlite_cgo.NewConnection(config)

	cleanup := func() {
		_ = os.RemoveAll(tmpDir)
	}

	return conn, dbPath, cleanup
}

// --------------------------------------------------------------------
// The rest of the test file is identical to the modernc version,
// just in a different package. You can copy the tests from the
// sqlite_modernc_test package and paste them here if there are changes.
// --------------------------------------------------------------------

func TestDriver_Connect(t *testing.T) {
	conn, _, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	defer func(conn *sqlite_cgo.Connection) {
		_ = conn.Close()
	}(conn)

	// Verify we can get both read and write connections
	readDB := conn.ReadDB()
	require.NotNil(t, readDB)

	writeDB := conn.WriteDB()
	require.NotNil(t, writeDB)

	// Test basic operations
	t.Run("basic operations", func(t *testing.T) {
		// Create a test table
		//goland:noinspection SqlNoDataSourceInspection
		_, err := writeDB.ExecContext(ctx, `
            CREATE TABLE test_table (
                id INTEGER PRIMARY KEY,
                name TEXT NOT NULL
            )
        `)
		require.NoError(t, err)

		// Insert data using write connection
		//goland:noinspection SqlNoDataSourceInspection
		_, err = writeDB.ExecContext(ctx,
			"INSERT INTO test_table (name) VALUES (?)",
			"test_name")
		require.NoError(t, err)

		// Read data using read connection
		var name string
		//goland:noinspection SqlNoDataSourceInspection
		err = readDB.QueryRowContext(ctx,
			"SELECT name FROM test_table WHERE id = 1").Scan(&name)
		require.NoError(t, err)
		require.Equal(t, "test_name", name)
	})

	t.Run("connection ping", func(t *testing.T) {
		err := conn.Ping(ctx)
		require.NoError(t, err)
	})
}

func TestDriver_ConcurrentAccess(t *testing.T) {
	conn, _, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	defer func(conn *sqlite_cgo.Connection) {
		_ = conn.Close()
	}(conn)

	writeDB := conn.WriteDB()
	readDB := conn.ReadDB()

	// Create test table
	//goland:noinspection SqlNoDataSourceInspection
	_, err := writeDB.ExecContext(ctx, `
        CREATE TABLE concurrent_test (
            id INTEGER PRIMARY KEY,
            counter INTEGER NOT NULL
        )`)
	require.NoError(t, err)

	// Insert initial record
	//goland:noinspection SqlNoDataSourceInspection
	_, err = writeDB.ExecContext(ctx,
		"INSERT INTO concurrent_test (id, counter) VALUES (1, 0)")
	require.NoError(t, err)

	// Test concurrent access
	t.Run("concurrent access", func(t *testing.T) {
		const numGoroutines = 10
		const iterations = 100

		done := make(chan bool)

		// Start multiple goroutines to increment counter
		for i := 0; i < numGoroutines; i++ {
			go func() {
				for j := 0; j < iterations; j++ {
					//goland:noinspection SqlNoDataSourceInspection
					_, err := writeDB.ExecContext(ctx,
						"UPDATE concurrent_test SET counter = counter + 1 WHERE id = 1")
					require.NoError(t, err)
				}
				done <- true
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		// Verify final counter value
		var counter int
		//goland:noinspection SqlNoDataSourceInspection
		err := readDB.QueryRowContext(ctx,
			"SELECT counter FROM concurrent_test WHERE id = 1").Scan(&counter)
		require.NoError(t, err)
		require.Equal(t, numGoroutines*iterations, counter)
	})
}

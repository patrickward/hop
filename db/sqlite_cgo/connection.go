// Package sqlite_cgo provides SQLite database connection and migration utilities using the mattn/go-sqlite3 driver.
package sqlite_cgo

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"runtime"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type Config struct {
	URI             string
	Timeout         time.Duration
	MaxIdleConns    int
	MaxIdleTime     time.Duration
	MaxConnLifetime time.Duration
	AutoMigrate     bool
	MigrationsFS    fs.FS
	MigrationsPath  string
}

type Connection struct {
	readDB  *sqlx.DB
	writeDB *sqlx.DB
}

func NewConnection(cfg Config) (*Connection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	conn, err := connect(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	// Verify connection
	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return conn, nil
}

func (c *Connection) Close() error {
	writeErr := c.writeDB.Close()
	readErr := c.readDB.Close()
	if writeErr != nil {
		return writeErr
	}
	return readErr
}

func (c *Connection) Ping(ctx context.Context) error {
	if err := c.writeDB.PingContext(ctx); err != nil {
		return fmt.Errorf("pinging write connection: %w", err)
	}
	return c.readDB.PingContext(ctx)
}

func (c *Connection) ReadDB() *sqlx.DB {
	return c.readDB
}

func (c *Connection) WriteDB() *sqlx.DB {
	return c.writeDB
}

// -- private methods --

func connect(ctx context.Context, cfg Config) (*Connection, error) {
	// Build connection parameters - using mattn/sqlite3 specific syntax
	connectionParams := make(url.Values)
	connectionParams.Add("_timeout", "10000")
	connectionParams.Add("_journal_mode", "WAL")
	connectionParams.Add("_journal_size_limit", "200000000")
	connectionParams.Add("_synchronous", "NORMAL")
	connectionParams.Add("_foreign_keys", "ON")
	connectionParams.Add("_temp_store", "MEMORY")
	connectionParams.Add("_cache_size", "-16000")
	connectionParams.Add("_txlock", "immediate")
	connectionURL := cfg.URI + "?" + connectionParams.Encode()

	// Create write connection
	writeDB, err := sqlx.ConnectContext(ctx, "sqlite3", connectionURL)
	if err != nil {
		return nil, fmt.Errorf("creating write connection: %w", err)
	}
	writeDB.SetMaxOpenConns(1)

	// Create read connection
	readDB, err := sqlx.ConnectContext(ctx, "sqlite3", connectionURL)
	if err != nil {
		_ = writeDB.Close()
		return nil, fmt.Errorf("creating read connection: %w", err)
	}

	// Configure read connection pool
	readDB.SetMaxOpenConns(max(10, runtime.NumCPU()))
	readDB.SetMaxIdleConns(cfg.MaxIdleConns)
	readDB.SetConnMaxIdleTime(cfg.MaxIdleTime)
	readDB.SetConnMaxLifetime(cfg.MaxConnLifetime)

	conn := &Connection{
		readDB:  readDB,
		writeDB: writeDB,
	}

	if cfg.AutoMigrate {
		if err := runMigrations(cfg); err != nil {
			_ = conn.Close()
			return nil, fmt.Errorf("running migrations: %w", err)
		}
	}

	return conn, nil
}

func runMigrations(cfg Config) error {
	iofsDriver, err := iofs.New(cfg.MigrationsFS, cfg.MigrationsPath)
	if err != nil {
		return fmt.Errorf("creating migration source: %w", err)
	}

	migrator, err := migrate.NewWithSourceInstance("iofs", iofsDriver, "sqlite3://"+cfg.URI)
	if err != nil {
		return fmt.Errorf("creating migrator: %w", err)
	}

	err = migrator.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("running migrations: %w", err)
	}

	return nil
}

// Package sqlite_modernc provides SQLite database connection and migration utilities using the modernc.org/sqlite driver.
package sqlite_modernc

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
	_ "modernc.org/sqlite"
)

// Config holds the configuration for the SQLite connection.
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

// Connection represents the database connection.
type Connection struct {
	config  *Config
	readDB  *sqlx.DB
	writeDB *sqlx.DB
}

// NewConnection creates a new SQLite connection.
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

// Close closes the database connections.
func (c *Connection) Close() error {
	writeErr := c.writeDB.Close()
	readErr := c.readDB.Close()
	if writeErr != nil {
		return writeErr
	}
	return readErr
}

// Ping checks the database connection.
func (c *Connection) Ping(ctx context.Context) error {
	if err := c.writeDB.PingContext(ctx); err != nil {
		return fmt.Errorf("pinging write connection: %w", err)
	}
	return c.readDB.PingContext(ctx)
}

// ReadDB returns the read database connection.
func (c *Connection) ReadDB() *sqlx.DB {
	return c.readDB
}

// WriteDB returns the write database connection.
func (c *Connection) WriteDB() *sqlx.DB {
	return c.writeDB
}

// -- private functions --

func connect(ctx context.Context, cfg Config) (*Connection, error) {
	connectionParams := make(url.Values)
	connectionParams.Add("_pragma", "busy_timeout(10000)")
	connectionParams.Add("_pragma", "journal_mode(WAL)")
	connectionParams.Add("_pragma", "journal_size_limit(200000000)")
	connectionParams.Add("_pragma", "synchronous(NORMAL)")
	connectionParams.Add("_pragma", "foreign_keys(ON)")
	connectionParams.Add("_pragma", "temp_store(MEMORY)")
	connectionParams.Add("_pragma", "cache_size(-16000)")
	connectionParams.Add("_txlock", "immediate")
	connectionURL := cfg.URI + "?" + connectionParams.Encode()

	writeDB, err := sqlx.ConnectContext(ctx, "sqlite", connectionURL)
	if err != nil {
		return nil, fmt.Errorf("creating write connection: %w", err)
	}
	writeDB.SetMaxOpenConns(1)

	readDB, err := sqlx.ConnectContext(ctx, "sqlite", connectionURL)
	if err != nil {
		_ = writeDB.Close()
		return nil, fmt.Errorf("creating read connection: %w", err)
	}

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

	migrator, err := migrate.NewWithSourceInstance("iofs", iofsDriver, "sqlite://"+cfg.URI)
	if err != nil {
		return fmt.Errorf("creating migrator: %w", err)
	}

	err = migrator.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("running migrations: %w", err)
	}

	return nil
}

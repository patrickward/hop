// Package postgres provides PostgreSQL database connection and migration utilities.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
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
	pool *pgxpool.Pool
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

// Close closes the connection pool
func (c *Connection) Close() error {
	c.pool.Close()
	return nil
}

// Ping checks the connection to the database
func (c *Connection) Ping(ctx context.Context) error {
	return c.pool.Ping(ctx)
}

// Pool returns the pgxpool.Pool instance
func (c *Connection) Pool() *pgxpool.Pool {
	return c.pool
}

func connect(ctx context.Context, cfg Config) (*Connection, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.URI)
	if err != nil {
		return nil, fmt.Errorf("parsing postgres config: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.MaxIdleConns)
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("creating postgres pool: %w", err)
	}

	conn := &Connection{pool: pool}

	if cfg.AutoMigrate {
		if err := runMigrations(cfg); err != nil {
			_ = conn.Close()
			return nil, fmt.Errorf("running migrations: %w", err)
		}
	}

	return conn, nil
}

// runMigrations runs database migrations
func runMigrations(cfg Config) error {

	iofsDriver, err := iofs.New(cfg.MigrationsFS, cfg.MigrationsPath)
	if err != nil {
		return err
	}

	migrator, err := migrate.NewWithSourceInstance("iofs", iofsDriver, "postgres://"+cfg.URI)
	if err != nil {
		return err
	}

	err = migrator.Up()
	switch {
	case errors.Is(err, migrate.ErrNoChange):
		break
	case err != nil:
		return err
	}

	return nil
}

package db

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/onefeed-th/onefeed-th-backend-api/config"
)

var pool *pgxpool.Pool

func InitDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dsn, err := buildPostgresDSN()
	if err != nil {
		return err
	}

	// Parse the DSN and configure connection pool
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("failed to parse database DSN: %w", err)
	}

	// Get pool configuration from config
	cfg := config.GetConfig()
	
	// Configure connection pool settings from config
	poolConfig.MaxConns = cfg.Postgres.Pool.MaxConns
	poolConfig.MinConns = cfg.Postgres.Pool.MinConns
	poolConfig.MaxConnLifetime = time.Duration(cfg.Postgres.Pool.MaxConnLifetime) * time.Minute
	poolConfig.MaxConnIdleTime = time.Duration(cfg.Postgres.Pool.MaxConnIdleTime) * time.Minute
	poolConfig.HealthCheckPeriod = time.Duration(cfg.Postgres.Pool.HealthCheckPeriod) * time.Minute
	poolConfig.ConnConfig.ConnectTimeout = time.Duration(cfg.Postgres.Pool.ConnectTimeout) * time.Second
	poolConfig.ConnConfig.RuntimeParams["application_name"] = "onefeed-backend"

	pool, err = pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	return nil
}

func GetPool() *pgxpool.Pool {
	return pool
}

func CloseDB() {
	if pool != nil {
		pool.Close()
	}
}

// GetPoolStats returns connection pool statistics for monitoring
func GetPoolStats() *pgxpool.Stat {
	if pool == nil {
		return nil
	}
	return pool.Stat()
}

func buildPostgresDSN() (string, error) {
	config := config.GetConfig()
	user := config.Postgres.User
	password := config.Postgres.Password
	host := config.Postgres.Host
	port := config.Postgres.Port
	db := config.Postgres.Dbname

	var missing []string
	if user == "" {
		missing = append(missing, "POSTGRES_USER")
	}
	if password == "" {
		missing = append(missing, "POSTGRES_PASSWORD")
	}
	if host == "" {
		missing = append(missing, "POSTGRES_HOST")
	}
	if port == 0 {
		missing = append(missing, "POSTGRES_PORT")
	}
	if db == "" {
		missing = append(missing, "POSTGRES_DB")
	}

	if len(missing) > 0 {
		return "", fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, password),
		Host:   fmt.Sprintf("%s:%d", host, port),
		Path:   db,
	}

	q := u.Query()
	q.Set("sslmode", "disable")
	u.RawQuery = q.Encode()

	return u.String(), nil
}

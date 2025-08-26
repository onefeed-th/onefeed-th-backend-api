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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dsn, err := buildPostgresDSN()
	if err != nil {
		return err
	}

	pool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		return err
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
		missing = append(missing, "POSTGREwS_DB")
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

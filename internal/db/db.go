package db

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
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
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	host := os.Getenv("POSTGRES_HOST")
	portStr := os.Getenv("POSTGRES_PORT")
	db := os.Getenv("POSTGRES_DB")

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
	if portStr == "" {
		missing = append(missing, "POSTGRES_PORT")
	}
	if db == "" {
		missing = append(missing, "POSTGRES_DB")
	}

	if len(missing) > 0 {
		return "", fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	if _, err := strconv.Atoi(portStr); err != nil {
		return "", fmt.Errorf("invalid POSTGRES_PORT: %s", portStr)
	}

	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, password),
		Host:   fmt.Sprintf("%s:%s", host, portStr),
		Path:   db,
	}

	q := u.Query()
	q.Set("sslmode", "disable")
	u.RawQuery = q.Encode()

	return u.String(), nil
}

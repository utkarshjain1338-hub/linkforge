package store

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrCodeGeneration = errors.New("could not generate a unique short code")

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(ctx context.Context) (*Postgres, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://linkforge:devpassword@localhost:5432/linkforge?sslmode=disable"
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse database URL: %w", err)
	}

	config.MaxConns = 10
	config.MinConns = 0
	config.MaxConnLifetime = time.Hour

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("connect to PostgreSQL: %w", err)
	}

	return &Postgres{pool: pool}, nil
}

func (p *Postgres) Close() {
	p.pool.Close()
}

func (p *Postgres) CreateLink(
	ctx context.Context,
	targetURL string,
) (string, error) {
	for attempt := 0; attempt < 5; attempt++ {
		randomBytes := make([]byte, 4)
		if _, err := rand.Read(randomBytes); err != nil {
			return "", fmt.Errorf("generate code: %w", err)
		}

		code := hex.EncodeToString(randomBytes)

		tag, err := p.pool.Exec(
			ctx,
			`INSERT INTO links (code, target_url)
			 VALUES ($1, $2)
			 ON CONFLICT (code) DO NOTHING`,
			code,
			targetURL,
		)
		if err != nil {
			return "", fmt.Errorf("insert link: %w", err)
		}

		if tag.RowsAffected() == 1 {
			return code, nil
		}
	}

	return "", ErrCodeGeneration
}

func (p *Postgres) GetLink(
	ctx context.Context,
	code string,
) (string, error) {
	var targetURL string

	err := p.pool.QueryRow(
		ctx,
		`SELECT target_url FROM links WHERE code = $1`,
		code,
	).Scan(&targetURL)

	if err != nil {
		return "", err
	}

	return targetURL, nil
}

func ValidateDatabaseURL(dsn string) error {
	parsed, err := url.Parse(dsn)
	if err != nil {
		return err
	}
	if parsed.Scheme != "postgres" && parsed.Scheme != "postgresql" {
		return errors.New("DATABASE_URL must use postgres or postgresql")
	}
	if parsed.Host == "" || parsed.Path == "/" || parsed.Path == "" {
		return errors.New("DATABASE_URL must include a host and database name")
	}
	return nil
}

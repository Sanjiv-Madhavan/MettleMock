package db

import (
	"context"
	"crypto/rand"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Factory interface {
	// Pool returns a shared *pgxpool.Pool, creating it on first use.
	Pool(ctx context.Context) (*pgxpool.Pool, error)
	// Close closes the underlying pool (useful in tests or on shutdown).
	Close()
}

type factory struct {
	once    sync.Once
	dsn     string
	pool    *pgxpool.Pool
	initErr error
}

// NewFactoryFromEnv builds a Factory using env vars.
// Preferred: PG_ADMIN_DSN (full DSN).
// Or it composes one from PG_HOST/PG_PORT/PG_SSLMODE + PG_ADMIN_USER/PG_ADMIN_PW.
func NewFactoryFromEnv() Factory {
	if dsn := os.Getenv("PG_ADMIN_DSN"); dsn != "" {
		return &factory{dsn: dsn}
	}
	host := getenvDefault("PG_HOST", "postgresql")
	port := getenvDefault("PG_PORT", "5432")
	ssl := getenvDefault("PG_SSLMODE", "disable")
	user := getenvDefault("PG_ADMIN_USER", "postgres")
	pw := os.Getenv("PG_ADMIN_PW")
	db := getenvDefault("PG_ADMIN_DB", "postgres")

	return &factory{dsn: buildAdminDSN(host, port, ssl, user, pw, db)}
}

func buildAdminDSN(host, port, sslmode, user, pw, db string) string {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, pw),
		Host:   joinHostPortSafe(host, (port)),
		Path:   "/" + db,
	}
	q := url.Values{}
	q.Set("sslmode", sslmode)
	u.RawQuery = q.Encode()
	return u.String()
}

func joinHostPortSafe(host, port string) string {
	if port == "" {
		return host
	}
	return net.JoinHostPort(host, port) // handles IPv6 correctly (adds [ ])
}

func (f *factory) Pool(ctx context.Context) (*pgxpool.Pool, error) {
	f.once.Do(func() {
		cfg, err := pgxpool.ParseConfig(f.dsn)
		if err != nil {
			f.initErr = err
			return
		}
		// Sensible small defaults; tune as needed.
		cfg.MaxConns = 10
		cfg.MinConns = 0
		cfg.HealthCheckPeriod = 30 * time.Second
		cfg.MaxConnIdleTime = 5 * time.Minute
		cfg.MaxConnLifetime = 30 * time.Minute

		p, err := pgxpool.NewWithConfig(ctx, cfg)
		if err != nil {
			f.initErr = err
			return
		}
		f.pool = p
	})
	if f.initErr != nil {
		return nil, f.initErr
	}
	return f.pool, nil
}

func (f *factory) Close() {
	if f.pool != nil {
		f.pool.Close()
	}
}

func EnsureDatabaseRole(ctx context.Context, pool *pgxpool.Pool, dbName, user, rotation string) (password string, rotated bool, err error) {
	// Create role if not exists
	_, err = pool.Exec(ctx, `
DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = $1) THEN
    EXECUTE format('CREATE ROLE %I LOGIN PASSWORD %L', $1, md5(random()::text));
  END IF;
END $$;`, user)
	if err != nil {
		return "", false, err
	}

	// Create DB if not exists; owner = user
	_, err = pool.Exec(ctx, `
DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_database WHERE datname = $1) THEN
    EXECUTE format('CREATE DATABASE %I OWNER %I', $1, $2);
  END IF;
END $$;`, dbName, user)
	if err != nil {
		return "", false, err
	}

	// Grant privileges (safe to re-run)
	_, _ = pool.Exec(ctx, `GRANT ALL PRIVILEGES ON DATABASE `+QuoteIdent(dbName)+` TO `+QuoteIdent(user))

	// Rotation
	d, _ := time.ParseDuration(rotation)
	if d <= 0 {
		// fetch current pw? we just generate a new to keep code simple for demo
		pw := RandPassword(24)
		_, err = pool.Exec(ctx, `ALTER ROLE `+QuoteIdent(user)+` WITH LOGIN PASSWORD $1`, pw)
		return pw, true, err
	}
	// For demo simplicity, rotate every reconcile when interval > 0
	pw := RandPassword(24)
	_, err = pool.Exec(ctx, `ALTER ROLE `+QuoteIdent(user)+` WITH LOGIN PASSWORD $1`, pw)
	return pw, true, err
}

func DropDatabaseAndRole(ctx context.Context, pool *pgxpool.Pool, dbName, user string) error {
	// Terminate sessions and drop
	_, _ = pool.Exec(ctx, `SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname=$1`, dbName)
	_, _ = pool.Exec(ctx, `REVOKE CONNECT ON DATABASE `+QuoteIdent(dbName)+` FROM PUBLIC`)
	_, _ = pool.Exec(ctx, `DROP DATABASE IF EXISTS `+QuoteIdent(dbName))
	_, _ = pool.Exec(ctx, `DROP ROLE IF EXISTS `+QuoteIdent(user))
	return nil
}

func RevokeAndKillSessions(ctx context.Context, pool *pgxpool.Pool, dbName, user string) error {
	_, _ = pool.Exec(ctx, `SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE usename=$1`, user)
	_, _ = pool.Exec(ctx, `REVOKE ALL PRIVILEGES ON DATABASE `+QuoteIdent(dbName)+` FROM `+QuoteIdent(user))
	_, _ = pool.Exec(ctx, `ALTER ROLE `+QuoteIdent(user)+` NOLOGIN`)
	return nil
}

func BuildAppDSN(user, pw, db string) string {
	host := getenvDefault("PG_HOST", "postgresql") // service name
	port := getenvDefault("PG_PORT", "5432")
	ssl := getenvDefault("PG_SSLMODE", "disable")
	return "postgres://" + user + ":" + urlQueryEscape(pw) + "@" + host + ":" + port + "/" + db + "?sslmode=" + ssl
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func urlQueryEscape(pass string) string {
	return url.QueryEscape(pass)
}

func NextRotationAfter(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0
	}
	return d
}

func QuoteIdent(s string) string { return `"` + strings.ReplaceAll(s, `"`, `""`) + `"` }

var pwAlphabet = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_")

// RandPassword returns a cryptographically secure random password of length n.
// Uses rejection sampling to avoid modulo bias over the alphabet.
func RandPassword(n int) string {
	if n <= 0 {
		n = 24
	}
	out := make([]byte, n)

	// Largest multiple of len(alphabet) less than 256.
	// Values >= limit are discarded to keep uniform distribution.
	limit := byte(256 - (256 % len(pwAlphabet)))

	i := 0
	for i < n {
		var b [1]byte
		if _, err := rand.Read(b[:]); err != nil {
			// This should be extremely rare; panic is acceptable for infra code.
			// If you prefer, bubble the error up instead of panicking.
			panic(fmt.Errorf("crypto/rand failure: %w", err))
		}
		if b[0] >= limit {
			continue // reject and retry
		}
		out[i] = pwAlphabet[int(b[0])%len(pwAlphabet)]
		i++
	}
	return string(out)
}

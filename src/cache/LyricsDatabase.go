package cache

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

// LyricsDatabase is a small Postgres-backed cache connector for Supabase.
type LyricsDatabase struct {
	connURL string
	conn    *pgx.Conn
	ttl     time.Duration
}

// LyricsDB is the global DB-backed cache instance (nil until initialized).
var LyricsDB *LyricsDatabase

// InitLyricsDBFromEnv initializes LyricsDB using DATABASE_URL and APP_LYRICS_CACHE_TTL_SECONDS.
// It returns an error if connection/table creation fails.
func InitLyricsDBFromEnv() error {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return fmt.Errorf("DATABASE_URL not set")
	}
	ttl := 1800 * time.Second
	if v := os.Getenv("APP_LYRICS_CACHE_TTL_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			ttl = time.Duration(n) * time.Second
		}
	}

	db := &LyricsDatabase{connURL: dbURL, ttl: ttl}
	if err := db.Connect(); err != nil {
		return err
	}
	if err := db.EnsureTable(); err != nil {
		db.Close()
		return err
	}
	LyricsDB = db
	return nil
}

// Connect opens the pgx connection.
func (d *LyricsDatabase) Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, d.connURL)
	if err != nil {
		return err
	}
	d.conn = conn
	return nil
}

// EnsureTable creates the cache table if it does not exist.
func (d *LyricsDatabase) EnsureTable() error {
	if d.conn == nil {
		return fmt.Errorf("not connected")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	q := `CREATE TABLE IF NOT EXISTS lyrics_cache (
        key TEXT PRIMARY KEY,
        value TEXT NOT NULL,
        expires_at TIMESTAMPTZ NOT NULL
    )`
	_, err := d.conn.Exec(ctx, q)
	return err
}

// Get returns the cached value if present and not expired.
func (d *LyricsDatabase) Get(key string) (string, bool) {
	if d == nil || d.conn == nil {
		return "", false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var value string
	var expires time.Time
	row := d.conn.QueryRow(ctx, "SELECT value, expires_at FROM lyrics_cache WHERE key=$1", key)
	if err := row.Scan(&value, &expires); err != nil {
		return "", false
	}
	if time.Now().After(expires) {
		// lazy delete
		go func(k string) {
			ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel2()
			d.conn.Exec(ctx2, "DELETE FROM lyrics_cache WHERE key=$1", k)
		}(key)
		return "", false
	}
	return value, true
}

// Set writes the value with TTL (upsert).
func (d *LyricsDatabase) Set(key, value string) error {
	if d == nil || d.conn == nil {
		return fmt.Errorf("db not initialized")
	}
	expires := time.Now().Add(d.ttl)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := d.conn.Exec(ctx, "INSERT INTO lyrics_cache (key, value, expires_at) VALUES ($1,$2,$3) ON CONFLICT (key) DO UPDATE SET value=EXCLUDED.value, expires_at=EXCLUDED.expires_at", key, value, expires)
	return err
}

// Close closes the DB connection.
func (d *LyricsDatabase) Close() {
	if d == nil || d.conn == nil {
		return
	}
	d.conn.Close(context.Background())
	d.conn = nil
}

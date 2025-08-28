package cache

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestLyricsDatabaseIntegration(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set; skipping integration test")
	}

	if err := InitLyricsDBFromEnv(); err != nil {
		t.Fatalf("InitLyricsDBFromEnv failed: %v", err)
	}
	defer func() {
		if LyricsDB != nil {
			LyricsDB.Close()
		}
	}()

	key := fmt.Sprintf("test-lyrics-%d", time.Now().UnixNano())
	val := "integration-value"

	if err := LyricsDB.Set(key, val); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	got, ok := LyricsDB.Get(key)
	if !ok {
		t.Fatalf("expected cache hit after set, got miss")
	}
	if got != val {
		t.Fatalf("value mismatch: expected %q got %q", val, got)
	}

	// Test TTL expiry quickly by temporarily setting a short TTL
	oldTTL := LyricsDB.ttl
	LyricsDB.ttl = 1 * time.Second
	key2 := key + "-ttl"
	if err := LyricsDB.Set(key2, "v2"); err != nil {
		t.Fatalf("Set key2 failed: %v", err)
	}
	if _, ok := LyricsDB.Get(key2); !ok {
		t.Fatalf("expected hit for key2 immediately after set")
	}
	time.Sleep(1200 * time.Millisecond)
	if _, ok := LyricsDB.Get(key2); ok {
		t.Fatalf("expected key2 to be expired but got hit")
	}
	LyricsDB.ttl = oldTTL
}

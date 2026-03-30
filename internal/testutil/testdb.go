package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	migrateOnce sync.Once
	migrateErr  error
)

func OpenTestDB(t testing.TB) *pgxpool.Pool {
	t.Helper()
	url := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL"))
	if url == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("create test db pool: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping test db: %v", err)
	}

	migrateOnce.Do(func() {
		migrateErr = applyMigrations(ctx, pool)
	})
	if migrateErr != nil {
		t.Fatalf("apply migrations: %v", migrateErr)
	}

	if err := resetAllTables(ctx, pool); err != nil {
		t.Fatalf("reset test db tables: %v", err)
	}

	return pool
}

func applyMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	root, err := repoRoot()
	if err != nil {
		return err
	}
	paths, err := filepath.Glob(filepath.Join(root, "migrations", "*.up.sql"))
	if err != nil {
		return fmt.Errorf("glob migrations: %w", err)
	}
	sort.Strings(paths)
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", p, err)
		}
		if strings.TrimSpace(string(b)) == "" {
			continue
		}
		if _, err := pool.Exec(ctx, string(b)); err != nil {
			if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "duplicate") {
				continue
			}
			return fmt.Errorf("exec migration %s: %w", filepath.Base(p), err)
		}
	}
	return nil
}

func resetAllTables(ctx context.Context, pool *pgxpool.Pool) error {
	const q = `
DO $$
DECLARE r RECORD;
BEGIN
  FOR r IN (
    SELECT tablename
    FROM pg_tables
    WHERE schemaname = 'public'
  ) LOOP
    EXECUTE 'TRUNCATE TABLE ' || quote_ident(r.tablename) || ' RESTART IDENTITY CASCADE';
  END LOOP;
END $$;
`
	_, err := pool.Exec(ctx, q)
	return err
}

func repoRoot() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("resolve caller")
	}
	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("repo root not found from %s", file)
		}
		dir = parent
	}
}

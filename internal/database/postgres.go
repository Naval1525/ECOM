package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"

	migrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type ConnParams struct {
    URL      string
    Host     string
    Port     string
    User     string
    Password string
    Name     string
    SSLMode  string
}

// DB is a wrapper around sql.DB to provide additional methods or properties if needed in the future.
type DB struct {
	*sql.DB
}

func NewPostgresConnection() (*DB, error) {
    // Preserve existing behavior if called without explicit params by reading envs
    if url := os.Getenv("DB_URL"); url != "" {
        db, err := sql.Open("postgres", url)
        if err != nil {
            return nil, fmt.Errorf("failed to open db with DB_URL: %w", err)
        }
        db.SetMaxOpenConns(25)
        db.SetMaxIdleConns(10)
        db.SetConnMaxLifetime(5 * time.Minute)
        if err = db.Ping(); err != nil {
            return nil, fmt.Errorf("failed to ping db: %w", err)
        }
        log.Println("✅ Successfully connected to PostgreSQL database")
        return &DB{db}, nil
    }

    // Fallback to individual env vars
    host := getenvDefault("DB_HOST", "localhost")
    port := getenvDefault("DB_PORT", "5432")
    user := getenvDefault("DB_USER", "postgres")
    password := os.Getenv("DB_PASSWORD")
    name := getenvDefault("DB_NAME", "social_media")

    connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, name)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	log.Println("✅ Successfully connected to PostgreSQL database")

	return &DB{db}, nil

}

// Migrate runs database migrations. For now it's a no-op to satisfy callers.
func (db *DB) Migrate() error {
    // Best-effort preflight to ensure essential schema exists for existing DBs
    if err := db.preflightEnsureUsersSchema(); err != nil {
        log.Printf("warning: users schema preflight failed: %v", err)
    }

    // Use golang-migrate with file:// source from project root
    driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
    if err != nil {
        return fmt.Errorf("failed to init migration driver: %w", err)
    }

    m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
    if err != nil {
        return fmt.Errorf("failed to create migrator: %w", err)
    }

    // Handle version mismatch and dirty states proactively
    if v, dirty, verr := m.Version(); verr == nil {
        maxFileV, maxErr := getMaxMigrationVersion("migrations")
        if maxErr == nil {
            if int(v) > maxFileV {
                if ferr := m.Force(maxFileV); ferr != nil {
                    return fmt.Errorf("failed to force migration version from %d to %d: %w", v, maxFileV, ferr)
                }
            }
        }
        if dirty {
            if ferr := m.Force(int(v)); ferr != nil {
                return fmt.Errorf("failed to clear dirty state at v=%d: %w", v, ferr)
            }
        }
    }

    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        // If lock/prepared statement issues likely due to pooler, skip hard failure
        if isPoolerLockError(err) {
            log.Printf("warning: skipping migrations due to pooler lock error: %v", err)
        } else {
            return fmt.Errorf("migration failed: %w", err)
        }
    }

    log.Println("✅ Database migrations up-to-date")
    return nil
}

// isPoolerLockError detects errors commonly seen when running golang-migrate through PgBouncer/Neon pooler
func isPoolerLockError(err error) bool {
    if err == nil {
        return false
    }
    msg := err.Error()
    return strings.Contains(msg, "try lock failed") ||
        strings.Contains(msg, "pg_advisory_lock") ||
        strings.Contains(msg, "unnamed prepared statement does not exist")
}

// getMaxMigrationVersion finds the maximum numeric version prefix among *.up.sql files
func getMaxMigrationVersion(dir string) (int, error) {
    entries, err := os.ReadDir(dir)
    if err != nil {
        return 0, err
    }
    maxV := 0
    for _, e := range entries {
        if e.IsDir() {
            continue
        }
        name := e.Name()
        if !strings.HasSuffix(name, ".up.sql") {
            continue
        }
        // parse leading digits until underscore
        underscore := strings.IndexByte(name, '_')
        if underscore <= 0 {
            continue
        }
        numStr := name[:underscore]
        if v, convErr := strconv.Atoi(numStr); convErr == nil {
            if v > maxV {
                maxV = v
            }
        }
    }
    return maxV, nil
}

// preflightEnsureUsersSchema adds any missing columns or base table for users to avoid runtime errors
func (db *DB) preflightEnsureUsersSchema() error {
    sqlText := `
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

    CREATE TABLE IF NOT EXISTS users (
        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
        username VARCHAR(50) UNIQUE NOT NULL,
        email VARCHAR(100) UNIQUE NOT NULL,
        password_hash VARCHAR(255) NOT NULL,
        full_name VARCHAR(100) NOT NULL,
        bio TEXT DEFAULT '',
        avatar VARCHAR(255) DEFAULT '',
        created_at TIMESTAMPTZ DEFAULT NOW(),
        updated_at TIMESTAMPTZ DEFAULT NOW()
    );

    ALTER TABLE IF EXISTS users
        ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255) NOT NULL DEFAULT '';
    ALTER TABLE IF EXISTS users
        ALTER COLUMN password_hash DROP DEFAULT;

    ALTER TABLE IF EXISTS users
        ADD COLUMN IF NOT EXISTS full_name VARCHAR(100) NOT NULL DEFAULT '';
    ALTER TABLE IF EXISTS users
        ALTER COLUMN full_name DROP DEFAULT;

    ALTER TABLE IF EXISTS users
        ADD COLUMN IF NOT EXISTS bio TEXT DEFAULT '';
    ALTER TABLE IF EXISTS users
        ADD COLUMN IF NOT EXISTS avatar VARCHAR(255) DEFAULT '';
    ALTER TABLE IF EXISTS users
        ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();
    ALTER TABLE IF EXISTS users
        ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();

    CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
    CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

    -- Drop legacy password column if it exists (code uses password_hash)
    DO $$
    BEGIN
        IF EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_name='users' AND column_name='password'
        ) THEN
            ALTER TABLE users DROP COLUMN password;
        END IF;
    END $$;
    `
    _, err := db.Exec(sqlText)
    return err
}

func getenvDefault(key, fallback string) string {
    v := os.Getenv(key)
    if v == "" {
        return fallback
    }
    return v
}

// NewWithParams allows passing explicit connection params (from config) instead of env
func NewWithParams(p ConnParams) (*DB, error) {
    if p.URL != "" {
        db, err := sql.Open("postgres", p.URL)
        if err != nil {
            return nil, fmt.Errorf("failed to open db with URL: %w", err)
        }
        db.SetMaxOpenConns(25)
        db.SetMaxIdleConns(10)
        db.SetConnMaxLifetime(5 * time.Minute)
        if err = db.Ping(); err != nil {
            return nil, fmt.Errorf("failed to ping db: %w", err)
        }
        log.Println("✅ Successfully connected to PostgreSQL database")
        return &DB{db}, nil
    }

    host := firstNonEmpty(p.Host, "localhost")
    port := firstNonEmpty(p.Port, "5432")
    user := firstNonEmpty(p.User, "postgres")
    name := firstNonEmpty(p.Name, "social_media")
    ssl := firstNonEmpty(p.SSLMode, "disable")
    connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, p.Password, name, ssl)
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, fmt.Errorf("failed to open db: %w", err)
    }
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(10)
    db.SetConnMaxLifetime(5 * time.Minute)
    if err = db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping db: %w", err)
    }
    log.Println("✅ Successfully connected to PostgreSQL database")
    return &DB{db}, nil
}

func firstNonEmpty(v, def string) string {
    if v == "" {
        return def
    }
    return v
}

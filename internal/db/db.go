package db

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

var _ Store = (*DB)(nil)

func New(dbPath string, migrationsFS embed.FS) (*DB, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := runMigrations(db, migrationsFS); err != nil {
		return nil, fmt.Errorf("migrations: %w", err)
	}

	return &DB{db}, nil
}

func runMigrations(db *sql.DB, fs embed.FS) error {
	entries, err := fs.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Slice(names, func(i, j int) bool {
		ni, _ := strconv.Atoi(strings.Split(names[i], "_")[0])
		nj, _ := strconv.Atoi(strings.Split(names[j], "_")[0])
		return ni < nj
	})

	for _, name := range names {
		data, err := fs.ReadFile("migrations/" + name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		log.Printf("running migration: %s", name)
		if _, err := db.Exec(string(data)); err != nil {
			return fmt.Errorf("exec migration %s: %w", name, err)
		}
	}

	return nil
}

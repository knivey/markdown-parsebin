package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/knivey/dave-web/internal/models"
)

var ErrNotFound = errors.New("not found")

func (d *DB) CreatePaste(paste *models.Paste) error {
	_, err := d.Exec(
		`INSERT INTO pastes (slug, title, content, rendered, created_at, expires_at, language)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		paste.Slug, paste.Title, paste.Content, paste.Rendered,
		paste.CreatedAt, paste.ExpiresAt, paste.Language,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return fmt.Errorf("duplicate slug %s: %w", paste.Slug, err)
		}
		return fmt.Errorf("insert paste: %w", err)
	}
	return nil
}

func (d *DB) GetPaste(slug string) (*models.Paste, error) {
	p := &models.Paste{}
	err := d.QueryRow(
		`SELECT slug, title, content, rendered, created_at, expires_at, language
		 FROM pastes WHERE slug = ? AND (expires_at IS NULL OR expires_at > ?)`, slug, time.Now(),
	).Scan(&p.Slug, &p.Title, &p.Content, &p.Rendered, &p.CreatedAt, &p.ExpiresAt, &p.Language)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get paste %s: %w", slug, err)
	}
	return p, nil
}

func (d *DB) ListPastes(limit int) ([]*models.Paste, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := d.Query(
		`SELECT slug, title, content, rendered, created_at, expires_at, language
		 FROM pastes WHERE expires_at IS NULL OR expires_at > ?
		 ORDER BY created_at DESC LIMIT ?`, time.Now(), limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list pastes: %w", err)
	}
	defer rows.Close()

	var pastes []*models.Paste
	for rows.Next() {
		p := &models.Paste{}
		if err := rows.Scan(&p.Slug, &p.Title, &p.Content, &p.Rendered, &p.CreatedAt, &p.ExpiresAt, &p.Language); err != nil {
			return nil, fmt.Errorf("scan paste: %w", err)
		}
		pastes = append(pastes, p)
	}
	return pastes, rows.Err()
}

func (d *DB) DeletePaste(slug string) error {
	res, err := d.Exec(`DELETE FROM pastes WHERE slug = ?`, slug)
	if err != nil {
		return fmt.Errorf("delete paste %s: %w", slug, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("paste %s not found", slug)
	}
	return nil
}

func (d *DB) DeleteExpired() (int64, error) {
	res, err := d.Exec(`DELETE FROM pastes WHERE expires_at IS NOT NULL AND expires_at < ?`, time.Now())
	if err != nil {
		return 0, fmt.Errorf("delete expired: %w", err)
	}
	return res.RowsAffected()
}

func IsDuplicateSlug(err error) bool {
	return err != nil && strings.Contains(err.Error(), "duplicate slug")
}

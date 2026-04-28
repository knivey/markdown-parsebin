package db

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

type APIKey struct {
	Key         string
	Description string
	CreatedAt   time.Time
}

func (d *DB) CreateAPIKey(description string) (*APIKey, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}
	key := "dave_" + hex.EncodeToString(raw)

	ak := &APIKey{
		Key:         key,
		Description: description,
		CreatedAt:   time.Now(),
	}

	_, err := d.Exec(
		`INSERT INTO api_keys (key, description, created_at) VALUES (?, ?, ?)`,
		ak.Key, ak.Description, ak.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert api key: %w", err)
	}

	return ak, nil
}

func (d *DB) GetAPIKey(key string) (*APIKey, error) {
	ak := &APIKey{}
	err := d.QueryRow(
		`SELECT key, description, created_at FROM api_keys WHERE key = ?`, key,
	).Scan(&ak.Key, &ak.Description, &ak.CreatedAt)
	if err != nil {
		return nil, err
	}
	return ak, nil
}

func (d *DB) ListAPIKeys() ([]*APIKey, error) {
	rows, err := d.Query(
		`SELECT key, description, created_at FROM api_keys ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}
	defer rows.Close()

	var keys []*APIKey
	for rows.Next() {
		ak := &APIKey{}
		if err := rows.Scan(&ak.Key, &ak.Description, &ak.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan api key: %w", err)
		}
		keys = append(keys, ak)
	}
	return keys, rows.Err()
}

func (d *DB) DeleteAPIKey(key string) error {
	res, err := d.Exec(`DELETE FROM api_keys WHERE key = ?`, key)
	if err != nil {
		return fmt.Errorf("delete api key: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("api key not found")
	}
	return nil
}

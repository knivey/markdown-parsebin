package db

import "github.com/knivey/dave-web/internal/models"

type PasteStore interface {
	CreatePaste(paste *models.Paste) error
	GetPaste(id string) (*models.Paste, error)
	ListPastes(limit int) ([]*models.Paste, error)
	DeletePaste(id string) error
	DeleteExpired() (int64, error)
}

type APIKeyStore interface {
	CreateAPIKey(description string) (*APIKey, error)
	GetAPIKey(key string) (*APIKey, error)
	ListAPIKeys() ([]*APIKey, error)
	DeleteAPIKey(key string) error
}

type Store interface {
	PasteStore
	APIKeyStore
}

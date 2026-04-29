package testutil

import (
	"fmt"

	"github.com/knivey/dave-web/internal/db"
	"github.com/knivey/dave-web/internal/models"
)

type MockStore struct {
	CreatePasteFunc         func(paste *models.Paste) error
	GetPasteFunc            func(id string) (*models.Paste, error)
	ListPastesFunc          func(limit int) ([]*models.Paste, error)
	DeletePasteFunc         func(id string) error
	DeleteExpiredFunc       func() (int64, error)
	CountPastesFunc         func() (*db.PasteStats, error)
	ListAllPastesFunc       func() ([]*models.Paste, error)
	UpdatePasteRenderedFunc func(slug string, rendered string) error
	CreateAPIKeyFunc        func(description string) (*db.APIKey, error)
	GetAPIKeyFunc           func(key string) (*db.APIKey, error)
	ListAPIKeysFunc         func() ([]*db.APIKey, error)
	DeleteAPIKeyFunc        func(key string) error
}

func (m *MockStore) CreatePaste(paste *models.Paste) error {
	if m.CreatePasteFunc != nil {
		return m.CreatePasteFunc(paste)
	}
	return fmt.Errorf("CreatePaste not mocked")
}

func (m *MockStore) GetPaste(slug string) (*models.Paste, error) {
	if m.GetPasteFunc != nil {
		return m.GetPasteFunc(slug)
	}
	return nil, db.ErrNotFound
}

func (m *MockStore) ListPastes(limit int) ([]*models.Paste, error) {
	if m.ListPastesFunc != nil {
		return m.ListPastesFunc(limit)
	}
	return nil, fmt.Errorf("ListPastes not mocked")
}

func (m *MockStore) DeletePaste(slug string) error {
	if m.DeletePasteFunc != nil {
		return m.DeletePasteFunc(slug)
	}
	return fmt.Errorf("DeletePaste not mocked")
}

func (m *MockStore) DeleteExpired() (int64, error) {
	if m.DeleteExpiredFunc != nil {
		return m.DeleteExpiredFunc()
	}
	return 0, fmt.Errorf("DeleteExpired not mocked")
}

func (m *MockStore) CountPastes() (*db.PasteStats, error) {
	if m.CountPastesFunc != nil {
		return m.CountPastesFunc()
	}
	return &db.PasteStats{}, nil
}

func (m *MockStore) ListAllPastes() ([]*models.Paste, error) {
	if m.ListAllPastesFunc != nil {
		return m.ListAllPastesFunc()
	}
	return nil, fmt.Errorf("ListAllPastes not mocked")
}

func (m *MockStore) UpdatePasteRendered(slug string, rendered string) error {
	if m.UpdatePasteRenderedFunc != nil {
		return m.UpdatePasteRenderedFunc(slug, rendered)
	}
	return fmt.Errorf("UpdatePasteRendered not mocked")
}

func (m *MockStore) CreateAPIKey(description string) (*db.APIKey, error) {
	if m.CreateAPIKeyFunc != nil {
		return m.CreateAPIKeyFunc(description)
	}
	return nil, fmt.Errorf("CreateAPIKey not mocked")
}

func (m *MockStore) GetAPIKey(key string) (*db.APIKey, error) {
	if m.GetAPIKeyFunc != nil {
		return m.GetAPIKeyFunc(key)
	}
	return nil, fmt.Errorf("GetAPIKey not mocked")
}

func (m *MockStore) ListAPIKeys() ([]*db.APIKey, error) {
	if m.ListAPIKeysFunc != nil {
		return m.ListAPIKeysFunc()
	}
	return nil, fmt.Errorf("ListAPIKeys not mocked")
}

func (m *MockStore) DeleteAPIKey(key string) error {
	if m.DeleteAPIKeyFunc != nil {
		return m.DeleteAPIKeyFunc(key)
	}
	return fmt.Errorf("DeleteAPIKey not mocked")
}

var _ db.Store = (*MockStore)(nil)

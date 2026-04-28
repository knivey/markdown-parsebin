package models

import "time"

type Paste struct {
	Slug      string     `json:"slug"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	Rendered  string     `json:"rendered"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at"`
	Language  string     `json:"language"`
}

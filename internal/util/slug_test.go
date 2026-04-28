package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSlug_Length(t *testing.T) {
	for i := 0; i < 100; i++ {
		slug, err := GenerateSlug()
		require.NoError(t, err)
		assert.Len(t, slug, 8, "slug should always be 8 characters")
	}
}

func TestGenerateSlug_Charset(t *testing.T) {
	for i := 0; i < 100; i++ {
		slug, err := GenerateSlug()
		require.NoError(t, err)
		for _, c := range slug {
			assert.True(t, isAlphanumeric(c), "character %c should be alphanumeric", c)
		}
	}
}

func TestGenerateSlug_Uniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		slug, err := GenerateSlug()
		require.NoError(t, err)
		assert.False(t, seen[slug], "slug %s should be unique", slug)
		seen[slug] = true
	}
}

func isAlphanumeric(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

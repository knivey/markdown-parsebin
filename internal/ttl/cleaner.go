package ttl

import (
	"log"
	"time"

	"github.com/knivey/dave-web/internal/db"
)

type PasteDeleter interface {
	DeleteExpired() (int64, error)
}

func StartCleaner(database PasteDeleter, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			n, err := database.DeleteExpired()
			if err != nil {
				log.Printf("ttl cleaner error: %v", err)
			} else if n > 0 {
				log.Printf("ttl cleaner: deleted %d expired paste(s)", n)
			}
		}
	}()
}

var _ PasteDeleter = (*db.DB)(nil)

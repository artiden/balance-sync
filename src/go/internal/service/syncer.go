package service

import (
	"log"
	"time"
)

// Syncer keeps the in-memory cache in sync with the DB.
type Syncer struct {
	cache *Cache
	repo  Repository
}

func NewSyncer(c *Cache, r Repository) *Syncer {
	return &Syncer{cache: c, repo: r}
}

// SyncOnce performs a single full refresh of the cache from the DB.
func (s *Syncer) SyncOnce() error {
	states, err := s.repo.LatestStates()
	if err != nil {
		return err
	}
	for _, st := range states {
		s.cache.SetIfNewer(st.UserID, st.WalletID, st.Balance, st.UpdatedAt)
	}
	log.Printf("cache sync: %d items", len(states))
	return nil
}

// Start runs periodic cache synchronisation in the background.
func (s *Syncer) Start() {
	if err := s.SyncOnce(); err != nil {
		log.Printf("initial cache sync failed: %v", err)
	}

	// We could store the interval in the config in real system. It's just for example
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := s.SyncOnce(); err != nil {
			log.Printf("periodic cache sync failed: %v", err)
		}
	}
}

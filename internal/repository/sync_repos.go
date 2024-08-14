package repository

import "sync"

type syncRepo struct {
	mu sync.Mutex
}

func (r *syncRepo) Lock() {
	r.mu.Lock()
}

func (r *syncRepo) Unlock() {
	r.mu.Unlock()
}

func newSyncRepo() syncRepo {
	return syncRepo{
		mu: sync.Mutex{},
	}
}

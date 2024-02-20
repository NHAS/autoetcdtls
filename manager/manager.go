package manager

import (
	"sync"
	"time"
)

type manager struct {
	sync.RWMutex

	storageDir string
	sessions   map[string]*time.Timer
}

func NewManager(certStore string) *manager {
	return &manager{
		storageDir: certStore,
		sessions:   make(map[string]*time.Timer),
	}
}

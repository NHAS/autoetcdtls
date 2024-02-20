package manager

import (
	"sync"
	"time"
)

type manager struct {
	sync.RWMutex

	storageDir string
	sessions   map[string]*time.Timer

	// On serve, we serve these additional extra bits of data
	additionals map[string]string

	// On sync/join we get data through these function handlers
	additionalsHandlers map[string]func(string, string)
}

func New(certStore string) *manager {
	return &manager{
		storageDir:          certStore,
		sessions:            make(map[string]*time.Timer),
		additionals:         make(map[string]string),
		additionalsHandlers: make(map[string]func(string, string)),
	}
}

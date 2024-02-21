package manager

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"
)

type ActiveToken struct {
	sync.RWMutex

	timer *time.Timer
	// On serve, we serve these additional extra bits of data
	additionals map[string]string

	Token string
}

// SetAdditional adds more configuration data, this should be used for basic information only as large data will struggle to be json marshalled
func (at *ActiveToken) SetAdditional(name, data string) {
	at.Lock()
	defer at.Unlock()

	// Alright not to check trampling here as we're expecting folk to update these values
	at.additionals[name] = data
}

func (m *Manager) serveAdditionals(w http.ResponseWriter, r *http.Request) {
	m.RLock()
	defer m.RUnlock()

	s, ok := m.sessions[r.Header.Get(AuthHeader)]
	if !ok {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(s.additionals)
	if err != nil {
		log.Println("unable to send additional configurations")
		http.Error(w, "Error marshalling additionals", http.StatusInternalServerError)
		return
	}
}

// HandleAdditional registers a function to listen when connecting via "Join" if one of the additionals matches then it executes a function with the extra config data
func (m *Manager) HandleAdditonal(name string, fnc func(name, data string)) error {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.additionalsHandlers[name]; ok {
		return errors.New("additional handler already registered")
	}

	m.additionalsHandlers[name] = fnc

	return nil
}

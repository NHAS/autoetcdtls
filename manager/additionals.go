package manager

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

func (m *manager) SetAdditional(name, data string) {
	m.Lock()
	defer m.Unlock()

	// Alright not to check trampling here as we're expecting folk to update these values
	m.additionals[name] = data
}

func (m *manager) HandleAdditonal(name string, fnc func(name, data string)) error {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.additionalsHandlers[name]; ok {
		return errors.New("additional handler already registered")
	}

	m.additionalsHandlers[name] = fnc

	return nil
}

func (m *manager) serveAdditionals(w http.ResponseWriter, r *http.Request) {
	m.RLock()
	defer m.RUnlock()

	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(m.additionals)
	if err != nil {
		log.Println("unable to send additional configurations")
		http.Error(w, "Error marshalling additionals", http.StatusInternalServerError)
		return
	}
}

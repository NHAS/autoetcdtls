package manager

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"
)

type Token struct {
	NewPeerURL      string `json:"d"`
	ExistingPeerURL string `json:"peer"`
	JoinPassword    string `json:"p"`
	CACertHash      string `json:"h"`
}

func parseToken(token string) (Token, error) {

	fullToken, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return Token{}, err
	}

	var t Token
	err = json.Unmarshal([]byte(fullToken), &t)
	if err != nil {
		return Token{}, err
	}

	return t, err
}

func (m *manager) isTokenValid(joinPassword string) bool {
	m.RLock()
	defer m.RUnlock()
	_, ok := m.sessions[joinPassword]
	return ok
}

func (m *manager) CreateToken(newNodeUrl string) (string, error) {
	m.Lock()
	defer m.Unlock()

	var t Token
	t.ExistingPeerURL = m.domain
	t.NewPeerURL = newNodeUrl
	for {
		t.JoinPassword = generateRandom(32)

		if _, ok := m.sessions[t.JoinPassword]; ok {
			continue
		}

		break
	}
	certBytes, err := os.ReadFile(filepath.Join(m.storageDir, CACertFileName))
	if err != nil {
		return "", err
	}

	hashResult := sha512.Sum512(certBytes)

	t.CACertHash = base64.RawURLEncoding.EncodeToString(hashResult[:])

	m.sessions[t.JoinPassword] = time.AfterFunc(30*time.Second, func() {
		m.Lock()
		defer m.Unlock()

		delete(m.sessions, t.JoinPassword)
	})

	b, err := json.Marshal(t)

	log.Println(string(b))

	return base64.RawURLEncoding.EncodeToString(b), err
}

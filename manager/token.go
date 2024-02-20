package manager

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Token struct {
	Domain       string `json:"d"`
	JoinPassword string `json:"p"`
	CACertHash   string `json:"h"`
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

func (m *manager) CreateToken(domain string) (string, error) {
	m.Lock()
	defer m.Unlock()

	var t Token
	t.Domain = domain
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

	return base64.RawURLEncoding.EncodeToString(b), err
}

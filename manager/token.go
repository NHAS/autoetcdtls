package manager

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	lc       sync.RWMutex
	sessions = map[string]*time.Timer{}
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

func isTokenValid(joinPassword string) bool {
	lc.RLock()
	defer lc.RUnlock()
	_, ok := sessions[joinPassword]
	return ok
}

func CreateToken(certificateStorage, domain string) (string, error) {
	lc.Lock()
	defer lc.Unlock()

	var t Token
	t.Domain = domain
	for {
		t.JoinPassword = generateRandom(32)

		if _, ok := sessions[t.JoinPassword]; ok {
			continue
		}

		break
	}
	certBytes, err := os.ReadFile(filepath.Join(certificateStorage, CACertFileName))
	if err != nil {
		return "", err
	}

	hashResult := sha512.Sum512(certBytes)

	t.CACertHash = base64.RawURLEncoding.EncodeToString(hashResult[:])

	sessions[t.JoinPassword] = time.AfterFunc(30*time.Second, func() {
		lc.Lock()
		defer lc.Unlock()

		delete(sessions, t.JoinPassword)
	})

	b, err := json.Marshal(t)

	return base64.RawURLEncoding.EncodeToString(b), err
}

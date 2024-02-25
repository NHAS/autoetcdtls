package manager

import (
	"errors"
	"net/url"
	"path/filepath"
	"sync"
)

type Manager struct {
	sync.RWMutex

	storageDir string
	domain     string

	sessions map[string]*ActiveToken

	// On sync/join we get data through these function handlers
	additionalsHandlers map[string]func(string, string)

	listenAddress string
}

func (m *Manager) GetCACertPath() string {
	return filepath.Join(m.storageDir, CACertFileName)
}

func (m *Manager) GetPeerCertPath() string {
	return filepath.Join(m.storageDir, PeerCertFileName)
}

func (m *Manager) GetPeerKeyPath() string {
	return filepath.Join(m.storageDir, PeerKeyFileName)
}

// New creates a webserver that syncs cluster certificates, provided it is either the first node, or it has "Joined" the cluster previously
// certStore is a path to where the certificates will be stored if they are generated, or where it will attempt to load the certs from
// urlAddress is where the server listens to, and its publicly accessible address
func New(certStore, urlAddress string) (*Manager, error) {

	m, err := createEmptyManager(certStore, urlAddress)
	if err != nil {
		return nil, err
	}

	err = createOrLoadCerts(certStore, urlAddress)
	if err != nil {
		return nil, err
	}

	return m, m.startListening()
}

// createEmptyManager just checks the url is https, and creates an empty manager object
func createEmptyManager(storageDir, urlAddress string) (*Manager, error) {

	u, err := url.Parse(urlAddress)
	if err != nil {
		return nil, err
	}

	if u.Scheme != "https" {
		return nil, errors.New("unsupported scheme for url, must be https://domain.something: " + urlAddress)
	}

	if u.Port() == "" {
		u.Host += ":443"
	}

	return &Manager{
		storageDir:          storageDir,
		domain:              urlAddress,
		listenAddress:       u.Host,
		sessions:            make(map[string]*ActiveToken),
		additionalsHandlers: make(map[string]func(string, string)),
	}, nil
}

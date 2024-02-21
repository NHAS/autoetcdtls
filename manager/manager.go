package manager

import (
	"errors"
	"net/url"
	"sync"
	"time"
)

type manager struct {
	sync.RWMutex

	storageDir string
	domain     string

	sessions map[string]*time.Timer

	// On serve, we serve these additional extra bits of data
	additionals map[string]string

	// On sync/join we get data through these function handlers
	additionalsHandlers map[string]func(string, string)

	listenAddress string
}

func New(certStore, urlAddress string) (*manager, error) {

	err := createOrLoadCerts(certStore, urlAddress)
	if err != nil {
		return nil, err
	}

	return createEmptyManager(certStore, urlAddress)
}

func createEmptyManager(storageDir, urlAddress string) (*manager, error) {

	u, err := url.Parse(urlAddress)
	if err != nil {
		return nil, err
	}

	if u.Scheme != "https" {
		return nil, errors.New("unsupported scheme for url: " + urlAddress)
	}

	if u.Port() == "" {
		u.Host += ":443"
	}

	return &manager{
		storageDir:          storageDir,
		domain:              urlAddress,
		listenAddress:       u.Host,
		sessions:            make(map[string]*time.Timer),
		additionals:         make(map[string]string),
		additionalsHandlers: make(map[string]func(string, string)),
	}, nil
}

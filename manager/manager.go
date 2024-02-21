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

// New creates a webserver that syncs cluster certificates, provided it is either the first node, or it has "Joined" the cluster previously
// certStore is a path to where the certificates will be stored if they are generated, or where it will attempt to load the certs from
// urlAddress is where the server listens to, and its publicly accessible address
func New(certStore, urlAddress string) (*manager, error) {

	err := createOrLoadCerts(certStore, urlAddress)
	if err != nil {
		return nil, err
	}

	return createEmptyManager(certStore, urlAddress)
}

// createEmptyManager just checks the url is https, and creates an empty manager object
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

package manager

import (
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
)

// startListening executes the TLS http server which has the api for providing a CA cert and configuration data to new clients
func (m *Manager) startListening() error {

	rootMux := http.NewServeMux()

	public := http.NewServeMux()
	public.HandleFunc(getCACert, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")

		f, err := os.Open(filepath.Join(m.storageDir, CACertFileName))
		if err != nil {
			log.Println("failed to serve cluster CA certificate: ", err)
			http.Error(w, "Server Error", http.StatusInternalServerError)
			return
		}
		defer f.Close()

		io.Copy(w, f)
	})

	rootMux.Handle("/public/", public)

	private := http.NewServeMux()
	private.HandleFunc(getCAPrivateKey, func(w http.ResponseWriter, r *http.Request) {

		log.Println("new cluster joiner", r.RemoteAddr, "is downloading CA private key")
		w.Header().Set("Content-Type", "text/plain")

		f, err := os.Open(filepath.Join(m.storageDir, CAKeyFileName))
		if err != nil {
			log.Println("failed to serve cluster CA private key: ", err)
			http.Error(w, "Server Error", http.StatusInternalServerError)
			return
		}
		defer f.Close()

		io.Copy(w, f)
	})

	private.HandleFunc(getAdditionals, m.serveAdditionals)

	rootMux.Handle("/private/", m.basicAuthorisation(private))

	listener, err := net.Listen("tcp", m.listenAddress)
	if err != nil {
		return err
	}

	go func() {
		log.Println("Started tls serving on: ", m.listenAddress)
		err := http.ServeTLS(listener, rootMux, filepath.Join(m.storageDir, PeerCertFileName), filepath.Join(m.storageDir, PeerKeyFileName))
		if err != nil {
			log.Println("Manager TLS crashed: ", err)
		}
	}()

	return nil
}

func (m *Manager) basicAuthorisation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if !m.isTokenValid(r.Header.Get(AuthHeader)) {
			http.NotFound(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

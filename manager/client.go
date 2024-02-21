package manager

import (
	"bytes"
	"crypto/sha512"
	"crypto/subtle"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

// Join connects an instance to existing clusters
// It gets the CA cert and private key, along with any additional configurations that are required
// This is done with a base64 encoded json blob called a "Join token"
func Join(token, certStorage string, additionals map[string]func(name string, data string)) (*manager, error) {

	tokenStruct, err := parseToken(token)
	if err != nil {
		return nil, err
	}

	m, err := createEmptyManager(certStorage, tokenStruct.NewPeerURL)
	if err != nil {
		return nil, err
	}

	for name, fnc := range additionals {
		err := m.HandleAdditonal(name, fnc)
		if err != nil {
			return nil, err
		}
	}

	httpsURL, err := url.Parse(tokenStruct.ExistingPeerURL)
	if err != nil {
		return nil, err
	}

	if httpsURL.Scheme == "" {
		httpsURL.Scheme = "https"
	}

	if httpsURL.Scheme != "https" {
		return nil, errors.New("address url scheme is not supported: " + httpsURL.Scheme)
	}

	originalURLPath := httpsURL.Path

	httpsURL.Path = filepath.Join(originalURLPath, getCACert)

	client := &http.Client{
		Transport: &http.Transport{
			// For the first request we wont have the CA cert, so just skip verifying
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	req, err := http.NewRequest(http.MethodGet, httpsURL.String(), nil)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		response, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("read all error bytes failed: %s", err)
		}

		return nil, errors.New("server returned error fetching ca: " + string(response))
	}

	err = os.MkdirAll(m.storageDir, 0700)
	if err != nil {
		return nil, fmt.Errorf("unable to create certificate storage path: %s", err)
	}

	hasher := sha512.New()

	f, err := os.Create(filepath.Join(m.storageDir, CACertFileName))
	if err != nil {
		return nil, fmt.Errorf("unable to create file for ca cert: %s", err)
	}

	caPEM := bytes.NewBuffer(nil)

	_, err = io.Copy(io.MultiWriter(hasher, f, caPEM), res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to copy ca cert: %s", err)
	}

	f.Close()

	hash := hasher.Sum(nil)
	expectedHash, err := base64.RawURLEncoding.DecodeString(tokenStruct.CACertHash)
	if err != nil {
		return nil, err
	}

	if subtle.ConstantTimeCompare(hash, expectedHash) == 0 {
		return nil, errors.New("ca cert bundle did not match token")
	}

	certpool := x509.NewCertPool()
	ok := certpool.AppendCertsFromPEM(caPEM.Bytes())
	if !ok {
		return nil, errors.New("unable to add ca to cert pool")
	}

	client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			// Now that we have a cert, and are about to issue creds we must check
			RootCAs: certpool,
		},
	}

	httpsURL.Path = filepath.Join(originalURLPath, getCAPrivateKey)
	req, err = http.NewRequest(http.MethodGet, httpsURL.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set(AuthHeader, tokenStruct.JoinPassword)

	res, err = client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		response, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("read all error bytes failed: %s", err)
		}

		return nil, errors.New("server returned error fetching ca key:" + string(response))
	}

	f, err = os.Create(filepath.Join(m.storageDir, CAKeyFileName))
	if err != nil {
		return nil, fmt.Errorf("unable to create file for ca key: %s", err)
	}

	_, err = io.Copy(f, res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to copy ca key: %s", err)
	}

	f.Close()

	if err := createOrLoadCerts(m.storageDir, tokenStruct.NewPeerURL); err != nil {
		return nil, err
	}

	// Do additionals

	m.RLock()
	defer m.RUnlock()

	httpsURL.Path = filepath.Join(originalURLPath, getAdditionals)
	req, err = http.NewRequest(http.MethodGet, httpsURL.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set(AuthHeader, tokenStruct.JoinPassword)

	res, err = client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		response, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("read all error bytes failed: %s", err)
		}

		return nil, errors.New("server returned error fetching additionals:" + string(response))
	}

	err = json.NewDecoder(res.Body).Decode(&m.additionals)
	if err != nil {
		return nil, err
	}

	for name, data := range m.additionals {
		if fnc, ok := m.additionalsHandlers[name]; ok {
			go fnc(name, data)
		}
	}

	return m, nil
}

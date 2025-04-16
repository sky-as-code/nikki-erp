package util

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
)

func LoadServerKey(certfile, certkey, certca string) (*http.Server, error) {
	if certfile == "" {
		return nil, fmt.Errorf("no certfile")
	}

	var cert tls.Certificate
	cert, err := tls.LoadX509KeyPair(certfile, certkey)
	if err != nil {
		return nil, fmt.Errorf("failed to load x509 key pair: %v", err)
	}
	tlsConfig := tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	// Require and verify user cert
	if certca != "" {
		pool := x509.NewCertPool()
		ca, err := ioutil.ReadFile(certca)
		if err != nil {
			return nil, fmt.Errorf("failed to read ca cert: %v", err)
		}
		if !pool.AppendCertsFromPEM(ca) {
			return nil, fmt.Errorf("failed to add ca cert: %v", err)
		}
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		tlsConfig.ClientCAs = pool
	}

	return &http.Server{TLSConfig: &tlsConfig}, nil
}

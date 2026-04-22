package client

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/core/constants"
	"golang.org/x/crypto/pkcs12"
)

const (
	DefaultHTTPTimeout = 30 * time.Second
)

type HTTPClient struct {
	http.Client
}

func NewCoreHTTPClient(cfg config.ConfigService) *HTTPClient {
	httpClientConfig := HTTPClientConfig{
		Timeout: cfg.GetInt64(constants.HttpClientTimeout),
		TLSConfig: TLSConfig{
			InsecureSkipVerify:     cfg.GetBool(constants.HttpClientSkipVerifyServer),
			IncludeSystemTrustedCA: cfg.GetBool(constants.HttpClientIncludeSystemTrustedCA),
			CustomTrustedCAs:       cfg.GetStrArr(constants.HttpClientCustomTrustedCACertPaths),
		},
		ClientCertConfig: ClientCertConfig{
			Enabled: cfg.GetBool(constants.HttpClientClientCertEnabled),
		},
	}

	if httpClientConfig.ClientCertConfig.Enabled {
		httpClientConfig.ClientCertConfig.Cert = cfg.GetStr(constants.HttpClientClientCert)
		httpClientConfig.ClientCertConfig.PrivateKey = cfg.GetStr(constants.HttpClientClientCertKey)
	}

	c, err := NewHTTPClient(&httpClientConfig)
	if err != nil {
		panic(err)
	}

	return c
}

func NewHTTPClient(cfg *HTTPClientConfig) (*HTTPClient, error) {
	rootClient, err := httpClientFromConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &HTTPClient{*rootClient}, nil
}

func httpClientFromConfig(cfg *HTTPClientConfig) (*http.Client, error) {
	if cfg == nil {
		return nil, errors.New("missing config")
	}

	tlsClientCfg := &tls.Config{}

	// ========== TLS Config ===============
	if cfg.TLSConfig.InsecureSkipVerify {
		tlsClientCfg.InsecureSkipVerify = true
	} else {
		var err error
		var certPool = x509.NewCertPool()

		//  Loads System trusted CAs
		if cfg.TLSConfig.IncludeSystemTrustedCA {
			certPool, err = x509.SystemCertPool()
			if err != nil {
				return nil, err
			}
		}

		// Appends custom CAs
		for _, caPath := range cfg.TLSConfig.CustomTrustedCAs {
			caCert, err := os.ReadFile(caPath)
			if err != nil {
				return nil, err
			}

			certPool.AppendCertsFromPEM(caCert)
		}

		tlsClientCfg.RootCAs = certPool
	}

	// ========== Client Cert config =============
	if cfg.ClientCertConfig.Enabled {
		// Create client certs from cert path config
		cert, err := loadClientCert(cfg.ClientCertConfig)
		if err != nil {
			return nil, err
		}

		tlsClientCfg.Certificates = []tls.Certificate{cert}
	}

	// Create transport from tlsClientConfig
	transport := &http.Transport{
		TLSClientConfig: tlsClientCfg,
	}

	timeout := DefaultHTTPTimeout
	if cfg.Timeout != 0 {
		timeout = time.Duration(cfg.Timeout) * time.Millisecond
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}, nil
}

func loadClientCert(crtCfg ClientCertConfig) (tls.Certificate, error) {
	if crtCfg.P12Raw != "" {
		private, cert, err := pkcs12.Decode([]byte(crtCfg.P12Raw), crtCfg.Cert)
		if err != nil {
			return tls.Certificate{}, err
		}

		return tls.Certificate{
			Certificate: [][]byte{cert.Raw},
			PrivateKey:  private,
			Leaf:        cert,
		}, nil
	}

	return tls.X509KeyPair([]byte(crtCfg.Cert), []byte(crtCfg.PrivateKey))
}

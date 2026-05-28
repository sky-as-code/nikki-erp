package client

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net/http"
	"time"

	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/core/constants"
	"golang.org/x/crypto/pkcs12"
)

const (
	DefaultHttpTimeout = 30 * time.Second
)

type HttpClient struct {
	http.Client
}

func NewCoreHttpClient(cfg config.ConfigService) *HttpClient {
	httpClientConfig := HttpClientConfig{
		Timeout: cfg.GetInt64(constants.HttpClientTimeout),
		TlsConfig: TlsConfig{
			InsecureSkipVerify:     cfg.GetBool(constants.HttpClientSkipVerifyServer, false),
			IncludeSystemTrustedCa: cfg.GetBool(constants.HttpClientIncludeSystemTrustedCA, true),
			CustomTrustedCa:        cfg.GetStr(constants.HttpClientCustomTrustedCACert, ""),
		},
		ClientCertConfig: ClientCertConfig{
			Enabled:     cfg.GetBool(constants.HttpClientClientCertEnabled, false),
			Cert:        cfg.GetStr(constants.HttpClientClientCert, ""),
			PrivateKey:  cfg.GetStr(constants.HttpClientClientCertKey, ""),
			P12Raw:      cfg.GetStr(constants.HttpClientClientP12, ""),
			P12Password: cfg.GetStr(constants.HttpClientClientP12Password, ""),
		},
	}

	c, err := NewHttpClient(&httpClientConfig)
	if err != nil {
		panic(err)
	}

	return c
}

func NewHttpClient(cfg *HttpClientConfig) (*HttpClient, error) {
	rootClient, err := httpClientFromConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &HttpClient{*rootClient}, nil
}

func httpClientFromConfig(cfg *HttpClientConfig) (*http.Client, error) {
	if cfg == nil {
		return nil, errors.New("missing config")
	}

	tlsClientCfg := &tls.Config{}

	// ========== TLS Config ===============
	if cfg.TlsConfig.InsecureSkipVerify {
		tlsClientCfg.InsecureSkipVerify = true
	} else {
		var err error
		var certPool = x509.NewCertPool()

		//  Loads System trusted CAs
		if cfg.TlsConfig.IncludeSystemTrustedCa {
			certPool, err = x509.SystemCertPool()
			if err != nil {
				return nil, err
			}
		}

		// Appends custom CAs
		if cfg.TlsConfig.CustomTrustedCa != "" {
			certPool.AppendCertsFromPEM([]byte(cfg.TlsConfig.CustomTrustedCa))
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

	timeout := DefaultHttpTimeout
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

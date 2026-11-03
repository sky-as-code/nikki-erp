package client

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"time"

	"go.bryk.io/pkg/errors"
	"golang.org/x/crypto/pkcs12"

	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/core/constants"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/core/telemetry"
)

const (
	DefaultHttpTimeout = 30 * time.Second
)

// HttpClient is the one outbound HTTP client. Its transport is instrumented (see
// newInstrumentedTransport), so a client derived from it must keep Transport, as WithTimeout does.
type HttpClient struct {
	http.Client
}

// WithTimeout derives a client that shares this one's instrumented transport and connection pool
// but bounds each request by timeout instead of the shared ceiling.
func (this *HttpClient) WithTimeout(timeout time.Duration) *http.Client {
	derived := this.Client
	derived.Timeout = timeout
	return &derived
}

// NewCoreHttpClient reads the propagation allowlist when DI first resolves the client, which is
// after telemetry.InitSubModule has loaded it.
func NewCoreHttpClient(cfg config.ConfigService, logger logging.LoggerService) *HttpClient {
	httpClientConfig := HttpClientConfig{
		Timeout:                  cfg.GetInt64(constants.HttpClientTimeout),
		PropagationHostAllowlist: telemetry.PropagationHostAllowlist(),
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

	c, err := NewHttpClient(&httpClientConfig, logger)
	if err != nil {
		panic(err)
	}

	return c
}

func NewHttpClient(cfg *HttpClientConfig, logger logging.LoggerService) (*HttpClient, error) {
	rootClient, err := httpClientFromConfig(cfg)
	if err != nil {
		return nil, err
	}
	rootClient.Transport = newInstrumentedTransport(rootClient.Transport, cfg.PropagationHostAllowlist, logger)
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

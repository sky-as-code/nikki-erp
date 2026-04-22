package client

type HTTPClientConfig struct {
	Timeout int64 // milisecond

	TLSConfig TLSConfig

	// for mTLS
	ClientCertConfig ClientCertConfig
}

type TLSConfig struct {
	InsecureSkipVerify     bool // Skip verify server, default is false
	IncludeSystemTrustedCA bool
	CustomTrustedCAs       []string // CAs cert paths
}

// ClientCert is either P12 or Key-pair
type ClientCertConfig struct {
	Enabled bool // enable verify client

	Cert       string
	PrivateKey string

	P12Raw      string
	P12Password string
}

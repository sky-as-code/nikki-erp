package client

type HttpClientConfig struct {
	Timeout int64 // milisecond

	TlsConfig TlsConfig

	// for mTLS
	ClientCertConfig ClientCertConfig

	// PropagationHostAllowlist names the hosts that receive traceparent, tracestate, baggage and
	// X-Idempotency-ID (CORE.TELEMETRY.PROPAGATION_HOST_ALLOWLIST). Empty: no host does.
	PropagationHostAllowlist []string
}

type TlsConfig struct {
	InsecureSkipVerify     bool // Skip verify server, default is false
	IncludeSystemTrustedCa bool
	CustomTrustedCa        string // CA cert
}

// ClientCert is either P12 or Key-pair
type ClientCertConfig struct {
	Enabled bool // enable verify client

	Cert       string
	PrivateKey string

	P12Raw      string
	P12Password string
}

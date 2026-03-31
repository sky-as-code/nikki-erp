package dialects

import (
	"fmt"
	"net/url"
)

func buildPgDsn(opts DialectOptions) string {
	var sslmode string

	if opts.IsTlsEnabled {
		sslmode = "require"
	} else {
		sslmode = "disable"
	}

	// Properly encode the password (optional but safer)
	escapedPassword := url.QueryEscape(opts.Password)

	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?TimeZone=UTC&sslmode=%s",
		opts.User, escapedPassword, opts.HostPort, opts.Database, sslmode)

	return dsn
}

package dialects

import (
	"fmt"
	"net/url"
)

func buildMysqlDsn(opts DialectOptions) string {
	var tls string

	if opts.IsTlsEnabled {
		tls = "true"
	} else {
		tls = "false"
	}

	// Properly encode the password (optional but safer)
	escapedPassword := url.QueryEscape(opts.Password)

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&tls=%s",
		opts.User, escapedPassword, opts.HostPort, opts.Database, tls)

	return dsn
}

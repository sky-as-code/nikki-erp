package orm

import "github.com/sky-as-code/nikki-erp/modules/core/database/dialects"

type PgDynamicOptions struct {
	dialects.DialectOptions

	DebugEnabled bool
}

// func InitPostgresDynamicOrm(opts EntDriverOptions) error {
// 	conn, err := dialects.OpenConnection(opts.DialectName, opts.DialectOptions)
// 	if err != nil {
// 		return err
// 	}
// }

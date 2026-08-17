package dynamicengines

import (
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
)

// The Currency engine.
//
// Currency is reference data with no lifecycle beyond archiving, so it carries no action of its
// own: the built-in CRUD is the whole of its surface. Withdrawing a currency from use is
// is_active, a plain field, rather than an action, because amounts already recorded in it must
// stay readable either way.
func currencyEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.CurrencySchemaName,
		DefaultFields: []string{
			models.CurrencyFieldCode,
			models.CurrencyFieldName,
			models.CurrencyFieldSymbol,
			models.CurrencyFieldDecimalPlaces,
			models.CurrencyFieldIsActive,
		},
	}
}

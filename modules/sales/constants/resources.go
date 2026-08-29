package constants

import "github.com/sky-as-code/nikki-erp/modules/sales/domain/models"

// Resource codes for authorization. Each must stay byte-identical to its schema name, since the
// dynamic resource engine derives the asserted resource code from the schema name; a code that
// drifts denies every request with no hint in the 403. Aliased rather than re-typed to prevent
// that drift.
const (
	SalesChannelResource = models.SalesChannelSchemaName
	SalesPointResource   = models.SalesPointSchemaName
)

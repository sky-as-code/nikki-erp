package constants

import "github.com/sky-as-code/nikki-erp/modules/sales/domain/models"

// Resource codes for authorization.
//
// Each is byte-identical to its schema name, and that is a hard requirement rather than a
// convention: the dynamic resource engine derives the resource code it asserts against from the
// schema name of the engine handling the request. A code that drifts from its schema denies every
// request to that resource, with nothing in the 403 pointing at the seed as the cause.
//
// They are aliases of the model constants rather than repeated string literals, so that the two
// cannot drift in the first place.
const (
	SalesChannelResource = models.SalesChannelSchemaName
	SalesPointResource   = models.SalesPointSchemaName
)

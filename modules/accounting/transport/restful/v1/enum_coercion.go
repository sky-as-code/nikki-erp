package v1

import (
	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
)

// A snapshot arrives from the caller's own storage, so its enum strings are whatever was written
// there — possibly by an older version of this service, possibly corrupted. The Wrap* functions
// answer nil for a value outside the set, and these coercions turn that nil into the zero value
// rather than letting an unrecognized string travel on as if it were meaningful. A reversal that
// silently rounded by an unknown method would be worse than one that rounds by none.

func roundingScopeOf(value string) models.RoundingScope {
	if scope := models.WrapRoundingScope(value); scope != nil {
		return *scope
	}
	return ""
}

func roundingMethodOf(value string) models.RoundingMethod {
	if method := models.WrapRoundingMethod(value); method != nil {
		return *method
	}
	return ""
}

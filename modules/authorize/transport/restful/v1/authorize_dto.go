package v1

import (
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize"
)

type IsAuthorizedRequest = it.IsAuthorizedQuery
type IsAuthorizedResponse it.IsAuthorizedResult

func (this *IsAuthorizedResponse) FromResult(result it.IsAuthorizedResult) {
	this.Decision = result.Decision
	this.ClientError = result.ClientError
}

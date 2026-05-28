package app

import deps "github.com/sky-as-code/nikki-erp/common/deps_inject"

func InitApplicationServices() error {
	return deps.Register(
		NewPurchaseOrderApplicationServiceImpl,
		NewPurchaseRequestApplicationServiceImpl,
		NewRequestForProposalApplicationServiceImpl,
		NewRequestForQuoteApplicationServiceImpl,
		NewVendorApplicationServiceImpl,
	)
}

package repository

import (
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

func InitRepositories() error {
	return deps.Register(
		NewPurchaseOrderDynamicRepository,
		NewPurchaseRequestDynamicRepository,
		NewRequestForProposalDynamicRepository,
		NewRequestForQuoteDynamicRepository,
		NewVendorDynamicRepository,
	)
}

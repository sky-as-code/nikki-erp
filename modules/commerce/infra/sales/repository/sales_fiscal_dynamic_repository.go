package repository

import (
	itFiscal "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/fiscal"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

func NewSalesFiscalRequestRepository(base dynamicresource.CrudRepository) itFiscal.SalesFiscalRequestRepository {
	return &SalesFiscalRequestRepositoryImpl{CrudRepository: base}
}

type SalesFiscalRequestRepositoryImpl struct {
	dynamicresource.CrudRepository
}

func NewSalesBillingInstructionRepository(base dynamicresource.CrudRepository) itFiscal.SalesBillingInstructionRepository {
	return &SalesBillingInstructionRepositoryImpl{CrudRepository: base}
}

type SalesBillingInstructionRepositoryImpl struct {
	dynamicresource.CrudRepository
}

func NewSalesBillingIssuanceAttemptRepository(base dynamicresource.CrudRepository) itFiscal.SalesBillingIssuanceAttemptRepository {
	return &SalesBillingIssuanceAttemptRepositoryImpl{CrudRepository: base}
}

type SalesBillingIssuanceAttemptRepositoryImpl struct {
	dynamicresource.CrudRepository
}

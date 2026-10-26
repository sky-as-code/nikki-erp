package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itFiscal "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/fiscal"
)

func NewSalesFiscalRequestRepository(base composable.CrudRepository) itFiscal.SalesFiscalRequestRepository {
	return &SalesFiscalRequestRepositoryImpl{CrudRepository: base}
}

type SalesFiscalRequestRepositoryImpl struct {
	composable.CrudRepository
}

func NewSalesBillingInstructionRepository(base composable.CrudRepository) itFiscal.SalesBillingInstructionRepository {
	return &SalesBillingInstructionRepositoryImpl{CrudRepository: base}
}

type SalesBillingInstructionRepositoryImpl struct {
	composable.CrudRepository
}

func NewSalesBillingIssuanceAttemptRepository(base composable.CrudRepository) itFiscal.SalesBillingIssuanceAttemptRepository {
	return &SalesBillingIssuanceAttemptRepositoryImpl{CrudRepository: base}
}

type SalesBillingIssuanceAttemptRepositoryImpl struct {
	composable.CrudRepository
}

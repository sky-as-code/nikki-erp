package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itQuotation "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/quotation"
)

func NewSalesQuotationRepository(base composable.CrudRepository) itQuotation.SalesQuotationRepository {
	return &SalesQuotationRepositoryImpl{CrudRepository: base}
}

type SalesQuotationRepositoryImpl struct {
	composable.CrudRepository
}

func NewSalesQuotationLineRepository(base composable.CrudRepository) itQuotation.SalesQuotationLineRepository {
	return &SalesQuotationLineRepositoryImpl{CrudRepository: base}
}

type SalesQuotationLineRepositoryImpl struct {
	composable.CrudRepository
}

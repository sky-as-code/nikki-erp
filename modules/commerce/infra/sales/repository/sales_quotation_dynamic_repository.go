package repository

import (
	itQuotation "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/quotation"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
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

package repository

import (
	itQuotation "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/quotation"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

func NewSalesQuotationRepository(base dynamicresource.CrudRepository) itQuotation.SalesQuotationRepository {
	return &SalesQuotationRepositoryImpl{CrudRepository: base}
}

type SalesQuotationRepositoryImpl struct {
	dynamicresource.CrudRepository
}

func NewSalesQuotationLineRepository(base dynamicresource.CrudRepository) itQuotation.SalesQuotationLineRepository {
	return &SalesQuotationLineRepositoryImpl{CrudRepository: base}
}

type SalesQuotationLineRepositoryImpl struct {
	dynamicresource.CrudRepository
}

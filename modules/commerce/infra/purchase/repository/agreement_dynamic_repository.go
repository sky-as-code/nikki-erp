package repository

import (
	itAgreement "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/agreement"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

func NewAgreementRepository(base dynamicresource.CrudRepository) itAgreement.AgreementCrudRepository {
	return &AgreementRepositoryImpl{CrudRepository: base}
}

type AgreementRepositoryImpl struct {
	dynamicresource.CrudRepository
}

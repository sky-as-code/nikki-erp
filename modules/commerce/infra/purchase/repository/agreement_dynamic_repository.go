package repository

import (
	itAgreement "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/agreement"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

func NewAgreementRepository(base composable.CrudRepository) itAgreement.AgreementCrudRepository {
	return &AgreementRepositoryImpl{CrudRepository: base}
}

type AgreementRepositoryImpl struct {
	composable.CrudRepository
}

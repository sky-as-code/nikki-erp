package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itAgreement "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/agreement"
)

func NewAgreementRepository(base composable.CrudRepository) itAgreement.AgreementCrudRepository {
	return &AgreementRepositoryImpl{CrudRepository: base}
}

type AgreementRepositoryImpl struct {
	composable.CrudRepository
}

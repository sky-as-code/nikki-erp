package repository

import (
	itAgreement "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/agreement"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

func NewAgreementLineRepository(base composable.CrudRepository) itAgreement.AgreementLineCrudRepository {
	return &AgreementLineRepositoryImpl{CrudRepository: base}
}

type AgreementLineRepositoryImpl struct {
	composable.CrudRepository
}

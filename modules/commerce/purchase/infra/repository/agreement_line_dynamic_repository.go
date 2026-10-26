package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itAgreement "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/agreement"
)

func NewAgreementLineRepository(base composable.CrudRepository) itAgreement.AgreementLineCrudRepository {
	return &AgreementLineRepositoryImpl{CrudRepository: base}
}

type AgreementLineRepositoryImpl struct {
	composable.CrudRepository
}

package role_suite

import "github.com/sky-as-code/nikki-erp/modules/core/crud"

type RoleSuiteService interface {
	CreateRoleSuite(ctx crud.Context, cmd CreateRoleSuiteCommand) (*CreateRoleSuiteResult, error)
	UpdateRoleSuite(ctx crud.Context, cmd UpdateRoleSuiteCommand) (*UpdateRoleSuiteResult, error)
	DeleteHardRoleSuite(ctx crud.Context, cmd DeleteRoleSuiteCommand) (*DeleteRoleSuiteResult, error)
	GetRoleSuiteById(ctx crud.Context, cmd GetRoleSuiteByIdQuery) (*GetRoleSuiteByIdResult, error)
	GetRoleSuitesBySubject(ctx crud.Context, query GetRoleSuitesBySubjectQuery) (*GetRoleSuitesBySubjectResult, error)
	SearchRoleSuites(ctx crud.Context, query SearchRoleSuitesCommand) (*SearchRoleSuitesResult, error)
}

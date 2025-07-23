package role_suite

import "context"

type RoleSuiteService interface {
	CreateRoleSuite(ctx context.Context, cmd CreateRoleSuiteCommand) (*CreateRoleSuiteResult, error)
	GetRoleSuiteById(ctx context.Context, cmd GetRoleSuiteByIdQuery) (*GetRoleSuiteByIdResult, error)
	GetRoleSuitesBySubject(ctx context.Context, query GetRoleSuitesBySubjectQuery) (*GetRoleSuitesBySubjectResult, error)
	SearchRoleSuites(ctx context.Context, query SearchRoleSuitesCommand) (*SearchRoleSuitesResult, error)
}

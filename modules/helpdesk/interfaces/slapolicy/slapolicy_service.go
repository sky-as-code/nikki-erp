package slapolicy

import corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

type SlaPolicyService interface {
	CreateSlaPolicy(ctx corectx.Context, cmd CreateSlaPolicyCommand) (*CreateSlaPolicyResult, error)
	DeleteSlaPolicy(ctx corectx.Context, cmd DeleteSlaPolicyCommand) (*DeleteSlaPolicyResult, error)
	GetSlaPolicy(ctx corectx.Context, query GetSlaPolicyQuery) (*GetSlaPolicyResult, error)
	SlaPolicyExists(ctx corectx.Context, query SlaPolicyExistsQuery) (*SlaPolicyExistsResult, error)
	SearchSlaPolicies(ctx corectx.Context, query SearchSlaPoliciesQuery) (*SearchSlaPoliciesResult, error)
	UpdateSlaPolicy(ctx corectx.Context, cmd UpdateSlaPolicyCommand) (*UpdateSlaPolicyResult, error)
	SetSlaPolicyIsArchived(ctx corectx.Context, cmd SetSlaPolicyIsArchivedCommand) (*SetSlaPolicyIsArchivedResult, error)
}

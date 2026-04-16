package teammembership

import corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

type TeamMembershipService interface {
	CreateTeamMembership(ctx corectx.Context, cmd CreateTeamMembershipCommand) (*CreateTeamMembershipResult, error)
	DeleteTeamMembership(ctx corectx.Context, cmd DeleteTeamMembershipCommand) (*DeleteTeamMembershipResult, error)
	GetTeamMembership(ctx corectx.Context, query GetTeamMembershipQuery) (*GetTeamMembershipResult, error)
	TeamMembershipExists(ctx corectx.Context, query TeamMembershipExistsQuery) (*TeamMembershipExistsResult, error)
	SearchTeamMemberships(ctx corectx.Context, query SearchTeamMembershipsQuery) (*SearchTeamMembershipsResult, error)
	UpdateTeamMembership(ctx corectx.Context, cmd UpdateTeamMembershipCommand) (*UpdateTeamMembershipResult, error)
}

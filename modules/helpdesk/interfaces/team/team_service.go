package team

import corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

type TeamService interface {
	CreateTeam(ctx corectx.Context, cmd CreateTeamCommand) (*CreateTeamResult, error)
	DeleteTeam(ctx corectx.Context, cmd DeleteTeamCommand) (*DeleteTeamResult, error)
	GetTeam(ctx corectx.Context, query GetTeamQuery) (*GetTeamResult, error)
	TeamExists(ctx corectx.Context, query TeamExistsQuery) (*TeamExistsResult, error)
	SearchTeams(ctx corectx.Context, query SearchTeamsQuery) (*SearchTeamsResult, error)
	UpdateTeam(ctx corectx.Context, cmd UpdateTeamCommand) (*UpdateTeamResult, error)
	SetTeamIsArchived(ctx corectx.Context, cmd SetTeamIsArchivedCommand) (*SetTeamIsArchivedResult, error)
}

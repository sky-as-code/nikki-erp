package requestforproposal

import corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

type RequestForProposalService interface {
	CreateRequestForProposal(ctx corectx.Context, cmd CreateRequestForProposalCommand) (*CreateRequestForProposalResult, error)
	DeleteRequestForProposal(ctx corectx.Context, cmd DeleteRequestForProposalCommand) (*DeleteRequestForProposalResult, error)
	RequestForProposalExists(ctx corectx.Context, query RequestForProposalExistsQuery) (*RequestForProposalExistsResult, error)
	GetRequestForProposal(ctx corectx.Context, query GetRequestForProposalQuery) (*GetRequestForProposalResult, error)
	SearchRequestForProposals(ctx corectx.Context, query SearchRequestForProposalsQuery) (*SearchRequestForProposalsResult, error)
	SetRequestForProposalIsArchived(
		ctx corectx.Context, cmd SetRequestForProposalIsArchivedCommand,
	) (*SetRequestForProposalIsArchivedResult, error)
	UpdateRequestForProposal(ctx corectx.Context, cmd UpdateRequestForProposalCommand) (*UpdateRequestForProposalResult, error)
}

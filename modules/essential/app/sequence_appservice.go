package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itSeq "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/sequence"
)

func NewSequenceAppServiceImpl(domainSvc itSeq.SequenceDomainService) itSeq.SequenceService {
	return &SequenceAppServiceImpl{domainSvc: domainSvc}
}

// SequenceAppServiceImpl is the cross-module port. It asserts no permission (AGENTS.md: a port's
// caller already authorized its own action); it translates the API algorithm name and settles the
// idempotency id: the command's, else the one the request carries, else a fresh one for a caller
// with no request behind it (a job, an org initialization).
type SequenceAppServiceImpl struct {
	domainSvc itSeq.SequenceDomainService
}

func (this *SequenceAppServiceImpl) UpsertDefinition(
	ctx corectx.Context, cmd itSeq.UpsertDefinitionCommand,
) (*itSeq.UpsertDefinitionResult, error) {
	if cmd.Algorithm != nil {
		stored := StoredAlgorithmOf(*cmd.Algorithm)
		cmd.Algorithm = &stored
	}
	return this.domainSvc.UpsertDefinition(ctx, cmd)
}

func (this *SequenceAppServiceImpl) NextValue(
	ctx corectx.Context, cmd itSeq.NextValueCommand,
) (*itSeq.NextValueResult, error) {
	if cmd.IdempotencyId == nil || *cmd.IdempotencyId == "" {
		id, found := corectx.IdempotencyIdOf(ctx)
		if !found {
			minted, err := corectx.NewIdempotencyId()
			if err != nil {
				return nil, err
			}
			id = minted
		}
		cmd.IdempotencyId = &id
	}
	return this.domainSvc.NextValue(ctx, cmd)
}

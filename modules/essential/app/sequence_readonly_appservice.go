package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	itSeq "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/sequence"
)

var (
	counterOutput  = sequenceOutput{stringified: []string{models.SeqCtrFieldLastNumber}}
	issuanceOutput = sequenceOutput{stringified: []string{models.SeqIssFieldAllocatedValue}}
)

func NewSequenceCounterApplicationService(
	base dynamicresource.CrudApplicationService,
) itSeq.SequenceCounterApplicationService {
	return &SequenceCounterApplicationServiceImpl{readOnlySequenceApp{CrudApplicationService: base, output: counterOutput}}
}

func NewSequenceIssuanceApplicationService(
	base dynamicresource.CrudApplicationService,
) itSeq.SequenceIssuanceApplicationService {
	return &SequenceIssuanceApplicationServiceImpl{readOnlySequenceApp{CrudApplicationService: base, output: issuanceOutput}}
}

type SequenceCounterApplicationServiceImpl struct {
	readOnlySequenceApp
}

type SequenceIssuanceApplicationServiceImpl struct {
	readOnlySequenceApp
}

// readOnlySequenceApp serves the counter and issuance reads: the default authorizes, this reshapes the
// output (BR §6.1: large integers as strings) and caps the page (BR §8.3).
type readOnlySequenceApp struct {
	dynamicresource.CrudApplicationService

	output sequenceOutput
}

func (this *readOnlySequenceApp) GetById(
	ctx corectx.Context, query dynamicresource.GetByIdQuery,
) (*dynamicresource.GetOneResult, error) {
	return this.output.one(this.CrudApplicationService.GetById(ctx, query))
}

func (this *readOnlySequenceApp) GetOne(
	ctx corectx.Context, query dynamicresource.GetOneQuery,
) (*dynamicresource.GetOneResult, error) {
	return this.output.one(this.CrudApplicationService.GetOne(ctx, query))
}

func (this *readOnlySequenceApp) Search(
	ctx corectx.Context, query dynamicresource.SearchQuery,
) (*dynamicresource.SearchResult, error) {
	if cErrs := guardQuery(query); cErrs != nil {
		return &dynamicresource.SearchResult{ClientErrors: *cErrs}, nil
	}
	return this.output.paged(this.CrudApplicationService.Search(ctx, query))
}

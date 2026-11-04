package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	itSeq "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/sequence"
)

var definitionOutput = sequenceOutput{}

// definitionServerOwned are refused on every write: the server computes or stamps them.
var definitionServerOwned = []string{models.SeqDefFieldScopeType, models.SeqDefFieldFirstIssuedAt}

func NewSequenceDefinitionApplicationService(
	base dynamicresource.CrudApplicationService,
) itSeq.SequenceDefinitionApplicationService {
	return &SequenceDefinitionApplicationServiceImpl{CrudApplicationService: base}
}

// SequenceDefinitionApplicationServiceImpl is the authorized CRUD of definitions, speaking the API
// vocabulary: the algorithm is "pseudorandom" for the stored FF1_FPE, and algorithm_params never
// crosses the boundary in either direction. The composable default asserts every built-in action.
type SequenceDefinitionApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}

func (this *SequenceDefinitionApplicationServiceImpl) Create(
	ctx corectx.Context, cmd dynamicresource.CreateCommand,
) (*dynamicresource.CreateResult, error) {
	if cErrs := refuseSecretWrite(cmd, definitionServerOwned...); cErrs != nil {
		return &dynamicresource.CreateResult{ClientErrors: *cErrs}, nil
	}
	mapAlgorithmInput(cmd)
	result, err := this.CrudApplicationService.Create(ctx, cmd)
	if err == nil && result != nil && result.HasData {
		definitionOutput.record(result.Data)
	}
	return result, err
}

func (this *SequenceDefinitionApplicationServiceImpl) Update(
	ctx corectx.Context, cmd dynamicresource.UpdateCommand,
) (*dynamicresource.MutateResult, error) {
	if cErrs := refuseSecretWrite(cmd, definitionServerOwned...); cErrs != nil {
		return &dynamicresource.MutateResult{ClientErrors: *cErrs}, nil
	}
	mapAlgorithmInput(cmd)
	return this.CrudApplicationService.Update(ctx, cmd)
}

func (this *SequenceDefinitionApplicationServiceImpl) GetById(
	ctx corectx.Context, query dynamicresource.GetByIdQuery,
) (*dynamicresource.GetOneResult, error) {
	return definitionOutput.one(this.CrudApplicationService.GetById(ctx, query))
}

func (this *SequenceDefinitionApplicationServiceImpl) GetOne(
	ctx corectx.Context, query dynamicresource.GetOneQuery,
) (*dynamicresource.GetOneResult, error) {
	return definitionOutput.one(this.CrudApplicationService.GetOne(ctx, query))
}

func (this *SequenceDefinitionApplicationServiceImpl) Search(
	ctx corectx.Context, query dynamicresource.SearchQuery,
) (*dynamicresource.SearchResult, error) {
	if cErrs := guardQuery(query); cErrs != nil {
		return &dynamicresource.SearchResult{ClientErrors: *cErrs}, nil
	}
	if err := storedAlgorithmGraph(query); err != nil {
		return nil, err
	}
	return definitionOutput.paged(this.CrudApplicationService.Search(ctx, query))
}

// GetSchema serves the schema with the algorithm enum in the API vocabulary (pseudorandom, not FF1_FPE).
func (this *SequenceDefinitionApplicationServiceImpl) GetSchema(
	ctx corectx.Context, query dynamicresource.GetSchemaQuery,
) (*dynamicresource.GetSchemaResult, error) {
	result, err := this.CrudApplicationService.GetSchema(ctx, query)
	if err != nil || result == nil || !result.HasData {
		return result, err
	}
	schema, err := apiAlgorithmSchema(result.Data)
	if err != nil {
		return nil, err
	}
	result.Data = schema
	return result, nil
}

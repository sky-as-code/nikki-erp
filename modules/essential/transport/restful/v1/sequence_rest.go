package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	itSeq "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/sequence"
)

type sequenceRestParams struct {
	dig.In

	Definition dynamicresource.DynamicResourceEngineOnion `name:"dynengine_essential_sequence_definition"`
	Counter    dynamicresource.DynamicResourceEngineOnion `name:"dynengine_essential_sequence_counter"`
	Issuance   dynamicresource.DynamicResourceEngineOnion `name:"dynengine_essential_sequence_issuance"`
}

// SequenceRest groups the handlers of the three Sequence resources.
type SequenceRest struct {
	Definition *SequenceDefinitionRest
	Counter    *SequenceReadRest
	Issuance   *SequenceReadRest
}

func NewSequenceRest(params sequenceRestParams) *SequenceRest {
	definition := &SequenceDefinitionRest{
		definitionSvc: params.Definition.ApplicationService().(itSeq.SequenceDefinitionApplicationService),
	}
	definition.SetApplicationService(params.Definition.ApplicationService())
	counter := &SequenceReadRest{}
	counter.SetApplicationService(params.Counter.ApplicationService())
	issuance := &SequenceReadRest{}
	issuance.SetApplicationService(params.Issuance.ApplicationService())
	return &SequenceRest{Definition: definition, Counter: counter, Issuance: issuance}
}

// SequenceDefinitionRest serves the definition CRUD (no delete) and the preview action.
type SequenceDefinitionRest struct {
	dynamicresource.CrudRestBase

	definitionSvc itSeq.SequenceDefinitionApplicationService
}

func (this *SequenceDefinitionRest) Preview(echoCtx *echo.Context, payload map[string]any) error {
	return dynamicresource.ServeAction(echoCtx, "preview sequence definition", payload,
		this.definitionSvc.Preview, dynamicresource.Identity[itSeq.PreviewResultData])
}

// SequenceReadRest serves the read-only counter and issuance listings.
type SequenceReadRest struct {
	dynamicresource.CrudRestBase
}

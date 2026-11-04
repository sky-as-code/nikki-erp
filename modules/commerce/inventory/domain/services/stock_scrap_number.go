package services

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	invsettings "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/external"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// ScrapSequence numbers an organization's scrap documents from the inventory.scrap definition,
// whose id the org keeps in Inventory's own org settings (scrap_sequence_definition_id).
type ScrapSequence struct {
	sequences itExt.SequenceExtService
	reader    itExt.OrgSettingsReaderExtService
}

func NewScrapSequence(sequences itExt.SequenceExtService, reader itExt.OrgSettingsReaderExtService) *ScrapSequence {
	return &ScrapSequence{sequences: sequences, reader: reader}
}

// Next issues the next scrap number of the org. Client errors (numbering not set up, the definition
// archived) must make the caller roll back: the number is reserved in the caller's transaction.
func (this *ScrapSequence) Next(ctx corectx.Context, orgId model.Id) (string, *ft.ClientErrors, error) {
	definitionId, cErrs, err := sequenceDefinitionOf(ctx, this.reader, orgId,
		invsettings.OrgSettingScrapSequenceDefinitionId)
	if err != nil || cErrs != nil {
		return "", cErrs, err
	}
	return nextSequenceValue(ctx, this.sequences, sequenceRequest{
		definitionId: definitionId, orgId: orgId,
		operationKey: itExt.SequenceOperationKey(ctx, models.StockScrapFieldScrapNumber),
	})
}

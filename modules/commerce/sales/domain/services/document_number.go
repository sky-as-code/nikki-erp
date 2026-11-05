package services

import (
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	modconstants "github.com/sky-as-code/nikki-erp/modules/commerce/sales/constants"
	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/models"
	salessettings "github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/external"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// DocumentNumbering issues Sales' document numbers from the Essential sequence definitions the org
// lifecycle registered and recorded in the org settings.
type DocumentNumbering struct {
	Sequences itExt.SequenceExtService
	Settings  itExt.OrgSettingsReaderExtService
}

var documentNumbering *DocumentNumbering

// SetDocumentNumbering installs the numbering ports, from the module's OnAppStarted. A package hook
// for the reason the warehouse hold port is one: the create paths are free functions reached from
// several call chains, none of which carries the ports.
func SetDocumentNumbering(numbering DocumentNumbering) {
	documentNumbering = &numbering
}

// DocumentSequence names one numbered field and the setting holding its definition id.
type DocumentSequence struct {
	Setting string
	Field   string
}

var (
	SeqOrderNumber = DocumentSequence{
		Setting: salessettings.OrgSettingOrderNumberSequenceId, Field: models.SalesOrderFieldOrderNumber,
	}
	SeqOrderPublicReference = DocumentSequence{
		Setting: salessettings.OrgSettingOrderPublicReferenceSequenceId, Field: models.SalesOrderFieldPublicReference,
	}
	SeqBillNumber = DocumentSequence{
		Setting: salessettings.OrgSettingBillNumberSequenceId, Field: models.SalesBillFieldBillNumber,
	}
	SeqReturnNumber = DocumentSequence{
		Setting: salessettings.OrgSettingReturnNumberSequenceId, Field: models.SalesReturnFieldReturnNumber,
	}
	SeqPromotionReference = DocumentSequence{
		Setting: salessettings.OrgSettingPromotionReferenceSequenceId, Field: models.SalesPromotionFieldReference,
	}
)

// numberingRefusal carries a client refusal (the org's numbering not set up, or Essential refusing
// the issue) out of the write's transaction as an error, so that transaction rolls back: a
// reservation NextValue may already have written must not survive. The public entry points turn it
// back into client errors with numberingRefusalOf.
type numberingRefusal struct {
	cErrs ft.ClientErrors
}

func (this *numberingRefusal) Error() string {
	return this.cErrs.ToError().Error()
}

// numberingRefusalOf returns the client errors a numbering refusal carries, or nil when err is not one.
func numberingRefusalOf(err error) *ft.ClientErrors {
	var refusal *numberingRefusal
	if errors.As(err, &refusal) {
		return &refusal.cErrs
	}
	return nil
}

// issueDocumentNumbers returns one formatted number per sequence, in order. ctx must carry the
// write's transaction: NextValue joins it, so a rolled-back document gives its numbers back. A
// refusal is returned as a *numberingRefusal error, which rolls that transaction back. discriminators name the source document when one
// request numbers several documents of a series (see itExt.SequenceOperationKey); never the new
// record's freshly minted id, which would defeat the replay of a retried request.
func issueDocumentNumbers(
	ctx corectx.Context, orgId string, discriminators []string, sequences ...DocumentSequence,
) ([]string, error) {
	if documentNumbering == nil {
		return nil, errors.New("issueDocumentNumbers: document numbering is not installed")
	}
	ids, err := SequenceIdsOf(ctx, documentNumbering.Settings, model.Id(orgId))
	if err != nil {
		return nil, err
	}
	numbers := make([]string, 0, len(sequences))
	for _, sequence := range sequences {
		definitionId := ids[sequence.Setting]
		if definitionId == "" {
			return nil, &numberingRefusal{cErrs: ft.ClientErrors{
				*itExt.SequenceNotConfiguredError(modconstants.SalesModuleName, sequence.Setting),
			}}
		}
		operationKey := itExt.SequenceOperationKey(ctx, sequence.Field, discriminators...)
		number, err := nextNumber(ctx, definitionId, orgId, operationKey)
		if err != nil {
			return nil, errors.Wrapf(err, "issue %s", sequence.Field)
		}
		numbers = append(numbers, number)
	}
	return numbers, nil
}

func nextNumber(ctx corectx.Context, definitionId model.Id, orgId string, operationKey string) (string, error) {
	org := model.Id(orgId)
	result, err := documentNumbering.Sequences.NextValue(ctx, itExt.NextSequenceValueCommand{
		DefinitionId: definitionId,
		OrgId:        &org,
		CallerModule: modconstants.SalesModuleName,
		OperationKey: &operationKey,
	})
	if err != nil {
		return "", err
	}
	if result.ClientErrors.Count() > 0 {
		return "", &numberingRefusal{cErrs: result.ClientErrors}
	}
	if !result.HasData || result.Data.FormattedValue == "" {
		return "", errors.Errorf("sequence %s issued no value", definitionId)
	}
	return result.Data.FormattedValue, nil
}

// SequenceIdsOf reads the sequence definition ids one organization's settings hold, keyed by
// setting name. An unset id is absent. A failed read is an error.
func SequenceIdsOf(
	ctx corectx.Context, reader itExt.OrgSettingsReaderExtService, orgId model.Id,
) (map[string]model.Id, error) {
	if reader == nil {
		return nil, errors.New("SequenceIdsOf: no settings reader")
	}
	values, err := reader.GetOrgSettingValues(ctx, orgId, modconstants.SettingsModuleKey)
	if err != nil {
		return nil, err
	}
	ids := map[string]model.Id{}
	for _, name := range salessettings.SequenceIdSettings() {
		if id, ok := values[name].(string); ok && id != "" {
			ids[name] = model.Id(id)
		}
	}
	return ids, nil
}

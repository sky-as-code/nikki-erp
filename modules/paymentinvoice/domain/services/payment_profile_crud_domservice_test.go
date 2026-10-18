package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
)

// recordingDomainService stands in for the default CRUD: it records what reached it and answers
// with the record it was given, so the profile overrides can be checked either side of it.
type recordingDomainService struct {
	composable.CrudDomainService
	created dmodel.DynamicFields
	search  composable.SearchQuery
}

func (this *recordingDomainService) Create(
	_ corectx.Context, cmd composable.CreateCommand, _ ...composable.CreateOptions,
) (*composable.CreateResult, error) {
	this.created = dmodel.DynamicFields{}
	for key, value := range cmd {
		this.created[key] = value
	}
	return &composable.CreateResult{HasData: true, Data: cmd}, nil
}

func (this *recordingDomainService) Search(
	_ corectx.Context, query composable.SearchQuery, _ ...corecrud.ServiceSearchOptions,
) (*composable.SearchResult, error) {
	this.search = query
	return &composable.SearchResult{HasData: true, Data: dyn.PagedResultData[dmodel.DynamicFields]{
		Items:         []dmodel.DynamicFields{},
		DesiredFields: []string{"name", models.PaymentProfileFieldEncryptedConfig},
	}}, nil
}

// The plain credentials must never reach the default CRUD, which persists what it is given; the
// caller still gets back what they sent.
func TestACreatedProfileIsEncryptedBeforeItIsPersisted(t *testing.T) {
	base := &recordingDomainService{}
	svc := NewPaymentProfileCrudDomainService(base, newTestProfileService(testEncryptionKey))

	result, err := svc.Create(nil, dmodel.DynamicFields{
		"name": "momo", models.PaymentProfileFieldConfig: map[string]any{"partnerCode": "MOMOABCD"},
	})

	require.NoError(t, err)
	assert.NotContains(t, base.created, models.PaymentProfileFieldConfig)
	assert.NotContains(t, base.created[models.PaymentProfileFieldEncryptedConfig], "MOMOABCD")
	assert.Equal(t, map[string]any{"partnerCode": "MOMOABCD"}, result.Data[models.PaymentProfileFieldConfig])
	assert.NotContains(t, result.Data, models.PaymentProfileFieldEncryptedConfig)
}

// A search selecting "config" reads the column that holds it, and echoes back the name asked for.
func TestASearchResolvesConfigAndEchoesItBack(t *testing.T) {
	base := &recordingDomainService{}
	svc := NewPaymentProfileCrudDomainService(base, newTestProfileService(testEncryptionKey))

	result, err := svc.Search(nil, dmodel.DynamicFields{
		"fields": []string{"name", models.PaymentProfileFieldConfig},
	})

	require.NoError(t, err)
	assert.Equal(t, []string{"name", models.PaymentProfileFieldEncryptedConfig}, base.search["fields"])
	assert.Equal(t, []string{"name", models.PaymentProfileFieldConfig}, result.Data.DesiredFields)
}

// A field list that arrived through a JSON body is []any; the swap has to handle both shapes.
func TestASelectedConfigResolvesFromAnUntypedList(t *testing.T) {
	params := dmodel.DynamicFields{"fields": []any{"name", models.PaymentProfileFieldConfig}}

	swapConfigForEncryptedConfig(params)

	assert.Equal(t, []any{"name", models.PaymentProfileFieldEncryptedConfig}, params["fields"])
}

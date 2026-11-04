package app

import (
	"fmt"
	"strconv"
	"strings"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/services"
	itSeq "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/sequence"
)

// seqMaxPageSize caps a counter or issuance page (BR §8.3).
const seqMaxPageSize = 200

// seqSecretField holds algorithm secrets on definitions and counters. It never leaves the server and
// can be neither written, filtered nor sorted by a client.
const seqSecretField = models.SeqDefFieldAlgorithmParams

// sequenceOutput reshapes records on their way out: the secrets go, the stored algorithm takes its
// API name, and the listed int64 fields become decimal strings so JavaScript keeps every digit.
type sequenceOutput struct {
	stringified []string
}

func (this sequenceOutput) record(fields dmodel.DynamicFields) {
	if fields == nil {
		return
	}
	stripSecrets(fields)
	if algorithm, ok := fields[models.SeqDefFieldAlgorithm].(string); ok {
		fields[models.SeqDefFieldAlgorithm] = ApiAlgorithmOf(algorithm)
	}
	for _, key := range this.stringified {
		if number, ok := models.FieldInt64(fields, key); ok {
			fields[key] = strconv.FormatInt(number, 10)
		}
	}
}

// stripSecrets removes the secret field from a record and from every edge record nested in it.
func stripSecrets(value any) {
	switch typed := value.(type) {
	case dmodel.DynamicFields:
		stripSecretsOf(typed)
	case map[string]any:
		stripSecretsOf(typed)
	case []dmodel.DynamicFields:
		for _, item := range typed {
			stripSecretsOf(item)
		}
	case []map[string]any:
		for _, item := range typed {
			stripSecretsOf(item)
		}
	case []any:
		for _, item := range typed {
			stripSecrets(item)
		}
	}
}

func stripSecretsOf(fields map[string]any) {
	delete(fields, seqSecretField)
	for _, nested := range fields {
		stripSecrets(nested)
	}
}

func (this sequenceOutput) one(result *dynamicresource.GetOneResult, err error) (*dynamicresource.GetOneResult, error) {
	if err == nil && result != nil && result.HasData {
		this.record(result.Data.Item)
	}
	return result, err
}

func (this sequenceOutput) paged(result *dynamicresource.SearchResult, err error) (*dynamicresource.SearchResult, error) {
	if err == nil && result != nil && result.HasData {
		for _, item := range result.Data.Items {
			this.record(item)
		}
	}
	return result, err
}

// guardQuery refuses a filter or sort naming the secret field and caps the page size.
func guardQuery(query dmodel.DynamicFields) *ft.ClientErrors {
	if graph, ok := query[basemodel.FieldGraph]; ok && graph != nil && strings.Contains(fmt.Sprint(graph), seqSecretField) {
		cErrs := ft.ClientErrors{*ft.NewValidationError(basemodel.FieldGraph, services.ErrKeySeqReadOnlyField,
			"algorithm_params cannot be searched")}
		return &cErrs
	}
	if size, ok := models.FieldInt64(query, basemodel.FieldSize); ok && size > seqMaxPageSize {
		query[basemodel.FieldSize] = seqMaxPageSize
	}
	return nil
}

// refuseSecretWrite rejects a write that carries a server-owned field.
func refuseSecretWrite(cmd dmodel.DynamicFields, serverOwned ...string) *ft.ClientErrors {
	cErrs := ft.ClientErrors{}
	for _, key := range append([]string{seqSecretField}, serverOwned...) {
		if _, present := cmd[key]; present {
			cErrs.Append(*ft.NewValidationError(key, services.ErrKeySeqReadOnlyField,
				"{{field}} is set by the server", map[string]any{"field": key}))
		}
	}
	if cErrs.Count() == 0 {
		return nil
	}
	return &cErrs
}

// ApiAlgorithmOf names a stored algorithm the way the API does.
func ApiAlgorithmOf(stored string) string {
	if stored == models.SeqAlgorithmFf1Fpe {
		return itSeq.AlgorithmPseudorandom
	}
	return stored
}

// StoredAlgorithmOf maps the API's algorithm onto the stored enum. The stored spelling FF1_FPE is not
// part of the API, so it is passed on as an unknown value the schema rejects.
func StoredAlgorithmOf(api string) string {
	switch api {
	case itSeq.AlgorithmPseudorandom:
		return models.SeqAlgorithmFf1Fpe
	case models.SeqAlgorithmFf1Fpe:
		return "unsupported:" + api
	}
	return api
}

// mapAlgorithmInput rewrites an API algorithm in place.
func mapAlgorithmInput(cmd dmodel.DynamicFields) {
	if algorithm, ok := cmd[models.SeqDefFieldAlgorithm].(string); ok {
		cmd[models.SeqDefFieldAlgorithm] = StoredAlgorithmOf(algorithm)
	}
}

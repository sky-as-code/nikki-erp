package composable

import (
	"sort"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/resourceaccess"
)

// labelBatchSize stays under the search page cap (model.MODEL_RULE_PAGE_MAX_SIZE).
const labelBatchSize = 500

// LabelsAsCaller reads the record_label_field of the given records through the target's own
// application Search, as the caller: the same convention the Web BFF's edge_labels use, and the
// same permission, org confinement and field masking a list of that resource applies. Archived
// records are included, since a history names whatever the record pointed at back then.
func (this *engineResourceAccess) LabelsAsCaller(
	ctx corectx.Context, query resourceaccess.LabelsQuery,
) (map[string]string, error) {
	labels := map[string]string{}
	schema := this.schemaOf(query.ResourceCode)
	if schema == nil || len(query.Ids) == 0 {
		return labels, nil
	}
	labelField := schema.RecordLabelField()
	if _, ok := schema.Field(labelField); labelField == "" || !ok {
		return labels, nil
	}
	appSvc, err := this.resolve(query.ResourceCode)
	if err != nil {
		return labels, nil
	}
	ids := array.Map(query.Ids, func(id model.Id) any { return string(id) })
	for start := 0; start < len(ids); start += labelBatchSize {
		batch := ids[start:min(start+labelBatchSize, len(ids))]
		err := this.readLabels(ctx, appSvc, labelField, labelSearchQuery(labelField, batch, query.OrgId), labels)
		if err != nil {
			return nil, err
		}
	}
	return labels, nil
}

func labelSearchQuery(labelField string, ids []any, orgId *model.Id) SearchQuery {
	query := SearchQuery{
		fieldNameGraph:                 map[string]any{"if": append([]any{basemodel.FieldId, "in"}, ids...)},
		fieldNameFields:                []string{basemodel.FieldId, labelField},
		basemodel.FieldPage:            0,
		basemodel.FieldSize:            len(ids),
		basemodel.FieldIncludeArchived: true,
	}
	if orgId != nil {
		query[basemodel.FieldOrgId] = string(*orgId)
	}
	return query
}

// readLabels records each readable label; a refused search leaves the ids unlabelled rather than
// failing the history it serves.
func (this *engineResourceAccess) readLabels(
	ctx corectx.Context, appSvc CrudApplicationService, labelField string, query SearchQuery,
	labels map[string]string,
) error {
	found, err := appSvc.Search(ctx, query)
	if err != nil {
		return errors.Wrap(err, "LabelsAsCaller.Search")
	}
	if found == nil || found.ClientErrors.Count() > 0 {
		return nil
	}
	language := dyn.ResolveLocale(ctx)
	for _, item := range found.Data.Items {
		id := readString(item, basemodel.FieldId)
		if label := labelText(item[labelField], language); id != "" && label != "" {
			labels[id] = label
		}
	}
	return nil
}

// labelText renders a label value. A localized label arrives either resolved to the caller's
// language or as the whole LangJson map, in which case the caller's language wins, then en-US,
// then the first translation in key order.
func labelText(value any, language *model.LanguageCode) string {
	switch typed := value.(type) {
	case string:
		return typed
	case *string:
		if typed != nil {
			return *typed
		}
	case model.LangJson:
		texts := map[string]string{}
		for key, text := range typed {
			texts[string(key)] = text
		}
		return langJsonText(texts, language)
	case map[string]any:
		texts := map[string]string{}
		for key, text := range typed {
			if str, ok := text.(string); ok {
				texts[key] = str
			}
		}
		return langJsonText(texts, language)
	}
	return ""
}

func langJsonText(texts map[string]string, language *model.LanguageCode) string {
	if language != nil && texts[string(*language)] != "" {
		return texts[string(*language)]
	}
	if text := texts[string(model.LanguageCodeEnUs)]; text != "" {
		return text
	}
	keys := make([]string, 0, len(texts))
	for key := range texts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if texts[key] != "" {
			return texts[key]
		}
	}
	return ""
}

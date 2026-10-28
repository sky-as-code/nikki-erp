package app

import (
	"encoding/json"
	"sort"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	c "github.com/sky-as-code/nikki-erp/modules/auditlog/constants"
	"github.com/sky-as-code/nikki-erp/modules/auditlog/domain/models"
	itAction "github.com/sky-as-code/nikki-erp/modules/auditlog/interfaces/action"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/resourceaccess"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

const (
	paramResourceCode = "resource_code"
	paramResourceId   = "resource_id"
	actorUserName     = models.ActionEdgeActorUser + ".display_name"
	actorUserAvatar   = models.ActionEdgeActorUser + ".avatar_file_id"
)

type historyRequest struct {
	access resourceaccess.ReadAsCallerQuery
	page   int
	size   int
}

// GetHistory is authorized by the target record, not by read:auditlog_action: whoever may read
// the record may see how it came to be, minus the fields they may not read. The check is the
// target resource's own GetById, run as the caller.
func (this *ActionApplicationServiceImpl) GetHistory(
	ctx corectx.Context, query itAction.GetHistoryQuery,
) (*itAction.GetHistoryResult, error) {
	request, cErrs := parseHistoryRequest(query)
	if cErrs != nil {
		return &itAction.GetHistoryResult{ClientErrors: *cErrs}, nil
	}
	allowed, refused, err := this.access.ReadAsCaller(ctx, request.access)
	if err != nil {
		return nil, errors.Wrap(err, "GetHistory.ReadAsCaller")
	}
	if refused.Count() > 0 {
		return &itAction.GetHistoryResult{ClientErrors: refused}, nil
	}
	found, err := this.DomainService().Search(ctx, historySearchQuery(request))
	if err != nil {
		return nil, errors.Wrap(err, "GetHistory.Search")
	}
	if found.ClientErrors.Count() > 0 {
		return nil, errors.Wrap(found.ClientErrors.ToError(), "GetHistory.Search")
	}
	masked := map[string]bool{}
	for _, field := range allowed.MaskedFields {
		masked[field] = true
	}
	items := make([]itAction.HistoryItem, 0, len(found.Data.Items))
	for _, row := range found.Data.Items {
		items = append(items, toHistoryItem(row, masked))
	}
	if err := this.attachEdgeLabels(ctx, request.access, items); err != nil {
		return nil, errors.Wrap(err, "GetHistory.attachEdgeLabels")
	}
	return &itAction.GetHistoryResult{
		Data: itAction.HistoryResultData{
			Items: items, Total: found.Data.Total, Page: request.page, Size: request.size,
		},
		HasData: true,
	}, nil
}

func parseHistoryRequest(query itAction.GetHistoryQuery) (*historyRequest, *ft.ClientErrors) {
	cErrs := ft.ClientErrors{}
	code := composable.ParamString(query, paramResourceCode)
	id := composable.ParamString(query, paramResourceId)
	for field, value := range map[string]string{paramResourceCode: code, paramResourceId: id} {
		if value == "" {
			cErrs.Append(*ft.NewValidationError(field, ft.ErrorKey("err_missing_required_field"), "field is required"))
		}
	}
	page, pageOk := readInt(query, basemodel.FieldPage, 0)
	size, sizeOk := readInt(query, basemodel.FieldSize, c.DefaultHistorySize)
	if !pageOk || page < 0 {
		cErrs.Append(*numberRangeError(basemodel.FieldPage, 0, model.MODEL_RULE_PAGE_INDEX_END))
	}
	if !sizeOk || size < 1 {
		cErrs.Append(*numberRangeError(basemodel.FieldSize, 1, c.MaxHistorySize))
	}
	if cErrs.Count() > 0 {
		return nil, &cErrs
	}
	return &historyRequest{
		access: resourceaccess.ReadAsCallerQuery{
			ResourceCode: code, ResourceId: model.Id(id), OrgId: composable.ParamOrgId(query),
		},
		page: page,
		size: min(size, c.MaxHistorySize),
	}, nil
}

// numberRangeError is the shared common:err_invalid_number_range contract, so a client reads a
// bad page like any other out-of-range number.
func numberRangeError(field string, minValue, maxValue int) *ft.ClientErrorItem {
	return ft.NewValidationError(field, ft.ErrorKey("err_invalid_number_range"),
		"value must be between {{min}} and {{max}}", map[string]any{"min": minValue, "max": maxValue})
}

func readInt(query dmodel.DynamicFields, field string, fallback int) (int, bool) {
	raw, ok := query[field]
	if !ok || raw == nil || raw == "" {
		return fallback, true
	}
	value, ok := util.AnyToInt64(raw)
	return int(value), ok
}

// historySearchQuery is not confined to the caller's org: the target check above already decided
// the caller may read this record, and a global record's entries carry no org at all.
func historySearchQuery(request *historyRequest) composable.SearchQuery {
	graph := map[string]any{
		"and": []any{
			map[string]any{"if": []any{models.ActionFieldResourceCode, "=", request.access.ResourceCode}},
			map[string]any{"if": []any{models.ActionFieldResourceId, "=", string(request.access.ResourceId)}},
		},
		"order": []any{
			[]any{models.ActionFieldCreatedAt, "desc"},
			[]any{models.ActionFieldId, "desc"},
		},
	}
	return composable.SearchQuery{
		basemodel.FieldGraph:  graph,
		basemodel.FieldFields: []string{"*", actorUserName, actorUserAvatar},
		basemodel.FieldPage:   request.page,
		basemodel.FieldSize:   request.size,
	}
}

func toHistoryItem(row dmodel.DynamicFields, masked map[string]bool) itAction.HistoryItem {
	changes, maskedFields := toChanges(jsonMapOf(row[models.ActionFieldBefore]),
		jsonMapOf(row[models.ActionFieldAfter]), masked)
	return itAction.HistoryItem{
		Id:           stringOf(row[models.ActionFieldId]),
		Action:       stringOf(row[models.ActionFieldAction]),
		CreatedAt:    stringOf(row[models.ActionFieldCreatedAt]),
		RequestId:    optionalString(row[models.ActionFieldRequestId]),
		Actor:        toHistoryActor(row),
		Changes:      changes,
		MaskedFields: maskedFields,
	}
}

// toHistoryActor prefers the user's current name and avatar, and falls back to the name taken
// when the entry was written, for a user since removed or an actor that was never a user.
func toHistoryActor(row dmodel.DynamicFields) itAction.HistoryActor {
	actor := itAction.HistoryActor{
		Id:          optionalString(row[models.ActionFieldActorId]),
		Type:        stringOf(row[models.ActionFieldActorType]),
		DisplayName: optionalString(row[models.ActionFieldActorDisplayName]),
	}
	user := jsonMapOf(row[models.ActionEdgeActorUser])
	if name := optionalString(user["display_name"]); name != nil {
		actor.DisplayName = name
	}
	actor.AvatarFileId = optionalString(user["avatar_file_id"])
	return actor
}

// toChanges lists the changed fields in name order, nulling the values of the masked ones.
func toChanges(before, after map[string]any, masked map[string]bool) ([]itAction.HistoryChange, []string) {
	names := make([]string, 0, len(before)+len(after))
	for name := range after {
		names = append(names, name)
	}
	for name := range before {
		if _, inAfter := after[name]; !inAfter {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	changes := make([]itAction.HistoryChange, 0, len(names))
	maskedFields := []string{}
	for _, name := range names {
		if masked[name] {
			changes = append(changes, itAction.HistoryChange{Field: name})
			maskedFields = append(maskedFields, name)
			continue
		}
		changes = append(changes, itAction.HistoryChange{Field: name, Old: before[name], New: after[name]})
	}
	return changes, maskedFields
}

// jsonMapOf reads a jsonmap column or a to-one edge, whichever Go shape the driver produced.
func jsonMapOf(value any) map[string]any {
	switch typed := value.(type) {
	case nil:
		return map[string]any{}
	case map[string]any:
		return typed
	case dmodel.DynamicFields:
		return typed
	}
	out := map[string]any{}
	raw, err := json.Marshal(value)
	if err == nil {
		if text, isText := value.(string); isText {
			raw = []byte(text)
		}
		_ = json.Unmarshal(raw, &out)
	}
	return out
}

// stringOf renders an id, a datetime or a string the way it serializes.
func stringOf(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case *string:
		return util.ValueOrZeroOf(typed)
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return text
	}
	return string(raw)
}

func optionalString(value any) *string {
	text := stringOf(value)
	if text == "" {
		return nil
	}
	return &text
}

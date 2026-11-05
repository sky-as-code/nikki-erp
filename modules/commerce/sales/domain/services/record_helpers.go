package services

import (
	"strings"
	"time"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// searchBy reads every row of a schema whose field holds the given value, in one page of
// MODEL_RULE_PAGE_MAX. A parent with more children than that is a misconfiguration rather than a
// case to paginate for.
func searchBy(
	ctx corectx.Context, schemaName, field, value string,
) ([]dmodel.DynamicFields, error) {
	engineRepo, err := repoFor(schemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(*dmodel.NewSearchNode().NewCondition(field, dmodel.Equals, value))

	found, err := engineRepo.Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  model.MODEL_RULE_PAGE_MAX_SIZE,
	})
	if err != nil {
		return nil, err
	}
	if found == nil || !found.HasData {
		return nil, nil
	}
	return found.Data.Items, nil
}

// int32Of reads an integer field, treating an absent one as zero. Every numeric shape is accepted
// because a value that went through a jsonb column arrives as whatever the JSON decoder chose — a
// whole number comes back a float64.
func int32Of(record dmodel.DynamicFields, field string) int32 {
	value, present := record[field]
	if !present || value == nil {
		return 0
	}
	switch typed := value.(type) {
	case int32:
		return typed
	case int64:
		return int32(typed)
	case int:
		return int32(typed)
	case float64:
		return int32(typed)
	case float32:
		return int32(typed)
	case *int32:
		if typed != nil {
			return *typed
		}
	}
	return 0
}

// decimalOf reads a decimal field, treating an absent one as zero. A decimal crosses JSON as a string
// so it does not lose precision, and a malformed one reads as zero rather than becoming a NaN
// threshold.
func decimalOf(record dmodel.DynamicFields, field string) decimal.Decimal {
	value, present := record[field]
	if !present || value == nil {
		return decimal.Zero
	}
	switch typed := value.(type) {
	case decimal.Decimal:
		return typed
	case *decimal.Decimal:
		if typed != nil {
			return *typed
		}
	case string:
		if parsed, err := decimal.NewFromString(typed); err == nil {
			return parsed
		}
	case float64:
		return decimal.NewFromFloat(typed)
	}
	return decimal.Zero
}

func dateTimeOf(record dmodel.DynamicFields, field string) *model.ModelDateTime {
	value, present := record[field]
	if !present || value == nil {
		return nil
	}
	switch typed := value.(type) {
	case model.ModelDateTime:
		return &typed
	case *model.ModelDateTime:
		return typed
	}
	return nil
}

// isUniqueViolation is matched on the message because the repository returns a wrapped driver error
// rather than a typed one; the same three spellings jobscheduler matches on cover every driver this
// deployment uses.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate key") ||
		strings.Contains(message, "unique constraint") ||
		strings.Contains(message, "sqlstate 23505")
}

func OrderNotFoundErrors(orderId string) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation("id", ft.ErrorKey("sales_order.not_found", "sales"),
		"no sales order exists with id '"+orderId+"'"))
	return vErrs
}

// NowUnix reads the clock once per request at the edge; every rule beneath transport takes the
// instant as a parameter, so two gates in one request cannot disagree about the hour.
func NowUnix() int64 {
	return time.Now().Unix()
}

package services

import (
	"time"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// Order stage transitions, in one place.
//
// A stage is the order's commercial lifecycle - draft, confirmed, processing, completed, cancelled -
// and nothing else. payment_status, fulfillment_status and invoice_status move on their own clocks
// and are NOT stages; neither is is_archived, which is a filing decision rather than a step in the
// sale. Announcing those here would tell a consumer the sale moved on when only the money did.
//
// Everything that changes a stage goes through this file, so that "every transition announces
// itself" is true by construction rather than by five call sites each remembering. The event is
// written into the outbox in the CALLER'S transaction: a stage change that committed without its
// event, or an event announcing a change that rolled back, are the two failures the outbox exists
// to prevent, and only the caller knows the transaction they share.

// OrderStageChangedParams is one stage transition, as it will be announced.
type OrderStageChangedParams struct {
	// Order is the record as it was BEFORE the transition: the payload's identifying fields are read
	// from it, and reading them from the updated row would need a re-read inside the transaction for
	// values that cannot have changed.
	Order dmodel.DynamicFields

	// PreviousStage is empty for an order entering its first stage, which is how a consumer tells a
	// creation from a move.
	PreviousStage string
	CurrentStage  string

	StageVersion int32

	// InitialBillId is carried on the transition into confirmed, where a bill was just raised, and
	// empty everywhere else.
	InitialBillId string

	// Reason explains a transition somebody chose rather than derived - why a sale was called off.
	// Empty for the transitions that need no explanation.
	Reason string

	// OccurredAt is when the transition happened, not when it is published; zero means now.
	OccurredAt time.Time
}

// StageEventExtras are the facts only one caller knows, added to the announcement.
type StageEventExtras struct {
	// Reason explains a transition somebody chose rather than derived - why a sale was called off.
	Reason string

	// InitialBillId names the bill a confirmation just raised.
	InitialBillId string

	// OccurredAt is when the transition happened; zero means now.
	OccurredAt time.Time
}

// TransitionOrderStage moves an order to a new stage and announces it, as one unit.
//
// Must be called inside the caller's transaction. It validates against the transition table, writes
// the status, bumps the stage version and records the integration event; extra is merged into the
// same update so a caller's own stamp - confirmed_at, cancelled_at - cannot land in a separate write
// that could fail on its own.
func TransitionOrderStage(
	ctx corectx.Context,
	order dmodel.DynamicFields,
	toStage string,
	extra dmodel.DynamicFields,
	event StageEventExtras,
) error {
	orderId := stringOf(order, models.SalesOrderFieldId)
	fromStage := stringOf(order, models.SalesOrderFieldStatus)

	if !CanTransitionOrderStatus(fromStage, toStage) {
		// An error rather than a refusal: every caller has already applied its own business gate, so
		// reaching here means the code asked for something the lifecycle forbids.
		return errors.New("sales order '" + orderId + "' cannot move from '" +
			fromStage + "' to '" + toStage + "'")
	}

	engineRepo, err := repoFor(models.SalesOrderSchemaName)
	if err != nil {
		return err
	}

	version := nextStageVersion(order)
	update := dmodel.DynamicFields{
		models.SalesOrderFieldId:           orderId,
		models.SalesOrderFieldStatus:       toStage,
		models.SalesOrderFieldStageVersion: version,
	}
	for field, value := range extra {
		update[field] = value
	}
	if _, err := engineRepo.Update(ctx, update); err != nil {
		return err
	}

	return RecordOrderStageChanged(ctx, OrderStageChangedParams{
		Order:         order,
		PreviousStage: fromStage,
		CurrentStage:  toStage,
		StageVersion:  version,
		InitialBillId: event.InitialBillId,
		Reason:        event.Reason,
		OccurredAt:    event.OccurredAt,
	})
}

// RecordOrderStageChanged writes the stage event into the outbox, in the caller's transaction.
//
// Separate from TransitionOrderStage because two callers write the stage themselves: creation,
// which inserts the order rather than updating it, and confirmation, which must write the bill link
// in the same update and name the bill in the event.
func RecordOrderStageChanged(ctx corectx.Context, params OrderStageChangedParams) error {
	order := params.Order

	_, err := RecordEvent(ctx, RecordEventParams{
		EventType:   models.EventSalesOrderStageChanged,
		AggregateId: stringOf(order, models.SalesOrderFieldId),
		OrgId:       stringOf(order, basemodel.FieldOrgId),
		OccurredAt:  stageEventTimeOf(params).Unix(),
		Payload:     stageEventPayloadFor(params),
	})
	return err
}

// stageEventPayloadFor builds what a consumer receives. Separate from the write so the contract can
// be pinned by a test without a database: the payload IS the public contract, and the row it goes
// into is Sales' own storage.
func stageEventPayloadFor(params OrderStageChangedParams) map[string]any {
	order := params.Order

	payload := map[string]any{
		"sales_order_id": stringOf(order, models.SalesOrderFieldId),
		"order_number":   stringOf(order, models.SalesOrderFieldOrderNumber),
		"currency_code":  stringOf(order, models.SalesOrderFieldCurrencyCode),

		// null rather than "" when the order is entering its first stage: a consumer distinguishing a
		// creation from a move should not have to know that an empty string means one of them.
		"previous_stage": stageOrNil(params.PreviousStage),
		"current_stage":  params.CurrentStage,
		"stage_version":  params.StageVersion,

		// The totals travel WITH the event so a consumer never reads back into Sales, and acts on
		// what was true at the transition.
		"grand_total": decimalOf(order, models.SalesOrderFieldGrandTotal),
		"tax_total":   decimalOf(order, models.SalesOrderFieldTaxTotal),

		"occurred_at": stageEventTimeOf(params).Unix(),
	}

	// Absent rather than empty: a consumer checking for the key would otherwise find one and read ""
	// as an identifier.
	if params.InitialBillId != "" {
		payload["initial_bill_id"] = params.InitialBillId
	}
	if params.Reason != "" {
		payload["reason"] = params.Reason
	}
	return payload
}

func stageEventTimeOf(params OrderStageChangedParams) time.Time {
	if params.OccurredAt.IsZero() {
		return time.Now().UTC()
	}
	return params.OccurredAt
}

// nextStageVersion counts the stage the order is about to enter. An order written before the column
// existed reads zero and so enters its next stage as 1, which is the same answer a new order gets:
// the version orders the events of ONE order against each other, and nothing reads it as a count of
// history that never happened.
func nextStageVersion(order dmodel.DynamicFields) int32 {
	return int32Of(order, models.SalesOrderFieldStageVersion) + 1
}

// stageOrNil keeps the JSON honest: the absence of a previous stage is null, not "".
func stageOrNil(stage string) any {
	if stage == "" {
		return nil
	}
	return stage
}

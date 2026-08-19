package services

import (
	"github.com/shopspring/decimal"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

// The per-organization purchase policy (BR §47), and what happens when an organization has not set
// one.

// PurchaseConfiguration is the policy an order is confirmed under.
type PurchaseConfiguration struct {
	ApprovalMode models.ApprovalMode

	// ApprovalThreshold is nil when no threshold is set, which is NOT the same as zero. Zero means
	// "approve everything"; nil means the organization chose two-step approval without saying from
	// what value, and the sensible reading of that is also "approve everything" — but the two
	// arrive here differently and a caller may want to tell them apart.
	ApprovalThreshold *decimal.Decimal

	PoModificationPolicy models.PoModificationPolicy
}

// DefaultPurchaseConfiguration is the policy of an organization that has not configured one.
//
// One-step and allow_edit are the permissive defaults, matching the schema's own default_value on
// both fields. Defaulting to two-step approval instead would mean that installing this module
// silently blocked every purchase in an organization that had never asked for approvals — a module
// must not start refusing work it was not configured to refuse.
func DefaultPurchaseConfiguration() PurchaseConfiguration {
	return PurchaseConfiguration{
		ApprovalMode:         models.ApprovalModeOneStep,
		ApprovalThreshold:    nil,
		PoModificationPolicy: models.PoModificationPolicyAllowEdit,
	}
}

// LoadConfiguration reads one organization's purchase policy, falling back to the default.
//
// An org with no row gets the default rather than an error: configuration is optional, and a
// module that refused to confirm an order until somebody visited a settings page would be broken
// out of the box.
func LoadConfiguration(ctx corectx.Context, orgId string) (PurchaseConfiguration, error) {
	config := DefaultPurchaseConfiguration()
	if orgId == "" {
		return config, nil
	}

	engine, err := engineFor(models.ConfigurationSchemaName)
	if err != nil {
		return config, err
	}
	rows, err := models.FindConfigurationForOrg(ctx, engine.ResourceRepository(), orgId)
	if err != nil {
		return config, err
	}
	if len(rows) == 0 {
		return config, nil
	}

	row := rows[0]
	if mode := stringOf(row, models.ConfigurationFieldApprovalMode); mode != "" {
		config.ApprovalMode = models.ApprovalMode(mode)
	}
	if policy := stringOf(row, models.ConfigurationFieldPoModificationPolicy); policy != "" {
		config.PoModificationPolicy = models.PoModificationPolicy(policy)
	}
	if raw, ok := row[models.ConfigurationFieldApprovalThreshold]; ok && raw != nil {
		threshold := decimalOf(row, models.ConfigurationFieldApprovalThreshold)
		config.ApprovalThreshold = &threshold
	}
	return config, nil
}

// RequiresApproval decides whether confirming an order of this total needs an approver (BR §47.1-2).
//
// Under one-step it never does, whatever the threshold says: a threshold left over from a previous
// policy must not keep gating orders after the organization has switched approvals off.
//
// Under two-step with no threshold, everything needs approval. That is the reading that matches
// what the setting is for — an organization that turned on two-step approval and named no value
// asked for approvals, not for none.
//
// The comparison is "at or above", so a threshold of 1000 catches an order of exactly 1000. A
// threshold is the value at which control begins, and the alternative makes the single most likely
// test case — an order for exactly the limit — fall the wrong side of it.
func RequiresApproval(config PurchaseConfiguration, total decimal.Decimal) bool {
	if config.ApprovalMode != models.ApprovalModeTwoStep {
		return false
	}
	if config.ApprovalThreshold == nil {
		return true
	}
	return total.GreaterThanOrEqual(*config.ApprovalThreshold)
}

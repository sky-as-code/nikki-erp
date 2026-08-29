package services

import (
	"github.com/shopspring/decimal"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

// The per-organization purchase policy, and what happens when an organization has not set one.

// PurchaseConfiguration is the policy an order is confirmed under.
type PurchaseConfiguration struct {
	ApprovalMode models.ApprovalMode

	// Nil when no threshold is set, which is not the same as zero. Both end up approving
	// everything, but they arrive differently and a caller may want to tell them apart.
	ApprovalThreshold *decimal.Decimal

	PoModificationPolicy models.PoModificationPolicy
}

// DefaultPurchaseConfiguration is the policy of an organization that has not configured one. The
// permissive values match the schema's own default_value on both fields; defaulting to two-step
// approval would silently block every purchase in an org that never asked for approvals.
func DefaultPurchaseConfiguration() PurchaseConfiguration {
	return PurchaseConfiguration{
		ApprovalMode:         models.ApprovalModeOneStep,
		ApprovalThreshold:    nil,
		PoModificationPolicy: models.PoModificationPolicyAllowEdit,
	}
}

// LoadConfiguration reads one organization's purchase policy. An org with no row gets the default
// rather than an error, since configuration is optional.
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

// RequiresApproval decides whether confirming an order of this total needs an approver. Under
// one-step it never does, whatever the threshold says, so a leftover threshold cannot keep gating
// orders after approvals are switched off. Under two-step with no threshold everything needs
// approval. The comparison is at-or-above, so a threshold of 1000 catches an order of exactly 1000.
func RequiresApproval(config PurchaseConfiguration, total decimal.Decimal) bool {
	if config.ApprovalMode != models.ApprovalModeTwoStep {
		return false
	}
	if config.ApprovalThreshold == nil {
		return true
	}
	return total.GreaterThanOrEqual(*config.ApprovalThreshold)
}

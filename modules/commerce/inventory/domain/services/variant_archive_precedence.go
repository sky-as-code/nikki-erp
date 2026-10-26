package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// A variant has one archive_source, standing in for a set of archive reasons. Precedence decides
// which reason the column keeps: user beats system_sync, which beats template_cascade. A weaker
// reason never overwrites a stronger one, and each restore path lifts only its own reason:
// template unarchive lifts template_cascade, reconcile lifts system_sync, and only a user lifts
// user.
//
// A user archive that takes a template's last selectable variant, and so archives the template, is
// stamped template_cascade rather than user (see archiveStamp): it is the cascade's trigger, and the
// template's unarchive must bring it back or the template returns with nothing to sell.

// ArchiveSourceStrength ranks a reason; 0 means "not archived".
func ArchiveSourceStrength(source models.ArchiveSource) int {
	switch source {
	case models.ArchiveSourceUser:
		return 3
	case models.ArchiveSourceSystemSync:
		return 2
	case models.ArchiveSourceTemplateCascade:
		return 1
	default:
		return 0
	}
}

// EffectiveArchiveSource is the reason an archived variant is held by, or "" when it is active.
// An archived row with no source predates the column and counts as the user's: resurrecting a
// deliberately withdrawn product is worse than leaving one a user can restore by hand.
func EffectiveArchiveSource(variant *models.ProductVariant) models.ArchiveSource {
	if archived := variant.IsArchived(); archived == nil || !*archived {
		return ""
	}
	if source := variant.GetArchiveSource(); source != nil && ArchiveSourceStrength(*source) > 0 {
		return *source
	}
	return models.ArchiveSourceUser
}

// StrongerArchiveSource returns whichever of the two reasons outranks the other.
func StrongerArchiveSource(current models.ArchiveSource, proposed models.ArchiveSource) models.ArchiveSource {
	if ArchiveSourceStrength(proposed) > ArchiveSourceStrength(current) {
		return proposed
	}
	return current
}

// RaiseArchiveFields is the delta that archives a variant for reason, or raises the reason it is
// already archived for when reason is stronger. ok is false when nothing needs writing: the
// variant is already archived for an equal or stronger reason. The reconcile executor calls it
// with system_sync for a variant its configuration no longer allows.
func RaiseArchiveFields(variant *models.ProductVariant, reason models.ArchiveSource) (dmodel.DynamicFields, bool) {
	current := EffectiveArchiveSource(variant)
	next := StrongerArchiveSource(current, reason)
	if current != "" && next == current {
		return nil, false
	}
	return dmodel.DynamicFields{
		models.ProductVariantFieldId:            derefString(variant.GetId()),
		basemodel.FieldIsArchived:               true,
		models.ProductVariantFieldArchiveSource: next.String(),
	}, true
}

// LiftArchiveFields is the delta that restores a variant archived for exactly reason. ok is false
// when the variant is active or held by any other reason, which this restore path must not lift.
func LiftArchiveFields(variant *models.ProductVariant, reason models.ArchiveSource) (dmodel.DynamicFields, bool) {
	if EffectiveArchiveSource(variant) != reason {
		return nil, false
	}
	return dmodel.DynamicFields{
		models.ProductVariantFieldId:            derefString(variant.GetId()),
		basemodel.FieldIsArchived:               false,
		models.ProductVariantFieldArchiveSource: nil,
	}, true
}

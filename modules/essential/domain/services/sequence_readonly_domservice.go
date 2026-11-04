package services

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	modconstants "github.com/sky-as-code/nikki-erp/modules/essential/constants"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/external"
	itSeq "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/sequence"
)

// Counters and issuances are written only by NextValue's statements; their onions serve reads.
func NewSequenceCounterDomainService(base dynamicresource.CrudDomainService) itSeq.SequenceCounterDomainService {
	return &SequenceCounterDomainServiceImpl{CrudDomainService: base}
}

type SequenceCounterDomainServiceImpl struct {
	dynamicresource.CrudDomainService
}

func NewSequenceIssuanceDomainService(base dynamicresource.CrudDomainService) itSeq.SequenceIssuanceDomainService {
	return &SequenceIssuanceDomainServiceImpl{CrudDomainService: base}
}

type SequenceIssuanceDomainServiceImpl struct {
	dynamicresource.CrudDomainService
}

// NewSequenceTimezoneDefaulter reads the acting user's effective system_timezone (their org's, else
// the tenant's). It answers empty without an acting user, as in an org's initialization by a job,
// and the definition then falls back to UTC. The zone is copied once; a later settings change does
// not move an existing definition (BR §4.2).
func NewSequenceTimezoneDefaulter(settingsSvc itExt.EffectiveSettingsExtService) SequenceTimezoneDefaulter {
	key := modconstants.EssentialModuleName + "." + settings.OrgSettingSystemTimezone
	return func(ctx corectx.Context) string {
		if settingsSvc == nil || ctx.GetPermissions().UserId == "" {
			return ""
		}
		result, err := settingsSvc.GetEffectiveSettings(ctx, itExt.GetEffectiveSettingsQuery{})
		if err != nil || result == nil || !result.HasData {
			return ""
		}
		zone, _ := result.Data.Values[key].(string)
		return zone
	}
}

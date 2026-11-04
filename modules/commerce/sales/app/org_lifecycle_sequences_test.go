package app

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/model"
	salessettings "github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/external"
	coreSettings "github.com/sky-as-code/nikki-erp/modules/core/settings"
)

func TestEverySequenceSettingGetsARegisteredSeries(t *testing.T) {
	settings := map[string]bool{}
	for _, sequence := range salesSequences {
		settings[sequence.setting] = true
	}
	for _, name := range salessettings.SequenceIdSettings() {
		require.Truef(t, settings[name], "no series is registered for %s", name)
	}
}

func TestSequenceRegistrationShapes(t *testing.T) {
	for _, sequence := range salesSequences {
		cmd := sequence.upsertDefinitionCommand(model.Id("ORG1"))
		require.Equal(t, "org", cmd.ConfigurationScope)
		require.Equal(t, model.Id("ORG1"), *cmd.OrgId)
		require.Equal(t, "never", *cmd.ResetPeriod)
		require.Equal(t, int64(1), *cmd.InitialValue)
		if sequence.code == "sales.order.public_reference" {
			require.Equal(t, itExt.SequenceAlgorithmPseudorandom, *cmd.Algorithm)
			require.Equal(t, int32(0), *cmd.Padding)
			require.Equal(t, "CORE", *cmd.PrefixTemplate)
			continue
		}
		require.Equal(t, itExt.SequenceAlgorithmIncrement, *cmd.Algorithm)
		require.Equal(t, int32(5), *cmd.Padding)
	}
}

func TestClientsCannotRepointASequence(t *testing.T) {
	vErrs, err := SequenceSettingsRule(nil, "org", []coreSettings.SetSettingItem{
		{Name: salessettings.OrgSettingOrderNumberSequenceId, Value: "X"},
		{Name: salessettings.OrgSettingReturnWindowDays, Value: 10},
	})
	require.NoError(t, err)
	require.Equal(t, 1, vErrs.Count())
}

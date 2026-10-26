package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/external"
)

type fakeUserSettings struct {
	seeded []model.Id
	refuse bool
}

func (this *fakeUserSettings) InitUserPreferences(
	_ corectx.Context, cmd itExt.InitOwnerSettingsCommand,
) (*itExt.InitOwnerSettingsResult, error) {
	this.seeded = append(this.seeded, cmd.OwnerId)
	result := &itExt.InitOwnerSettingsResult{}
	if this.refuse {
		result.ClientErrors = ft.ClientErrors{*ft.NewValidationError("owner_id", "err_test", "refused")}
	}
	return result, nil
}

func TestUserCreateSeedsPreferencesAndRebuildsPermissions(t *testing.T) {
	base := newFakeCrudDomain()
	settings := &fakeUserSettings{}
	perms := &fakePermRepo{}
	svc := NewUserCrudDomainService(base, settings, perms, nil, nil)

	result, err := svc.Create(testContext(), dmodel.DynamicFields{
		domain.UserFieldDisplayName: "Jane", domain.UserFieldIsOwner: true,
	})

	require.NoError(t, err)
	require.True(t, result.HasData)
	_, hasOwner := base.created[domain.UserFieldIsOwner]
	assert.False(t, hasOwner, "a created user must never carry is_owner")
	assert.Len(t, settings.seeded, 1)
	assert.Equal(t, settings.seeded, perms.rebuiltUsers)
	assert.True(t, base.repo.tranx.committed)
}

func TestUserCreateRejectedSeedsNothing(t *testing.T) {
	base := newFakeCrudDomain()
	base.rejectNext = true
	settings := &fakeUserSettings{}
	perms := &fakePermRepo{}
	svc := NewUserCrudDomainService(base, settings, perms, nil, nil)

	result, err := svc.Create(testContext(), dmodel.DynamicFields{})

	require.NoError(t, err)
	assert.Positive(t, result.ClientErrors.Count())
	assert.Empty(t, settings.seeded)
	assert.Empty(t, perms.rebuiltUsers)
}

func TestUserCreateRefusedSeedingFailsTheCreate(t *testing.T) {
	base := newFakeCrudDomain()
	svc := NewUserCrudDomainService(base, &fakeUserSettings{refuse: true}, &fakePermRepo{}, nil, nil)

	_, err := svc.Create(testContext(), dmodel.DynamicFields{domain.UserFieldDisplayName: "Jane"})

	require.Error(t, err)
	assert.False(t, base.repo.tranx.committed)
}

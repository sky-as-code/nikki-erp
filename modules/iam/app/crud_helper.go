package app

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// Permission action codes of the IAM custom actions, beside the built-in dynamicresource.Permission*.
const (
	actionManageUsers           = "manage_users"
	actionManageRoles           = "manage_roles"
	actionManageEntitlements    = "manage_entitlements"
	actionManageRoleAssignments = "manage_role_assignments"
)

// mutateRefused turns an authorization outcome into the mutation result a custom action answers.
func mutateRefused(cErrs *ft.ClientErrors, err error) (*dynamicresource.MutateResult, error) {
	if err != nil {
		return nil, err
	}
	return &dynamicresource.MutateResult{ClientErrors: *cErrs}, nil
}

// searchRefused is mutateRefused for a read.
func searchRefused(cErrs *ft.ClientErrors, err error) (*dynamicresource.SearchResult, error) {
	if err != nil {
		return nil, err
	}
	return &dynamicresource.SearchResult{ClientErrors: *cErrs}, nil
}

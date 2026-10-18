package app

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// Permission action codes of the IAM custom actions, beside the built-in composable.Permission*.
const (
	actionManageUsers           = "manage_users"
	actionManageRoles           = "manage_roles"
	actionManageEntitlements    = "manage_entitlements"
	actionManageRoleAssignments = "manage_role_assignments"
)

// mutateRefused turns an authorization outcome into the mutation result a custom action answers.
func mutateRefused(cErrs *ft.ClientErrors, err error) (*composable.MutateResult, error) {
	if err != nil {
		return nil, err
	}
	return &composable.MutateResult{ClientErrors: *cErrs}, nil
}

// searchRefused is mutateRefused for a read.
func searchRefused(cErrs *ft.ClientErrors, err error) (*composable.SearchResult, error) {
	if err != nil {
		return nil, err
	}
	return &composable.SearchResult{ClientErrors: *cErrs}, nil
}

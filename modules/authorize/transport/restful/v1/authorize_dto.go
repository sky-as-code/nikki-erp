package v1

import (
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces"
)

type IsAuthorizedRequest = it.IsAuthorizedQuery
type IsAuthorizedResponse it.IsAuthorizedResult
type PermissionSnapshotRequest = it.PermissionSnapshotQuery
type PermissionSnapshotResponse it.PermissionSnapshotResult

func (this *IsAuthorizedResponse) FromResult(result it.IsAuthorizedResult) {
	this.Decision = result.Decision
	this.ClientError = result.ClientError
}

func (this *PermissionSnapshotResponse) FromResult(result it.PermissionSnapshotResult) {
	this.Permissions = result.Permissions
	this.ClientError = result.ClientError
}

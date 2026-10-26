// Package storage is the core port through which any module tells the storage module which files
// its records use. It holds interfaces and DTOs only: the storage module implements the port and
// apptrait registers that implementation, so a consumer never depends on the storage module itself.
package storage

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// MetadataFileField is the field metadata key that marks a `<name>_file_id` field as a reference
// to a stored file. The dynamic resource engine reads it to drive the file lifecycle below, and
// the frontend reads it from meta/schema to know which fields a form submits as files.
const MetadataFileField = "file_field"

// FileFieldSuffix is the naming convention of a file field.
const FileFieldSuffix = "_file_id"

const (
	FileStatusPending = "pending"
	FileStatusInUse   = "in_use"
)

// StorageConsumer is how a module reports the files its records reference. A file is never shared
// between records: a record that needs another record's file duplicates it.
//
// Calls run in the caller's transaction when there is one, so the status change commits or rolls
// back together with the record that references the file. The port asserts no permission: the
// caller's own application service has already authorized the write of that record.
type StorageConsumer interface {
	// AssertFilesUsable refuses, per field, any id that does not exist, is outside orgId's space
	// (the tenant's when OrgId is nil), is not pending, or was released by a record before: a
	// released file is on its way to expiry and never comes back into use.
	AssertFilesUsable(ctx corectx.Context, query AssertFilesUsableQuery) (ft.ClientErrors, error)
	// MarkFilesInUse moves pending files to in_use and tags their objects in_use before returning,
	// so an error must abort the caller's save. Every id must move: one that is no longer pending
	// (taken by a concurrent write, or released) is an error, which is what keeps a file from
	// being shared between two records.
	MarkFilesInUse(ctx corectx.Context, cmd MarkFilesInUseCommand) error
	// ReleaseFiles moves files a record no longer references back to pending, so their objects
	// expire. Unknown ids are ignored. The objects are re-tagged pending only after the rows
	// commit, best-effort; inside the caller's transaction that happens once it is seen committed.
	ReleaseFiles(ctx corectx.Context, cmd ReleaseFilesCommand) error
	// DuplicateFile clones a file into a new pending file in OrgId's space and returns its id. The
	// copy is reserved against the quota like an upload. Client errors (source missing, quota
	// exceeded) come back as a composable ClientErrorSignal.
	DuplicateFile(ctx corectx.Context, cmd DuplicateFileCommand) (*model.Id, error)
}

type AssertFilesUsableQuery struct {
	// Fields maps each file id to the record field that holds it, so a refusal names the field.
	Fields map[model.Id]string
	OrgId  *model.Id
}

type MarkFilesInUseCommand struct {
	FileIds []model.Id
	OrgId   *model.Id
}

type ReleaseFilesCommand struct {
	FileIds []model.Id
}

type DuplicateFileCommand struct {
	FileId model.Id
	OrgId  *model.Id
}

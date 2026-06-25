package models

// Các kiểu enum dùng string cố định (không iota), giá trị khớp bản drive/enum + cột string trên DB.

type ScopeType string

const (
	ScopeTypeNone     ScopeType = ""
	ScopeTypeDomain   ScopeType = "domain"
	ScopeTypeOrg      ScopeType = "org"
	ScopeTypeHierachy ScopeType = "hierachy"
	ScopeTypePrivate  ScopeType = "private"

	ScopeTypeDefault = ScopeTypePrivate
)

type DriveFileStorage string

const (
	DriveFileStorageNone DriveFileStorage = ""
	DriveFileStorageS3   DriveFileStorage = "s3"

	DriveFileStorageDefault = DriveFileStorageS3
)

type DriveFileVisibility string

const (
	DriveFileVisibilityNone   DriveFileVisibility = ""
	DriveFileVisibilityPublic DriveFileVisibility = "public"
	DriveFileVisibilityOwner  DriveFileVisibility = "owner"
	DriveFileVisibilityShared DriveFileVisibility = "shared"

	DriveFileVisibilityDefault = DriveFileVisibilityOwner
)

type DriveFilePerm string

const (
	DriveFilePermNone               DriveFilePerm = ""
	DriveFilePermView               DriveFilePerm = "view"
	DriveFilePermEdit               DriveFilePerm = "edit"
	DriveFilePermEditTrash          DriveFilePerm = "edit-trash"
	DriveFilePermInheritedView      DriveFilePerm = "inherited-view"
	DriveFilePermInheritedEdit      DriveFilePerm = "inherited-edit"
	DriveFilePermInheritedEditTrash DriveFilePerm = "inherited-edit-trash"
	DriveFilePermAncestorOwner      DriveFilePerm = "ancestor-owner"
	DriveFilePermOwner              DriveFilePerm = "owner"

	DriveFilePermDefault = DriveFilePermView
)

type DriveFileStatus string

const (
	DriveFileStatusNone          DriveFileStatus = ""
	DriveFileStatusActive        DriveFileStatus = "active"
	DriveFileStatusInTrash       DriveFileStatus = "in-trash"
	DriveFileStatusParentInTrash DriveFileStatus = "parent-in-trash"
	DriveFileStatusPendingDelete DriveFileStatus = "pending-delete"

	DriveFileStatusDefault = DriveFileStatusActive
)

func driveFileStorageEnumValues() []string {
	return []string{string(DriveFileStorageS3)}
}

func driveFileVisibilityEnumValues() []string {
	return []string{
		string(DriveFileVisibilityPublic),
		string(DriveFileVisibilityOwner),
		string(DriveFileVisibilityShared),
	}
}

func driveFilePermEnumValues() []string {
	return []string{
		string(DriveFilePermView),
		string(DriveFilePermEdit),
		string(DriveFilePermEditTrash),
		string(DriveFilePermInheritedView),
		string(DriveFilePermInheritedEdit),
		string(DriveFilePermInheritedEditTrash),
		string(DriveFilePermAncestorOwner),
		string(DriveFilePermOwner),
	}
}

func driveFileStatusEnumValues() []string {
	return []string{
		string(DriveFileStatusActive),
		string(DriveFileStatusInTrash),
		string(DriveFileStatusParentInTrash),
		string(DriveFileStatusPendingDelete),
	}
}

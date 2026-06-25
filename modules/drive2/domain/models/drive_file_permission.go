package models

type Permission struct {
	IsFolder   bool
	Permission DriveFilePerm
}

func driveFilePermRank(p DriveFilePerm) int {
	switch p {
	case DriveFilePermNone:
		return 0
	case DriveFilePermView:
		return 1
	case DriveFilePermInheritedView:
		return 2
	case DriveFilePermEdit:
		return 3
	case DriveFilePermInheritedEdit:
		return 4
	case DriveFilePermEditTrash:
		return 5
	case DriveFilePermInheritedEditTrash:
		return 6
	case DriveFilePermAncestorOwner:
		return 7
	case DriveFilePermOwner:
		return 8
	default:
		return -1
	}
}

func (this Permission) CanView() bool {
	return driveFilePermRank(this.Permission) >= driveFilePermRank(DriveFilePermView)
}

func (this Permission) CanCreateTo() bool {
	return this.IsFolder && driveFilePermRank(this.Permission) >= driveFilePermRank(DriveFilePermEdit)
}

func (this Permission) CanUpdate() bool {
	return driveFilePermRank(this.Permission) >= driveFilePermRank(DriveFilePermEdit)
}

func (this Permission) CanDelete() bool {
	return driveFilePermRank(this.Permission) >= driveFilePermRank(DriveFilePermAncestorOwner)
}

func (this Permission) CanMoveToTrash() bool {
	r := driveFilePermRank(this.Permission)
	if !this.IsFolder && r >= driveFilePermRank(DriveFilePermEditTrash) {
		return true
	}
	return r >= driveFilePermRank(DriveFilePermInheritedEditTrash)
}

func (this Permission) CanRestore() bool {
	return this.CanMoveToTrash()
}

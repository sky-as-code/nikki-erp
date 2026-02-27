package enum

import (
	"github.com/sky-as-code/nikki-erp/common/enum_util"
)

type DriveFileStatus uint8

const (
	DriveFileStatusNone DriveFileStatus = iota
	DriveFileStatusActive
	DriveFileStatusInTrash
	DriveFileStatusParentInTrash
	DriveFileStatusPendingDelete

	DriveFileStatusDefault = DriveFileStatusActive
)

var DriveFileStatusName = map[DriveFileStatus]string{
	DriveFileStatusNone:          "",
	DriveFileStatusActive:        "active",
	DriveFileStatusInTrash:       "in-trash",
	DriveFileStatusParentInTrash: "parent-in-trash",
	DriveFileStatusPendingDelete: "pending-delete",
}

var DriveFileStatusValue = func() map[string]DriveFileStatus {
	m := map[string]DriveFileStatus{}
	for k, v := range DriveFileStatusName {
		m[v] = k
	}

	return m
}()

func (e DriveFileStatus) EnumDescriptions() string {
	return enum_util.DescriptionFromMap(DriveFileStatusName)
}

func (e *DriveFileStatus) UnmarshalJSON(data []byte) error {
	v, err := enum_util.UnmarshalJSON(data, DriveFileStatusValue, DriveFileStatusName)
	if err != nil {
		return err
	}

	*e = v
	return nil
}

func (e DriveFileStatus) MarshalJSON() ([]byte, error) {
	return enum_util.MarshalJSON(&e, DriveFileStatusName)
}

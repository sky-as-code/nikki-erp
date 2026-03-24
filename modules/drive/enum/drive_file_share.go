package enum

import (
	"database/sql/driver"

	"github.com/sky-as-code/nikki-erp/common/enum_util"
)

type DriveFilePerm uint8

const (
	DriveFilePermNone DriveFilePerm = iota
	DriveFilePermView
	DriveFilePermInheritedView
	DriveFilePermEdit
	DriveFilePermInheritedEdit
	DriveFilePermEditTrash
	DriveFilePermInheritedEditTrash
	DriveFilePermAncestorOwner
	DriveFilePermOwner

	DriveFilePermDefault = DriveFilePermView
)

var DriveFileSharePermName = map[DriveFilePerm]string{
	DriveFilePermNone:               "",
	DriveFilePermView:               "view",
	DriveFilePermEdit:               "edit",
	DriveFilePermEditTrash:          "edit-trash",
	DriveFilePermInheritedView:      "inherited-view",
	DriveFilePermInheritedEdit:      "inherited-edit",
	DriveFilePermInheritedEditTrash: "inherited-edit-trash",
	DriveFilePermAncestorOwner:      "ancestor-owner",
	DriveFilePermOwner:              "owner",
}

var DriveFileSharePermValue = func() map[string]DriveFilePerm {
	m := map[string]DriveFilePerm{}
	for k, v := range DriveFileSharePermName {
		m[v] = k
	}

	return m
}()

func (e DriveFilePerm) EnumDescriptions() string {
	return enum_util.DescriptionFromMap(DriveFileSharePermName)
}

func (e *DriveFilePerm) UnmarshalJSON(data []byte) error {
	v, err := enum_util.UnmarshalJSON(data, DriveFileSharePermValue, DriveFileSharePermName)
	if err != nil {
		return err
	}

	*e = v
	return nil
}

func (e DriveFilePerm) MarshalJSON() ([]byte, error) {
	return enum_util.MarshalJSON(&e, DriveFileSharePermName)
}

func (e DriveFilePerm) Value() (driver.Value, error) {
	return enum_util.ValueSQL(&e, DriveFileSharePermName)
}

func (e *DriveFilePerm) Scan(src any) error {
	v, err := enum_util.ScanSQL(src, DriveFileSharePermValue, DriveFileSharePermName)
	if err != nil {
		return err
	}
	*e = v
	return nil
}

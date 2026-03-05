package enum

import (
	"database/sql/driver"

	"github.com/sky-as-code/nikki-erp/common/enum_util"
)

type DriveFileSharePerm uint8

const (
	DriveFileSharePermView DriveFileSharePerm = iota + 1
	DriveFileSharePermEdit
	DriveFileSharePermEditTrash

	DriveFileSharePermDefault = DriveFileSharePermView
)

var DriveFileSharePermName = map[DriveFileSharePerm]string{
	DriveFileSharePermView:      "view",
	DriveFileSharePermEdit:      "edit",
	DriveFileSharePermEditTrash: "edit-trash",
}

var DriveFileSharePermValue = func() map[string]DriveFileSharePerm {
	m := map[string]DriveFileSharePerm{}
	for k, v := range DriveFileSharePermName {
		m[v] = k
	}

	return m
}()

func (e DriveFileSharePerm) EnumDescriptions() string {
	return enum_util.DescriptionFromMap(DriveFileSharePermName)
}

func (e *DriveFileSharePerm) UnmarshalJSON(data []byte) error {
	v, err := enum_util.UnmarshalJSON(data, DriveFileSharePermValue, DriveFileSharePermName)
	if err != nil {
		return err
	}

	*e = v
	return nil
}

func (e DriveFileSharePerm) MarshalJSON() ([]byte, error) {
	return enum_util.MarshalJSON(&e, DriveFileSharePermName)
}

func (e DriveFileSharePerm) Value() (driver.Value, error) {
	return enum_util.ValueSQL(&e, DriveFileSharePermName)
}

func (e *DriveFileSharePerm) Scan(src any) error {
	v, err := enum_util.ScanSQL(src, DriveFileSharePermValue, DriveFileSharePermName)
	if err != nil {
		return err
	}
	*e = v
	return nil
}

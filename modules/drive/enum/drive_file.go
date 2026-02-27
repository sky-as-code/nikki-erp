package enum

import (
	"database/sql/driver"

	"github.com/sky-as-code/nikki-erp/common/enum_util"
)

type ScopeType uint8

const (
	ScopeTypeDomain ScopeType = iota + 1
	ScopeTypeOrg
	ScopeTypeHierachy
	ScopeTypePrivate

	ScopeTypeDefault = ScopeTypePrivate
)

var ScopeTypeName = map[ScopeType]string{
	ScopeTypeDomain:   "domain",
	ScopeTypeOrg:      "org",
	ScopeTypeHierachy: "hierachy",
	ScopeTypePrivate:  "private",
}

var ScopeTypeValue = func() map[string]ScopeType {
	m := map[string]ScopeType{}
	for k, v := range ScopeTypeName {
		m[v] = k
	}

	return m
}()

func (e *ScopeType) EnumDescriptions() string {
	return enum_util.DescriptionFromMap(ScopeTypeName)
}

func (e *ScopeType) UnmarshalJSON(data []byte) error {
	v, err := enum_util.UnmarshalJSON(data, ScopeTypeValue, ScopeTypeName)
	if err != nil {
		return err
	}

	*e = v

	return nil
}

func (e *ScopeType) MarshalJSON() ([]byte, error) {
	return enum_util.MarshalJSON(e, ScopeTypeName)
}

func (e ScopeType) Value() (driver.Value, error) {
	return enum_util.ValueSQL(&e, ScopeTypeName)
}

func (e *ScopeType) Scan(src any) error {
	v, err := enum_util.ScanSQL(src, ScopeTypeValue, ScopeTypeName)
	if err != nil {
		return err
	}
	*e = v
	return nil
}

type DriveFileStorage uint8

const (
	DriveFileStorageS3 DriveFileStorage = iota + 1

	DriveFileStorageDefault = DriveFileStorageS3
)

var DriveFileStorageName = map[DriveFileStorage]string{
	DriveFileStorageS3: "s3",
}

var DriveFileStorageValue = func() map[string]DriveFileStorage {
	m := map[string]DriveFileStorage{}
	for k, v := range DriveFileStorageName {
		m[v] = k
	}

	return m
}()

func (e *DriveFileStorage) EnumDescriptions() string {
	return enum_util.DescriptionFromMap(DriveFileStorageName)
}

func (e *DriveFileStorage) UnmarshalJSON(data []byte) error {
	v, err := enum_util.UnmarshalJSON(data, DriveFileStorageValue, DriveFileStorageName)
	if err != nil {
		return err
	}

	*e = v
	return nil
}

func (e *DriveFileStorage) MarshalJSON() ([]byte, error) {
	return enum_util.MarshalJSON(e, DriveFileStorageName)
}

func (e DriveFileStorage) Value() (driver.Value, error) {
	return enum_util.ValueSQL(&e, DriveFileStorageName)
}

func (e *DriveFileStorage) Scan(src any) error {
	v, err := enum_util.ScanSQL(src, DriveFileStorageValue, DriveFileStorageName)
	if err != nil {
		return err
	}
	*e = v
	return nil
}

type DriveFileVisibility uint8

const (
	DriveFileVisibilityPublic DriveFileVisibility = iota + 1
	DriveFileVisibilityOwner
	DriveFileVisibilityShared

	DriveFileVisibilityDefault = DriveFileVisibilityOwner
)

var DriveFileVisibilityName = map[DriveFileVisibility]string{
	DriveFileVisibilityPublic: "public",
	DriveFileVisibilityOwner:  "owner",
	DriveFileVisibilityShared: "shared",
}

var DriveFileVisibilityValue = func() map[string]DriveFileVisibility {
	m := map[string]DriveFileVisibility{}
	for k, v := range DriveFileVisibilityName {
		m[v] = k
	}

	return m
}()

func (e DriveFileVisibility) EnumDescriptions() string {
	return enum_util.DescriptionFromMap(DriveFileVisibilityName)
}

func (e *DriveFileVisibility) UnmarshalJSON(data []byte) error {
	v, err := enum_util.UnmarshalJSON(data, DriveFileVisibilityValue, DriveFileVisibilityName)
	if err != nil {
		return err
	}

	*e = v
	return nil
}

func (e *DriveFileVisibility) MarshalJSON() ([]byte, error) {
	return enum_util.MarshalJSON(e, DriveFileVisibilityName)
}

func (e DriveFileVisibility) Value() (driver.Value, error) {
	return enum_util.ValueSQL(&e, DriveFileVisibilityName)
}

func (e *DriveFileVisibility) Scan(src any) error {
	v, err := enum_util.ScanSQL(src, DriveFileVisibilityValue, DriveFileVisibilityName)
	if err != nil {
		return err
	}
	*e = v
	return nil
}

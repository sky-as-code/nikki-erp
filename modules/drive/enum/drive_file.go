package enum

import (
	"github.com/sky-as-code/nikki-erp/common/enum_util"
)

type ScopeType uint8

const (
	ScopeTypeNone ScopeType = iota
	ScopeTypeDomain
	ScopeTypeOrg
	ScopeTypeHierachy
	ScopeTypePrivate

	ScopeTypeDefault = ScopeTypePrivate
)

var ScopeTypeName = map[ScopeType]string{
	ScopeTypeNone:     "",
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

func (e *ScopeType) UnmarshalText(text []byte) error {
	v, err := enum_util.UnmarshalText(text, ScopeTypeValue, ScopeTypeName)
	if err != nil {
		return err
	}
	*e = v
	return nil
}

func (e ScopeType) MarshalJSON() ([]byte, error) {
	return enum_util.MarshalJSON(&e, ScopeTypeName)
}

type DriveFileStorage uint8

const (
	DriveFileStorageNone DriveFileStorage = iota
	DriveFileStorageS3

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

func (e *DriveFileStorage) UnmarshalText(text []byte) error {
	v, err := enum_util.UnmarshalText(text, DriveFileStorageValue, DriveFileStorageName)
	if err != nil {
		return err
	}
	*e = v
	return nil
}

func (e DriveFileStorage) MarshalJSON() ([]byte, error) {
	return enum_util.MarshalJSON(&e, DriveFileStorageName)
}

type DriveFileVisibility uint8

const (
	DriveFileVisibilityNone DriveFileVisibility = iota
	DriveFileVisibilityPublic
	DriveFileVisibilityOwner
	DriveFileVisibilityShared

	DriveFileVisibilityDefault = DriveFileVisibilityOwner
)

var DriveFileVisibilityName = map[DriveFileVisibility]string{
	DriveFileVisibilityNone:   "",
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

func (e *DriveFileVisibility) UnmarshalText(text []byte) error {
	v, err := enum_util.UnmarshalText(text, DriveFileVisibilityValue, DriveFileVisibilityName)
	if err != nil {
		return err
	}
	*e = v
	return nil
}

func (e DriveFileVisibility) MarshalJSON() ([]byte, error) {
	return enum_util.MarshalJSON(&e, DriveFileVisibilityName)
}

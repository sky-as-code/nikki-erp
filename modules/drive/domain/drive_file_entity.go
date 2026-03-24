package domain

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/invopop/validation"
	"github.com/samber/lo"
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/drive/enum"
)

type DriveFile struct {
	model.ModelBase     `json:",inline"`
	model.AuditableBase `json:",inline"`

	OwnerRef           *model.Id `json:"owner_ref"`
	ParentDriveFileRef *model.Id `json:"parent_drive_file_ref"`
	MaterializedPath   *string    `json:"materialized_path,omitempty"`

	Name        string                   `json:"name"`
	MINE        string                   `json:"mine"`
	IsFolder    bool                     `json:"isFolder"`
	Size        uint64                   `json:"size"`
	StoragePath string                   `json:"storagePath"`
	Storage     enum.DriveFileStorage    `json:"storage"`
	Visibility  enum.DriveFileVisibility `json:"visiblity"`
	Status      enum.DriveFileStatus     `json:"status"`
	StorageKey  string                   `json:"-"`

	File       multipart.File
	FileHeader *multipart.FileHeader

	Children []*DriveFile `json:"-"`

	DeletedAt *time.Time `json:"deletedAt,omitempty"`

	// UserId is the acting user for permission checks on mutations; not persisted.
	UserId model.Id `json:"-"`
}

var (
	driveFileNameRegex = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
	driveFileReserved  = []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}
)

func (d *DriveFile) Validate(forEdit bool) fault.ValidationErrors {
	var nameValidator *validator.FieldRules
	if d.IsFolder {
		nameValidator = validator.Field(&d.Name,
			validator.Length(3, 200))
	} else {
		nameValidator = validator.Field(&d.Name,
			validator.Length(3, 200),
			validation.By(driveFileNameValidator))
	}

	rules := []*validator.FieldRules{nameValidator}

	return validator.ApiBased.ValidateStruct(d, rules...)
}

func driveFileNameValidator(value interface{}) error {
	s, ok := value.(string)
	if !ok || len(s) == 0 {
		return nil
	}
	if !driveFileNameRegex.MatchString(s) {
		return validation.NewError("drive_file_name",
			"name may only contain letters, numbers, dots, underscores and hyphens")
	}
	if s[len(s)-1] == '.' || s[len(s)-1] == ' ' {
		return validation.NewError("drive_file_name", "name must not end with dot or space")
	}
	upper := strings.ToUpper(s)
	for _, reserved := range driveFileReserved {
		if upper == reserved {
			return validation.NewError("drive_file_name", "name is a reserved Windows filename")
		}
		if len(upper) > len(reserved) &&
			strings.HasPrefix(upper, reserved) &&
			upper[len(reserved)] == '.' {
			return validation.NewError("drive_file_name", "name is a reserved Windows filename")
		}
	}
	return nil
}

func (d *DriveFile) Process() {
	if d.FileHeader == nil {
		return
	}

	d.Size = uint64(d.FileHeader.Size)

	// MIME detection
	if d.File != nil {
		buffer := make([]byte, 512)
		_, err := d.File.Read(buffer)
		if err != nil {
			panic(err)
		}

		MIME := http.DetectContentType(buffer)
		d.File.Seek(0, 0)

		d.MINE = MIME
	}
}

func (d *DriveFile) BuildStorageInfo(endpoint string) {
	if d.OwnerRef == nil || d.Id == nil {
		return
	}

	d.StorageKey = fmt.Sprintf("%s/%s", *d.OwnerRef, *d.Id)

	switch d.Storage {
	case enum.DriveFileStorageS3:
		d.StoragePath = fmt.Sprintf("%s/%s", endpoint, d.StorageKey)
	}
}

func (d *DriveFile) BuildTree(children []*DriveFile) {
	children = append(children, d)

	childrenMap := lo.SliceToMap(children, func(driveFile *DriveFile) (model.Id, *DriveFile) {
		return *driveFile.Id, driveFile
	})

	for _, driveFile := range children {
		if driveFile.ParentDriveFileRef != nil {
			if parent, ok := childrenMap[*driveFile.ParentDriveFileRef]; ok {
				parent.Children = append(parent.Children, driveFile)
			}
		}
	}
}

package domain

import (
	"mime/multipart"
	"net/http"
	"regexp"
	"time"

	"github.com/invopop/validation"
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/drive/enum"
)

type DriveFile struct {
	model.ModelBase     `json:",inline"`
	model.AuditableBase `json:",inline"`

	ScopeType enum.ScopeType `json:"scope_type"`
	ScopeRef  model.Id       `json:"scope_ref"`

	OwnerRef           model.Id `json:"owner_ref"`
	ParentDriveFileRef model.Id `json:"parent_drive_file_ref"`

	Name      string                   `json:"name"`
	MINE      string                   `json:"mine"`
	IsFolder  bool                     `json:"is_folder"`
	Size      uint64                   `json:"size"`
	Path      string                   `json:"path"`
	Storage   enum.DriveFileStorage    `json:"storage"`
	Visibility enum.DriveFileVisibility `json:"visiblity"`

	File       multipart.File
	FileHeader multipart.FileHeader

	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

func (d *DriveFile) Validate(forEdit bool) fault.ValidationErrors {
	reg := regexp.MustCompile("^(?!.{1,2}$)(?!(?i:CON|PRN|AUX|NUL|COM[1-9]|LPT[1-9])(..*)?$)[A-Za-z0-9._-]+(?<![. ])$")

	rules := []*validator.FieldRules{
		validator.Field(&d.Name,
			validation.Length(1, 200),
			validation.Match(reg),
		),
	}

	return validator.ApiBased.ValidateStruct(d, rules...)
}

func (d *DriveFile) Process() {
	d.Size = uint64(d.FileHeader.Size)

	// MIME detection
	buffer := make([]byte, 512)
	_, err := d.File.Read(buffer)
	if err != nil {
		panic(err)
	}

	MIME := http.DetectContentType(buffer)
	d.File.Seek(0, 0)

	d.MINE = MIME
}

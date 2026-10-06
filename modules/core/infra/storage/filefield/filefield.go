package filefield

import (
	"mime/multipart"
	"time"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/filestorage"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/objectkey"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/upload"
)

const PresignedUrlTtl = time.Hour

var ImageMimes = []string{"image/png", "image/jpeg", "image/gif", "image/webp", "image/svg+xml"}

const MaxImageSize = 10 << 20

type FileField struct {
	Name         string
	KeyField     string
	UrlField     string
	KeyPrefix    string
	MaxSize      int64
	AllowedMimes *[]string
	MimeField    string
}

func FindFileField(fields []FileField, name string) (FileField, bool) {
	for _, field := range fields {
		if field.Name == name {
			return field, true
		}
	}

	return FileField{}, false
}

func PresignFileFields(
	ctx corectx.Context, storage filestorage.FileStorageAdapter,
	record dmodel.DynamicFields, fields []FileField,
) error {
	for _, field := range fields {
		key := record.GetString(field.KeyField)
		if key == nil || *key == "" {
			continue
		}

		url, err := storage.GeneratePresignedUrl(ctx.InnerContext(), *key, PresignedUrlTtl)
		if err != nil {
			return errors.Wrapf(err, "failed to presign '%s'", field.KeyField)
		}
		record[field.UrlField] = url
	}

	return nil
}

func PresignPage(
	ctx corectx.Context, storage filestorage.FileStorageAdapter,
	records []dmodel.DynamicFields, fields []FileField,
) error {
	for i := range records {
		if err := PresignFileFields(ctx, storage, records[i], fields); err != nil {
			return err
		}
	}

	return nil
}

func StoreFile(
	ctx corectx.Context, storage filestorage.FileStorageAdapter,
	field FileField, file *multipart.FileHeader, vErrs *ft.ClientErrors,
) (string, string, error) {
	key, err := objectkey.BuildFromFileHeader(field.KeyPrefix, file)
	if err != nil {
		return "", "", errors.Wrapf(err, "failed to name the object for '%s'", field.Name)
	}

	before := vErrs.Count()
	prepared, err := upload.Prepare(&upload.PrepareParam{
		FieldName:    field.Name,
		FileHeader:   file,
		MaxSize:      field.MaxSize,
		AllowedMimes: field.AllowedMimes,
	}, vErrs)
	if err != nil {
		return "", "", err
	}
	if prepared == nil {
		if vErrs.Count() > before {
			return "", "", nil
		}

		return "", "", errors.Errorf("failed to prepare the upload of '%s'", field.Name)
	}
	defer prepared.Close()

	if err := storage.Put(ctx.InnerContext(), key, prepared.Reader,
		filestorage.NewPutOptions(prepared.MimeType, prepared.Size)); err != nil {
		return "", "", errors.Wrapf(err, "failed to store '%s'", field.Name)
	}

	return key, prepared.MimeType, nil
}

func RemoveObjects(ctx corectx.Context, storage filestorage.FileStorageAdapter, keys []string) {
	for _, key := range keys {
		_ = storage.Remove(ctx.InnerContext(), key)
	}
}

func ManagedFieldViolations(names []string, fields []FileField) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	named := make(map[string]bool, len(names))
	for _, name := range names {
		named[name] = true
	}
	for _, field := range fields {
		for _, column := range []string{field.KeyField, field.MimeField} {
			if column != "" && named[column] {
				vErrs.Append(*ft.NewBusinessViolation(column, "file.err_field_managed",
					"'"+column+"' is set through files/"+field.Name+", not through create or update"))
			}
		}
	}

	return vErrs
}

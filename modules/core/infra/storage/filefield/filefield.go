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

const (
	PresignedUrlTtl = time.Hour
	UploadParamKey  = "__file_uploads"
	ClearParamKey   = "fields"
)

var ImageMimes = []string{"image/png", "image/jpeg", "image/gif", "image/webp", "image/svg+xml"}

const MaxImageSize = 10 << 20

type FileField struct {
	KeyField     string
	UrlField     string
	UploadField  string
	KeyPrefix    string
	MaxSize      int64
	AllowedMimes *[]string
	MimeField    string
}

type PendingUpload struct {
	Field FileField
	File  *multipart.FileHeader
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

func ReadClearFileParams(params dmodel.DynamicFields, allowed []FileField) ([]FileField, *ft.ClientErrors) {
	raw, present := params[ClearParamKey]
	if !present || raw == nil {
		return resolveClearFileFields(nil, allowed)
	}

	rawList, ok := raw.([]any)
	if !ok {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewValidationError(ClearParamKey, "file.err_clear_fields_not_a_list",
			"'fields' must be a list of file field names"))
		return nil, vErrs
	}

	names := make([]string, 0, len(rawList))
	for _, item := range rawList {
		name, ok := item.(string)
		if !ok {
			vErrs := ft.NewClientErrors()
			vErrs.Append(*ft.NewValidationError(ClearParamKey, "file.err_clear_field_not_a_string",
				"every entry of 'fields' must be a field name"))
			return nil, vErrs
		}
		names = append(names, name)
	}

	return resolveClearFileFields(names, allowed)
}

func resolveClearFileFields(requested []string, allowed []FileField) ([]FileField, *ft.ClientErrors) {
	vErrs := ft.NewClientErrors()
	if len(requested) == 0 {
		vErrs.Append(*ft.NewValidationError(ClearParamKey, "file.err_clear_fields_required",
			"name at least one file field to clear"))
		return nil, vErrs
	}

	byUrlField := make(map[string]FileField, len(allowed))
	for _, field := range allowed {
		byUrlField[field.UrlField] = field
	}

	seen := make(map[string]bool, len(requested))
	resolved := make([]FileField, 0, len(requested))
	for _, name := range requested {
		field, known := byUrlField[name]
		if !known {
			vErrs.Append(*ft.NewValidationError(name, "file.err_clear_field_unknown",
				"'"+name+"' is not a file field of this resource"))
			continue
		}
		if seen[name] {
			continue
		}
		seen[name] = true
		resolved = append(resolved, field)
	}

	if vErrs.Count() > 0 {
		return nil, vErrs
	}

	return resolved, vErrs
}

func StoreUploads(
	ctx corectx.Context, storage filestorage.FileStorageAdapter, params dmodel.DynamicFields,
) ([]string, *ft.ClientErrors, error) {
	raw, present := params[UploadParamKey]
	delete(params, UploadParamKey)

	vErrs := ft.NewClientErrors()
	if !present || raw == nil {
		return nil, vErrs, nil
	}

	uploads, ok := raw.([]PendingUpload)
	if !ok {
		return nil, vErrs, errors.New("the upload parameter carries something other than bound files")
	}

	stored := make([]string, 0, len(uploads))
	for _, pending := range uploads {
		key, mime, err := storeOne(ctx, storage, pending, vErrs)
		if err != nil {
			RemoveObjects(ctx, storage, stored)
			return nil, vErrs, err
		}
		if key == "" {
			continue
		}

		stored = append(stored, key)
		params.SetString(pending.Field.KeyField, &key)
		if pending.Field.MimeField != "" {
			params.SetString(pending.Field.MimeField, &mime)
		}
	}

	if vErrs.Count() > 0 {
		RemoveObjects(ctx, storage, stored)
		return nil, vErrs, nil
	}

	return stored, vErrs, nil
}

func storeOne(
	ctx corectx.Context, storage filestorage.FileStorageAdapter,
	pending PendingUpload, vErrs *ft.ClientErrors,
) (string, string, error) {
	key, err := objectkey.BuildFromFileHeader(pending.Field.KeyPrefix, pending.File)
	if err != nil {
		return "", "", errors.Wrapf(err, "failed to name the object for '%s'", pending.Field.UploadField)
	}

	before := vErrs.Count()
	prepared, err := upload.Prepare(&upload.PrepareParam{
		FieldName:    pending.Field.UploadField,
		FileHeader:   pending.File,
		MaxSize:      pending.Field.MaxSize,
		AllowedMimes: pending.Field.AllowedMimes,
	}, vErrs)
	if err != nil {
		return "", "", err
	}
	if prepared == nil {
		if vErrs.Count() > before {
			return "", "", nil
		}

		return "", "", errors.Errorf("failed to prepare the upload of '%s'", pending.Field.UploadField)
	}
	defer prepared.Close()

	if err := storage.Put(ctx.InnerContext(), key, prepared.Reader,
		filestorage.NewPutOptions(prepared.MimeType, prepared.Size)); err != nil {
		return "", "", errors.Wrapf(err, "failed to store '%s'", pending.Field.UploadField)
	}

	return key, prepared.MimeType, nil
}

func RemoveObjects(ctx corectx.Context, storage filestorage.FileStorageAdapter, keys []string) {
	for _, key := range keys {
		_ = storage.Remove(ctx.InnerContext(), key)
	}
}

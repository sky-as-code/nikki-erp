package composable

import (
	"path"
	"strings"

	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/tabular"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/filestorage"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/objectkey"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/upload"
)

// importObjectPrefix is the bucket folder transient import files live in until processed.
const importObjectPrefix = "imports"

// importMimesByExtension is what the sniffed content may be for each accepted extension. A
// csv sniffs as plain text (with a charset suffix, hence the wildcard) and an xlsx as its own
// type or as the zip container it is.
var importMimesByExtension = map[string][]string{
	tabular.ExtCsv: {"text/*", "application/csv"},
	tabular.ExtXlsx: {
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"application/zip", "application/x-zip-compressed",
	},
}

// Import is POST {resource}/import: after the same permission check as create, the file is
// bounded by the configured limits, parked in object storage under a random name, parsed from
// there and written through bulk create. The object is removed whatever the outcome.
//
// The order is deliberate - nothing touches storage before the permission check - for the
// reason coremart's vending-machine upload states: a stored file is a side effect a refused
// caller must not be able to cause.
func (this *DefaultApplicationServiceImpl) Import(
	ctx corectx.Context, cmd ImportCommand,
) (*BulkCreateResult, error) {
	if cErrs := assertActionSupported(this.crudActions, this.ResourceCode(), CrudActionImport); cErrs != nil {
		return &BulkCreateResult{ClientErrors: *cErrs}, nil
	}
	orgId, cErrs := this.guardCreateLike(ctx, cmd.Params)
	if cErrs != nil {
		return &BulkCreateResult{ClientErrors: *cErrs}, nil
	}
	if this.storage == nil {
		return importRefused(ft.NewAnonymousBusinessViolation(ErrImportStorageUnavailable,
			"object storage is not configured, import is unavailable")), nil
	}

	prepared, ext, cErrs := this.prepareImportFile(cmd)
	if cErrs != nil {
		return &BulkCreateResult{ClientErrors: *cErrs}, nil
	}
	defer prepared.Close()

	key, err := this.storeImportFile(ctx, prepared, ext)
	if err != nil {
		return nil, err
	}
	defer this.removeImportFile(ctx, key)

	table, cErrs, err := this.readImportFile(ctx, key, ext)
	if err != nil {
		return nil, err
	}
	if cErrs != nil {
		return &BulkCreateResult{ClientErrors: *cErrs}, nil
	}
	return importRows(ctx, this.domSvc, table, cmd.Mapping, orgId, LookupSourceOnion)
}

// prepareImportFile applies the extension, size and content-type limits. Every refusal is a
// client error naming the "file" part.
func (this *DefaultApplicationServiceImpl) prepareImportFile(cmd ImportCommand) (*upload.PreparedFile, string, *ft.ClientErrors) {
	if cmd.FileHeader == nil {
		return nil, "", singleClientError(ft.NewValidationError(ImportFormFile, ErrImportFileRequired, "a file is required"))
	}
	ext := strings.ToLower(strings.TrimPrefix(path.Ext(cmd.FileHeader.Filename), "."))
	if !this.importLimits.IsAllowed(ext) {
		return nil, "", singleClientError(ft.NewValidationError(ImportFormFile, "err_file_type_not_allowed",
			"file type {{actual}} is not one of {{allowed}}",
			map[string]any{"allowed": this.importLimits.AllowedExtensions, "actual": ext}))
	}
	mimes := importMimesByExtension[ext]
	cErrs := ft.NewClientErrors()
	prepared, err := upload.Prepare(&upload.PrepareParam{
		FieldName:    ImportFormFile,
		FileHeader:   cmd.FileHeader,
		MaxSize:      this.importLimits.MaxFileSizeBytes,
		AllowedMimes: &mimes,
	}, cErrs)
	if err != nil {
		return nil, "", singleClientError(ft.NewValidationError(ImportFormFile, ErrImportFileUnreadable, err.Error()))
	}
	if cErrs.Count() > 0 {
		return nil, "", cErrs
	}
	return prepared, ext, nil
}

func (this *DefaultApplicationServiceImpl) storeImportFile(
	ctx corectx.Context, prepared *upload.PreparedFile, ext string,
) (string, error) {
	key, err := objectkey.BuildRandom(importObjectPrefix, ext)
	if err != nil {
		return "", errors.Wrap(err, "import object key")
	}
	err = this.storage.Put(ctx, key, prepared.Reader, filestorage.NewPutOptions(prepared.MimeType, prepared.Size))
	if err != nil {
		return "", errors.Wrapf(err, "store import file '%s'", key)
	}
	return key, nil
}

// removeImportFile is the deferred cleanup. A failure is logged with the key so an operator
// can sweep the bucket; it never changes the outcome the caller already got.
func (this *DefaultApplicationServiceImpl) removeImportFile(ctx corectx.Context, key string) {
	if err := this.storage.Remove(ctx, key); err != nil && this.logger != nil {
		this.logger.Error("failed to remove import file '"+key+"' from object storage", err)
	}
}

// readImportFile parses the stored object. A row overflow and an unparseable file are the
// caller's problems (400); a storage read failure is the server's.
func (this *DefaultApplicationServiceImpl) readImportFile(
	ctx corectx.Context, key string, ext string,
) (*tabular.Table, *ft.ClientErrors, error) {
	object, err := this.storage.Open(ctx, key, "")
	if err != nil {
		return nil, nil, errors.Wrapf(err, "open import file '%s'", key)
	}
	defer object.Body.Close()

	table, err := tabular.Read(object.Body, ext, tabular.Options{MaxRows: this.importLimits.MaxRows})
	if errors.Is(err, tabular.ErrTooManyRows) {
		return nil, singleClientError(ft.NewValidationError(ImportFormFile, ErrImportTooManyRows,
			"the file has more than {{max_rows}} rows", map[string]any{"max_rows": this.importLimits.MaxRows})), nil
	}
	if err != nil {
		return nil, singleClientError(ft.NewValidationError(ImportFormFile, ErrImportFileUnreadable,
			"the file could not be read as {{extension}}", map[string]any{"extension": ext})), nil
	}
	return table, nil, nil
}

func importRefused(item *ft.ClientErrorItem) *BulkCreateResult {
	return &BulkCreateResult{ClientErrors: *singleClientError(item)}
}

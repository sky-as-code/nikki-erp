package composable

import (
	"encoding/json"
	"mime/multipart"
	"strings"

	"github.com/labstack/echo/v5"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
)

// maxImportFormMemory is how much of the multipart body stays in memory before spilling to a
// temp file. The file part itself is bounded later by the configured size limit.
const maxImportFormMemory = 8 << 20

// Import serves POST {resource}/import: a multipart form with exactly two parts, "file" and
// "mapping" (a JSON document). org_id travels in the query string, as on every read route.
func (this *CrudRestBase) Import(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST import"); e != nil {
			err = e
		}
	}()

	reqCtx, err := corectx.AsRequestContext(echoCtx)
	if err != nil {
		return err
	}
	cmd, err := bindImportCommand(echoCtx, this.schema())
	if err != nil {
		return bindFailure(echoCtx, err)
	}

	result, err := this.appSvc.Import(reqCtx, *cmd)
	if err != nil {
		return err
	}
	if result.ClientErrors != nil && result.ClientErrors.Count() > 0 {
		return httpserver.JsonBadRequest(echoCtx, result.ClientErrors)
	}
	if !result.HasData {
		return httpserver.JsonBadRequest(echoCtx, ft.ClientErrors{*ft.NewAnonymousNotFoundError()})
	}
	return httpserver.JsonOk(echoCtx, result.Data)
}

// bindImportCommand reads the form. A part the route does not know is refused rather than
// dropped, for the same reason create refuses an undeclared field.
func bindImportCommand(echoCtx *echo.Context, schema *dmodel.ModelSchema) (*ImportCommand, error) {
	form, err := echoCtx.MultipartForm()
	if err != nil {
		return nil, echo.NewHTTPError(400, "a multipart form is required: "+err.Error())
	}
	if cErrs := unknownImportParts(form); cErrs != nil {
		return nil, &unknownFieldsError{errors: *cErrs}
	}

	cmd := &ImportCommand{Params: dmodel.DynamicFields{}}
	mergeOrgId(echoCtx, schema, cmd.Params, ActionTypeUpload)
	if files := form.File[ImportFormFile]; len(files) > 0 {
		cmd.FileHeader = files[0]
	}

	rawMapping := strings.TrimSpace(strings.Join(form.Value[ImportFormMapping], ""))
	if rawMapping == "" {
		return nil, &unknownFieldsError{errors: *singleClientError(
			ft.NewValidationError(ImportFormMapping, ErrImportMappingInvalid, "the mapping part is required"))}
	}
	if err := json.Unmarshal([]byte(rawMapping), &cmd.Mapping); err != nil {
		return nil, &unknownFieldsError{errors: *singleClientError(
			ft.NewValidationError(ImportFormMapping, ErrImportMappingInvalid, "the mapping part is not valid JSON"))}
	}
	return cmd, nil
}

func unknownImportParts(form *multipart.Form) *ft.ClientErrors {
	cErrs := ft.NewClientErrors()
	for name := range form.Value {
		if name != ImportFormMapping {
			cErrs.Append(*ft.NewValidationError(name, "err_unknown_field", "unknown form part"))
		}
	}
	for name := range form.File {
		if name != ImportFormFile {
			cErrs.Append(*ft.NewValidationError(name, "err_unknown_field", "unknown form part"))
		}
	}
	if cErrs.Count() == 0 {
		return nil
	}
	return cErrs
}

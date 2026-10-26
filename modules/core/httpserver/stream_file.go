package httpserver

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v5"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/filestorage"
)

func ServeStreamFile[THttpReq FileStreamRequest, TSvcCommand any, TSvcResultData any](
	echoCtx *echo.Context,
	serviceFn func(ctx corectx.Context, cmd TSvcCommand) (*dyn.OpResult[TSvcResultData], error),
	requestToCommandFn func(req THttpReq) TSvcCommand,
	buildFileStreamResponse func(data TSvcResultData) GetFileStreamResponse,
	fallbackNameHeader string,
	skipNotFoundError ...bool,
) error {
	var request THttpReq
	if err := echoCtx.Bind(&request); err != nil {
		return JsonBadRequest(echoCtx, []any{ft.NewAnonymousValidationError(ft.ErrorKey("err_malformed_request"), "malformed request")})
	}
	// Bind reads path, query and body only; the Range header (FileStreamRequestBase) needs its own pass.
	if err := echo.BindHeaders(echoCtx, &request); err != nil {
		return JsonBadRequest(echoCtx, []any{ft.NewAnonymousValidationError(ft.ErrorKey("err_malformed_request"), "malformed request")})
	}

	reqCtx := echoCtx.Request().Context().(corectx.Context)
	result, err := serviceFn(reqCtx, requestToCommandFn(request))
	if err != nil {
		var unsatisfiable filestorage.ErrRangeNotSatisfiable
		if errors.As(err, &unsatisfiable) {
			return rangeNotSatisfiable(echoCtx, unsatisfiable.Size)
		}
		return err
	}
	if result == nil {
		return errors.New("service result is nil")
	}
	if result.ClientErrors != nil && result.ClientErrors.Count() > 0 {
		return JsonBadRequest(echoCtx, result.ClientErrors)
	}
	if !result.HasData && (len(skipNotFoundError) == 0 || !skipNotFoundError[0]) {
		cErrs := ft.ClientErrors{*ft.NewAnonymousNotFoundError()}
		return JsonBadRequest(echoCtx, cErrs)
	}
	return writeFileStream(echoCtx, request, buildFileStreamResponse(result.Data), fallbackNameHeader)
}

// rangeNotSatisfiable answers 416 with the Content-Range RFC 9110 §15.5.17 requires, so a client
// learns the length it may ask for.
func rangeNotSatisfiable(echoCtx *echo.Context, size int64) error {
	echoCtx.Response().Header().Set("Content-Range", fmt.Sprintf("bytes */%d", size))
	return echoCtx.NoContent(http.StatusRequestedRangeNotSatisfiable)
}

// writeFileStream answers 206 only when storage served a part: a Range header storage ignored
// (unparseable, or one the backend does not honour) yields the full body, which must be a 200.
func writeFileStream(
	echoCtx *echo.Context, request FileStreamRequest, streamInfo GetFileStreamResponse, fallbackName string,
) error {
	defer streamInfo.Body.Close()

	disposition := "inline"
	if request.IsDownloadRequest() {
		disposition = "attachment"
	}
	name := streamInfo.Name
	if name == "" {
		name = fallbackName
	}
	mimeType := "application/octet-stream"
	if streamInfo.MimeType != "" {
		mimeType = streamInfo.MimeType
	}

	resp := echoCtx.Response()
	resp.Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("%s; filename=%q", disposition, name))
	resp.Header().Set("Accept-Ranges", "bytes")
	resp.Header().Set(echo.HeaderContentType, mimeType)

	httpStatus := http.StatusOK
	if request.GetRangeHeader() != "" && streamInfo.ContentRange != nil && *streamInfo.ContentRange != "" {
		resp.Header().Set("Content-Range", *streamInfo.ContentRange)
		httpStatus = http.StatusPartialContent
	}
	if streamInfo.ContentLength != nil {
		resp.Header().Set(echo.HeaderContentLength, fmt.Sprintf("%d", *streamInfo.ContentLength))
	}
	return echoCtx.Stream(httpStatus, mimeType, streamInfo.Body)
}

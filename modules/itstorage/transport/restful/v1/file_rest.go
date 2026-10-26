package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itFile "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/file"
)

// FileEngineName is the container name of the onion, checked against
// composable.EngineDependencyName by a test.
const FileEngineName = "dynengine_itstorage_file"

type fileRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_itstorage_file"`
}

func NewFileRest(params fileRestParams) *FileRest {
	rest := &FileRest{fileSvc: params.Engine.ApplicationService().(itFile.FileApplicationService)}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

// FileRest serves the read and delete built-ins, plus the three file routes whose surface the
// engine cannot express: a streamed multipart upload and a streamed download.
type FileRest struct {
	composable.CrudRestBase
	fileSvc itFile.FileApplicationService
}

// Upload never buffers the file: the text parts are read first, then the file part's stream is
// handed to the service as is. The form is not parsed with ParseMultipartForm, which would spool
// the whole body to memory or disk before the handler saw a byte.
func (this *FileRest) Upload(echoCtx *echo.Context) (err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "handle REST upload file"); e != nil {
			err = e
		}
	}()
	reqCtx, err := corectx.AsRequestContext(echoCtx)
	if err != nil {
		return err
	}
	cmd, cErrs, err := readUploadForm(echoCtx.Request())
	if err != nil {
		return err
	}
	if cErrs.Count() > 0 {
		return httpserver.JsonBadRequest(echoCtx, cErrs)
	}
	result, err := this.fileSvc.UploadFile(reqCtx, *cmd)
	if err != nil {
		return err
	}
	if result.ClientErrors.Count() > 0 {
		return httpserver.JsonBadRequest(echoCtx, result.ClientErrors)
	}
	return httpserver.JsonCreated(echoCtx, newUploadFileResponse(result.Data))
}

func (this *FileRest) GetMeta(echoCtx *echo.Context) error {
	return httpserver.ServeRequest2(
		echoCtx, this.fileSvc.GetFileMeta,
		func(request GetFileMetaRequest) itFile.GetFileMetaQuery { return itFile.GetFileMetaQuery(request) },
		composable.Identity[dmodel.DynamicFields],
		httpserver.JsonOk,
	)
}

func (this *FileRest) Download(echoCtx *echo.Context) error {
	return httpserver.ServeStreamFile(
		echoCtx, this.fileSvc.DownloadFile,
		func(request DownloadFileRequest) itFile.DownloadFileQuery {
			return itFile.DownloadFileQuery{Path: request.Path, RangeHeader: request.GetRangeHeader()}
		},
		func(data itFile.DownloadFileResultData) httpserver.GetFileStreamResponse {
			return httpserver.GetFileStreamResponse{
				Name:          data.Name,
				MimeType:      data.MimeType,
				Body:          data.Body,
				ContentLength: data.ContentLength,
				ContentRange:  data.ContentRange,
			}
		},
		"file",
	)
}

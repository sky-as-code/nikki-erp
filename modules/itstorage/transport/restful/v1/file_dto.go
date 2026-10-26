package v1

import (
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	c "github.com/sky-as-code/nikki-erp/modules/itstorage/constants"
	itFile "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/file"
)

// Text part names of the upload form; the file part (c.FieldFilePart) must come last.
const (
	partOrgId     = "org_id"
	partSpaceId   = "space_id"
	partName      = "name"
	partMimeType  = "mime_type"
	partSizeBytes = "size_bytes"

	// maxTextPartBytes bounds one text part, so a client cannot stream an unbounded "name".
	maxTextPartBytes = 4096
)

type GetFileMetaRequest struct {
	Path string `json:"path" query:"path"`
}

type DownloadFileRequest struct {
	httpserver.FileStreamRequestBase
	Path string `json:"path" query:"path"`
}

// UploadFileResponse is the built-in create answer plus the path to store in `*_file_path`.
type UploadFileResponse struct {
	httpserver.RestCreateResponse
	FilePath string `json:"file_path"`
}

func newUploadFileResponse(data itFile.UploadFileResultData) UploadFileResponse {
	return UploadFileResponse{
		RestCreateResponse: *httpserver.NewRestCreateResponseDyn(data.Created),
		FilePath:           data.FilePath,
	}
}

// readUploadForm walks the parts in order and stops at the file part, leaving its body unread for
// the service to stream. The declared name and media type fall back to the file part's own.
func readUploadForm(request *http.Request) (*itFile.UploadFileCommand, ft.ClientErrors, error) {
	reader, err := request.MultipartReader()
	if err != nil {
		return nil, malformed("the request must be multipart/form-data"), nil
	}
	texts := map[string]string{}
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			return nil, ft.ClientErrors{*ft.NewValidationError(c.FieldFilePart,
				ft.ErrorKey("err_missing_required_field"), "the file part is required")}, nil
		}
		if err != nil {
			return nil, malformed("the multipart body is malformed"), nil
		}
		if part.FormName() == c.FieldFilePart {
			return buildUploadCommand(texts, part)
		}
		value, err := io.ReadAll(io.LimitReader(part, maxTextPartBytes+1))
		if err != nil {
			return nil, nil, errors.Wrap(err, "read upload form part")
		}
		if len(value) > maxTextPartBytes {
			return nil, ft.ClientErrors{*ft.NewValidationError(part.FormName(),
				ft.ErrorKey("err_invalid_value"), "the form value is too long")}, nil
		}
		texts[part.FormName()] = strings.TrimSpace(string(value))
	}
}

func buildUploadCommand(texts map[string]string, part *multipart.Part) (*itFile.UploadFileCommand, ft.ClientErrors, error) {
	cmd := &itFile.UploadFileCommand{
		OrgId:    optionalId(texts[partOrgId]),
		SpaceId:  optionalId(texts[partSpaceId]),
		Name:     firstNonEmpty(texts[partName], part.FileName()),
		MimeType: firstNonEmpty(texts[partMimeType], part.Header.Get("Content-Type")),
		Body:     part,
	}
	cErrs := ft.ClientErrors{}
	for field, value := range map[string]string{partName: cmd.Name, partMimeType: cmd.MimeType} {
		if value == "" {
			cErrs.Append(*dmodel.NewMissingFieldErr(field))
		}
	}
	size, err := strconv.ParseInt(texts[partSizeBytes], 10, 64)
	if err != nil {
		// Without a declared size the quota cannot be checked before the stream starts.
		cErrs.Append(*ft.NewValidationError(partSizeBytes, ft.ErrorKey("err_invalid_value"),
			"size_bytes must be the file's length in bytes, sent before the file part"))
	}
	cmd.SizeBytes = size
	return cmd, cErrs, nil
}

func optionalId(raw string) *model.Id {
	if raw == "" {
		return nil
	}
	id := model.Id(raw)
	return &id
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func malformed(message string) ft.ClientErrors {
	return ft.ClientErrors{*ft.NewAnonymousValidationError(ft.ErrorKey("err_malformed_request"), message)}
}

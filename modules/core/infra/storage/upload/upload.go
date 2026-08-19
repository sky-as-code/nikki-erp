package upload

import (
	"bytes"
	"io"
	"mime/multipart"
	"slices"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
)

// PrepareParam describes one inbound file and the rules it must satisfy.
type PrepareParam struct {
	// FieldName is the request field name reported in validation errors.
	FieldName string

	FileHeader *multipart.FileHeader

	// MaxSize is the maximum accepted size in bytes. Zero means no limit.
	MaxSize int64

	// AllowedMimes restricts the sniffed MIME type. Empty means no restriction.
	AllowedMimes *[]string
}

// PreparedFile carries everything filestorage.Put needs, and owns the file
// opened by Prepare. The caller must always Close it, including on success.
type PreparedFile struct {
	Reader   io.Reader
	MimeType string
	Size     int64

	file multipart.File
}

func (this *PreparedFile) Close() error {
	if this == nil || this.file == nil {
		return nil
	}

	return this.file.Close()
}

// Prepare opens the file, sniffs its MIME type from the content, and checks it
// against param. Rule violations are appended to cErrs and make Prepare return
// a nil PreparedFile, so a caller cannot upload a file that failed validation.
// Prepare closes the file itself whenever it returns a nil PreparedFile.
func Prepare(param *PrepareParam, cErrs *ft.ClientErrors) (prepared *PreparedFile, err error) {
	if param == nil || param.FileHeader == nil {
		return nil, errors.New("file header is required to prepare an upload")
	}

	if cErrs == nil {
		return nil, errors.New("client errors collector is required to prepare an upload")
	}

	file, openErr := param.FileHeader.Open()
	if openErr != nil {
		return nil, errors.WithStack(openErr)
	}

	defer func() {
		if prepared == nil {
			file.Close()
		}
	}()

	mimeType, sniffErr := sniff(file)
	if sniffErr != nil {
		return nil, sniffErr
	}

	appendRuleErrors(param, mimeType, cErrs)
	if cErrs.Count() > 0 {
		return nil, nil
	}

	// multipart.File always implements io.Seeker, so a failure here is a real
	// I/O error: returning an unrewound reader would silently upload a file
	// truncated by the bytes sniff() consumed.
	if _, seekErr := file.Seek(0, io.SeekStart); seekErr != nil {
		return nil, seekErr
	}

	return &PreparedFile{
		Reader:   file,
		MimeType: mimeType,
		Size:     param.FileHeader.Size,
		file:     file,
	}, nil
}

func matchMime(allowed, actual string) bool {
	if allowed == "*/*" {
		return true
	}

	allowedType, allowedSubType, ok := strings.Cut(allowed, "/")
	if !ok {
		return false
	}

	actualType, actualSubType, ok := strings.Cut(actual, "/")
	if !ok {
		return false
	}

	return (allowedType == "*" || allowedType == actualType) &&
		(allowedSubType == "*" || allowedSubType == actualSubType)
}

func appendRuleErrors(param *PrepareParam, mimeType string, cErrs *ft.ClientErrors) {
	if param.AllowedMimes != nil && len(*param.AllowedMimes) > 0 {
		allowed := slices.ContainsFunc(*param.AllowedMimes, func(mime string) bool {
			return matchMime(mime, mimeType)
		})

		if !allowed {
			cErrs.Append(*ft.NewValidationError(
				param.FieldName,
				ft.ErrorKey("err_file_type_not_allowed"),
				"file type {{actual}} is not one of {{allowed}}",
				map[string]any{
					"allowed": *param.AllowedMimes,
					"actual":  mimeType,
				},
			))
		}
	}
	if param.MaxSize > 0 && param.FileHeader.Size > param.MaxSize {
		cErrs.Append(*ft.NewValidationError(param.FieldName,
			ft.ErrorKey("err_file_too_large"),
			"file size {{actual_size}} exceeds the maximum of {{max_size}} bytes",
			map[string]any{"max_size": param.MaxSize, "actual_size": param.FileHeader.Size}))
	}
}

func sniff(file multipart.File) (string, error) {
	head := make([]byte, 512)
	n, readErr := io.ReadFull(file, head)
	if readErr != nil && readErr != io.ErrUnexpectedEOF && readErr != io.EOF {
		return "", errors.WithStack(readErr)
	}

	mt := "application/octet-stream"
	if d := mimetype.Detect(head[:n]); d != nil {
		mt = d.String()
	}

	return mt, nil
}

// Deprecated: use upload.Prepare instead
func SniffContentTypeAndRewind(
	file multipart.File, header *multipart.FileHeader,
) (contentType string, reader io.Reader, err error) {
	if header != nil {
		if ct := header.Header.Get("Content-Type"); ct != "" && ct != "application/octet-stream" {
			if seeker, ok := file.(io.Seeker); ok {
				_, err = seeker.Seek(0, io.SeekStart)
				if err != nil {
					return "", nil, err
				}
			}
			return ct, file, nil
		}
	}

	head := make([]byte, 512)
	n, readErr := io.ReadFull(file, head)
	if readErr != nil && readErr != io.ErrUnexpectedEOF && readErr != io.EOF {
		return "", nil, readErr
	}

	mt := "application/octet-stream"
	if d := mimetype.Detect(head[:n]); d != nil {
		mt = d.String()
	}

	if seeker, ok := file.(io.Seeker); ok {
		_, err = seeker.Seek(0, io.SeekStart)
		if err != nil {
			return "", nil, err
		}
		return mt, file, nil
	}

	return mt, io.MultiReader(bytes.NewReader(head[:n]), file), nil
}

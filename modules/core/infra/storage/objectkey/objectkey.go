package objectkey

import (
	"mime/multipart"
	"path"
	"strings"

	"github.com/sky-as-code/nikki-erp/common/model"
)

const defaultBlobName = "blob"

func SanitizeFilename(raw string) string {
	name := defaultBlobName
	if raw != "" {
		name = path.Base(strings.ReplaceAll(raw, "\\", "/"))
	}
	if name == "." || name == "/" {
		name = defaultBlobName
	}
	return name
}

func SanitizeFileHeaderName(header *multipart.FileHeader) string {
	if header == nil || header.Filename == "" {
		return defaultBlobName
	}
	return SanitizeFilename(header.Filename)
}

func Build(prefix, filename string) (string, error) {
	uuid, err := model.NewUUID()
	if err != nil {
		return "", err
	}
	return path.Join(prefix, *uuid, SanitizeFilename(filename)), nil
}

func BuildFromFileHeader(prefix string, header *multipart.FileHeader) (string, error) {
	return Build(prefix, SanitizeFileHeaderName(header))
}

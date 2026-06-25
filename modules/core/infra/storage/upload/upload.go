package upload

import (
	"bytes"
	"io"
	"mime/multipart"

	"github.com/gabriel-vasile/mimetype"
)

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

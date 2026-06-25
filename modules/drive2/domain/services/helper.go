package services

import (
	"mime/multipart"
	"net/http"

	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/database"
)

func extractMIME(file multipart.File) string {
	if file != nil {
		buffer := make([]byte, 512)
		_, err := file.Read(buffer)
		if err != nil {
			panic(err)
		}

		MIME := http.DetectContentType(buffer)
		file.Seek(0, 0)

		return MIME
	}

	return ""
}

func resolveTransaction(
	tx database.DbTransaction,
	err *error,
	clientErr fault.ClientErrors,
	shouldRollback ...bool) error {
	if err != nil ||
		len(clientErr) > 0 {
		return tx.Rollback()
	}

	for _, item := range shouldRollback {
		if item {
			return tx.Rollback()
		}
	}

	return tx.Commit()
}

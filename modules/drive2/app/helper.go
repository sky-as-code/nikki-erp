package app

import (
	"mime/multipart"
	"net/http"

	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/database"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
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

func extractOpData[T any](clientErrs *fault.ClientErrors, fn func() (*dynamicmodel.OpResult[T], error), field ...string) (T, error, bool) {
	var zero T
	if clientErrs == nil {
		clientErrs = fault.NewClientErrors()
	}

	concreteRes, err := fn()
	if err != nil {
		return zero, err, false
	}

	if concreteRes.ClientErrors.Count() > 0 {
		*clientErrs = concreteRes.ClientErrors
		return zero, nil, false
	}

	if !concreteRes.HasData {
		if len(field) > 0 {
			clientErrs.Append(*fault.NewNotFoundError(field[0]))
		} else {
			clientErrs.Append(*fault.NewAnonymousNotFoundError())
		}
		return zero, nil, false
	}

	return concreteRes.Data, nil, true
}

func resolveOpResult[T any](fn func() (*dynamicmodel.OpResult[T], error), field ...string) (*dynamicmodel.OpResult[T], error, bool) {
	concreteRes, err := fn()
	if err != nil {
		return nil, err, false
	}

	if concreteRes.ClientErrors.Count() > 0 {
		return concreteRes, nil, false
	}

	if !concreteRes.HasData {
		return concreteRes, nil, false
	}

	return concreteRes, nil, true
}

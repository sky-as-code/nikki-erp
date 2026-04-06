package domain

import ft "github.com/sky-as-code/nikki-erp/common/fault"

func ValidationErrorsToClientErrors(vErrs ft.ValidationErrors) ft.ClientErrors {
	if vErrs.Count() == 0 {
		return nil
	}

	result := make(ft.ClientErrors, 0, vErrs.Count())
	for field, message := range vErrs {
		result = append(result, *ft.NewValidationError(field, "", message))
	}
	return result
}

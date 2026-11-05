package services

import ft "github.com/sky-as-code/nikki-erp/common/fault"

func hasReasonKey(vErrs *ft.ClientErrors, key string) bool {
	if vErrs == nil {
		return false
	}
	for _, item := range *vErrs {
		if item.Key == ft.ErrorKey(key, "sales") {
			return true
		}
	}
	return false
}

package domain

import (
	"github.com/sky-as-code/nikki-erp/common/model"
)

type UserSummary struct {
	Id          model.Id `json:"id,omitempty"`
	DisplayName *string `json:"displayName,omitempty"`
}

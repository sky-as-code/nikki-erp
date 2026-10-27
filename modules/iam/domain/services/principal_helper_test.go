package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
)

func TestUserTypeOf(t *testing.T) {
	assert.Equal(t, models.UserTypeUser, userTypeOf(nil), "a principal that is not an IAM user is a person")
	assert.Equal(t, models.UserTypeUser, userTypeOf(&loginPrincipal{}))
	assert.Equal(t, models.UserTypeService, userTypeOf(&loginPrincipal{Type: models.UserTypeService}))
}

package v1

import (
	it "github.com/sky-as-code/nikki-erp/modules/core/interfaces/user"
	"github.com/sky-as-code/nikki-erp/modules/core/transport/restful"
)

type CreateUserRequest = it.CreateUserCommand

type CreateUserResponse restful.RestResponse[it.CreateUserResult]

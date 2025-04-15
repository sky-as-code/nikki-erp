package user

import "github.com/ThreeDotsLabs/watermill"

type CreateUserCommand struct {
	ID                 string `watermill:"command_id"`
	Username           string
	Email              string
	DisplayName        string
	Password           string
	AvatarURL          string
	Status             string
	MustChangePassword bool
	CreatedBy          string
}

type UpdateUserCommand struct {
	ID                 string `watermill:"command_id"`
	DisplayName        string
	AvatarURL          string
	Status             string
	MustChangePassword bool
	UpdatedBy          string
}

type DeleteUserCommand struct {
	ID        string `watermill:"command_id"`
	DeletedBy string
}

type GetUserByIDQuery struct {
	ID string `watermill:"command_id"`
}

type GetUserByUsernameQuery struct {
	Username string `watermill:"command_id"`
}

type GetUserByEmailQuery struct {
	Email string `watermill:"command_id"`
}

func NewEventID() string {
	return watermill.NewUUID()
}

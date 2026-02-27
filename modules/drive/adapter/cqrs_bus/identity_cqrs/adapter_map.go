package identity_cqrs

import (
	"errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	identityclient "github.com/sky-as-code/nikki-erp/modules/drive/adapter/cqrs_bus/identity_cqrs/client"
)

var errNilUserExistsClientResult = errors.New("identity_cqrs: user exists client returned nil result")

// userExistsMapped is the drive-side view produced from the identity client response (anti-corruption boundary).
type userExistsMapped struct {
	exists      bool
	clientError *ft.ClientError
}

func mapUserExistsFromClient(res *identityclient.UserExistsResult) (userExistsMapped, error) {
	if res == nil {
		return userExistsMapped{}, errNilUserExistsClientResult
	}
	if res.ClientError != nil {
		return userExistsMapped{clientError: res.ClientError}, nil
	}
	return userExistsMapped{exists: res.HasData && res.Data}, nil
}

var errNilSearchUsersResultClientResult = errors.New("identity_cqrs: search users client returned nil result")

type userSearchUsersMapped struct {
	users       map[model.Id]*UserSummary
	clientError *ft.ClientError
}

type userGetUserByIdMapped struct {
	user        *UserSummary
	clientError *ft.ClientError
}

type UserSummary struct {
	Id          model.Id `json:"id"`
	DisplayName *string  `json:"displayName,omitempty"`
	Email       *string  `json:"email,omitempty"`
	AvatarUrl   *string  `json:"avatarUrl,omitempty"`
}

func mapUserSummaryFromClientUserEntity(entity *identityclient.UserEntity) (*UserSummary, error) {
	if entity == nil {
		return nil, nil
	}

	if entity.Id == nil {
		return nil, nil
	}

	return &UserSummary{
		Id:          *entity.Id,
		DisplayName: entity.DisplayName,
		Email:       entity.Email,
		AvatarUrl:   entity.AvatarUrl,
	}, nil
}

func mapUsersFromClient(res *identityclient.SearchUsersResult) (userSearchUsersMapped, error) {
	if res == nil {
		return userSearchUsersMapped{}, errNilSearchUsersResultClientResult
	}

	if res.ClientError != nil {
		return userSearchUsersMapped{clientError: res.ClientError}, nil
	}

	users := map[model.Id]*UserSummary{}
	if res.Data == nil || len(res.Data.Items) == 0 {
		return userSearchUsersMapped{users: users}, nil
	}

	for i := range res.Data.Items {
		user, mapErr := mapUserSummaryFromClientUserEntity(&res.Data.Items[i])
		if mapErr != nil {
			return userSearchUsersMapped{}, mapErr
		}
		if user == nil {
			continue
		}
		users[user.Id] = user
	}

	return userSearchUsersMapped{users: users}, nil
}

func mapUserByIdFromClient(res *identityclient.SearchUsersResult) (*UserSummary, *ft.ClientError, error) {
	mapped, err := mapUsersFromClient(res)
	if err != nil {
		return nil, nil, err
	}
	if mapped.clientError != nil {
		return nil, mapped.clientError, nil
	}

	// If backend returns 0 items -> not found => nil user.
	// If it returns multiple (shouldn't happen for "= id"), pick the first deterministic one.
	for _, u := range mapped.users {
		return u, nil, nil
	}
	return nil, nil, nil
}

func mapUserByIdFromGetUserByIdClient(res *identityclient.GetUserByIdResult) (userGetUserByIdMapped, error) {
	if res == nil {
		return userGetUserByIdMapped{}, errors.New("identity_cqrs: getUserById client returned nil result")
	}

	if res.ClientError != nil {
		return userGetUserByIdMapped{clientError: res.ClientError}, nil
	}

	if res.Data == nil {
		return userGetUserByIdMapped{user: nil}, nil
	}

	user, mapErr := mapUserSummaryFromClientUserEntity(res.Data)
	if mapErr != nil {
		return userGetUserByIdMapped{}, mapErr
	}

	return userGetUserByIdMapped{user: user}, nil
}

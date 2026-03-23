package identity_cqrs

import (
	"context"
	"fmt"
	"strings"

	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	identityclient "github.com/sky-as-code/nikki-erp/modules/drive/adapter/cqrs_bus/identity_cqrs/client"
)

// IdentityCqrsAdapter is the drive-facing facade for cross-module calls (maps to drive use cases).
type IdentityCqrsAdapter interface {
	UserExists(ctx context.Context, userId model.Id) (bool, error, *fault.ClientError)
	GetUserById(ctx context.Context, userId model.Id) (*UserSummary, error, *fault.ClientError)
	GetUsersByIds(ctx context.Context, userIds []model.Id) (map[model.Id]*UserSummary, error, *fault.ClientError)
}

func NewIdentityCqrsAdapter(c identityclient.IdentityCqrsClient) IdentityCqrsAdapter {
	return &identityCqrsAdapter{client: c}
}

type identityCqrsAdapter struct {
	client identityclient.IdentityCqrsClient
}

func (this *identityCqrsAdapter) UserExists(ctx context.Context, userId model.Id) (bool, error, *fault.ClientError) {
	res, err := this.client.UserExists(ctx, identityclient.UserExistsQuery{Id: userId})
	if err != nil {
		return false, err, nil
	}

	mapped, mapErr := mapUserExistsFromClient(res)
	if mapErr != nil {
		return false, mapErr, nil
	}
	if mapped.clientError != nil {
		return false, nil, mapped.clientError
	}

	return mapped.exists, nil, nil
}

func (this *identityCqrsAdapter) GetUserById(ctx context.Context, userId model.Id) (*UserSummary, error, *fault.ClientError) {
	res, err := this.client.GetUserById(ctx, identityclient.GetUserByIdQuery{
		Id: userId,
		// Keep all expansions off to minimize response payload.
		WithGroup:     false,
		WithHierarchy: false,
		WithOrg:       false,
		ScopeRef:      nil,
		Status:        nil,
	})
	if err != nil {
		return nil, err, nil
	}

	mapped, mapErr := mapUserByIdFromGetUserByIdClient(res)
	if mapErr != nil {
		return nil, mapErr, nil
	}
	if mapped.clientError != nil {
		return nil, nil, mapped.clientError
	}

	return mapped.user, nil, nil
}

func (this *identityCqrsAdapter) GetUsersByIds(
	ctx context.Context,
	userIds []model.Id,
) (map[model.Id]*UserSummary, error, *fault.ClientError) {
	if len(userIds) == 0 {
		return map[model.Id]*UserSummary{}, nil, nil
	}

	// crud.SearchQuery.Size is validated by common/model rules (max 500).
	// We chunk to avoid invalid request payload and keep one request per chunk.
	maxSize := model.MODEL_RULE_PAGE_MAX_SIZE
	page := model.MODEL_RULE_PAGE_INDEX_START

	result := make(map[model.Id]*UserSummary, len(userIds))

	for start := 0; start < len(userIds); start += maxSize {
		end := start + maxSize
		if end > len(userIds) {
			end = len(userIds)
		}

		chunk := userIds[start:end]

		quotedIds := make([]string, 0, len(chunk))
		for _, id := range chunk {
			quotedIds = append(quotedIds, fmt.Sprintf("%q", id))
		}

		// graph = {"if":["id","in",<id1>,<id2>,...]}
		graph := fmt.Sprintf("{\"if\":[\"id\",\"in\",%s]}", strings.Join(quotedIds, ","))
		size := len(chunk)

		query := identityclient.SearchUsersQuery{
			WithGroups:    false,
			WithOrgs:      false,
			WithHierarchy: false,
			ScopeRef:      nil,
		}
		// Promoted fields from embedded crud.SearchQuery
		query.Page = &page
		query.Size = &size
		query.Graph = &graph

		res, err := this.client.SearchUsers(ctx, query)
		if err != nil {
			return nil, err, nil
		}

		mapped, mapErr := mapUsersFromClient(res)
		if mapErr != nil {
			return nil, mapErr, nil
		}
		if mapped.clientError != nil {
			return nil, nil, mapped.clientError
		}

		for id, user := range mapped.users {
			result[id] = user
		}
	}

	return result, nil, nil
}

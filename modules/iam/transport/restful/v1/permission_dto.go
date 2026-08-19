package v1

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	it "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/permission"
)

// TestMyPermissionsRequest carries only the expression under test. There is no
// user field: the subject is always the caller, so there is no parameter an
// attacker could point at somebody else.
type TestMyPermissionsRequest = it.TestMyPermissionsQuery

type PermissionMatchResponse struct {
	SourceKind    string   `json:"source_kind"`
	SourceId      model.Id `json:"source_id"`
	SourceName    string   `json:"source_name"`
	EntExpression string   `json:"ent_expression"`
}

type TestMyPermissionsResponse struct {
	IsGranted bool `json:"is_granted"`
	// Always a list, never null - a denial is an empty list, so a client can read
	// the field without a nil check and learns nothing extra from a refusal.
	Matches []PermissionMatchResponse `json:"matches"`
}

func NewTestMyPermissionsResponse(data it.TestMyPermissionsResultData) TestMyPermissionsResponse {
	matches := array.Map(data.Matches, func(match it.PermissionMatch) PermissionMatchResponse {
		return PermissionMatchResponse{
			SourceKind:    match.SourceKind,
			SourceId:      match.SourceId,
			SourceName:    match.SourceName,
			EntExpression: match.EntExpression,
		}
	})
	if matches == nil {
		matches = []PermissionMatchResponse{}
	}
	return TestMyPermissionsResponse{
		IsGranted: data.IsGranted,
		Matches:   matches,
	}
}

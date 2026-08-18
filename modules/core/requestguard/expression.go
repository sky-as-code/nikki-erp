package requestguard

import (
	"fmt"
	"strings"

	"github.com/sky-as-code/nikki-erp/common/model"
)

// Wildcard is the token that matches any action or any resource in an
// entitlement expression.
const Wildcard = "*"

// maxSegmentLength bounds each segment of an expression.
//
// Action and resource codes are short identifiers and scope ids are ULIDs, so
// nothing legitimate comes close to this. The bound exists because the parser is
// fed untrusted request bodies whose segments end up as query parameters: without
// it a caller could hand over a megabyte-long "resource code" that no real
// resource could ever have, and it would still be matched against the wildcard
// grants and carried into the database.
const maxSegmentLength = 128

// An entitlement expression has the shape `{action}:{resource}:{scope}[/{scopeId}]`
// where action and resource may be the wildcard `*`, and the scope segment carries
// an optional id for the org / orgunit scopes.
//
// BuildExpression is the ONLY producer of expressions in the codebase: the
// entitlement model persists what it returns, the guard matches against what it
// returns, and the SQL matcher builds its IN-list from it. Any second
// implementation is a defect - the two sides silently drift apart (see plan D1/D4).
func BuildExpression(actionCode string, resourceCode string, scope ResourceScope, scopeId *model.Id) string {
	action := orWildcard(actionCode)
	resource := orWildcard(resourceCode)
	if scopeId != nil && *scopeId != "" {
		return fmt.Sprintf("%s:%s:%s/%s", action, resource, string(scope), string(*scopeId))
	}
	return fmt.Sprintf("%s:%s:%s", action, resource, string(scope))
}

func orWildcard(code string) string {
	if code == "" {
		return Wildcard
	}
	return code
}

// OmnipotentExpression grants everything, everywhere. It is the only expression
// whose scope segment is a wildcard.
func OmnipotentExpression() string {
	return Wildcard + ":" + Wildcard + ":" + Wildcard
}

// ParsedExpression is the decomposition of an entitlement expression.
type ParsedExpression struct {
	ActionCode   string
	ResourceCode string
	Scope        ResourceScope
	ScopeId      *model.Id
}

// HasWildcard reports whether the expression grants across all actions or all
// resources. A *question* asked of the permission probe must never have one: a
// real requirement is always concrete, and a wildcard question would turn the
// probe into a grant-enumeration tool.
func (this ParsedExpression) HasWildcard() bool {
	return this.ActionCode == Wildcard || this.ResourceCode == Wildcard || string(this.Scope) == Wildcard
}

// String rebuilds the canonical expression, so that parsing then rebuilding is
// the identity for every valid input.
func (this ParsedExpression) String() string {
	if string(this.Scope) == Wildcard {
		return OmnipotentExpression()
	}
	return BuildExpression(this.ActionCode, this.ResourceCode, this.Scope, this.ScopeId)
}

// ParseExpression decomposes an entitlement expression. It is strict on purpose:
// callers hand it untrusted input (the permission probe accepts an expression in
// a request body), so every malformed shape must come back as a plain error the
// transport turns into a 400, never a panic.
func ParseExpression(expr string) (*ParsedExpression, error) {
	segments := strings.Split(expr, ":")
	if len(segments) != 3 {
		return nil, fmt.Errorf("expression must have exactly 3 colon-separated segments, got %d", len(segments))
	}

	action, resource, scopeSegment := segments[0], segments[1], segments[2]
	if action == "" || resource == "" || scopeSegment == "" {
		return nil, fmt.Errorf("expression segments must not be empty")
	}
	for _, segment := range segments {
		if len(segment) > maxSegmentLength {
			return nil, fmt.Errorf("expression segment exceeds %d characters", maxSegmentLength)
		}
	}

	if scopeSegment == Wildcard {
		if action != Wildcard || resource != Wildcard {
			return nil, fmt.Errorf("a wildcard scope is only valid in the omnipotent expression %q", OmnipotentExpression())
		}
		return &ParsedExpression{ActionCode: Wildcard, ResourceCode: Wildcard, Scope: ResourceScope(Wildcard)}, nil
	}

	scopeName, scopeId, err := splitScopeSegment(scopeSegment)
	if err != nil {
		return nil, err
	}

	return &ParsedExpression{
		ActionCode:   action,
		ResourceCode: resource,
		Scope:        scopeName,
		ScopeId:      scopeId,
	}, nil
}

func splitScopeSegment(segment string) (ResourceScope, *model.Id, error) {
	name, rawId, hasId := strings.Cut(segment, "/")
	scope := ResourceScope(name)
	if !IsKnownScope(scope) {
		return "", nil, fmt.Errorf("unknown scope %q", name)
	}
	if !hasId {
		return scope, nil, nil
	}
	if rawId == "" {
		return "", nil, fmt.Errorf("scope %q carries an empty id", name)
	}
	if strings.Contains(rawId, "/") {
		return "", nil, fmt.Errorf("scope %q carries a malformed id", name)
	}
	if scope != ResourceScopeOrg && scope != ResourceScopeOrgUnit {
		return "", nil, fmt.Errorf("scope %q must not carry an id", name)
	}
	id := model.Id(rawId)
	return scope, &id, nil
}

// IsKnownScope reports whether the scope is one this system evaluates.
func IsKnownScope(scope ResourceScope) bool {
	switch scope {
	case ResourceScopeDomain, ResourceScopeOrg, ResourceScopeOrgUnit, ResourceScopePrivate:
		return true
	}
	return false
}

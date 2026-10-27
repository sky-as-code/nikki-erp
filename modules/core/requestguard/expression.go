package requestguard

import (
	"fmt"
	"regexp"
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

// FieldSeparator splits a field-level resource segment `{resource}.{field}`.
const FieldSeparator = "."

// fieldNamePattern is the shape of a field name in a field-level expression. It matches the snake
// lower case every schema field name follows, and rules out the wildcard and a second separator.
var fieldNamePattern = regexp.MustCompile(`^[a-z0-9_]+$`)

// An entitlement expression has the shape `{action}:{resource}[.{field}]:{scope}[/{scopeId}]`
// where action and resource may be the wildcard `*`, the optional field narrows the grant to one
// field of the resource, and the scope segment carries an optional id for the org / orgunit scopes.
//
// BuildExpression and BuildFieldExpression are the ONLY producers of expressions in the codebase:
// the entitlement model persists what they return, the guard matches against what they return,
// and the SQL matcher builds its IN-list from it. Any second implementation is a defect - the two
// sides silently drift apart (see plan D1/D4).
func BuildExpression(actionCode string, resourceCode string, scope ResourceScope, scopeId *model.Id) string {
	return BuildFieldExpression(actionCode, resourceCode, "", scope, scopeId)
}

// BuildFieldExpression is BuildExpression for a field-level grant. An empty fieldName builds the
// resource-level expression. A field-level expression always names its resource: a field of "any
// resource" means nothing, so the caller must not pass a wildcard resource with a field.
func BuildFieldExpression(
	actionCode string, resourceCode string, fieldName string, scope ResourceScope, scopeId *model.Id,
) string {
	action := orWildcard(actionCode)
	resource := ResourceSegment(orWildcard(resourceCode), fieldName)
	if scopeId != nil && *scopeId != "" {
		return fmt.Sprintf("%s:%s:%s/%s", action, resource, string(scope), string(*scopeId))
	}
	return fmt.Sprintf("%s:%s:%s", action, resource, string(scope))
}

// ResourceSegment is the middle segment of an expression: the resource code, followed by the
// field name when there is one.
func ResourceSegment(resourceCode string, fieldName string) string {
	if fieldName == "" {
		return resourceCode
	}
	return resourceCode + FieldSeparator + fieldName
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
	// FieldName is empty for a resource-level expression.
	FieldName string
	Scope     ResourceScope
	ScopeId   *model.Id
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
	return BuildFieldExpression(this.ActionCode, this.ResourceCode, this.FieldName, this.Scope, this.ScopeId)
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

	resourceCode, fieldName, err := splitResourceSegment(resource)
	if err != nil {
		return nil, err
	}
	scopeName, scopeId, err := splitScopeSegment(scopeSegment)
	if err != nil {
		return nil, err
	}

	return &ParsedExpression{
		ActionCode:   action,
		ResourceCode: resourceCode,
		FieldName:    fieldName,
		Scope:        scopeName,
		ScopeId:      scopeId,
	}, nil
}

func splitResourceSegment(segment string) (string, string, error) {
	resourceCode, fieldName, hasField := strings.Cut(segment, FieldSeparator)
	if !hasField {
		return segment, "", nil
	}
	if resourceCode == "" || resourceCode == Wildcard {
		return "", "", fmt.Errorf("a field-level expression must name its resource")
	}
	if !fieldNamePattern.MatchString(fieldName) {
		return "", "", fmt.Errorf("field %q is not a valid field name", fieldName)
	}
	return resourceCode, fieldName, nil
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
	case ResourceScopeTenant, ResourceScopeOrg, ResourceScopeOrgUnit, ResourceScopePrivate:
		return true
	}
	return false
}

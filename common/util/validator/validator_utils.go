package validator

import (
	"fmt"
	"regexp"
)

const urn = `arn|grn`
const partition = `gjc|aws|aws-cn|aws-us-gov`
const lowAlphaNumDash = `([a-z0-9]-?)*[a-z0-9]`
const service = `([a-z0-9]-?)*[a-z0-9]` // Valid: a, a-b, a1-b2 ; Invalid: -a, b-, aA, -
const region = `([a-z0-9]-?)*[a-z0-9]`  // Valid: a, a-b, a1-b2 ; Invalid: -a, b-, aA, -
const account = `([a-z0-9]-?)*[a-z0-9]` // Valid: a, a-b, a1-b2 ; Invalid: -a, b-, aA, -
const resource = `.*\S`

// Some part allows single `*`
var grnPattern = fmt.Sprintf(`((%s):(%s):\*|(%s):\*|(%s):\*|(%s):\*|(%s))|\*`, urn, partition, service, region, account, resource)
var grnRegexp = regexp.MustCompile(grnPattern)

// IsValidGrn checks if the given string is a valid ARN
func IsValidGrn(grn string) bool {
	return grnRegexp.MatchString(grn)
}

// Valid:
//
//	bla:bla
//	234%@##$:bl@#$
//
// Invalid:
//
//	:
//	bla:bla blah
var actionRegexp = regexp.MustCompile("(\\S+:\\S+)|\\*")

// IsValidActionPattern checks if the given string is a valid action for IAM policy
func IsValidActionPattern(action string) bool {
	return actionRegexp.MatchString(action)
}

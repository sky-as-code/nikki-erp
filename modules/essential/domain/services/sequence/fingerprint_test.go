package sequence

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func baseFingerprintInput() FingerprintInput {
	return FingerprintInput{
		DefinitionId:  "01JDEFINITION0000000000000",
		TenantId:      "01JTENANT000000000000000000",
		OrgId:         "01JORG000000000000000000000",
		OwnerModule:   "commerce",
		CallerModule:  "commerce",
		ScopeObjectId: "",
		OperationKey:  "default",
	}
}

func TestFingerprintStable(t *testing.T) {
	first := Fingerprint(baseFingerprintInput())
	assert.Len(t, first, 64)
	assert.Equal(t, first, Fingerprint(baseFingerprintInput()), "a retry with the server default is identical")

	instant := time.Date(2026, 11, 3, 10, 0, 0, 0, time.UTC)
	location, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	sameInstant := instant.In(location)
	withUtc, withLocal := baseFingerprintInput(), baseFingerprintInput()
	withUtc.EffectiveAt, withLocal.EffectiveAt = &instant, &sameInstant
	assert.Equal(t, Fingerprint(withUtc), Fingerprint(withLocal), "same instant in another zone")
	assert.NotEqual(t, first, Fingerprint(withUtc), "explicit effective_at differs from SERVER_DEFAULT")
}

func TestFingerprintEveryFieldMatters(t *testing.T) {
	base := Fingerprint(baseFingerprintInput())
	later := time.Date(2026, 11, 3, 10, 0, 1, 0, time.UTC)
	mutations := map[string]func(*FingerprintInput){
		"definition": func(in *FingerprintInput) { in.DefinitionId += "X" },
		"tenant":     func(in *FingerprintInput) { in.TenantId += "X" },
		"org":        func(in *FingerprintInput) { in.OrgId = "" },
		"owner":      func(in *FingerprintInput) { in.OwnerModule = "sales" },
		"caller":     func(in *FingerprintInput) { in.CallerModule = "sales" },
		"scope":      func(in *FingerprintInput) { in.ScopeObjectId = "01JSCOPE" },
		"operation":  func(in *FingerprintInput) { in.OperationKey = "line:01J" },
		"effective":  func(in *FingerprintInput) { in.EffectiveAt = &later },
	}
	seen := map[string]string{base: "base"}
	for name, mutate := range mutations {
		input := baseFingerprintInput()
		mutate(&input)
		value := Fingerprint(input)
		previous, duplicate := seen[value]
		assert.False(t, duplicate, "%s collides with %s", name, previous)
		seen[value] = name
	}
}

func TestFingerprintDelimiterSafe(t *testing.T) {
	left, right := baseFingerprintInput(), baseFingerprintInput()
	left.OwnerModule, left.CallerModule = "ab", "c"
	right.OwnerModule, right.CallerModule = "a", "bc"
	assert.NotEqual(t, Fingerprint(left), Fingerprint(right))

	marker := baseFingerprintInput()
	marker.OperationKey = "default;14:SERVER_DEFAULT"
	assert.NotEqual(t, Fingerprint(baseFingerprintInput()), Fingerprint(marker))
}

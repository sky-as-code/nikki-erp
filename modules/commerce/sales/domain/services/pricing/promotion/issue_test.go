package promotion

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func nextOrderProgram(id string) Program {
	issue := applyRule("issue-1", "0", "200000")
	issue.Phase = PhaseIssue
	program := automatic(id, 10, []Rule{issue, applyRule("apply-1", "0", "0")}, fixedReward("rw", ScopeOrder, "20000"))
	program.ProgramType = ProgramNextOrderCoupon
	return program
}

func TestIssuePhase(t *testing.T) {
	cases := []struct {
		name      string
		customer  string
		amount    string
		qualified bool
		reason    string
	}{
		{name: "AC-31 qualifying source order with a customer", customer: "cus-1", amount: "250000", qualified: true},
		{name: "AC-30 qualifying but no customer", customer: "", amount: "250000", reason: ReasonCustomerRequired},
		{name: "issue rule not met", customer: "cus-1", amount: "100000", reason: ReasonConditionNotMet},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			input := orderOf([]Line{buyLine("l1", 1, "1", testCase.amount)}, nextOrderProgram("nx"))
			input.CustomerId = testCase.customer
			decisions := EvaluateIssues(input)
			require.Len(t, decisions, 1)
			require.Equal(t, testCase.qualified, decisions[0].Qualified)
			require.Equal(t, testCase.reason, decisions[0].ReasonCode)

			// AC-31: the source order never receives the program's own reward.
			applied := programResult(t, Apply(input), "nx")
			require.Equal(t, OutcomeNotEligible, applied.Outcome)
			require.Equal(t, ReasonVoucherRequired, applied.ReasonCode)
		})
	}
}

func TestIssueOnlyForTypesWithAnIssuer(t *testing.T) {
	input := orderOf([]Line{buyLine("l1", 1, "1", "1000")},
		automatic("auto", 10, anyRule(), fixedReward("rw", ScopeOrder, "1")), nextOrderProgram("nx"))
	input.CustomerId = "cus-1"
	decisions := EvaluateIssues(input)
	require.Len(t, decisions, 1)
	require.Equal(t, "nx", decisions[0].PromotionId)
}

func TestIssueValidity(t *testing.T) {
	confirmedAt := time.Date(2026, 11, 4, 10, 0, 0, 0, time.UTC)
	days := func(value int32) *int32 { return &value }
	at := func(value time.Time) *time.Time { return &value }
	cases := []struct {
		name    string
		days    *int32
		endsAt  *time.Time
		expires *time.Time
		ok      bool
	}{
		{name: "days only", days: days(30), expires: at(confirmedAt.AddDate(0, 0, 30)), ok: true},
		{name: "promotion ends first", days: days(30), endsAt: at(confirmedAt.AddDate(0, 0, 7)),
			expires: at(confirmedAt.AddDate(0, 0, 7)), ok: true},
		{name: "both null: no own expiry", ok: true},
		{name: "promotion already ended: nothing issued", days: days(30), endsAt: at(confirmedAt)},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			validFrom, expiresAt, ok, err := IssueValidity(confirmedAt, testCase.days, testCase.endsAt)
			require.NoError(t, err)
			require.Equal(t, testCase.ok, ok)
			if !ok {
				return
			}
			require.Equal(t, confirmedAt, validFrom)
			require.Equal(t, testCase.expires, expiresAt)
		})
	}
	_, _, _, err := IssueValidity(confirmedAt, days(0), nil)
	require.Error(t, err)
}

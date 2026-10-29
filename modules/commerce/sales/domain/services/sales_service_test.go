package services

import (
	"testing"
)

// The code rules are pure functions, so they are tested without a database. A channel code is
// immutable once written, so a code that should have been rejected is a permanent wrong entry, and
// a code normalised inconsistently resolves for one caller and not another.

func TestNormalizeChannelCode(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"already canonical", "VDMC", "VDMC"},
		{"legacy lowercase is raised before create", "vdmc", "VDMC"},
		{"mixed case", "VdMc", "VDMC"},
		{"underscore becomes a hyphen", "pos_01", "POS-01"},
		{"surrounding space is trimmed", "  vdmc  ", "VDMC"},
		{"tab and newline are trimmed", "\tvdmc\n", "VDMC"},
		{"empty stays empty", "", ""},
		{"only space becomes empty", "   ", ""},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := NormalizeChannelCode(testCase.input); got != testCase.want {
				t.Errorf("NormalizeChannelCode(%q) = %q, want %q",
					testCase.input, got, testCase.want)
			}
		})
	}
}

func TestIsValidChannelCode(t *testing.T) {
	valid := []string{"VDMC", "BO", "POS01", "ECOM", "A", "0", "ABC123", "POS-01"}
	for _, code := range valid {
		if !IsValidChannelCode(code) {
			t.Errorf("IsValidChannelCode(%q) = false, want true", code)
		}
	}

	// Every one of these would otherwise end up in a URL or another module's configuration file,
	// where it would need escaping nobody would remember to apply.
	invalid := map[string]string{
		"":       "empty",
		"vdmc":   "lowercase — normalize before validating, never after",
		"VD MC":  "whitespace",
		"VD_MC":  "underscore",
		"VD/MC":  "slash",
		"VD.MC":  "dot",
		"VDMC!":  "punctuation",
		"VDMC\n": "newline",
		"KÊNH":   "non-ascii",
	}
	for code, reason := range invalid {
		if IsValidChannelCode(code) {
			t.Errorf("IsValidChannelCode(%q) = true, want false (%s)", code, reason)
		}
	}
}

// Normalising then validating is the order the create path uses, and the pair must agree: a code
// typed in any case is valid, because it is raised first, and the legacy underscore becomes a hyphen.
func TestNormalizeThenValidate(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"VDMC", true},
		{"  Bo  ", true},
		{"POS01", true},
		{"VD_MC", true},
		{"VD MC", false},
		{"", false},
	}
	for _, testCase := range cases {
		got := IsValidChannelCode(NormalizeChannelCode(testCase.input))
		if got != testCase.want {
			t.Errorf("validate(normalize(%q)) = %v, want %v", testCase.input, got, testCase.want)
		}
	}
}

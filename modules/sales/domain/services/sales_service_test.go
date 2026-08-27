package services

import (
	"testing"
)

// The code rules are pure functions, so they are tested without a database. They matter more than
// their size suggests: a channel code is immutable once written, so a code that should have been
// rejected is a permanent wrong entry, and a code normalised inconsistently resolves for one caller
// and not another.

func TestNormalizeChannelCode(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"already canonical", "vdmc", "vdmc"},
		{"uppercase is lowered before create", "VDMC", "vdmc"},
		{"mixed case", "VdMc", "vdmc"},
		{"surrounding space is trimmed", "  vdmc  ", "vdmc"},
		{"tab and newline are trimmed", "\tvdmc\n", "vdmc"},
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
	valid := []string{"vdmc", "bo", "pos01", "ecom", "a", "0", "abc123"}
	for _, code := range valid {
		if !IsValidChannelCode(code) {
			t.Errorf("IsValidChannelCode(%q) = false, want true", code)
		}
	}

	// Every one of these would otherwise end up in a URL or another module's configuration file,
	// where it would need escaping or quoting that nobody would remember to apply.
	invalid := map[string]string{
		"":       "empty",
		"VDMC":   "uppercase — normalize before validating, never after",
		"vd mc":  "whitespace",
		"vd_mc":  "underscore",
		"vd-mc":  "hyphen",
		"vd/mc":  "slash",
		"vd.mc":  "dot",
		"vdmc!":  "punctuation",
		"vdmc\n": "newline",
		"kênh":   "non-ascii",
	}
	for code, reason := range invalid {
		if IsValidChannelCode(code) {
			t.Errorf("IsValidChannelCode(%q) = true, want false (%s)", code, reason)
		}
	}
}

// Normalising then validating is the order the create path uses, and the pair must agree: a code
// the user typed in uppercase is valid, because it is lowered first.
func TestNormalizeThenValidate(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"VDMC", true},
		{"  Bo  ", true},
		{"POS01", true},
		{"VD_MC", false},
		{"", false},
	}
	for _, testCase := range cases {
		got := IsValidChannelCode(NormalizeChannelCode(testCase.input))
		if got != testCase.want {
			t.Errorf("validate(normalize(%q)) = %v, want %v", testCase.input, got, testCase.want)
		}
	}
}

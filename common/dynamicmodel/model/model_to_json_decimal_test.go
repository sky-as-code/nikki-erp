package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ToModelJson and ParseModelJson are inverses, and a type supported by one and not the other is a
// schema that registers but cannot be read back.
//
// decimal was exactly that until Sales declared a decimal setting: the settings module serialises a
// registered schema into settings_schemas, and the serialisation failed at start-up with
// "unsupported data type 'decimal'". These tests pin the round trip so the pair cannot drift apart
// again.

func parseOneField(t *testing.T, dataTypeJson string) *ModelSchema {
	t.Helper()
	return ParseModelJson(`{
		"name": "test_decimal_schema",
		"fields": [{"name": "rate", "data_type": ` + dataTypeJson + `}]
	}`).Build()
}

// The bounds must be emitted as STRINGS and the scale as an integer, because that is what the
// parser requires. Emitting the bounds as JSON numbers would round-trip through float64 and lose
// the precision the string form exists to preserve.
func TestToModelJson_DecimalShape(t *testing.T) {
	schema := parseOneField(t, `{"type":"decimal","min":"0","max":"1","scale":4}`)

	document, err := schema.ToModelJson()
	require.NoError(t, err, "a decimal must be renderable: the settings module does this on boot")

	fields, ok := document["fields"].([]any)
	require.Truef(t, ok, "fields is %T", document["fields"])
	require.Len(t, fields, 1)

	field, ok := fields[0].(map[string]any)
	require.Truef(t, ok, "field is %T", fields[0])

	dataType, ok := field["data_type"].(map[string]any)
	require.Truef(t, ok, "data_type is %T, want the object form", field["data_type"])

	assert.Equal(t, "decimal", dataType["type"])
	assert.Equal(t, "0", dataType["min"],
		"min must be the STRING \"0\": the parser reads it with jsonToString")
	assert.Equal(t, "1", dataType["max"], "max must be a string")
	assert.Equal(t, 4, dataType["scale"], "scale must be an integer")
}

// The real requirement: a schema carrying a decimal survives being rendered, marshalled and parsed
// back with its bounds and scale intact. This is what the settings module does on every boot.
func TestDecimalSurvivesRoundTrip(t *testing.T) {
	cases := []struct {
		name     string
		declared string
	}{
		{"a rate between zero and one", `{"type":"decimal","min":"0","max":"1","scale":4}`},
		{"a money bound", `{"type":"decimal","min":"0","max":"999999999999","scale":4}`},
		// Bounds too large for a float64 to hold exactly, which is the case the string form exists
		// for: a numeric round trip would quietly return a different number.
		{"beyond float64 precision",
			`{"type":"decimal","min":"-99999999999999999999.123456","max":"99999999999999999999.123456","scale":6}`},
		{"a negative range", `{"type":"decimal","min":"-100.5","max":"-0.5","scale":2}`},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			original := parseOneField(t, testCase.declared)

			document, err := original.ToModelJson()
			require.NoError(t, err)
			encoded, err := json.Marshal(document)
			require.NoError(t, err, "the rendered document must be marshallable")

			rebuilt := ParseModelJson(string(encoded)).Build()

			originalField, ok := original.Field("rate")
			require.True(t, ok)
			rebuiltField, ok := rebuilt.Field("rate")
			require.True(t, ok, "the round trip dropped the field")

			assert.Equal(t, originalField.DataType().String(), rebuiltField.DataType().String())
			assert.Equal(t, originalField.DataType().Options(), rebuiltField.DataType().Options(),
				"bounds or scale changed crossing JSON")
		})
	}
}

// A decimal whose options are missing must fail loudly rather than emit a document the parser will
// reject later, at a point far from the cause.
func TestDecimalWithoutScaleIsRefused(t *testing.T) {
	broken := fieldDataTypeDecimal{fieldDataTypeBase{
		name:    FieldDataTypeNameDecimal,
		options: FieldDataTypeOptions{FieldDataTypeOptRange: []string{"0", "1"}},
	}}

	_, err := dataTypeToModelJson(broken)
	assert.Error(t, err, "a decimal with no scale must be refused: emitting it would store a "+
		"document the parser cannot read back")
}

func TestDecimalWithoutBoundsIsRefused(t *testing.T) {
	broken := fieldDataTypeDecimal{fieldDataTypeBase{
		name:    FieldDataTypeNameDecimal,
		options: FieldDataTypeOptions{FieldDataTypeOptScale: uint(4)},
	}}

	_, err := dataTypeToModelJson(broken)
	assert.Error(t, err, "a decimal with no range must be refused")
}

// Every data type the PARSER accepts must be renderable, or a schema declaring it registers and
// then cannot be read back — which is precisely the failure decimal caused. This walks the same
// list the parser's own type test uses, so a type added to one and forgotten in the other fails
// here rather than at somebody's start-up.
func TestEveryParsedDataTypeIsRenderable(t *testing.T) {
	declarations := []string{
		`"boolean"`,
		`"ulid"`,
		`"uuid"`,
		`"email"`,
		`"phone"`,
		`"url"`,
		`"etag"`,
		`"slug"`,
		`"date"`,
		`"time"`,
		`"datetime"`,
		`"jsonmap"`,
		`{"type":"int32","min":0,"max":100}`,
		`{"type":"int64","min":0,"max":900}`,
		`{"type":"decimal","min":"0","max":"99.9","scale":2}`,
		`{"type":"string","min":1,"max":50}`,
		`{"type":"secret","min":1,"max":50}`,
		`{"type":"langjson","min":1,"max":200}`,
		`{"type":"enum_string","values":["a","b"]}`,
		`{"type":"enum_int32","values":[1,2]}`,
	}
	for _, declared := range declarations {
		t.Run(declared, func(t *testing.T) {
			schema := parseOneField(t, declared)
			_, err := schema.ToModelJson()
			assert.NoErrorf(t, err,
				"%s parses but cannot be rendered: a schema declaring it would register and then "+
					"fail to be read back", declared)
		})
	}
}

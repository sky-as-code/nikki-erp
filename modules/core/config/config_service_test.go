package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.bryk.io/pkg/errors"

	c "github.com/sky-as-code/nikki-erp/modules/core/constants"
)

type mapLoader map[string]string

func (mapLoader) Init() error { return nil }

func (this mapLoader) Get(name string) (string, error) {
	if value, ok := this[name]; ok {
		return value, nil
	}
	return "", errors.Errorf("config %s not found", name)
}

const testKey c.ConfigName = "TEST.NUMBER"

func serviceWith(value string) *configServiceImpl {
	return NewConfigService(mapLoader{string(testKey): value})
}

func TestNumericGetters(t *testing.T) {
	cases := []struct {
		name  string
		value string
		get   func(svc *configServiceImpl) any
		want  any
	}{
		{"int", "-42", func(svc *configServiceImpl) any { return svc.GetInt(testKey) }, -42},
		{"int32 max", "2147483647", func(svc *configServiceImpl) any { return svc.GetInt32(testKey) }, int32(2147483647)},
		{"int64 above int32", "2147483648", func(svc *configServiceImpl) any { return svc.GetInt64(testKey) }, int64(2147483648)},
		{"int64 4 GiB", "4294967296", func(svc *configServiceImpl) any { return svc.GetInt64(testKey) }, int64(4294967296)},
		{"int64 negative", "-5", func(svc *configServiceImpl) any { return svc.GetInt64(testKey) }, int64(-5)},
		{"uint", "4294967296", func(svc *configServiceImpl) any { return svc.GetUint(testKey) }, uint(4294967296)},
		{"uint64 max", "18446744073709551615", func(svc *configServiceImpl) any { return svc.GetUint64(testKey) },
			uint64(18446744073709551615)},
		{"float32", "1.5", func(svc *configServiceImpl) any { return svc.GetFloat32(testKey) }, float32(1.5)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.get(serviceWith(tc.value)))
		})
	}
}

func TestNumericGettersFallBackToTheDefault(t *testing.T) {
	svc := NewConfigService(mapLoader{})
	assert.Equal(t, int64(104857600), svc.GetInt64(testKey, "104857600"))
	assert.Equal(t, int32(7), svc.GetInt32(testKey, "7"))
}

func TestNumericGettersPanicOnOverflow(t *testing.T) {
	cases := map[string]func(svc *configServiceImpl){
		"int32":  func(svc *configServiceImpl) { svc.GetInt32(testKey) },
		"int64":  func(svc *configServiceImpl) { svc.GetInt64(testKey) },
		"uint64": func(svc *configServiceImpl) { svc.GetUint64(testKey) },
	}
	values := map[string]string{
		"int32":  "2147483648",
		"int64":  "9223372036854775808",
		"uint64": "18446744073709551616",
	}
	for name, get := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Panics(t, func() { get(serviceWith(values[name])) })
		})
	}
}

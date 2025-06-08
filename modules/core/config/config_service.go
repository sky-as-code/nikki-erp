package config

import (
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/xgfone/go-cast"
	"go.bryk.io/pkg/errors"

	. "github.com/sky-as-code/nikki-erp/common/fault"
	c "github.com/sky-as-code/nikki-erp/modules/core/constants"
)

var CommitId = func() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value
			}
		}
	}
	return "N/A"
}()

func NewConfigService(loader ConfigLoader) *configServiceImpl {
	return &configServiceImpl{
		loader,
	}
}

type configServiceImpl struct {
	Loader ConfigLoader
}

func (this *configServiceImpl) Init() error {
	return this.Loader.Init()
}

func (this *configServiceImpl) GetAppVersion() string {
	return CommitId
}

func (this *configServiceImpl) GetStr(name c.ConfigName, defaultVal ...any) string {
	val, err := this.Loader.Get(string(name))
	if err == nil {
		return val
	}
	if len(defaultVal) > 0 {
		return defaultVal[0].(string)
	}
	panic(err)
}

func (this *configServiceImpl) GetStrArr(name c.ConfigName, defaultVal ...any) []string {
	str := this.GetStr(name, defaultVal...)
	strArr := strings.Split(str, ",")
	return strArr
}

func (this *configServiceImpl) GetDuration(name c.ConfigName, defaultVal ...any) time.Duration {
	str := this.GetStr(name, defaultVal...)
	val, err := cast.ToDuration(str)
	this.panicOnConversionError("GetDuration()", name, err)
	return val
}

func (this *configServiceImpl) GetBool(name c.ConfigName, defaultVal ...interface{}) bool {
	str := this.GetStr(name, defaultVal...)
	val, err := cast.ToBool(str)
	this.panicOnConversionError("GetBool()", name, err)
	return val
}

func (this *configServiceImpl) GetUint(name c.ConfigName, defaultVal ...interface{}) uint {
	str := this.GetStr(name, defaultVal...)
	val64, err := strconv.ParseUint(str, 10, 64)
	val := uint(val64)
	this.panicOnConversionError("GetUint()", name, err)
	return val
}

func (this *configServiceImpl) GetUint64(name c.ConfigName, defaultVal ...interface{}) uint64 {
	str := this.GetStr(name, defaultVal...)
	val64, err := strconv.ParseUint(str, 10, 64)
	this.panicOnConversionError("GetUint64()", name, err)
	return val64
}

func (this *configServiceImpl) GetInt(name c.ConfigName, defaultVal ...interface{}) int {
	str := this.GetStr(name, defaultVal...)
	val, err := strconv.Atoi(str)
	this.panicOnConversionError("GetInt()", name, err)
	return val
}

func (this *configServiceImpl) GetInt32(name c.ConfigName, defaultVal ...interface{}) int32 {
	str := this.GetStr(name, defaultVal...)
	val64, err := strconv.ParseInt(str, 10, 32)
	val32 := int32(val64)
	this.panicOnConversionError("GetInt32()", name, err)
	return val32
}

func (this *configServiceImpl) GetInt64(name c.ConfigName, defaultVal ...interface{}) int64 {
	str := this.GetStr(name, defaultVal...)
	val64, err := strconv.ParseInt(str, 10, 32)
	this.panicOnConversionError("GetInt64()", name, err)
	return val64
}

func (this *configServiceImpl) GetFloat32(name c.ConfigName, defaultVal ...interface{}) float32 {
	str := this.GetStr(name, defaultVal...)
	val64, err := strconv.ParseFloat(str, 32)
	val32 := float32(val64)
	this.panicOnConversionError("GetFloat32()", name, err)
	return val32
}

func (*configServiceImpl) panicOnConversionError(funcName string, name c.ConfigName, err error) {
	PanicOnErr(errors.Wrap(err, fmt.Sprintf("configServiceImpl.%s failed to convert config '%s'", funcName, name)))
}

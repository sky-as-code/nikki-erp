package config

import (
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/xgfone/go-cast"

	// . "github.com/sky-as-code/nikki-erp/common/config/types"
	. "github.com/sky-as-code/nikki-erp/common/util/fault"
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

func (this *configServiceImpl) Init() AppError {
	return this.Loader.Init()
}

func (this *configServiceImpl) GetAppVersion() string {
	return CommitId
}

func (this *configServiceImpl) GetStr(name ConfigName, defaultVal ...interface{}) string {
	val, err := this.Loader.Get(string(name))
	if err == nil {
		return val
	}
	if len(defaultVal) > 0 {
		return defaultVal[0].(string)
	}
	panic(err)
}

func (this *configServiceImpl) GetStrArr(name ConfigName, defaultVal ...interface{}) []string {
	str := this.GetStr(name, defaultVal...)
	strArr := strings.Split(str, ",")
	return strArr
}

func (this *configServiceImpl) GetDuration(configName ConfigName, defaultVal ...interface{}) time.Duration {
	str := this.GetStr(configName, defaultVal...)
	val, err := cast.ToDuration(str)
	this.panicOnConversionError("GetDuration()", configName, err)
	return val
}

func (this *configServiceImpl) GetBool(configName ConfigName, defaultVal ...interface{}) bool {
	str := this.GetStr(configName, defaultVal...)
	val, err := cast.ToBool(str)
	this.panicOnConversionError("GetBool()", configName, err)
	return val
}

func (this *configServiceImpl) GetUint(configName ConfigName, defaultVal ...interface{}) uint {
	str := this.GetStr(configName, defaultVal...)
	val64, err := strconv.ParseUint(str, 10, 64)
	val := uint(val64)
	this.panicOnConversionError("GetUint()", configName, err)
	return val
}

func (this *configServiceImpl) GetUint64(configName ConfigName, defaultVal ...interface{}) uint64 {
	str := this.GetStr(configName, defaultVal...)
	val64, err := strconv.ParseUint(str, 10, 64)
	this.panicOnConversionError("GetUint64()", configName, err)
	return val64
}

func (this *configServiceImpl) GetInt(configName ConfigName, defaultVal ...interface{}) int {
	str := this.GetStr(configName, defaultVal...)
	val, err := strconv.Atoi(str)
	this.panicOnConversionError("GetInt()", configName, err)
	return val
}

func (this *configServiceImpl) GetInt32(configName ConfigName, defaultVal ...interface{}) int32 {
	str := this.GetStr(configName, defaultVal...)
	val64, err := strconv.ParseInt(str, 10, 32)
	val32 := int32(val64)
	this.panicOnConversionError("GetInt32()", configName, err)
	return val32
}

func (this *configServiceImpl) GetInt64(configName ConfigName, defaultVal ...interface{}) int64 {
	str := this.GetStr(configName, defaultVal...)
	val64, err := strconv.ParseInt(str, 10, 32)
	this.panicOnConversionError("GetInt64()", configName, err)
	return val64
}

func (this *configServiceImpl) GetFloat32(configName ConfigName, defaultVal ...interface{}) float32 {
	str := this.GetStr(configName, defaultVal...)
	val64, err := strconv.ParseFloat(str, 32)
	val32 := float32(val64)
	this.panicOnConversionError("GetFloat32()", configName, err)
	return val32
}

func (*configServiceImpl) panicOnConversionError(funcName string, configName ConfigName, err error) {
	PanicOnErr(WrapTechnicalError(
		err,
		"configServiceImpl.%s failed to convert config '%s'",
		funcName, configName,
	))
}

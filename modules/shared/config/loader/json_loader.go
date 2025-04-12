package loader

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sky-as-code/nikki-erp/modules/shared/logging"
	"github.com/sky-as-code/nikki-erp/utility/env"
	. "github.com/sky-as-code/nikki-erp/utility/fault"
	"github.com/sky-as-code/nikki-erp/utility/json"
	"github.com/tidwall/gjson"
)

var jsonPath = filepath.Join("config", "config.json")

func NewJsonConfigLoader(logger logging.LoggerService) *JsonConfigLoader {
	return &JsonConfigLoader{
		logger: logger,
	}
}

// JsonConfigLoader gets configuration from JSON file.
type JsonConfigLoader struct {
	logger     logging.LoggerService
	jsonResult gjson.Result
}

func (loader *JsonConfigLoader) Init() (err AppError) {
	loader.jsonResult, err = loader.load()
	logging.Logger().Infof("JsonConfigLoader.Init conf: %v", loader.jsonResult)
	return err
}

func (loader *JsonConfigLoader) CreateConfigDB(mapCfg interface{}) AppError {
	// logging.Logger().Infof("JsonConfigLoader.CreateConfigDB mapCfg: %+v", mapCfg)
	if mapCfg == nil {
		return WrapTechnicalError(fmt.Errorf("mapCfg nil"), "JsonConfigLoader.CreateConfigDB()")
	}
	currentEnv := env.AppEnv()
	if len(currentEnv) == 0 {
		currentEnv = "default"
	}

	dataBase, err := json.Marshal(mapCfg)
	if err != nil {
		return WrapTechnicalError(err, "JsonConfigLoader.CreateConfigDB Marshal()")
	}
	data := fmt.Sprintf(`{"%s":%s}`, currentEnv, dataBase)
	// logging.Logger().Infof("JsonConfigLoader.CreateConfigDB: %v", string(data))
	loader.jsonResult = gjson.Parse(data)
	return nil
}

func (loader *JsonConfigLoader) Get(name string) (string, AppError) {
	currentEnv := env.AppEnv()

	val := loader.jsonResult.Get(fmt.Sprintf("%s.%s", currentEnv, name)) // Eg: dev.DB_HOST, prod.HTTP_PORT
	noEnvSpecificConfig := (!val.Exists() || len(val.String()) == 0)
	if noEnvSpecificConfig {
		val = loader.jsonResult.Get(fmt.Sprintf("default.%s", name)) // Eg: default.DB_HOST, default.HTTP_PORT
	}

	noDefaultConfig := (!val.Exists() || len(val.String()) == 0)
	if noDefaultConfig {
		return "", NewTechnicalError("JsonConfigLoader.Get('%s') found nothing", name)
	}

	return val.String(), nil
}

func (loader *JsonConfigLoader) load() (result gjson.Result, appErr AppError) {
	defer func() {
		if err := recover(); err != nil {
			appErr = WrapTechnicalError(err.(error), "JsonConfigLoader.readJsonFile()")
		}
	}()

	result = gjson.Result{}
	workDir := env.Cwd()
	bytes, err := os.ReadFile(filepath.Join(workDir, jsonPath))
	PanicOnErr(err)

	if !json.IsValidBytes(bytes) {
		panic(fmt.Errorf("Content of %s is invalid JSON", jsonPath))
	}

	return json.ParseBytes(bytes), appErr
}

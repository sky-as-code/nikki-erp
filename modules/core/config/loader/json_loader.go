package loader

import (
	"fmt"
	"os"
	"path/filepath"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/env"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/json"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
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

func (loader *JsonConfigLoader) Init() (err error) {
	loader.jsonResult, err = loader.load()
	logging.Logger().Infof("JsonConfigLoader.Init conf: %v", loader.jsonResult)
	return err
}

func (loader *JsonConfigLoader) CreateConfigDB(mapCfg interface{}) error {
	// logging.Logger().Infof("JsonConfigLoader.CreateConfigDB mapCfg: %+v", mapCfg)
	if mapCfg == nil {
		return errors.New("mapCfg nil")
	}
	currentEnv := env.AppEnv()
	if len(currentEnv) == 0 {
		currentEnv = "default"
	}

	dataBase, err := json.Marshal(mapCfg)
	if err != nil {
		return errors.Wrap(err, "JsonConfigLoader.CreateConfigDB Marshal()")
	}
	data := fmt.Sprintf(`{"%s":%s}`, currentEnv, dataBase)
	// logging.Logger().Infof("JsonConfigLoader.CreateConfigDB: %v", string(data))
	loader.jsonResult = gjson.Parse(data)
	return nil
}

func (loader *JsonConfigLoader) Get(name string) (string, error) {
	currentEnv := env.AppEnv()

	val := loader.jsonResult.Get(fmt.Sprintf("%s.%s", currentEnv, name)) // Eg: dev.DB_HOST, prod.HTTP_PORT
	noEnvSpecificConfig := (!val.Exists() || len(val.String()) == 0)
	if noEnvSpecificConfig {
		val = loader.jsonResult.Get(fmt.Sprintf("default.%s", name)) // Eg: default.DB_HOST, default.HTTP_PORT
	}

	noDefaultConfig := (!val.Exists() || len(val.String()) == 0)
	if noDefaultConfig {
		return "", errors.New(fmt.Errorf("JsonConfigLoader.Get('%s') found nothing", name))
	}

	return val.String(), nil
}

func (loader *JsonConfigLoader) load() (result gjson.Result, appErr error) {
	defer func() {
		appErr = ft.RecoverPanic(recover(), "JsonConfigLoader.load()")
	}()

	result = gjson.Result{}
	workDir := env.Cwd()
	bytes, err := os.ReadFile(filepath.Join(workDir, jsonPath))
	ft.PanicOnErr(err)

	if !json.IsValidBytes(bytes) {
		panic(fmt.Errorf("Content of %s is invalid JSON", jsonPath))
	}

	return json.ParseBytes(bytes), appErr
}

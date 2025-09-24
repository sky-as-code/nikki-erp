package loader

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/env"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	rootConfig "github.com/sky-as-code/nikki-erp/config"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

var configFilePath = filepath.Join("config", "config.yaml")

func NewYamlConfigLoader(logger logging.LoggerService) *yamlConfigLoader {
	return &yamlConfigLoader{
		logger: logger,
	}
}

type yamlConfigLoader struct {
	logger         logging.LoggerService
	defaultYamlAst *ast.File
	currentYamlAst *ast.File
}

func (this *yamlConfigLoader) Init() (err error) {
	this.defaultYamlAst, err = this.loadDefaultYaml()
	this.currentYamlAst, err = this.loadCurrentYaml()
	return err
}

func (this *yamlConfigLoader) Get(name string) (string, error) {
	yamlPath := fmt.Sprintf("$.%s", name)

	if this.currentYamlAst != nil {
		// Try to get value from current environment first
		if value, found := this.getValueByPath(yamlPath, this.currentYamlAst); found {
			return value, nil
		}
	}

	// Fallback to default environment
	if value, found := this.getValueByPath(yamlPath, this.defaultYamlAst); found {
		return value, nil
	}

	return "", errors.Errorf("yamlConfigLoader.Get('%s') found nothing", name)
}

func (this *yamlConfigLoader) getValueByPath(yamlPath string, yamlAst *ast.File) (string, bool) {
	path, err := yaml.PathString(yamlPath)
	if err == nil {
		if node, err := path.FilterFile(yamlAst); err == nil {
			if scalarNode, ok := node.(ast.ScalarNode); ok {
				return fmt.Sprintf("%v", scalarNode.GetValue()), true
			}
		}
	}
	return "", false
}

func (this *yamlConfigLoader) loadDefaultYaml() (result *ast.File, appErr error) {
	return this.load([]byte(rootConfig.DefaultConfigYaml), "default YAML file")
}

func (this *yamlConfigLoader) loadCurrentYaml() (result *ast.File, appErr error) {
	workDir := env.Cwd()
	configPath := filepath.Join(workDir, configFilePath)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// If the config file does not exist, return nil without error
		return nil, nil
	}

	bytes, err := os.ReadFile(filepath.Join(workDir, configFilePath))
	ft.PanicOnErr(err)
	return this.load(bytes, "config YAML file")
}

func (this *yamlConfigLoader) load(bytes []byte, yamlType string) (result *ast.File, appErr error) {
	defer func() {
		if e := ft.RecoverPanicf(recover(), "failed to load %s", yamlType); e != nil {
			appErr = e
		}
	}()

	astFile, err := parser.ParseBytes(bytes, 0)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse YAML")
	}

	return astFile, appErr
}

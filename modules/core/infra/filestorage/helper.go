package filestorage

import (
	"path"
	"strings"

	"github.com/google/uuid"
)

type BuildObjectKeyParams struct {
	Env     string
	Feature string
	Id      string
	Name    string
}

func BuildObjectKey(params BuildObjectKeyParams) string {
	if params.Name != "" {
		params.Name = strings.ReplaceAll(params.Name, "\\", "/")
	}
	if params.Name == "." || params.Name == "" {
		params.Name = "blod"
	}

	if params.Env == "" {
		params.Env = "test"
	}

	if params.Id == "" {
		params.Id = uuid.NewString()
	}

	if params.Feature == "" {
		params.Feature = "test"
	}

	return path.Join(params.Env, params.Feature, params.Id, params.Name)
}

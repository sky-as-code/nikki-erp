package promotion

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// AC-01: no point, credit or loyalty concept may be earned, accumulated, redeemed or carried over.
// The point of sale ("sales_point", and its junction "sales_promotion_point") is the one legitimate
// use of the word.
var forbiddenConcept = regexp.MustCompile(`(?i)(loyalty|credit|points?)`)

func isForbidden(name string) bool {
	cleaned := strings.ToLower(name)
	for _, allowed := range []string{"sales_promotion_point", "sales_point", "salespoint"} {
		cleaned = strings.ReplaceAll(cleaned, allowed, "")
	}
	return forbiddenConcept.MatchString(cleaned)
}

// The engine's own identifiers and string values (comments excluded) carry no such concept.
func TestNoLoyaltyConceptInEngine(t *testing.T) {
	files, err := filepath.Glob("*.go")
	require.NoError(t, err)
	fileSet := token.NewFileSet()
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(fileSet, file, nil, 0)
		require.NoError(t, err)
		ast.Inspect(parsed, func(node ast.Node) bool {
			switch value := node.(type) {
			case *ast.Ident:
				require.Falsef(t, isForbidden(value.Name), "%s: identifier %s", file, value.Name)
			case *ast.BasicLit:
				require.Falsef(t, isForbidden(value.Value), "%s: literal %s", file, value.Value)
			}
			return true
		})
	}
}

// The promotion and voucher JSON models carry no such field. Skipped until the models exist, since
// they are written by another lane.
func TestNoLoyaltyConceptInPromotionModels(t *testing.T) {
	patterns := []string{"sales_promotion*.json", "sales_voucher*.json", "sales_order_voucher*.json"}
	modelsDir := filepath.Join("..", "..", "..", "models")
	checked := 0
	for _, pattern := range patterns {
		files, err := filepath.Glob(filepath.Join(modelsDir, pattern))
		require.NoError(t, err)
		for _, file := range files {
			checked++
			for _, name := range jsonFieldNames(t, file) {
				require.Falsef(t, isForbidden(name), "%s: field %s", filepath.Base(file), name)
			}
		}
	}
	if checked == 0 {
		t.Skip("no promotion models found")
	}
}

func jsonFieldNames(t *testing.T, file string) []string {
	content, err := os.ReadFile(file)
	require.NoError(t, err)
	var model struct {
		Name   string `json:"name"`
		Fields []struct {
			Name string `json:"name"`
		} `json:"fields"`
	}
	require.NoError(t, json.Unmarshal(content, &model), file)
	names := []string{model.Name}
	for _, field := range model.Fields {
		names = append(names, field.Name)
	}
	return names
}

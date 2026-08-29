package sales

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// Every Sales table is written twice — once here, once in coremart with a tenant_id on every table
// and unique constraint. The trees drift silently, so these tests parse and compare both. The
// coremart tree is reached by a relative path; when it is absent the cross-tree checks skip.

// postgresIdentifierLimit is where Postgres truncates a name silently: the constraint is created
// under a shortened name, so every later DROP CONSTRAINT by the declared name fails instead.
const postgresIdentifierLimit = 63

func TestMigrationIdentifiersFitPostgres(t *testing.T) {
	pattern := regexp.MustCompile(`(?:CREATE (?:UNIQUE )?INDEX|CONSTRAINT)\s+"([^"]+)"`)

	for _, tree := range migrationTrees(t) {
		for _, file := range salesMigrations(t, tree) {
			content := readMigration(t, tree, file)

			for _, match := range pattern.FindAllStringSubmatch(content, -1) {
				name := match[1]
				if len(name) > postgresIdentifierLimit {
					t.Errorf("%s/%s declares %q (%d chars); Postgres truncates at %d SILENTLY, so "+
						"the constraint exists under a different name than this file claims",
						filepath.Base(tree), file, name, len(name), postgresIdentifierLimit)
				}
			}
		}
	}
}

// A table added to one tree and forgotten in the other breaks only a fresh install of the tree that
// missed it, so it does not show up locally.
func TestBothTreesCarryTheSameSalesMigrations(t *testing.T) {
	trees := migrationTrees(t)
	if len(trees) < 2 {
		t.Skip("the coremart tree is not present in this checkout")
	}

	first := salesMigrations(t, trees[0])
	second := salesMigrations(t, trees[1])

	if strings.Join(first, ",") != strings.Join(second, ",") {
		t.Errorf("the trees carry different Sales migrations:\n  %s: %v\n  %s: %v",
			filepath.Base(trees[0]), first, filepath.Base(trees[1]), second)
	}
}

// Every coremart Sales table carries tenant_id, and every unique constraint is tenant-prefixed: an
// unprefixed unique makes a value globally unique across tenants, so one tenant's "SO-1" blocks
// every other tenant's.
func TestCoremartTablesAreTenantScoped(t *testing.T) {
	trees := migrationTrees(t)
	if len(trees) < 2 {
		t.Skip("the coremart tree is not present in this checkout")
	}
	coremart := trees[1]

	tablePattern := regexp.MustCompile(`(?s)CREATE TABLE "(\w+)" \((.*?)\);`)
	uniquePattern := regexp.MustCompile(`CONSTRAINT "(\w+)" UNIQUE \(([^)]*)\)`)

	for _, file := range salesMigrations(t, coremart) {
		if strings.Contains(file, "iam") {
			// IAM rows are global rather than tenant-scoped: a resource and its actions describe
			// what the software can do, not what any one tenant's data contains.
			continue
		}
		content := readMigration(t, coremart, file)

		for _, table := range tablePattern.FindAllStringSubmatch(content, -1) {
			name, body := table[1], table[2]

			if !strings.Contains(body, `"tenant_id"`) {
				t.Errorf("%s: table %s has no tenant_id; every row in coremart belongs to a tenant, "+
					"and a table without one cannot be filtered by the multi-tenant repository",
					file, name)
			}

			for _, unique := range uniquePattern.FindAllStringSubmatch(body, -1) {
				if !strings.Contains(unique[2], "tenant_id") {
					t.Errorf("%s: %s's unique %q covers (%s) without tenant_id, which makes the "+
						"value unique ACROSS tenants — one tenant's row would block another's",
						file, name, unique[1], unique[2])
				}
			}
		}
	}
}

// The Sales migrations own their numeric block exclusively: another module numbering into it would
// interleave on a fresh install, creating a Sales table before something it references. Existing
// installs are unaffected, so it only shows up when provisioning a new environment.
func TestTheSalesMigrationBlockIsExclusivelySales(t *testing.T) {
	for _, tree := range migrationTrees(t) {
		entries, err := os.ReadDir(tree)
		if err != nil {
			t.Fatalf("reading %s: %v", tree, err)
		}

		for _, entry := range entries {
			name := entry.Name()
			if !strings.HasPrefix(name, salesMigrationPrefix) {
				continue
			}
			if !strings.Contains(name, "sales") {
				t.Errorf("%s/%s is in the %s block but is not a Sales migration; the block must "+
					"stay exclusively Sales so a fresh install orders them correctly",
					filepath.Base(tree), name, salesMigrationPrefix)
			}
		}
	}
}

// salesMigrationPrefix is the numeric block Sales owns.
const salesMigrationPrefix = "1007"

// migrationTrees returns the migration directories to check, nikkierp first.
func migrationTrees(t *testing.T) []string {
	t.Helper()

	trees := []string{filepath.Join("..", "..", "scripts", "migrations")}
	if _, err := os.Stat(trees[0]); err != nil {
		t.Fatalf("the nikkierp migrations directory must be readable: %v", err)
	}

	coremart := filepath.Join("..", "..", "..", "coremart", "scripts", "migrations")
	if _, err := os.Stat(coremart); err == nil {
		trees = append(trees, coremart)
	}
	return trees
}

// salesMigrations lists one tree's Sales migration filenames, sorted.
func salesMigrations(t *testing.T, tree string) []string {
	t.Helper()

	entries, err := os.ReadDir(tree)
	if err != nil {
		t.Fatalf("reading %s: %v", tree, err)
	}

	files := make([]string, 0)
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, salesMigrationPrefix) && strings.HasSuffix(name, ".sql") {
			files = append(files, name)
		}
	}
	sort.Strings(files)

	if len(files) == 0 {
		t.Fatalf("no Sales migrations found in %s", tree)
	}
	return files
}

func readMigration(t *testing.T, tree, file string) string {
	t.Helper()

	content, err := os.ReadFile(filepath.Join(tree, file))
	if err != nil {
		t.Fatalf("reading %s: %v", file, err)
	}
	return string(content)
}

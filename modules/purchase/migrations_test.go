package purchase

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// Migration invariants across every module the pricing change touched, hence the module block as a
// parameter rather than one module hard-coded.
//
// Every table is written twice, once per tree, and the two drift silently: a table added to one and
// forgotten in the other breaks only a fresh install of the tree that missed it.

// checkedBlocks are the module blocks the pricing change wrote into, with the token that must
// appear in a filename for it to belong to that block.
var checkedBlocks = []struct{ prefix, module string }{
	{"1002", "contacts"},
	{"1005", "inventory"},
	{"1006", "purchase"},
	{"1007", "sales"},
}

// postgresIdentifierLimit is where Postgres truncates a name, with no error or warning: the DDL
// applies under a shortened name and every later DROP CONSTRAINT by the declared name fails.
const postgresIdentifierLimit = 63

func TestPricingMigrationIdentifiersFitPostgres(t *testing.T) {
	pattern := regexp.MustCompile(`(?:CREATE (?:UNIQUE )?INDEX|CONSTRAINT)\s+"([^"]+)"`)

	for _, tree := range pricingMigrationTrees(t) {
		for _, block := range checkedBlocks {
			for _, file := range blockMigrations(t, tree, block.prefix) {
				content := readPricingMigration(t, tree, file)

				for _, match := range pattern.FindAllStringSubmatch(content, -1) {
					name := match[1]
					if len(name) > postgresIdentifierLimit {
						t.Errorf("%s/%s declares %q (%d chars); Postgres truncates at %d SILENTLY, "+
							"so the constraint exists under a different name than this file claims",
							filepath.Base(tree), file, name, len(name), postgresIdentifierLimit)
					}
				}
			}
		}
	}
}

// Both trees carry the same migration files, block by block.
func TestBothTreesCarryTheSamePricingMigrations(t *testing.T) {
	trees := pricingMigrationTrees(t)
	if len(trees) < 2 {
		t.Skip("the coremart tree is not present in this checkout")
	}

	for _, block := range checkedBlocks {
		first := migrationKinds(t, trees[0], block.prefix)
		second := migrationKinds(t, trees[1], block.prefix)

		if strings.Join(first, ",") != strings.Join(second, ",") {
			t.Errorf("the trees carry different %s migrations:\n  %s: %v\n  %s: %v",
				block.module, filepath.Base(trees[0]), first, filepath.Base(trees[1]), second)
		}
	}
}

// migrationKinds reduces a block's filenames to their module-and-kind suffixes, dropping the
// sequence number.
//
// The number is dropped because the trees legitimately disagree on one — the contacts IAM migration
// is 1002003 in nikkierp and 1002002 in coremart — and renaming either would rewrite a migration
// already applied and recorded in atlas.sum. What matters is that neither tree is missing a
// migration the other has, since that breaks a fresh install of the tree that lacks it.
func migrationKinds(t *testing.T, tree, prefix string) []string {
	t.Helper()

	kinds := make([]string, 0)
	for _, file := range blockMigrations(t, tree, prefix) {
		name := strings.TrimSuffix(file, ".sql")
		if underscore := strings.Index(name, "_"); underscore >= 0 {
			name = name[underscore+1:]
		}
		kinds = append(kinds, name)
	}
	sort.Strings(kinds)
	return kinds
}

// Every coremart table carries tenant_id, and every unique constraint is tenant-prefixed. An
// unprefixed unique makes a value globally unique across tenants, so one tenant's row blocks
// another's with a duplicate-key error in unrelated data.
func TestCoremartPricingTablesAreTenantScoped(t *testing.T) {
	trees := pricingMigrationTrees(t)
	if len(trees) < 2 {
		t.Skip("the coremart tree is not present in this checkout")
	}
	coremart := trees[1]

	tablePattern := regexp.MustCompile(`(?s)CREATE TABLE "(\w+)" \((.*?)\);`)
	uniquePattern := regexp.MustCompile(`CONSTRAINT "(\w+)" UNIQUE \(([^)]*)\)`)

	for _, block := range checkedBlocks {
		for _, file := range blockMigrations(t, coremart, block.prefix) {
			if strings.Contains(file, "iam") || strings.Contains(file, "seed") {
				// IAM rows and seeds insert rather than declare a table.
				continue
			}
			content := readPricingMigration(t, coremart, file)

			for _, table := range tablePattern.FindAllStringSubmatch(content, -1) {
				name, body := table[1], table[2]

				if !strings.Contains(body, `"tenant_id"`) {
					t.Errorf("%s: table %s has no tenant_id; every row in coremart belongs to a "+
						"tenant, and a table without one cannot be filtered by the multi-tenant "+
						"repository", file, name)
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
}

// Every coremart IAM and seed row carries a tenant_id. The column is NOT NULL in coremart, so a row
// without one fails the INSERT on a fresh install only — an existing database already has the rows
// and skips them via ON CONFLICT, which is why this is checked here and not left to the database.
func TestCoremartPricingInsertsCarryATenant(t *testing.T) {
	trees := pricingMigrationTrees(t)
	if len(trees) < 2 {
		t.Skip("the coremart tree is not present in this checkout")
	}
	coremart := trees[1]

	// The tables whose coremart copies declare a NOT NULL tenant_id.
	tenantScoped := []string{"iam_resources", "iam_actions", "iam_roles", "iam_entitlements"}
	insertPattern := regexp.MustCompile(`INSERT INTO "(\w+)" \(([^)]*)\)`)

	for _, block := range checkedBlocks {
		for _, file := range blockMigrations(t, coremart, block.prefix) {
			content := readPricingMigration(t, coremart, file)

			for _, insert := range insertPattern.FindAllStringSubmatch(content, -1) {
				table, columns := insert[1], insert[2]
				if !contains(tenantScoped, table) {
					continue
				}
				if !strings.Contains(columns, "tenant_id") {
					t.Errorf("%s: INSERT INTO %s names no tenant_id column. The column is NOT NULL "+
						"in coremart, so this breaks a FRESH install while every existing database "+
						"stays green — see SALES-046", file, table)
				}
			}
		}
	}
}

// Each module block stays exclusively its own module: another module numbering into a block would
// interleave on a fresh install, creating a table before something it references.
func TestPricingMigrationBlocksAreExclusive(t *testing.T) {
	for _, tree := range pricingMigrationTrees(t) {
		entries, err := os.ReadDir(tree)
		if err != nil {
			t.Fatalf("reading %s: %v", tree, err)
		}

		for _, entry := range entries {
			name := entry.Name()
			for _, block := range checkedBlocks {
				if !strings.HasPrefix(name, block.prefix) {
					continue
				}
				if !strings.Contains(name, block.module) {
					t.Errorf("%s/%s is in the %s block but is not a %s migration; the block must "+
						"stay exclusive so a fresh install orders them correctly",
						filepath.Base(tree), name, block.prefix, block.module)
				}
			}
		}
	}
}

// The vendor price table exists in both trees. Its absence makes the seeds a silent no-op, because
// their `IF EXISTS` guard reports success while inserting nothing.
func TestTheVendorPriceTableExistsInBothTrees(t *testing.T) {
	for _, tree := range pricingMigrationTrees(t) {
		found := false
		for _, file := range blockMigrations(t, tree, "1006") {
			if strings.Contains(readPricingMigration(t, tree, file),
				`CREATE TABLE "purchase_vendor_product_prices"`) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s declares no purchase_vendor_product_prices table; its seeds would insert "+
				"nothing and report success", filepath.Base(tree))
		}
	}
}

// pricingMigrationTrees returns the migration directories to check, nikkierp first. A missing
// coremart tree skips rather than fails, since the invariant does not apply to a one-tree checkout.
func pricingMigrationTrees(t *testing.T) []string {
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

// blockMigrations lists one tree's migration filenames for one numeric block, sorted.
func blockMigrations(t *testing.T, tree, prefix string) []string {
	t.Helper()

	entries, err := os.ReadDir(tree)
	if err != nil {
		t.Fatalf("reading %s: %v", tree, err)
	}

	files := make([]string, 0)
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, prefix) && strings.HasSuffix(name, ".sql") {
			files = append(files, name)
		}
	}
	sort.Strings(files)

	if len(files) == 0 {
		t.Fatalf("no migrations found for block %s in %s", prefix, tree)
	}
	return files
}

func readPricingMigration(t *testing.T, tree, file string) string {
	t.Helper()

	content, err := os.ReadFile(filepath.Join(tree, file))
	if err != nil {
		t.Fatalf("reading %s: %v", file, err)
	}
	return string(content)
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

# Config for the dbpatch scripts only. Kept apart from scripts/atlas.hcl, whose required `module`/`cwd`
# variables mean nothing here.

# Diff for patch.sql: everything that cannot lose data. Drops are left to the second diff, which
# writes them to destructive.review.sql for a human to decide on.
env "patch" {
  dev = "docker://postgres/17/dev?search_path=public"

  diff {
    skip {
      drop_schema      = true
      drop_table       = true
      drop_column      = true
      drop_index       = true
      drop_foreign_key = true
    }
  }
}

# Full diff: used for the residual after patch.sql, and for verifying live after the apply.
env "full" {
  dev = "docker://postgres/17/dev?search_path=public"
}

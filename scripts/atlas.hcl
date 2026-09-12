env "local" {
  dev = "docker://postgres/17/test?search_path=public"

  diff {
    skip {
      drop_schema = true
      drop_table  = true
    }
  }
  migration {
    exclude = [""]
  }
}

variable "module" {
  type = string
}

variable "cwd" {
  type = string
}

# Tables the generator emits for this module's dependencies (-withdeps). They must be present
# for a cross-module foreign key to resolve, but they belong to another module's migration, so
# they are excluded from the diff. db-migrations.sh fills this from -listdeptables.
variable "dep_tables" {
  type    = list(string)
  default = []
}

data "external_schema" "nikki" {
  program = [
    "go",
    "run",
    "-tags=staticmods",
    "${var.cwd}main.go",
    "-createsql",
    "-dialect=postgres",
    "-withdeps",
    "-module=${var.module}"
  ]
}

env "nikki" {
  src = data.external_schema.nikki.url
  dev = "docker://postgres/17/test?search_path=public"

  migration {
    exclude = var.dep_tables
  }
}

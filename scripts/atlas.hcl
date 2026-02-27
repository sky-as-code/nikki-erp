env "local" {
  dev = "postgresql://nikki_admin:nikki_password@127.0.0.1:5432/nikki_atlas_dev?sslmode=disable&search_path=public"

  diff {
    skip {
      drop_schema = true
      drop_table  = true
    }
  }
  migration {
    exclude = []
  }
}

variable "module" {
  type = string
}

variable "cwd" {
  type = string
}

data "external_schema" "nikki" {
  program = [
    "go",
    "run",
    "-tags=staticmods",
    "${var.cwd}main.go",
    "-createsql",
    "-dialect=postgres",
    "-module=${var.module}"
  ]
}

env "nikki" {
  src = data.external_schema.nikki.url
  dev = "docker://postgres/17/test?search_path=public"
}
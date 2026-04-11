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
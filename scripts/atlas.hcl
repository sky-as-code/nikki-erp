env "local" {
  # dev = "docker://postgres/17/test?search_path=public"
  dev = "postgresql://nikki_admin:nikki_password@localhost:5432/nikki_erp?sslmode=disable"

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

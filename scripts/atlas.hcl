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

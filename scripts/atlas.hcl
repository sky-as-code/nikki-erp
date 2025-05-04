env "local" {
  dev = "docker://postgres/17/test?search_path=public"

  diff {
    skip {
      drop_schema = true
	  drop_table  = true
    }
  }
}

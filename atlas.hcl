data "external_schema" "gorm" {
  program = [
    "go",
    "run",
    "./cmd/tinyurl/main.go",
    "migrate"
  ]
}

env "local" {
  src = data.external_schema.gorm.url
  dev = "docker://postgres/latest/dev?search_path=public"
  url = "postgres://tinyurl:secret@127.0.0.1:5432/tinyurl?sslmode=disable"
  migration {
    dir = "file://migrations"
  }
}

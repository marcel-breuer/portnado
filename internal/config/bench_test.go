package config

import "testing"

var benchmarkRepositoryConfig = []byte(`
version: 1
project:
  name: webguard
services:
  app:
    protocol: http
    route:
      host: app.webguard.localhost
    target:
      discovery: auto
  api:
    protocol: http
    route:
      host: api.webguard.localhost
    target:
      discovery: docker-compose
      service: api
      containerPort: 8080
  db:
    protocol: tcp
    route:
      host: db.webguard.localhost
      preferredPort: 15432
    target:
      discovery: docker-compose
      service: postgres
      containerPort: 5432
`)

func BenchmarkParseRepository(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, err := ParseRepository(benchmarkRepositoryConfig); err != nil {
			b.Fatalf("parse repository: %v", err)
		}
	}
}

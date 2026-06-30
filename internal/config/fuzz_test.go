package config

import "testing"

func FuzzParseRepository(f *testing.F) {
	f.Add([]byte(`
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
`))
	f.Add([]byte(`version: 1`))
	f.Add([]byte(`{not yaml`))

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = ParseRepository(data)
	})
}

func FuzzValidateLocalhost(f *testing.F) {
	f.Add("app.webguard.localhost")
	f.Add("app.webguard.localhost:4780")
	f.Add("../app.localhost")
	f.Add("127.0.0.1.localhost")

	f.Fuzz(func(t *testing.T, host string) {
		_ = ValidateLocalhost(host)
	})
}

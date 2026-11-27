# Local development

Install Go 1.22 or newer. Run `go test ./...` and `go run ./cmd/inkveil`, then open http://127.0.0.1:8080.

This prototype uses Go standard library and an in-memory fictional catalog to keep initial setup dependency-free. Gin, PostgreSQL persistence, reader and authorized source adapters are future milestones. No external source scraping or chapter mirroring is implemented.

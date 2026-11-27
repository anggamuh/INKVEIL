# INKVEIL — Obsidian
Beyond the pages.

Go-based personal reader for original and authorized third-party content.

- [Setup](docs/setup.md), [MySQL](docs/mysql.md), [Suwayomi](docs/suwayomi.md), [private mobile](docs/private-mobile.md)
- [Global search](docs/global-search.md), [chapter release feed](docs/recent-releases.md), [reading experience](docs/obsidian-v2.md)
- [Release checklist](docs/release-checklist.md), [source pipeline](docs/package-pipeline.md)

Test locally with private mobile disabled in the test terminal: `go test ./...`, `go vet ./...`. Never expose an unauthenticated instance through a tunnel. Use existing authorized source configuration. Not production-verified.

## Repository layout

- `cmd/inkveil/`: executable package, colocated Go tests and embedded `web/` templates/assets.
- `docs/`: setup, provider and maintenance documentation.
- `start.ps1`: Windows startup, still run from repository root.
- `sources.example.json`: source configuration example.
- `data/`: ignored runtime data, unchanged.

Run from repository root: `go run ./cmd/inkveil`, `go test ./...`, `go vet ./...`, `go build -o INKVEIL.exe ./cmd/inkveil`. CLI subcommands use `go run ./cmd/inkveil <command>`. This is a structural move only; package boundaries are unchanged to avoid breaking interdependent handlers. Older historical docs may still show root command examples.

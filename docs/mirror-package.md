# Mirror package import

Stop the running server first (single writer). Back up database and data/pages. Prepare a ZIP containing manifest.json and pages/ with authorized images. Example manifest:

```json
{"Source":"licensed-partner","Permission":"license or written permission reference","Comics":[{"Slug":"example","Title":"Example","Type":"Manga","Cover":"cover.jpg","Chapters":[{"Number":"1","Images":["page-001.jpg","page-002.jpg"]}]}]}
```

PowerShell, in the same configured MySQL environment:

```powershell
$env:INKVEIL_IMPORT_PERMISSION='yes'
go run . import-package 'C:\\path\\mirror.zip'
go run .
```

Import merges existing titles from the same source, skips existing chapters, uses content-hashed filenames and rejects unsafe paths. Limits 20 comics, 256 MiB uncompressed, 10 MiB per file. No web scraping, remote fetching or automatic acquisition. Permission field records operator-provided reference in the package, not independent verification. Keep the original package for provenance. Existing metadata is retained. Run tests before use.

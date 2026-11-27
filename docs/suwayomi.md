# Suwayomi localhost integration

Use only titles for which integration and reading are authorized. Keep both servers on loopback. Enable optional import of your authorized Suwayomi library (not all installed-source catalogs), or specify explicit manga IDs. Library mode treats library membership as your selection of authorized titles. Refresh after adding/removing titles. Covers and chapter pages are loaded directly by the browser from the fixed local server; no upstream scraper, image proxy, bypass, or watermark modification. Reader is separately opt-in. The API schema and paths match current upstream source; installed-version compatibility remains to be tested.

```powershell
$env:INKVEIL_SUWAYOMI_URL="http://127.0.0.1:4567"
$env:INKVEIL_SUWAYOMI_PERMISSION="yes"
$env:INKVEIL_SUWAYOMI_LIBRARY="yes"
$env:INKVEIL_SUWAYOMI_READER="yes"
$env:INKVEIL_UNIFIED_INDEX="yes"
$env:INKVEIL_LIVE_PROVIDER="mangadex"
go test ./...
go run ./cmd/inkveil refresh-catalog
go run ./cmd/inkveil
```

Alternatively set INKVEIL_SUWAYOMI_MANGA_IDS to approved IDs and omit library mode. Import up to 1000 library records, only Indonesian sources. Existing MySQL catalog, conservative deduplication, and MangaDex fallback remain. Browser must have access to the same local Suwayomi server. Reader requires stored chapter metadata and returns an error if page URLs deviate from the expected local route. No automatic remote catalog enumeration.

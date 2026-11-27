# Suwayomi local metadata adapter

This opt-in adapter reads metadata for explicitly approved manga IDs from an existing local Suwayomi installation. Only Indonesian sources are accepted. Use only titles for which your use is authorized. No upstream scraper, image proxy, downloads or anti-bot workaround is included.

PowerShell:
```powershell
$env:INKVEIL_SUWAYOMI_URL="http://127.0.0.1:4567"
$env:INKVEIL_SUWAYOMI_PERMISSION="yes"
$env:INKVEIL_SUWAYOMI_MANGA_IDS="2"
$env:INKVEIL_UNIFIED_INDEX="yes"
$env:INKVEIL_LIVE_PROVIDER="mangadex"
go test ./...
go run . refresh-catalog
go run .
```

Replace 2 with the local ID(s) of titles you are authorized to integrate. The catalog snapshot uses existing MySQL storage and conservative author/title matching. The detail route lists stored chapters and links to the local Suwayomi UI; embedded reading and automatic source discovery are not implemented. Suwayomi outages do not prevent a MangaDex snapshot. Unset permission to hide entries. Keep both services bound to loopback.

# Configured adapter registry

Use the existing sources.local.json with an enabled source and an accurate permission reference. For offline-json:

`go run . run-source sources.local.json licensed-partner manga.json chapters.json chapter.json`

For local-package, stop the server and back up storage first. Set INKVEIL_IMPORT_PERMISSION=yes only for authorized material. The ZIP manifest Source and Permission must match the configured ID and permission reference.

`go run . run-source sources.local.json local-project package.zip`

Disabled and unknown sources are rejected. Existing standalone CLI commands remain available for backward compatibility and do not consult the registry. Base URL remains descriptive; no remote requests or automatic sync. Configuration permission is operator-supplied, not verified. No web UI changes.

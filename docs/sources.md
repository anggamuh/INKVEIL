# Source Manager

Open /admin, enter a source label and paste a JSON array using the existing catalog schema. Confirm publishing permission. Example: `[{"Slug":"new-title","Title":"New Title","Type":"Manga","Chapters":[{"Number":"1","Panels":["Original demo"]}]}]`.

Existing comics retain metadata; existing chapter numbers are skipped. Slugs owned by another source are rejected. Images must already exist as regular files in the configured pages directory. Successful imports append a report to imports.jsonl next to the catalog. If logging fails after persistence, the UI reports the failure separately. This is a manual manifest importer, not an automated scraper, remote source registry, or scheduler.

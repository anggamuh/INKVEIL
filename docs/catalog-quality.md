# Catalog quality

Set INKVEIL_INDEX_LIMIT to a multiple of 100 between 100 and 10000 (default 1000), then run refresh-catalog. Actual coverage depends on upstream availability and database size; the existing 12 MiB snapshot cap remains. Matching uses normalized alias/author buckets and preserves conservative metadata checks. Search ignores punctuation. Removed scraper hosts are filtered before merge. Settings shows stored snapshot counts and age, not live health. No new external provider activated; no cover or chapter scraper. Tests added, not run.

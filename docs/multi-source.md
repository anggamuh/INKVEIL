# Multi-source discovery

Set the comma-separated authorized Indonesian source IDs through start.ps1. At most three installed sources are queried. The explicit allowlist also limits library-derived discovery and the recent release feed; previously bookmarked/library titles remain readable. Old snapshots without source IDs are hidden when the allowlist is set. Search sources concurrently within a shared ten-second budget, with individual eight-second limits; failed sources are logged, and successful results are preserved. Filter by named source in advanced search (enter a title; a blank query does not crawl the source). Settings remain server-controlled. No change to data or authorization for existing library readers.

Run .\test.ps1 and verify each source separately before combining them.

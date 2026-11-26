# Admin source manager

Open /admin/sources or use the Source Manager link in /admin. Existing admin Basic Auth applies. Edit JSON and save. Defaults to sources.local.json; override with INKVEIL_SOURCE_CONFIG. Use the same file with run-source. Writes are atomic where supported by the filesystem, and stale browser revisions are rejected; external concurrent writers are not coordinated. Existing invalid files require local repair. This is a configuration editor, not an import job UI, and performs no remote network requests. Back up config before editing. Use HTTPS for remote admin access.

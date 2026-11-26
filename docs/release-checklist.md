# v1.0 release candidate checklist

This is a candidate, not a production certification.

- Run go test ./..., go vet ./..., go test -race ./... on supported platforms, and go build ./... .
- Test MySQL migration/startup, backup and restore of database, data/pages, team.json and source configuration.
- Smoke test public catalog, reader, authentication, bookmarks, admin, source management and team workflow on desktop/mobile.
- Use HTTPS reverse proxy, strong unique admin password, restrictive file permissions and firewall; never expose MySQL directly.
- Review existing manifest sync endpoint and network policy before enabling. Local inbox only trusted writers. One server writer only.
- Verify publication rights per title. No external image acquisition implemented.
- Load test and audit before production. Queue remains in memory. Team assignments are admin-managed labels, not user roles, and ready tasks require manual authorized publication. No independent translator logins or private draft file management yet.
- Review accessibility and visual layout on actual devices. CSS polish is not a full visual redesign.

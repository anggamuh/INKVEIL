# Reader library and local container

Bookmarks, recent chapters and reader width are stored in browser localStorage. They do not sync between devices or track exact scroll position. Clearing browser data removes them. Deleted catalog entries can remain in history until it is cleared. Use left/right arrows in reader to switch chapters.

Optional Docker: docker compose up --build. Port is bound to localhost only; named data volume is separate from existing local data. Do not delete the volume if you want to preserve content. Public deployment still requires TLS, admin hardening and backups. No live source integration or SQL database included.

# Production preparation

Set INKVEIL_PUBLIC_URL to the final HTTPS origin to enable sitemap.xml and public robots.txt. Without it crawlers are asked not to index the site. Robots directives are not access control. Pagination is 24 comics per page. Chapter-count sort is not popularity ranking. Keep the admin private, terminate TLS at a reverse proxy, apply rate limiting, back up data and test restore. Current JSON persistence is single-process only; no SQL or public account system. Run go test ./... and go test -race ./... before deployment. Docker remains optional.

# Optional MySQL storage

Back up data/catalog.json and data/pages before switching storage. Create an empty database called inkveil with utf8mb4 encoding in Laragon. For a local root account without a password:

```powershell
$env:INKVEIL_MYSQL_DSN = 'root@tcp(127.0.0.1:3306)/inkveil?charset=utf8mb4&parseTime=true'
go mod tidy
go test ./...
go run .
```

If the MySQL user has a password, configure the DSN accordingly. Never commit credentials. Use a dedicated limited database user for deployment.

At startup the application creates inkveil_catalog if needed. An empty table is seeded from the existing validated JSON catalog (or the original demo if no file exists). A nonempty SQL table takes precedence; it is never overwritten from JSON automatically. The original JSON file and images are not deleted. All subsequent catalog changes use SQL transactions; the old JSON file is not updated. Images and import logs remain local. Removing the DSN switches back to the old JSON snapshot, not the latest SQL contents. Export the current JSON through the authenticated admin editor before switching back.

This release stores one JSON payload per comic in MySQL, preserving chapter and page ordering; it is not a normalized relational schema. A single server process is required. For an integration test, use a separate empty disposable database and INKVEIL_TEST_MYSQL_DSN. Never use production data. Docker users must provide a reachable external database; localhost inside the container is not the host database.

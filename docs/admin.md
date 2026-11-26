# Local admin

Set INKVEIL_ADMIN_PASSWORD to a unique password of at least 12 characters before running Go. Open /admin and sign in with username admin. Admin is disabled when the password is unset. HTTP Basic authentication is intended for localhost development; require HTTPS and rate limiting before public deployment. Forms have CSRF protection and request size limits. Uploaded images are identified by MIME signature, not fully decoded. The advanced JSON editor replaces metadata, not image files. No remote mirroring or SQL database yet.

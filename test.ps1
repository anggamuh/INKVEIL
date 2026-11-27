$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot
if (-not (Get-Command go -ErrorAction SilentlyContinue)) { $env:Path += ';C:\Program Files\Go\bin' }
$previous = $env:INKVEIL_PRIVATE_MOBILE
try {
 $env:INKVEIL_PRIVATE_MOBILE = 'no'
 go test ./...
 if ($LASTEXITCODE -ne 0) { throw 'Go tests failed' }
 go vet ./...
 if ($LASTEXITCODE -ne 0) { throw 'Go vet failed' }
} finally { $env:INKVEIL_PRIVATE_MOBILE = $previous }

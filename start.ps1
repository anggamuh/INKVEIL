param([switch]$AuthorizedSuwayomiLibrary,[string]$AuthorizedSearchSources="",[switch]$Refresh,[switch]$PrivateMobile)
$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot
if (-not (Get-Command go -ErrorAction SilentlyContinue)) { $env:Path += ';C:\Program Files\Go\bin' }
$env:INKVEIL_LIVE_PROVIDER = 'mangadex'
$env:INKVEIL_ID_ONLY = 'yes'
if ($AuthorizedSuwayomiLibrary) {
 $env:INKVEIL_SUWAYOMI_URL = 'http://127.0.0.1:4567'
 $env:INKVEIL_SUWAYOMI_PERMISSION = 'yes'
 $env:INKVEIL_SUWAYOMI_LIBRARY = 'yes'
 $env:INKVEIL_SUWAYOMI_READER = 'yes'
}
if ($AuthorizedSearchSources) {
 $env:INKVEIL_SUWAYOMI_URL = "http://127.0.0.1:4567"
 $env:INKVEIL_SUWAYOMI_PERMISSION = "yes"
 $env:INKVEIL_SUWAYOMI_SEARCH_SOURCES = $AuthorizedSearchSources
 $env:INKVEIL_SUWAYOMI_READER = "yes"
}
if ($env:INKVEIL_MYSQL_DSN) { $env:INKVEIL_UNIFIED_INDEX = 'yes' }
if ($Refresh) {
 if (-not $env:INKVEIL_MYSQL_DSN) { throw 'Set INKVEIL_MYSQL_DSN before refreshing the MySQL index.' }
 go run . refresh-catalog
 if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}
if ($PrivateMobile) {
 $env:INKVEIL_PRIVATE_MOBILE = 'yes'
 if (-not $env:INKVEIL_MOBILE_PASSWORD) {
  $secret = Read-Host 'Password akses pribadi (minimal 16 karakter)' -AsSecureString
  $ptr = [Runtime.InteropServices.Marshal]::SecureStringToBSTR($secret)
  try { $env:INKVEIL_MOBILE_PASSWORD = [Runtime.InteropServices.Marshal]::PtrToStringBSTR($ptr) } finally { [Runtime.InteropServices.Marshal]::ZeroFreeBSTR($ptr) }
 }
}
go run .
exit $LASTEXITCODE

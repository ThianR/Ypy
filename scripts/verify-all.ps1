$ErrorActionPreference = 'Stop'
$root = Split-Path $PSScriptRoot -Parent
$env:GOROOT = $null
$go = Join-Path $root '.tools\go\bin'
if (-not (Test-Path (Join-Path $go 'go.exe'))) { throw 'Go runtime not found at .tools/go/bin' }
$env:Path = "$go;$env:Path"
Push-Location $root
try {
  gofmt -d cmd internal | Out-Null
  if ($LASTEXITCODE -ne 0) { throw 'gofmt failed' }
  Get-ChildItem db\migrations -Filter '*.sql' | ForEach-Object {
    $checksum = Join-Path $_.DirectoryName ($_.BaseName + '.sha256')
    if (-not (Test-Path -LiteralPath $checksum)) { throw "Falta checksum para $($_.Name)" }
    $expected = ((Get-Content -LiteralPath $checksum -Raw).Trim() -split '\s+')[0].ToUpperInvariant()
    $actual = (Get-FileHash -Algorithm SHA256 -LiteralPath $_.FullName).Hash.ToUpperInvariant()
    if ($expected -ne $actual) { throw "Checksum inválido para $($_.Name)" }
  }
  Write-Output 'Migration checksums passed.'
  go test ./...
  go vet ./...
  go build ./cmd/...
  powershell -NoProfile -ExecutionPolicy Bypass -File scripts\check-boundaries.ps1
  Get-ChildItem contracts -Recurse -Filter '*.json' | ForEach-Object { Get-Content -Raw $_.FullName | ConvertFrom-Json | Out-Null }
  Write-Output 'Contract JSON syntax passed.'
  python scripts\validate-contracts.py
  Push-Location web
  npm run typecheck
  npm test
  npm run build
  Pop-Location
  powershell -NoProfile -ExecutionPolicy Bypass -File scripts\verify-initial.ps1
  Write-Output 'Ypy full verification passed.'
} finally { Pop-Location }

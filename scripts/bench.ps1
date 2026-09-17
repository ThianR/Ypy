$ErrorActionPreference = 'Stop'
$root = Split-Path $PSScriptRoot -Parent
$go = Join-Path $root '.tools\go\bin\go.exe'
if (-not (Test-Path -LiteralPath $go)) { throw "Go portable runtime not found: $go" }
$env:GOROOT = $null
& $go test ./internal/poscentral -run '^$' -bench '^BenchmarkMemoryStoreApplyIdempotent$' -benchmem -count=3
if ($LASTEXITCODE -ne 0) { throw 'benchmark failed' }

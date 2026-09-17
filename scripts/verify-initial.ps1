$ErrorActionPreference = 'Stop'
$root = Split-Path $PSScriptRoot -Parent
$env:GOROOT = $null
$go = Join-Path $root '.tools\go\bin'
if (Test-Path (Join-Path $go 'go.exe')) { $env:Path = "$go;$env:Path"; go test (Join-Path $root 'cmd\...') }
$baseline = Join-Path $root 'db\migrations\0001_baseline_v4.sql'
$expected = (Get-Content (Join-Path $root 'db\migrations\0001_baseline_v4.sha256') -Raw).Trim()
$actual = (Get-FileHash $baseline -Algorithm SHA256).Hash
if ($actual -ne $expected) { throw 'Baseline V4 hash mismatch' }
Get-Content (Join-Path $root 'contracts\fixtures\pos-operation.json') | ConvertFrom-Json | Out-Null
Push-Location (Join-Path $root 'web')
npm run build
Pop-Location
Write-Output 'Initial Ypy verification passed: baseline, fixture and frontend.'

$ErrorActionPreference = 'Stop'
if (-not $env:YPY_DATABASE_URL) { throw 'Define YPY_DATABASE_URL antes de cargar datos de demostracion.' }
$psql = Join-Path $PSScriptRoot '..\..\herramientas\v4\runtime\pgsql\bin\psql.exe'
if (-not (Test-Path -LiteralPath $psql)) { $psql = (Get-Command psql -ErrorAction SilentlyContinue).Source }
if (-not $psql) { throw 'No se encontro psql.exe.' }
& $psql $env:YPY_DATABASE_URL -v ON_ERROR_STOP=1 -f (Join-Path $PSScriptRoot '..\db\seed\demo.sql')
if ($LASTEXITCODE -ne 0) { throw 'No se pudo cargar el seed de demostracion.' }
Write-Output 'Seed cargado. Login: demo / demo123. Empresa: DEMO. Terminal: POS-01.'

$ErrorActionPreference = 'Stop'

if (-not $env:YPY_DATABASE_URL) {
    throw 'Define YPY_DATABASE_URL antes de aplicar migraciones.'
}

$psql = Join-Path $PSScriptRoot '..\..\herramientas\v4\runtime\pgsql\bin\psql.exe'
if (-not (Test-Path -LiteralPath $psql)) {
    $psql = (Get-Command psql -ErrorAction SilentlyContinue).Source
}
if (-not $psql) { throw 'No se encontró psql.exe.' }

& $psql $env:YPY_DATABASE_URL -v ON_ERROR_STOP=1 -c 'CREATE TABLE IF NOT EXISTS public.ypy_schema_migration (version text PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now());'
if ($LASTEXITCODE -ne 0) { throw "No se pudo conectar a YPY_DATABASE_URL: $env:YPY_DATABASE_URL" }

$migrationDir = Join-Path $PSScriptRoot '..\db\migrations'
Get-ChildItem -LiteralPath $migrationDir -Filter '*.sql' | Sort-Object Name | ForEach-Object {
    $version = $_.BaseName
    $checksum = Join-Path $migrationDir ($version + '.sha256')
    if (Test-Path -LiteralPath $checksum) {
        $expected = ((Get-Content -LiteralPath $checksum -Raw).Trim() -split '\s+')[0].ToUpperInvariant()
        $actual = (Get-FileHash -Algorithm SHA256 -LiteralPath $_.FullName).Hash.ToUpperInvariant()
        if ($expected -ne $actual) { throw "Checksum inválido para $version" }
    }
    $appliedOutput = & $psql $env:YPY_DATABASE_URL -At -v ON_ERROR_STOP=1 -c "SELECT 1 FROM public.ypy_schema_migration WHERE version='$version'"
    if ($LASTEXITCODE -ne 0) { throw "No se pudo consultar el registro de migraciones para $version" }
    $applied = ($appliedOutput -join "`n").Trim()
    if ($applied -eq '1') {
        Write-Host "Skipping $version"
        return
    }
    Write-Host "Applying $version"
    & $psql $env:YPY_DATABASE_URL -v ON_ERROR_STOP=1 -f $_.FullName
    if ($LASTEXITCODE -ne 0) { throw "Falló la migración $version" }
    & $psql $env:YPY_DATABASE_URL -v ON_ERROR_STOP=1 -c "INSERT INTO public.ypy_schema_migration(version) VALUES ('$version')"
    if ($LASTEXITCODE -ne 0) { throw "No se pudo registrar la migración $version" }
}

& $psql $env:YPY_DATABASE_URL -At -v ON_ERROR_STOP=1 -c "SELECT 'erp_v4 migration count=' || count(*) FROM public.ypy_schema_migration"

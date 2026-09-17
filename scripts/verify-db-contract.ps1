$ErrorActionPreference = 'Stop'
if (-not $env:YPY_DATABASE_URL) { throw 'Define YPY_DATABASE_URL antes de verificar el contrato PostgreSQL.' }

$psql = Join-Path $PSScriptRoot '..\..\herramientas\v4\runtime\pgsql\bin\psql.exe'
if (-not (Test-Path -LiteralPath $psql)) { $psql = (Get-Command psql -ErrorAction SilentlyContinue).Source }
if (-not $psql) { throw 'No se encontró psql.exe.' }

$migrationCount = (& $psql $env:YPY_DATABASE_URL -At -v ON_ERROR_STOP=1 -c "SELECT count(*) FROM public.ypy_schema_migration")
if ($LASTEXITCODE -ne 0 -or [int](($migrationCount -join '').Trim()) -lt 6) { throw 'Migraciones Ypy incompletas.' }

$tables = @('gp_evento_entrada','gs_evento_salida','gp_sync_cursor','gv_comprobante_cabecera')
foreach ($table in $tables) {
    $found = (& $psql $env:YPY_DATABASE_URL -At -v ON_ERROR_STOP=1 -c "SELECT to_regclass('erp_v4.$table') IS NOT NULL")
    if ($LASTEXITCODE -ne 0 -or ($found -join '').Trim() -ne 't') { throw "Falta tabla erp_v4.$table" }
}

$function = (& $psql $env:YPY_DATABASE_URL -At -v ON_ERROR_STOP=1 -c "SELECT to_regprocedure('erp_v4.gp_integrar_venta(bigint,bigint,bigint,uuid,timestamp with time zone,jsonb)') IS NOT NULL")
if ($LASTEXITCODE -ne 0 -or ($function -join '').Trim() -ne 't') { throw 'Falta función erp_v4.gp_integrar_venta' }
$crypto = (& $psql $env:YPY_DATABASE_URL -At -v ON_ERROR_STOP=1 -c "SELECT to_regprocedure('erp_v4.crypt(text,text)') IS NOT NULL")
if ($LASTEXITCODE -ne 0 -or ($crypto -join '').Trim() -ne 't') { throw 'Falta función erp_v4.crypt() de pgcrypto' }
Write-Output 'Ypy PostgreSQL commercial contract verified.'

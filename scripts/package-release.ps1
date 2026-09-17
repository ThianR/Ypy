$ErrorActionPreference = 'Stop'
$root = Split-Path $PSScriptRoot -Parent
$output = if ($args.Count -gt 0) { [IO.Path]::GetFullPath($args[0]) } else { Join-Path $root 'dist\ypy-release.zip' }
$staging = Join-Path ([IO.Path]::GetTempPath()) ('ypy-package-' + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Force $staging | Out-Null
try {
    $items = @('README.md', 'deploy', 'contracts', 'db/migrations', 'db/seed', 'scripts/apply-migrations.ps1', 'scripts/verify-db-contract.ps1', 'scripts/seed-demo.ps1')
    foreach ($item in $items) {
        $source = Join-Path $root $item
        if (Test-Path -LiteralPath $source) {
            $destination = Join-Path $staging $item
            New-Item -ItemType Directory -Force (Split-Path $destination) | Out-Null
            Copy-Item -LiteralPath $source -Destination $destination -Recurse -Force
        }
    }
    $binDir = Join-Path $staging 'bin'
    New-Item -ItemType Directory -Force $binDir | Out-Null
    $env:GOROOT = $null
    $go = Join-Path $root '.tools\go\bin\go.exe'
    & $go build -o (Join-Path $binDir 'ypy-erp.exe') (Join-Path $root 'cmd\erp')
    if ($LASTEXITCODE -ne 0) { throw 'No se pudo compilar ypy-erp.exe.' }
    & $go build -o (Join-Path $binDir 'ypy-pos-agent.exe') (Join-Path $root 'cmd\pos-agent')
    if ($LASTEXITCODE -ne 0) { throw 'No se pudo compilar ypy-pos-agent.exe.' }
    $version = if ($env:YPY_RELEASE_VERSION) { $env:YPY_RELEASE_VERSION } else { 'development' }
    $manifest = [ordered]@{ version = $version; required_files = @('contracts/http/openapi.yaml', 'contracts/events/fiscal-document.schema.json', 'contracts/fixtures/fiscal-document.json', 'db/migrations/0001_baseline_v4.sql', 'db/migrations/0002_identidad.sql', 'db/migrations/0003_eventos_idempotencia.sql', 'db/migrations/0004_auditoria_contexto.sql', 'db/migrations/0005_sync_cursor.sql', 'db/migrations/0006_identidad_autenticacion.sql', 'db/migrations/0007_venta_cobro.sql', 'db/migrations/0008_movimientos_caja.sql', 'db/migrations/0009_cierre_movimientos_caja.sql', 'db/seed/demo.sql', 'scripts/seed-demo.ps1'); artifacts = @(
        [ordered]@{ path = 'bin/ypy-erp.exe'; sha256 = (Get-FileHash -Algorithm SHA256 (Join-Path $binDir 'ypy-erp.exe')).Hash },
        [ordered]@{ path = 'bin/ypy-pos-agent.exe'; sha256 = (Get-FileHash -Algorithm SHA256 (Join-Path $binDir 'ypy-pos-agent.exe')).Hash }
    ) }
    $manifest | ConvertTo-Json -Depth 4 | Set-Content -Encoding UTF8 -LiteralPath (Join-Path $staging 'release-manifest.json')
    New-Item -ItemType Directory -Force (Split-Path $output) | Out-Null
    if (Test-Path -LiteralPath $output) { Remove-Item -LiteralPath $output -Force }
    Compress-Archive -Path (Join-Path $staging '*') -DestinationPath $output
    (Get-FileHash -Algorithm SHA256 -LiteralPath $output).Hash | Set-Content -LiteralPath ($output + '.sha256')
    Write-Host "Package: $output"
    Write-Host "SHA256: $output.sha256"
} finally {
    if (Test-Path -LiteralPath $staging) { Remove-Item -LiteralPath $staging -Recurse -Force }
}

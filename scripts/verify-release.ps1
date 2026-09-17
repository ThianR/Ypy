$ErrorActionPreference = 'Stop'
$archive = if ($args.Count -gt 0) { [IO.Path]::GetFullPath($args[0]) } else { Join-Path (Split-Path $PSScriptRoot -Parent) 'dist\ypy-release.zip' }
if (-not (Test-Path -LiteralPath $archive)) { throw "No existe el release: $archive" }
$root = Join-Path ([IO.Path]::GetTempPath()) ('ypy-release-verify-' + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Force -Path $root | Out-Null
try {
    Expand-Archive -LiteralPath $archive -DestinationPath $root
    $manifestPath = Join-Path $root 'release-manifest.json'
    if (-not (Test-Path -LiteralPath $manifestPath)) { throw 'Falta release-manifest.json' }
    $manifest = Get-Content -Raw -LiteralPath $manifestPath | ConvertFrom-Json
    foreach ($required in $manifest.required_files) {
        $requiredPath = Join-Path $root ($required -replace '/', '\')
        if (-not (Test-Path -LiteralPath $requiredPath)) { throw "Falta archivo requerido $required" }
    }
    foreach ($artifact in $manifest.artifacts) {
        $path = Join-Path $root ($artifact.path -replace '/', '\')
        if (-not (Test-Path -LiteralPath $path)) { throw "Falta artefacto $($artifact.path)" }
        $actual = (Get-FileHash -Algorithm SHA256 -LiteralPath $path).Hash
        if ($actual -ne $artifact.sha256) { throw "SHA-256 inválido para $($artifact.path)" }
    }
    Write-Output "Ypy release verified: version $($manifest.version)."
} finally {
    if (Test-Path -LiteralPath $root) { Remove-Item -LiteralPath $root -Recurse -Force }
}

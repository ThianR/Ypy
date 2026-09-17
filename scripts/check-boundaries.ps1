$ErrorActionPreference = 'Stop'
$root = Split-Path $PSScriptRoot -Parent
$violations = @()
Get-ChildItem (Join-Path $root 'internal\modules') -Recurse -Filter '*.go' | ForEach-Object {
  $sourceModule = $_.FullName -replace '.*internal\\modules\\([^\\]+).*','$1'
  $content = Get-Content -Raw $_.FullName
  [regex]::Matches($content, 'internal/modules/([^/]+)/internal') | ForEach-Object {
    if ($_.Groups[1].Value -ne $sourceModule) { $violations += "$($_.Name): $sourceModule imports $($_.Groups[1].Value)/internal" }
  }
}
if ($violations.Count) { $violations | ForEach-Object { Write-Error $_ }; exit 1 }
Write-Output 'Module boundary check passed.'


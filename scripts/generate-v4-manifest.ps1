$ErrorActionPreference = 'Stop'
$root = Split-Path $PSScriptRoot -Parent
$sql = Get-Content -Raw (Join-Path $root 'db\migrations\0001_baseline_v4.sql')
$rows = @('# Manifiesto V4 inicial', '', '| Objeto | Propietario lógico | Procedencia |', '|---|---|---|')
[regex]::Matches($sql, 'CREATE TABLE (?:IF NOT EXISTS )?([a-zA-Z0-9_]+)') | ForEach-Object {
  $name=$_.Groups[1].Value; $owner=if($name -like 'gi_*'){'inventario/catalogo'}elseif($name -like 'gv_*'){'ventas'}elseif($name -like 'gc_*'){'compras'}elseif($name -like 'gp_*'){'pos/numeracion'}elseif($name -like 'gs_*'){'plataforma/identidad'}elseif($name -like 'gf_*'){'tesoreria'}elseif($name -like 'gl_*'){'logistica'}else{'revisar'}; $rows += "| $name | $owner | V4 |"
}
[regex]::Matches($sql, 'CREATE FUNCTION (?:IF NOT EXISTS )?([a-zA-Z0-9_]+)') | ForEach-Object { $rows += "| $($_.Groups[1].Value)() | revisar (función V4) | V4 |" }
$rows += '', '## Regla CI', '', 'Los módulos no importan `internal` de otro módulo. Toda tabla o función nueva requiere propietario único y revisión de dependencias.'
$rows -join "`n" | Set-Content (Join-Path $root 'docs\planes\MANIFIESTO_V4.md')
Write-Output "Manifest generated: $($rows.Count) lines"


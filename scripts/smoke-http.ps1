$ErrorActionPreference = 'Stop'
$root = Split-Path $PSScriptRoot -Parent
$env:GOROOT = $null
$go = Join-Path $root '.tools\go\bin'
$env:Path = "$go;$env:Path"
$process = Start-Process -FilePath (Join-Path $go 'go.exe') -ArgumentList @('run', '.\cmd\erp') -WorkingDirectory $root -PassThru -WindowStyle Hidden
try {
    $ready = $false
    for ($i = 0; $i -lt 30; $i++) {
        try { if ((Invoke-RestMethod 'http://127.0.0.1:8080/ready').status -eq 'ready') { $ready = $true; break } } catch { Start-Sleep -Milliseconds 200 }
    }
    if (-not $ready) { throw 'ERP no alcanzó estado ready.' }
    if ((Invoke-RestMethod 'http://127.0.0.1:8080/health').status -ne 'ok') { throw 'Health check falló.' }
    $id = '00000000-0000-0000-0000-000000000201'
    $body = @{ id = $id; empresa_id = '1'; terminal_id = '2'; payload = '{"sequence":1,"item":"A"}' } | ConvertTo-Json -Compress
    $response = Invoke-WebRequest 'http://127.0.0.1:8080/api/v1/pos/operations' -Method Post -ContentType 'application/json; charset=utf-8' -Headers @{ 'X-Request-ID' = '00000000-0000-0000-0000-000000000202' } -Body $body
    $ack = $response.Content | ConvertFrom-Json
    if ($response.StatusCode -ne 200 -or $ack.operation_id -ne $id -or $ack.status -ne 'ACCEPTED' -or -not $ack.content_hash) { throw 'ACK POS inválido.' }
    if ($response.Headers['X-Request-ID'] -ne '00000000-0000-0000-0000-000000000202') { throw 'X-Request-ID no fue preservado.' }
    Write-Output 'Ypy HTTP smoke test passed.'
} finally {
    if (-not $process.HasExited) { Stop-Process -Id $process.Id -Force -ErrorAction SilentlyContinue }
    $listeners = Get-NetTCPConnection -LocalPort 8080 -State Listen -ErrorAction SilentlyContinue
    foreach ($listener in $listeners) {
        $child = Get-CimInstance Win32_Process -Filter "ProcessId = $($listener.OwningProcess)" -ErrorAction SilentlyContinue
        if ($child -and $child.Name -eq 'erp.exe' -and $child.CommandLine -like '*go-build*') {
            Stop-Process -Id $listener.OwningProcess -Force -ErrorAction SilentlyContinue
        }
    }
}

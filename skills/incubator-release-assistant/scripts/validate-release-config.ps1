#Requires -Version 5.1

[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [string]$Config
)

$ErrorActionPreference = "Stop"
$engine = Join-Path $PSScriptRoot "ira"
$cache = Join-Path ([IO.Path]::GetTempPath()) "ira-go-build-cache"
$previousTelemetry = $env:GOTELEMETRY
$previousCache = $env:GOCACHE

try {
    $env:GOTELEMETRY = "off"
    $env:GOCACHE = $cache
    $configPath = $ExecutionContext.SessionState.Path.GetUnresolvedProviderPathFromPSPath($Config)
    & go -C $engine run .\cmd\ira validate --config $configPath
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
}
finally {
    if ($null -eq $previousTelemetry) { Remove-Item Env:\GOTELEMETRY -ErrorAction SilentlyContinue } else { $env:GOTELEMETRY = $previousTelemetry }
    if ($null -eq $previousCache) { Remove-Item Env:\GOCACHE -ErrorAction SilentlyContinue } else { $env:GOCACHE = $previousCache }
}

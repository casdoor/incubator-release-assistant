#Requires -Version 5.1

[CmdletBinding()]
param(
    [Parameter(Position = 0, Mandatory = $true)]
    [ValidateSet("validate", "plan", "prepare", "sign", "stage", "verify-public", "version")]
    [string]$Command,

    [string]$Config,

    [string]$Confirm,

    [switch]$Clean
)

$ErrorActionPreference = "Stop"
$engine = Join-Path $PSScriptRoot "skills\incubator-release-assistant\scripts\ira"
$cache = Join-Path $PSScriptRoot ".ira\go-build-cache"
$previousTelemetry = $env:GOTELEMETRY
$previousCache = $env:GOCACHE

try {
    $env:GOTELEMETRY = "off"
    $env:GOCACHE = $cache
    $arguments = @("run", ".\cmd\ira", $Command)
    if ($Config) {
        $configPath = $ExecutionContext.SessionState.Path.GetUnresolvedProviderPathFromPSPath($Config)
        $arguments += @("--config", $configPath)
    }
    if ($Confirm) { $arguments += @("--confirm", $Confirm) }
    if ($Clean) { $arguments += "--clean" }
    & go -C $engine @arguments
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
}
finally {
    if ($null -eq $previousTelemetry) { Remove-Item Env:\GOTELEMETRY -ErrorAction SilentlyContinue } else { $env:GOTELEMETRY = $previousTelemetry }
    if ($null -eq $previousCache) { Remove-Item Env:\GOCACHE -ErrorAction SilentlyContinue } else { $env:GOCACHE = $previousCache }
}

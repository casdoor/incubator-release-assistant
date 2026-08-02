#Requires -Version 5.1

[CmdletBinding()]
param(
    [Parameter(Position = 0, Mandatory = $true)]
    [ValidateSet("validate", "plan", "prepare", "sign", "stage", "verify-public", "version")]
    [string]$Command,

    [string]$Config,

    [string]$Confirm,

    [string]$SecretDirectory,

    [switch]$Clean
)

$ErrorActionPreference = "Stop"
$engine = Join-Path $PSScriptRoot "ira"
$cache = Join-Path ([IO.Path]::GetTempPath()) "incubator-release-assistant-go-build-cache"
$previousTelemetry = $env:GOTELEMETRY
$previousCache = $env:GOCACHE
$previousSecretDirectory = $env:IRA_SECRET_DIR
$previousGpgHome = $env:GNUPGHOME

try {
    $env:GOTELEMETRY = "off"
    $env:GOCACHE = $cache
    if ($Command -eq "sign") {
        if (-not $SecretDirectory) {
            $SecretDirectory = Join-Path (Get-Location).Path "secretkey"
        }
        $secretPath = $ExecutionContext.SessionState.Path.GetUnresolvedProviderPathFromPSPath($SecretDirectory)
        $repositoryRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..\..")).Path
        $repositoryPrefix = $repositoryRoot.TrimEnd([IO.Path]::DirectorySeparatorChar) + [IO.Path]::DirectorySeparatorChar
        if ($secretPath.Equals($repositoryRoot, [StringComparison]::OrdinalIgnoreCase) -or
            $secretPath.StartsWith($repositoryPrefix, [StringComparison]::OrdinalIgnoreCase)) {
            throw "SecretDirectory must be outside the repository. Run from its parent workspace or pass an external -SecretDirectory."
        }
        $secretParent = Split-Path -Parent $secretPath
        if (-not (Test-Path -LiteralPath $secretParent -PathType Container)) {
            throw "The parent of SecretDirectory must already exist: $secretParent"
        }
        $gitProbe = $secretPath
        while ($true) {
            if (Test-Path -LiteralPath (Join-Path $gitProbe ".git")) {
                throw "SecretDirectory must not be inside any Git worktree: $gitProbe"
            }
            $gitParent = Split-Path -Parent $gitProbe
            if (-not $gitParent -or $gitParent -eq $gitProbe) { break }
            $gitProbe = $gitParent
        }
        New-Item -ItemType Directory -Force -Path $secretPath | Out-Null
        $resolvedSecretPath = (Resolve-Path $secretPath).Path
        if ($resolvedSecretPath.Equals($repositoryRoot, [StringComparison]::OrdinalIgnoreCase) -or
            $resolvedSecretPath.StartsWith($repositoryPrefix, [StringComparison]::OrdinalIgnoreCase)) {
            throw "Resolved SecretDirectory must be outside the repository."
        }
        $gitProbe = $resolvedSecretPath
        while ($true) {
            if (Test-Path -LiteralPath (Join-Path $gitProbe ".git")) {
                throw "SecretDirectory must not be inside any Git worktree: $gitProbe"
            }
            $gitParent = Split-Path -Parent $gitProbe
            if (-not $gitParent -or $gitParent -eq $gitProbe) { break }
            $gitProbe = $gitParent
        }
        $env:IRA_SECRET_DIR = $resolvedSecretPath
        $env:GNUPGHOME = $resolvedSecretPath
    }
    $arguments = @("run", ".\cmd\ira", $Command)
    if ($Config) {
        $configPath = $ExecutionContext.SessionState.Path.GetUnresolvedProviderPathFromPSPath($Config)
        $arguments += @("--config", $configPath)
    }
    if ($Confirm) { $arguments += @("--confirm", $Confirm) }
    if ($Clean) { $arguments += "--clean" }
    & go -C $engine @arguments
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}
finally {
    if ($null -eq $previousTelemetry) { Remove-Item Env:\GOTELEMETRY -ErrorAction SilentlyContinue } else { $env:GOTELEMETRY = $previousTelemetry }
    if ($null -eq $previousCache) { Remove-Item Env:\GOCACHE -ErrorAction SilentlyContinue } else { $env:GOCACHE = $previousCache }
    if ($null -eq $previousSecretDirectory) { Remove-Item Env:\IRA_SECRET_DIR -ErrorAction SilentlyContinue } else { $env:IRA_SECRET_DIR = $previousSecretDirectory }
    if ($null -eq $previousGpgHome) { Remove-Item Env:\GNUPGHOME -ErrorAction SilentlyContinue } else { $env:GNUPGHOME = $previousGpgHome }
}

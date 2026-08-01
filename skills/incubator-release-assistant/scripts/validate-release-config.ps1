#Requires -Version 5.1

[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [string]$Config
)

Set-StrictMode -Version 2.0
$ErrorActionPreference = "Stop"

function Assert-Property {
    param(
        [Parameter(Mandatory = $true)]
        [object]$Object,
        [Parameter(Mandatory = $true)]
        [string]$Name,
        [Parameter(Mandatory = $true)]
        [string]$Path
    )

    if ($Object.PSObject.Properties.Name -notcontains $Name) {
        throw "Missing required property: $Path.$Name"
    }

    return $Object.$Name
}

function Assert-HttpsUrl {
    param(
        [string]$Value,
        [string]$Path
    )

    $uri = $null
    if (-not [Uri]::TryCreate($Value, [UriKind]::Absolute, [ref]$uri)) {
        throw "$Path must be an absolute URL"
    }
    if ($uri.Scheme -ne "https") {
        throw "$Path must use HTTPS"
    }
}

$configPath = $ExecutionContext.SessionState.Path.
    GetUnresolvedProviderPathFromPSPath($Config)
if (-not (Test-Path -LiteralPath $configPath -PathType Leaf)) {
    throw "Configuration file not found: $configPath"
}

try {
    $releaseConfig = Get-Content -LiteralPath $configPath -Raw -Encoding UTF8 |
        ConvertFrom-Json
}
catch {
    throw "Configuration is not valid JSON: $($_.Exception.Message)"
}

foreach ($section in @(
    "schemaVersion",
    "project",
    "source",
    "release",
    "checks",
    "signing",
    "distribution",
    "votes"
)) {
    Assert-Property -Object $releaseConfig -Name $section -Path "config" |
        Out-Null
}

if ([string]$releaseConfig.schemaVersion -ne "1") {
    throw "schemaVersion must be 1"
}

$source = $releaseConfig.source
$release = $releaseConfig.release
$checks = $releaseConfig.checks
$signing = $releaseConfig.signing
$distribution = $releaseConfig.distribution
$votes = $releaseConfig.votes

foreach ($name in @("repository", "commit", "archivePrefix")) {
    Assert-Property -Object $source -Name $name -Path "source" | Out-Null
}
foreach ($name in @("version", "rc", "artifactBaseName")) {
    Assert-Property -Object $release -Name $name -Path "release" | Out-Null
}
foreach ($name in @("requiredFiles", "commands", "rat")) {
    Assert-Property -Object $checks -Name $name -Path "checks" | Out-Null
}
foreach ($name in @(
    "apacheId",
    "fingerprint",
    "keysUrl",
    "requireApacheUid",
    "minimumRsaBits"
)) {
    Assert-Property -Object $signing -Name $name -Path "signing" | Out-Null
}

if ([string]$source.commit -notmatch "^[0-9a-fA-F]{40}$") {
    throw "source.commit must be a full 40-character commit"
}
if ([string]$source.commit -match "^0{40}$") {
    throw "source.commit is still a placeholder"
}
if ([int]$release.rc -lt 1) {
    throw "release.rc must be at least 1"
}
if ([string]$signing.fingerprint -notmatch "^[0-9a-fA-F]{40}$") {
    throw "signing.fingerprint must be a full 40-character fingerprint"
}
if ([string]$signing.fingerprint -match "^0{40}$") {
    throw "signing.fingerprint is still a placeholder"
}
if ([int]$signing.minimumRsaBits -lt 2048) {
    throw "signing.minimumRsaBits must be at least 2048"
}

Assert-HttpsUrl -Value ([string]$source.repository) -Path "source.repository"
Assert-HttpsUrl -Value ([string]$signing.keysUrl) -Path "signing.keysUrl"

foreach ($name in @("devUrl", "releaseUrl")) {
    $value = Assert-Property `
        -Object $distribution `
        -Name $name `
        -Path "distribution"
    Assert-HttpsUrl -Value ([string]$value) -Path "distribution.$name"
}

foreach ($name in @("devList", "generalList", "minimumHours")) {
    Assert-Property -Object $votes -Name $name -Path "votes" | Out-Null
}
if ([int]$votes.minimumHours -lt 72) {
    throw "votes.minimumHours must be at least 72"
}

if (@($checks.requiredFiles).Count -eq 0) {
    throw "checks.requiredFiles must not be empty"
}
foreach ($command in @($checks.commands)) {
    Assert-Property -Object $command -Name "name" -Path "checks.commands[]" |
        Out-Null
    Assert-Property -Object $command -Name "run" -Path "checks.commands[]" |
        Out-Null
}

Write-Host "Configuration structure is valid." -ForegroundColor Green
Write-Host "Project: $($releaseConfig.project.displayName)"
Write-Host "Release: $($release.version) RC$($release.rc)"
Write-Host "Commit: $($source.commit)"
Write-Host "Signer: $($signing.fingerprint)"

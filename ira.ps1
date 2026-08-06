#Requires -Version 5.1

[CmdletBinding()]
param(
    [Parameter(Position = 0, Mandatory = $true)]
    [ValidateSet("doctor", "validate", "plan", "prepare", "sign", "stage", "verify-public", "queue-status", "queue-prepare", "adapters", "version")]
    [string]$Command,

    [string]$Config,

    [string]$Queue,

    [string]$Confirm,

    [string]$SecretDirectory,

    [switch]$Clean
)

$runner = Join-Path $PSScriptRoot "skills\incubator-release-assistant\scripts\run.ps1"
& $runner @PSBoundParameters
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

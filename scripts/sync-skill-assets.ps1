#Requires -Version 5.1

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
$skillAssets = Join-Path $root "skills\incubator-release-assistant\assets"

Copy-Item -LiteralPath (Join-Path $root "config\release.schema.json") `
    -Destination (Join-Path $skillAssets "release.schema.json") -Force
Copy-Item -LiteralPath (Join-Path $root "config\examples\casbin-go.json") `
    -Destination (Join-Path $skillAssets "examples\casbin-go.json") -Force
Copy-Item -LiteralPath (Join-Path $root "config\release-queue.schema.json") `
    -Destination (Join-Path $skillAssets "release-queue.schema.json") -Force
Copy-Item -LiteralPath (Join-Path $root "config\examples\casbin-release-queue.json") `
    -Destination (Join-Path $skillAssets "examples\casbin-release-queue.json") -Force

Write-Host "Skill assets synchronized."

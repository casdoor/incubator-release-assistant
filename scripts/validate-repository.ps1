#Requires -Version 5.1

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
$skill = Join-Path $root "skills\incubator-release-assistant"

$pairs = @(
    @((Join-Path $root "config\release.schema.json"), (Join-Path $skill "assets\release.schema.json")),
    @((Join-Path $root "config\examples\casbin-go.json"), (Join-Path $skill "assets\examples\casbin-go.json"))
)

foreach ($pair in $pairs) {
    $left = (Get-FileHash -LiteralPath $pair[0] -Algorithm SHA256).Hash
    $right = (Get-FileHash -LiteralPath $pair[1] -Algorithm SHA256).Hash
    if ($left -ne $right) {
        throw "Skill asset drift: $($pair[0]) != $($pair[1]); run scripts/sync-skill-assets.ps1"
    }
}

Get-Content -LiteralPath (Join-Path $root "config\release.schema.json") -Raw | ConvertFrom-Json | Out-Null
Get-Content -LiteralPath (Join-Path $root "config\examples\casbin-go.json") -Raw | ConvertFrom-Json | Out-Null

$skillText = Get-Content -LiteralPath (Join-Path $skill "SKILL.md") -Raw
if ($skillText -notmatch "(?s)^---\s*\r?\nname:\s*incubator-release-assistant\s*\r?\ndescription:\s*.+?\r?\n---") {
    throw "SKILL.md frontmatter is missing or invalid"
}

foreach ($required in @(
    "scripts\ira\go.mod",
    "scripts\ira\cmd\ira\main.go",
    "assets\release.schema.json",
    "assets\examples\casbin-go.json",
    "references\configuration.md",
    "references\apache-release-gates.md"
)) {
    if (-not (Test-Path -LiteralPath (Join-Path $skill $required) -PathType Leaf)) {
        throw "Self-contained Skill resource is missing: $required"
    }
}

Write-Host "Repository contract is valid."

#Requires -Version 5.1

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
$releaseSkill = Join-Path $root "skills\incubator-release-assistant"
$handbookSkill = Join-Path $root "skills\apache-incubator-handbook"

$pairs = @(
    @((Join-Path $root "config\release.schema.json"), (Join-Path $releaseSkill "assets\release.schema.json")),
    @((Join-Path $root "config\examples\casbin-go.json"), (Join-Path $releaseSkill "assets\examples\casbin-go.json")),
    @((Join-Path $root "config\release-queue.schema.json"), (Join-Path $releaseSkill "assets\release-queue.schema.json")),
    @((Join-Path $root "config\examples\casbin-release-queue.json"), (Join-Path $releaseSkill "assets\examples\casbin-release-queue.json"))
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
Get-Content -LiteralPath (Join-Path $root "config\release-queue.schema.json") -Raw | ConvertFrom-Json | Out-Null
Get-Content -LiteralPath (Join-Path $root "config\examples\casbin-release-queue.json") -Raw | ConvertFrom-Json | Out-Null
Get-Content -LiteralPath (Join-Path $releaseSkill "evals\evals.json") -Encoding UTF8 -Raw | ConvertFrom-Json | Out-Null
Get-Content -LiteralPath (Join-Path $handbookSkill "evals\evals.json") -Encoding UTF8 -Raw | ConvertFrom-Json | Out-Null

$releaseSkillText = Get-Content -LiteralPath (Join-Path $releaseSkill "SKILL.md") -Raw
if ($releaseSkillText -notmatch "(?s)^---\s*\r?\nname:\s*incubator-release-assistant\s*\r?\ndescription:\s*.+?\r?\n---") {
    throw "Release Skill frontmatter is missing or invalid"
}

foreach ($required in @(
    "scripts\ira\go.mod",
    "scripts\ira\cmd\ira\main.go",
    "scripts\ira\internal\release\doctor.go",
    "scripts\run.ps1",
    "scripts\run.sh",
    "assets\release.schema.json",
    "assets\examples\casbin-go.json",
    "assets\release-queue.schema.json",
    "assets\examples\casbin-release-queue.json",
    "assets\examples\key-metadata.example.json",
    "assets\examples\doctor-report.example.json",
    "references\configuration.md",
    "references\workspace-bootstrap.md",
    "references\signing-key-setup.md",
    "references\asf-keys-publication.md",
    "references\prerequisites.md",
    "references\release-recovery.md"
)) {
    if (-not (Test-Path -LiteralPath (Join-Path $releaseSkill $required) -PathType Leaf)) {
        throw "Release Skill resource is missing: $required"
    }
}

if (-not (Test-Path -LiteralPath (Join-Path $root ".github\workflows\ci.yml") -PathType Leaf)) {
    throw "Lightweight CI workflow is missing: .github/workflows/ci.yml"
}

$engineText = Get-Content -LiteralPath (Join-Path $releaseSkill "scripts\ira\internal\release\engine.go") -Raw
foreach ($forbidden in @("go test", "runCasbinGoTests", "casbinGoTestCommand")) {
    if ($engineText -match [regex]::Escape($forbidden)) {
        throw "IRA prepare must not execute target-project tests: found $forbidden in engine.go"
    }
}

$ciText = Get-Content -LiteralPath (Join-Path $root ".github\workflows\ci.yml") -Raw
foreach ($requiredCheck in @("go test ./...", "go vet ./...", "gofmt -l", "bash -n", "validate-repository.ps1")) {
    if ($ciText -notmatch [regex]::Escape($requiredCheck)) {
        throw "Lightweight CI is missing required repository check: $requiredCheck"
    }
}

Get-Content -LiteralPath (Join-Path $releaseSkill "assets\examples\key-metadata.example.json") -Encoding UTF8 -Raw | ConvertFrom-Json | Out-Null
Get-Content -LiteralPath (Join-Path $releaseSkill "assets\examples\doctor-report.example.json") -Encoding UTF8 -Raw | ConvertFrom-Json | Out-Null

$powerShellWrapper = Get-Content -LiteralPath (Join-Path $releaseSkill "scripts\run.ps1") -Raw
$bashWrapper = Get-Content -LiteralPath (Join-Path $releaseSkill "scripts\run.sh") -Raw
foreach ($contract in @(
    @($powerShellWrapper, "IRA_SECRET_DIR", "PowerShell wrapper does not export the external secret directory"),
    @($powerShellWrapper, "IRA_REPOSITORY_ROOT", "PowerShell wrapper does not identify the repository for doctor"),
    @($powerShellWrapper, '"doctor"', "PowerShell wrapper does not expose doctor"),
    @($powerShellWrapper, "GNUPGHOME", "PowerShell wrapper does not isolate the GPG home"),
    @($powerShellWrapper, "outside the repository", "PowerShell wrapper does not reject repository-contained secrets"),
    @($bashWrapper, "IRA_SECRET_DIR", "Bash wrapper does not export the external secret directory"),
    @($bashWrapper, "IRA_REPOSITORY_ROOT", "Bash wrapper does not identify the repository for doctor"),
    @($bashWrapper, '"doctor"', "Bash wrapper does not expose doctor"),
    @($bashWrapper, "GNUPGHOME", "Bash wrapper does not isolate the GPG home"),
    @($bashWrapper, "outside the repository", "Bash wrapper does not reject repository-contained secrets")
)) {
    if ($contract[0] -notmatch [regex]::Escape($contract[1])) {
        throw $contract[2]
    }
}

$gitIgnore = Get-Content -LiteralPath (Join-Path $root ".gitignore") -Raw
if ($gitIgnore -notmatch "(?m)^secretkey/\r?$") {
    throw ".gitignore must exclude a mistakenly created repository-local secretkey directory"
}

$handbookSkillText = Get-Content -LiteralPath (Join-Path $handbookSkill "SKILL.md") -Raw
if ($handbookSkillText -notmatch "(?s)^---\s*\r?\nname:\s*apache-incubator-handbook\s*\r?\ndescription:\s*.+?\r?\n---") {
    throw "Handbook Skill frontmatter is missing or invalid"
}

foreach ($required in @(
    "references\README.md",
    "references\01-lifecycle-and-roles.md",
    "references\02-governance-and-reporting.md",
    "references\03-releases-voting-and-distribution.md",
    "references\04-ip-and-licensing.md",
    "references\05-branding-and-websites.md",
    "references\06-community-graduation-and-retirement.md",
    "references\07-infrastructure-accounts-and-security.md",
    "references\official-sources.md"
)) {
    if (-not (Test-Path -LiteralPath (Join-Path $handbookSkill $required) -PathType Leaf)) {
        throw "Handbook Skill resource is missing: $required"
    }
}

Write-Host "Repository and both Skill contracts are valid."

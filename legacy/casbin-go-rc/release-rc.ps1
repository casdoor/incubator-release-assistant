#Requires -Version 5.1

[CmdletBinding()]
param(
    [Parameter(Mandatory = $false)]
    [string]$Config = ".\release.local.json",

    [Parameter(Mandatory = $false)]
    [switch]$Stage,

    [Parameter(Mandatory = $false)]
    [switch]$Clean
)

Set-StrictMode -Version 2.0
$ErrorActionPreference = "Stop"

$ScriptRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$ToolsRoot = Join-Path $ScriptRoot "tools"
$WorkRoot = Join-Path $ScriptRoot "work"
$OutputRoot = Join-Path $ScriptRoot "output"
$EvidenceRoot = Join-Path $ScriptRoot "evidence"
$isolatedKeyring = $null

$RatVersion = "0.18"
$RatZipName = "apache-rat-$RatVersion-bin.zip"
$RatBaseUrl = "https://dist.apache.org/repos/dist/release/creadur/apache-rat-$RatVersion"
$DistDevUrl = "https://dist.apache.org/repos/dist/dev/incubator/casbin"
$KeysUrl = "https://dist.apache.org/repos/dist/release/incubator/casbin/KEYS"

function Write-Step {
    param(
        [int]$Number,
        [string]$Text
    )

    Write-Host ""
    Write-Host "[$Number/15] $Text" -ForegroundColor Cyan
}

function Write-Ok {
    param([string]$Text)
    Write-Host "[OK] $Text" -ForegroundColor Green
}

function Assert-Command {
    param([string]$Name)

    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        throw "找不到命令：$Name"
    }
}

function Invoke-Native {
    param(
        [string]$Name,
        [string[]]$Arguments,
        [string]$LogPath = ""
    )

    $previousPreference = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    try {
        $output = @(& $Name @Arguments 2>&1)
        $exitCode = $LASTEXITCODE
    }
    finally {
        $ErrorActionPreference = $previousPreference
    }

    if ($LogPath) {
        $output | Out-File -LiteralPath $LogPath -Encoding UTF8
    }

    foreach ($line in $output) {
        Write-Host $line
    }

    if ($exitCode -ne 0) {
        throw "$Name 执行失败，退出码：$exitCode"
    }

    return $output
}

function Invoke-NativeQuiet {
    param(
        [string]$Name,
        [string[]]$Arguments
    )

    $previousPreference = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    try {
        $output = @(& $Name @Arguments 2>&1)
        $exitCode = $LASTEXITCODE
    }
    finally {
        $ErrorActionPreference = $previousPreference
    }

    if ($exitCode -ne 0) {
        throw "$Name 执行失败，退出码：$exitCode"
    }

    return $output
}

function Invoke-NativeInteractive {
    param(
        [string]$Name,
        [string[]]$Arguments
    )

    & $Name @Arguments
    $exitCode = $LASTEXITCODE
    if ($exitCode -ne 0) {
        throw "$Name 执行失败，退出码：$exitCode"
    }
}

function Invoke-Download {
    param(
        [string]$Uri,
        [string]$OutFile
    )

    Invoke-Native -Name "svn" -Arguments @(
        "export",
        "--force",
        $Uri,
        $OutFile
    ) | Out-Null
    if (-not (Test-Path -LiteralPath $OutFile)) {
        throw "下载后找不到文件：$OutFile"
    }
}

function Invoke-DownloadWithRetry {
    param(
        [string]$Uri,
        [string]$OutFile
    )

    for ($attempt = 1; $attempt -le 6; $attempt++) {
        try {
            Invoke-Download -Uri $Uri -OutFile $OutFile
            return
        }
        catch {
            if ($attempt -eq 6) {
                throw
            }
            Write-Host "ASF文件暂时不可用，5秒后重试（$attempt/6）" `
                -ForegroundColor Yellow
            Start-Sleep -Seconds 5
        }
    }
}

function Remove-RunDirectory {
    param(
        [string]$Target,
        [string]$AllowedRoot
    )

    if (-not (Test-Path -LiteralPath $Target)) {
        return
    }

    $resolvedTarget = (Resolve-Path -LiteralPath $Target).Path
    $resolvedRoot = (Resolve-Path -LiteralPath $AllowedRoot).Path.TrimEnd("\")
    if (-not $resolvedTarget.StartsWith(
        $resolvedRoot + "\",
        [StringComparison]::OrdinalIgnoreCase
    )) {
        throw "拒绝清理预期目录以外的路径：$resolvedTarget"
    }

    Remove-Item -LiteralPath $resolvedTarget -Recurse -Force
}

function Test-Sha512File {
    param(
        [string]$ArtifactPath,
        [string]$ChecksumPath,
        [string]$ArtifactName
    )

    $bytes = [IO.File]::ReadAllBytes($ChecksumPath)
    if ($bytes.Length -eq 0) {
        throw "SHA512文件为空"
    }
    if ($bytes -contains 13) {
        throw "SHA512文件包含CR，不是LF-only格式"
    }
    if ($bytes[$bytes.Length - 1] -ne 10) {
        throw "SHA512文件末尾不是LF"
    }
    if (
        $bytes.Length -ge 3 -and
        $bytes[0] -eq 239 -and
        $bytes[1] -eq 187 -and
        $bytes[2] -eq 191
    ) {
        throw "SHA512文件包含UTF-8 BOM"
    }

    $checksumText = [Text.Encoding]::ASCII.GetString($bytes)
    $expectedPattern = "^[0-9a-f]{128}  $([regex]::Escape($ArtifactName))`n$"
    if ($checksumText -notmatch $expectedPattern) {
        throw "SHA512文件格式不正确"
    }

    $expectedHash = $checksumText.Substring(0, 128)
    $actualHash = (
        Get-FileHash -LiteralPath $ArtifactPath -Algorithm SHA512
    ).Hash.ToLowerInvariant()

    if ($expectedHash -ne $actualHash) {
        throw "SHA512数值与源码包不匹配"
    }
}

try {
    Set-Location -LiteralPath $ScriptRoot

    Write-Step 1 "读取并检查配置"

    $configPath = $ExecutionContext.SessionState.Path.
        GetUnresolvedProviderPathFromPSPath($Config)
    if (-not (Test-Path -LiteralPath $configPath)) {
        throw "找不到配置文件：$configPath"
    }

    $releaseConfig = Get-Content -LiteralPath $configPath -Raw -Encoding UTF8 |
        ConvertFrom-Json

    $requiredConfigFields = @(
        "repository",
        "commit",
        "version",
        "rc",
        "apacheId",
        "gpgFingerprint"
    )
    foreach ($field in $requiredConfigFields) {
        if ($releaseConfig.PSObject.Properties.Name -notcontains $field) {
            throw "配置缺少字段：$field"
        }
    }

    if ([string]::IsNullOrWhiteSpace([string]$releaseConfig.repository)) {
        throw "repository不能为空"
    }
    if ([string]$releaseConfig.commit -notmatch "^[0-9a-fA-F]{40}$") {
        throw "commit必须是40位完整Git commit"
    }
    if ([string]$releaseConfig.version -notmatch "^[0-9]+\.[0-9]+\.[0-9]+-incubating$") {
        throw "version格式应类似3.11.0-incubating"
    }
    if ([int]$releaseConfig.rc -lt 1) {
        throw "rc必须大于等于1"
    }
    if ([string]::IsNullOrWhiteSpace([string]$releaseConfig.apacheId)) {
        throw "apacheId不能为空"
    }
    $fingerprint = (
        [string]$releaseConfig.gpgFingerprint -replace "\s", ""
    ).ToUpperInvariant()
    if ($fingerprint -notmatch "^[0-9A-F]{40}$") {
        throw "gpgFingerprint必须是40位完整fingerprint"
    }

    $repository = [string]$releaseConfig.repository
    $commit = ([string]$releaseConfig.commit).ToLowerInvariant()
    $version = [string]$releaseConfig.version
    $rc = [int]$releaseConfig.rc
    $apacheId = [string]$releaseConfig.apacheId

    $runName = "$version-rc$rc"
    $sourceDirectoryName = "apache-casbin-$version-src"
    $artifactName = "$sourceDirectoryName.tar.gz"
    $signatureName = "$artifactName.asc"
    $checksumName = "$artifactName.sha512"

    $workDirectory = Join-Path $WorkRoot $runName
    $sourceRepository = Join-Path $workDirectory "source"
    $extractRoot = Join-Path $workDirectory "extracted"
    $extractedSource = Join-Path $extractRoot $sourceDirectoryName
    $distWorkingCopy = Join-Path $workDirectory "dist-dev"
    $downloadCheck = Join-Path $workDirectory "public-download"
    $outputDirectory = Join-Path $OutputRoot $runName
    $artifactPath = Join-Path $outputDirectory $artifactName
    $signaturePath = Join-Path $outputDirectory $signatureName
    $checksumPath = Join-Path $outputDirectory $checksumName

    $evidenceDirectory = Join-Path $EvidenceRoot (
        "$runName-" + (Get-Date -Format "yyyyMMdd-HHmmss")
    )
    $isolatedKeyring = Join-Path ([IO.Path]::GetTempPath()) (
        "cgr" + [guid]::NewGuid().ToString("N").Substring(0, 8)
    )

    Write-Ok "配置有效：$runName / $commit"

    Write-Step 2 "检查本机工具"

    foreach ($command in @("git", "tar", "go", "java", "gpg", "svn")) {
        Assert-Command $command
        Write-Ok "找到$command"
    }

    Write-Step 3 "准备相对工作目录"

    foreach ($root in @($ToolsRoot, $WorkRoot, $OutputRoot, $EvidenceRoot)) {
        if (-not (Test-Path -LiteralPath $root)) {
            New-Item -ItemType Directory -Path $root | Out-Null
        }
    }

    if ($Clean) {
        Remove-RunDirectory -Target $workDirectory -AllowedRoot $WorkRoot
        Remove-RunDirectory -Target $outputDirectory -AllowedRoot $OutputRoot
    }

    if (Test-Path -LiteralPath $workDirectory) {
        throw "工作目录已经存在：$workDirectory；确认旧内容无用后加-Clean重试"
    }
    if (Test-Path -LiteralPath $outputDirectory) {
        throw "输出目录已经存在：$outputDirectory；确认旧内容无用后加-Clean重试"
    }

    New-Item -ItemType Directory -Path $workDirectory | Out-Null
    New-Item -ItemType Directory -Path $outputDirectory | Out-Null
    New-Item -ItemType Directory -Path $evidenceDirectory | Out-Null
    Write-Ok "工作目录：.\work\$runName"
    Write-Ok "输出目录：.\output\$runName"

    Write-Step 4 "取得配置指定的源码commit"

    $archiveRepository = $sourceRepository
    if (Test-Path -LiteralPath $repository) {
        $localRepository = (
            Resolve-Path -LiteralPath $repository
        ).Path
        $archiveRepository = $localRepository
        "Using local repository: $localRepository" | Out-File -LiteralPath (
            Join-Path $evidenceDirectory "git-source.txt"
        ) -Encoding UTF8
        Write-Ok "使用本地仓库，只读取指定commit"
    }
    else {
        Invoke-Native -Name "git" -Arguments @(
            "-c",
            "http.sslBackend=openssl",
            "clone",
            "--no-checkout",
            "--no-tags",
            $repository,
            $sourceRepository
        ) -LogPath (Join-Path $evidenceDirectory "git-clone.txt") | Out-Null
    }

    $resolvedCommitOutput = Invoke-NativeQuiet -Name "git" -Arguments @(
        "-C",
        $archiveRepository,
        "rev-parse",
        "$commit`^{commit}"
    )
    $resolvedCommit = (
        [string]($resolvedCommitOutput | Select-Object -First 1)
    ).Trim().ToLowerInvariant()
    if ($resolvedCommit -ne $commit) {
        throw "仓库解析出的commit与配置不一致：$resolvedCommit"
    }
    Write-Ok "取得commit：$resolvedCommit"

    Write-Step 5 "生成源码包"

    Invoke-Native -Name "git" -Arguments @(
        "-C",
        $archiveRepository,
        "archive",
        "--format=tar.gz",
        "--prefix=$sourceDirectoryName/",
        "--output=$artifactPath",
        $commit
    ) -LogPath (Join-Path $evidenceDirectory "git-archive.txt") | Out-Null

    if (
        -not (Test-Path -LiteralPath $artifactPath) -or
        (Get-Item -LiteralPath $artifactPath).Length -eq 0
    ) {
        throw "源码包没有生成或文件为空"
    }
    Write-Ok "已生成：.\output\$runName\$artifactName"

    Write-Step 6 "解压并检查源码包内容"

    New-Item -ItemType Directory -Path $extractRoot | Out-Null
    Invoke-Native -Name "tar" -Arguments @(
        "-xzf",
        $artifactPath,
        "-C",
        $extractRoot
    ) -LogPath (Join-Path $evidenceDirectory "extract.txt") | Out-Null

    if (-not (Test-Path -LiteralPath $extractedSource -PathType Container)) {
        throw "源码包顶层目录不正确，应为：$sourceDirectoryName"
    }

    foreach ($requiredFile in @(
        "LICENSE",
        "NOTICE",
        "DISCLAIMER",
        "go.mod",
        "go.sum",
        ".rat-excludes"
    )) {
        $requiredPath = Join-Path $extractedSource $requiredFile
        if (-not (Test-Path -LiteralPath $requiredPath -PathType Leaf)) {
            throw "源码包缺少必要文件：$requiredFile"
        }
        Write-Ok "存在$requiredFile"
    }
    Write-Ok "按配置约定，前置许可证、RAT排除和logo检查已经完成"

    Write-Step 7 "下载并运行Apache RAT"

    $ratZipPath = Join-Path $ToolsRoot $RatZipName
    $ratChecksumPath = "$ratZipPath.sha512"
    $ratExtractRoot = Join-Path $ToolsRoot "apache-rat-$RatVersion"
    $ratJarPath = Join-Path $ratExtractRoot (
        "apache-rat-$RatVersion\apache-rat-$RatVersion.jar"
    )

    if (-not (Test-Path -LiteralPath $ratZipPath)) {
        Invoke-DownloadWithRetry -Uri "$RatBaseUrl/$RatZipName" `
            -OutFile $ratZipPath
    }
    Invoke-DownloadWithRetry -Uri "$RatBaseUrl/$RatZipName.sha512" `
        -OutFile $ratChecksumPath

    $ratExpectedHash = (
        (Get-Content -LiteralPath $ratChecksumPath -Raw).Split()[0]
    ).ToLowerInvariant()
    $ratActualHash = (
        Get-FileHash -LiteralPath $ratZipPath -Algorithm SHA512
    ).Hash.ToLowerInvariant()
    if ($ratExpectedHash -ne $ratActualHash) {
        throw "Apache RAT下载文件的SHA512不匹配"
    }
    Write-Ok "Apache RAT下载校验通过"

    if (-not (Test-Path -LiteralPath $ratJarPath)) {
        if (Test-Path -LiteralPath $ratExtractRoot) {
            Remove-RunDirectory -Target $ratExtractRoot -AllowedRoot $ToolsRoot
        }
        New-Item -ItemType Directory -Path $ratExtractRoot | Out-Null
        Expand-Archive -LiteralPath $ratZipPath -DestinationPath $ratExtractRoot
    }
    if (-not (Test-Path -LiteralPath $ratJarPath)) {
        throw "解压后找不到Apache RAT jar：$ratJarPath"
    }

    $ratReport = Join-Path $evidenceDirectory "rat-report.txt"
    Invoke-Native -Name "java" -Arguments @(
        "-jar",
        $ratJarPath,
        "--output-file",
        $ratReport,
        "--input-exclude-file",
        (Join-Path $extractedSource ".rat-excludes"),
        $extractedSource
    ) -LogPath (Join-Path $evidenceDirectory "rat-console.txt") | Out-Null

    $ratText = Get-Content -LiteralPath $ratReport -Raw
    $unapprovedMatch = [regex]::Match(
        $ratText,
        "(?m)^\s*Unapproved:\s*(\d+)\b"
    )
    $unknownMatch = [regex]::Match(
        $ratText,
        "(?m)^\s*Unknown:\s*(\d+)\b"
    )
    if (-not $unapprovedMatch.Success -or -not $unknownMatch.Success) {
        throw "无法从RAT报告中读取Unapproved和Unknown结果"
    }
    if (
        [int]$unapprovedMatch.Groups[1].Value -ne 0 -or
        [int]$unknownMatch.Groups[1].Value -ne 0
    ) {
        throw "RAT未通过，请查看：$ratReport"
    }
    Write-Ok "RAT：Unapproved 0 / Unknown 0"

    Write-Step 8 "运行Go测试"

    $previousGoCache = $env:GOCACHE
    $previousGoModuleCache = $env:GOMODCACHE
    $env:GOCACHE = Join-Path $workDirectory "go-build-cache"
    $env:GOMODCACHE = Join-Path $workDirectory "go-module-cache"
    Push-Location -LiteralPath $extractedSource
    try {
        Invoke-Native -Name "go" -Arguments @(
            "test",
            "./..."
        ) -LogPath (Join-Path $evidenceDirectory "go-test.txt") | Out-Null
    }
    finally {
        Pop-Location
        if ($null -eq $previousGoCache) {
            Remove-Item Env:\GOCACHE -ErrorAction SilentlyContinue
        }
        else {
            $env:GOCACHE = $previousGoCache
        }
        if ($null -eq $previousGoModuleCache) {
            Remove-Item Env:\GOMODCACHE -ErrorAction SilentlyContinue
        }
        else {
            $env:GOMODCACHE = $previousGoModuleCache
        }
    }
    Write-Ok "go test ./...通过"

    Write-Step 9 "生成并验证LF-only SHA512"

    $artifactHash = (
        Get-FileHash -LiteralPath $artifactPath -Algorithm SHA512
    ).Hash.ToLowerInvariant()
    $checksumLine = "$artifactHash  $artifactName`n"
    $utf8WithoutBom = New-Object Text.UTF8Encoding($false)
    [IO.File]::WriteAllText(
        $checksumPath,
        $checksumLine,
        $utf8WithoutBom
    )

    Test-Sha512File -ArtifactPath $artifactPath `
        -ChecksumPath $checksumPath `
        -ArtifactName $artifactName
    Write-Ok "SHA512匹配，格式为无BOM、无CR、结尾一个LF"

    Write-Step 10 "检查GPG私钥、Apache UID和官方KEYS"

    $previousPreference = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    try {
        $secretKeyOutput = @(& gpg `
            --batch `
            --with-colons `
            --list-secret-keys `
            $fingerprint 2>$null)
    }
    finally {
        $ErrorActionPreference = $previousPreference
    }
    $secretKeyLines = @($secretKeyOutput | Where-Object {
        [string]$_ -like "sec:*"
    })
    if ($secretKeyLines.Count -eq 0) {
        throw "本机找不到配置fingerprint对应的GPG私钥"
    }

    $secretKeyFields = ([string]$secretKeyLines[0]) -split ":"
    $secretKeyLength = [int]$secretKeyFields[2]
    $secretKeyAlgorithm = [int]$secretKeyFields[3]
    $secretKeyValidity = [string]$secretKeyFields[1]
    $secretKeyCapabilities = [string]$secretKeyFields[11]
    if (
        $secretKeyAlgorithm -notin @(1, 2, 3) -or
        $secretKeyLength -lt 2048
    ) {
        throw "ASF新发布签名key必须是至少2048位RSA；当前key不符合"
    }
    if ($secretKeyValidity -match "[re]" -or $secretKeyCapabilities -notmatch "s") {
        throw "配置的GPG私钥已失效、已撤销或不能用于签名"
    }

    $primaryUidLine = @($secretKeyOutput | Where-Object {
        [string]$_ -like "uid:*"
    } | Select-Object -First 1)
    if ($primaryUidLine.Count -eq 0) {
        throw "GPG私钥没有UID"
    }
    $primaryUidFields = ([string]$primaryUidLine[0]) -split ":"
    $primaryUid = $primaryUidFields[9]
    if ($primaryUid -notmatch "@apache\.org") {
        throw "GPG primary UID不是apache.org邮箱：$primaryUid"
    }
    Write-Ok "GPG私钥存在，primary UID使用apache.org邮箱"
    Write-Ok "签名key符合ASF要求：RSA $secretKeyLength bit"

    $keysPath = Join-Path $workDirectory "KEYS"
    Invoke-Download -Uri $KeysUrl -OutFile $keysPath
    New-Item -ItemType Directory -Path $isolatedKeyring | Out-Null

    Invoke-Native -Name "gpg" -Arguments @(
        "--homedir",
        $isolatedKeyring,
        "--batch",
        "--import",
        $keysPath
    ) -LogPath (Join-Path $evidenceDirectory "keys-import.txt") | Out-Null

    $officialFingerprintsOutput = Invoke-NativeQuiet -Name "gpg" -Arguments @(
        "--homedir",
        $isolatedKeyring,
        "--batch",
        "--with-colons",
        "--fingerprint"
    )
    $officialFingerprints = @(
        $officialFingerprintsOutput |
            Where-Object { [string]$_ -like "fpr:*" } |
            ForEach-Object {
                (([string]$_ -split ":")[9]).ToUpperInvariant()
            }
    )
    if ($officialFingerprints -notcontains $fingerprint) {
        throw "官方KEYS中找不到配置的完整fingerprint"
    }
    Write-Ok "官方KEYS包含完整fingerprint"

    $officialKeyOutput = Invoke-NativeQuiet -Name "gpg" -Arguments @(
        "--homedir",
        $isolatedKeyring,
        "--batch",
        "--with-colons",
        "--list-keys",
        $fingerprint
    )
    $officialUids = @(
        $officialKeyOutput |
            Where-Object { [string]$_ -like "uid:*" } |
            ForEach-Object {
                ([string]$_ -split ":")[9]
            }
    )
    if ($officialUids -notcontains $primaryUid) {
        throw "官方KEYS尚未包含本机key的primary UID：$primaryUid"
    }
    Write-Ok "官方KEYS包含本机primary UID"

    Write-Step 11 "签名并使用官方KEYS验证"

    Write-Host "GPG可能弹出密码输入窗口，请由私钥持有人本人输入。" `
        -ForegroundColor Yellow
    Invoke-NativeInteractive -Name "gpg" -Arguments @(
        "--armor",
        "--local-user",
        $fingerprint,
        "--output",
        $signaturePath,
        "--detach-sign",
        $artifactPath
    )

    Invoke-Native -Name "gpg" -Arguments @(
        "--homedir",
        $isolatedKeyring,
        "--verify",
        $signaturePath,
        $artifactPath
    ) -LogPath (Join-Path $evidenceDirectory "gpg-verify.txt") | Out-Null
    Write-Ok "GPG签名有效，并且签名key来自官方KEYS"

    Write-Step 12 "确认三个RC文件"

    foreach ($releaseFile in @(
        $artifactPath,
        $signaturePath,
        $checksumPath
    )) {
        if (
            -not (Test-Path -LiteralPath $releaseFile) -or
            (Get-Item -LiteralPath $releaseFile).Length -eq 0
        ) {
            throw "发布文件不存在或为空：$releaseFile"
        }
        $item = Get-Item -LiteralPath $releaseFile
        Write-Ok "$($item.Name) / $($item.Length) bytes"
    }

    $manifestPath = Join-Path $evidenceDirectory "release-files.txt"
    @(
        "Repository: $repository",
        "Commit: $commit",
        "Version: $version",
        "RC: $rc",
        "Fingerprint: $fingerprint",
        "Artifact SHA512: $artifactHash",
        "Artifact: $artifactName",
        "Signature: $signatureName",
        "Checksum: $checksumName"
    ) | Out-File -LiteralPath $manifestPath -Encoding UTF8

    if (-not $Stage) {
        Write-Host ""
        Write-Host "RC文件已经准备并验证完成，尚未上传ASF dist dev。" `
            -ForegroundColor Green
        Write-Host "确认后运行："
        Write-Host ".\release-rc.ps1 -Config .\release.local.json -Stage"
        exit 0
    }

    Write-Step 13 "检查ASF dist dev目标"

    $remoteEntries = Invoke-NativeQuiet -Name "svn" -Arguments @(
        "list",
        $DistDevUrl
    )
    $rcDirectoryName = "$runName/"
    if (@($remoteEntries) -contains $rcDirectoryName) {
        throw "ASF dist dev中已经存在$rcDirectoryName，禁止覆盖"
    }
    Write-Ok "远程RC目录不存在，可以创建"

    Write-Step 14 "准备并提交ASF dist dev"

    Invoke-Native -Name "svn" -Arguments @(
        "checkout",
        $DistDevUrl,
        $distWorkingCopy
    ) -LogPath (Join-Path $evidenceDirectory "svn-checkout.txt") | Out-Null

    $stageDirectory = Join-Path $distWorkingCopy $runName
    New-Item -ItemType Directory -Path $stageDirectory | Out-Null
    Copy-Item -LiteralPath $artifactPath -Destination $stageDirectory
    Copy-Item -LiteralPath $signaturePath -Destination $stageDirectory
    Copy-Item -LiteralPath $checksumPath -Destination $stageDirectory

    Invoke-Native -Name "svn" -Arguments @(
        "add",
        $stageDirectory
    ) | Out-Null
    $svnStatus = Invoke-NativeQuiet -Name "svn" -Arguments @(
        "status",
        $stageDirectory
    )
    $svnStatus | Out-File -LiteralPath (
        Join-Path $evidenceDirectory "svn-status-before-commit.txt"
    ) -Encoding UTF8
    foreach ($line in $svnStatus) {
        Write-Host $line
    }

    foreach ($releaseName in @(
        $artifactName,
        $signatureName,
        $checksumName
    )) {
        if (-not (@($svnStatus) -match [regex]::Escape($releaseName))) {
            throw "svn status中找不到发布文件：$releaseName"
        }
    }

    $confirmationText = "STAGE RC$rc"
    $confirmation = Read-Host (
        "确认上面只有本次RC三个文件后，输入 $confirmationText"
    )
    if ($confirmation -cne $confirmationText) {
        throw "没有收到正确确认文字，未提交ASF dist dev"
    }

    Write-Host "SVN将询问Apache账号密码；不要把密码写入配置文件。" `
        -ForegroundColor Yellow
    Invoke-NativeInteractive -Name "svn" -Arguments @(
        "commit",
        $stageDirectory,
        "--username",
        $apacheId,
        "--no-auth-cache",
        "-m",
        "[casbin] Stage Apache Casbin $version RC$rc"
    )

    Invoke-Native -Name "svn" -Arguments @(
        "log",
        "$DistDevUrl/$runName",
        "-l",
        "1"
    ) -LogPath (Join-Path $evidenceDirectory "svn-commit.txt") | Out-Null
    Write-Ok "ASF dist dev提交完成"

    Write-Step 15 "从公开地址重新下载并验证"

    New-Item -ItemType Directory -Path $downloadCheck | Out-Null
    $publicRcUrl = "$DistDevUrl/$runName"

    foreach ($releaseName in @(
        $artifactName,
        $signatureName,
        $checksumName
    )) {
        Invoke-DownloadWithRetry -Uri "$publicRcUrl/$releaseName" `
            -OutFile (Join-Path $downloadCheck $releaseName)
    }

    $downloadedArtifact = Join-Path $downloadCheck $artifactName
    $downloadedSignature = Join-Path $downloadCheck $signatureName
    $downloadedChecksum = Join-Path $downloadCheck $checksumName

    Test-Sha512File -ArtifactPath $downloadedArtifact `
        -ChecksumPath $downloadedChecksum `
        -ArtifactName $artifactName

    Invoke-Native -Name "gpg" -Arguments @(
        "--homedir",
        $isolatedKeyring,
        "--verify",
        $downloadedSignature,
        $downloadedArtifact
    ) -LogPath (
        Join-Path $evidenceDirectory "public-gpg-verify.txt"
    ) | Out-Null

    $downloadedHash = (
        Get-FileHash -LiteralPath $downloadedArtifact -Algorithm SHA512
    ).Hash.ToLowerInvariant()
    if ($downloadedHash -ne $artifactHash) {
        throw "公开下载的源码包与本地源码包不是同一字节"
    }

    @(
        "Public URL: $publicRcUrl/",
        "Artifact SHA512: $downloadedHash",
        "SHA512 format: OK",
        "GPG signature with official KEYS: OK"
    ) | Out-File -LiteralPath (
        Join-Path $evidenceDirectory "public-verification.txt"
    ) -Encoding UTF8

    Write-Ok "公开下载的SHA512、GPG签名和源码包字节全部验证通过"
    Write-Host ""
    Write-Host "RC已完成制作、签名、dist dev上传和公开复验。" `
        -ForegroundColor Green
    Write-Host "投票地址：$publicRcUrl/"
}
catch {
    Write-Host ""
    Write-Host "[FAILED] $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "请查看 EXPECTED_OUTPUT.md 中对应步骤，修复后再运行。"
    exit 1
}
finally {
    if (
        -not [string]::IsNullOrWhiteSpace([string]$isolatedKeyring) -and
        (Test-Path -LiteralPath $isolatedKeyring)
    ) {
        try {
            Remove-RunDirectory -Target $isolatedKeyring `
                -AllowedRoot ([IO.Path]::GetTempPath())
        }
        catch {
            Write-Host "[WARN] 临时GPG目录未能清理：$isolatedKeyring" `
                -ForegroundColor Yellow
        }
    }
    Set-Location -LiteralPath $ScriptRoot
}

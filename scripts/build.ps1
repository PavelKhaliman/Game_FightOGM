param([switch]$SkipTests)
$ErrorActionPreference = 'Stop'
$projectRoot = Split-Path -Parent $PSScriptRoot
Set-Location -LiteralPath $projectRoot
$env:GOMODCACHE = Join-Path $projectRoot '.tools\gomodcache'
$env:GOCACHE = Join-Path $projectRoot '.tools\gocache'
$env:GOTMPDIR = Join-Path $projectRoot '.tools\tmp'
$env:CGO_ENABLED = '1'
New-Item -ItemType Directory -Path $env:GOTMPDIR -Force | Out-Null
if (-not (Get-Command go -ErrorAction SilentlyContinue)) { throw 'Go 1.23+ is required. See README.md.' }
if (-not (Get-Command gcc -ErrorAction SilentlyContinue)) { throw 'MinGW-w64 GCC is required. See README.md.' }
if (-not $SkipTests) {
    & go test ./...
    if ($LASTEXITCODE -ne 0) { throw 'Tests failed.' }
}
& go build ./...
if ($LASTEXITCODE -ne 0) { throw 'Compilation failed.' }
$distPath = Join-Path $projectRoot 'dist'
New-Item -ItemType Directory -Path $distPath -Force | Out-Null
& go build -trimpath -ldflags '-s -w -H=windowsgui' -o (Join-Path $distPath 'FightOGM.exe') ./cmd/fightogm
if ($LASTEXITCODE -ne 0) { throw 'Executable build failed.' }
$assetsPath = Join-Path $distPath 'assets'
New-Item -ItemType Directory -Path $assetsPath -Force | Out-Null
foreach ($folder in @('sprites', 'backgrounds')) {
    Copy-Item -LiteralPath (Join-Path $projectRoot "assets\$folder") -Destination $assetsPath -Recurse -Force
}
# Retire only the generated legacy model copy inside this build's dist folder.
$legacyPath = Join-Path $distPath 'assets\models'
if (Test-Path -LiteralPath $legacyPath) {
    $resolvedLegacy = (Resolve-Path -LiteralPath $legacyPath).Path
    $resolvedDist = (Resolve-Path -LiteralPath $distPath).Path.TrimEnd('\') + '\'
    if (-not $resolvedLegacy.StartsWith($resolvedDist, [StringComparison]::OrdinalIgnoreCase)) { throw 'Legacy build path is outside dist.' }
    Remove-Item -LiteralPath $resolvedLegacy -Recurse -Force
}
$audioPath = Join-Path $distPath 'assets\audio'
New-Item -ItemType Directory -Path $audioPath -Force | Out-Null
Get-ChildItem -LiteralPath (Join-Path $projectRoot 'assets\audio') -File |
    Where-Object { $_.Extension -in '.wav', '.ogg' } |
    ForEach-Object { Copy-Item -LiteralPath $_.FullName -Destination $audioPath -Force }
Copy-Item -LiteralPath (Join-Path $projectRoot 'README.md') -Destination $distPath -Force
Copy-Item -LiteralPath (Join-Path $projectRoot 'THIRD_PARTY_NOTICES.md') -Destination $distPath -Force
Copy-Item -LiteralPath (Join-Path $projectRoot 'licenses') -Destination $distPath -Recurse -Force
Write-Host 'Build ready: dist\FightOGM.exe'

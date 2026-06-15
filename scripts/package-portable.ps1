$ErrorActionPreference = "Stop"

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$uiDir = Join-Path $repoRoot "ui-go"
$distDir = Join-Path $repoRoot "dist"
$outExe = Join-Path $distDir "l4d2-autobhop-ui.exe"

New-Item -ItemType Directory -Force -Path $distDir | Out-Null

Push-Location $uiDir
try {
    go build -trimpath -ldflags "-H=windowsgui -s -w" -o $outExe .
}
finally {
    Pop-Location
}

Write-Host "Portable exe created:"
Write-Host $outExe

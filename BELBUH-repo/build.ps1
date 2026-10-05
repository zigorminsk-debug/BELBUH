$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $MyInvocation.MyCommand.Path
$Build = Join-Path $Root "build"
$Dist = Join-Path $Root "dist"
Remove-Item $Build,$Dist -Recurse -Force -ErrorAction SilentlyContinue
New-Item $Build,$Dist -ItemType Directory | Out-Null

if (-not (Get-Command go -ErrorAction SilentlyContinue)) { throw "Go не найден" }
if (-not (Get-Command rsrc -ErrorAction SilentlyContinue)) {
    Write-Host "Установка rsrc..."
    go install github.com/akavel/rsrc@latest
    $env:PATH += ";$(go env GOPATH)\bin"
}
if (-not (Get-Command makensis -ErrorAction SilentlyContinue)) { throw "NSIS (makensis) не найден" }

# Основной launcher
$Launcher = Join-Path $Build "launcher"
Copy-Item "$Root\cmd\launcher" $Launcher -Recurse
Push-Location $Launcher
rsrc -manifest app.manifest -ico "$Root\assets\BELHUB.ico" -arch amd64 -o rsrc_windows_amd64.syso
$env:GOOS="windows"; $env:GOARCH="amd64"; $env:CGO_ENABLED="0"
go build -trimpath -ldflags="-s -w -H windowsgui" -o "$Dist\BELHUB-3.2.exe" .
Pop-Location

# Виджет
$Widget = Join-Path $Build "widget"
Copy-Item "$Root\cmd\widget" $Widget -Recurse
Copy-Item "$Root\cmd\launcher\app.manifest" "$Widget\app.manifest"
Push-Location $Widget
rsrc -manifest app.manifest -ico "$Root\assets\BELHUB.ico" -arch amd64 -o rsrc_windows_amd64.syso
go build -trimpath -ldflags="-s -w -H windowsgui" -o "$Dist\BELHUB-Widget.exe" .
Pop-Location

Copy-Item "$Root\web\index.html" "$Dist\BELHUB-3.2.html"
Copy-Item "$Root\web\widget.html" "$Dist\widget.html"
Copy-Item "$Root\assets" "$Dist\assets" -Recurse
Copy-Item "$Root\installer\installer.nsi" "$Dist\installer.nsi"

Push-Location $Dist
makensis installer.nsi
Pop-Location
Write-Host "Готово: $Dist" -ForegroundColor Green

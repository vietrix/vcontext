param(
  [string]$Repo = "vietrix/vcontext",
  [string]$Version = "latest",
  [string]$InstallDir = "$env:LOCALAPPDATA\\vcontext\\bin"
)

$os = "windows"
$arch = if ($env:PROCESSOR_ARCHITECTURE -match "ARM64") { "arm64" } else { "amd64" }
$asset = "vcontext_${os}_${arch}.exe"

$releaseUrl = if ($Version -eq "latest") {
  "https://api.github.com/repos/$Repo/releases/latest"
} else {
  "https://api.github.com/repos/$Repo/releases/tags/$Version"
}

$release = Invoke-RestMethod -Uri $releaseUrl -Headers @{ "User-Agent" = "vcontext" }
$assetInfo = $release.assets | Where-Object { $_.name -eq $asset } | Select-Object -First 1

if (-not $assetInfo) {
  Write-Error "asset not found for $asset"
  exit 1
}

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
$dest = Join-Path $InstallDir "vcontext.exe"

Invoke-WebRequest -Uri $assetInfo.browser_download_url -OutFile $dest

Write-Host "installed $dest"

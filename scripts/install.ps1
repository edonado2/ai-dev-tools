<#
.SYNOPSIS
    Install ai-dev-tools on Windows

.DESCRIPTION
    Downloads and installs the latest ai-dev-tools release from GitHub.

.EXAMPLE
    irm https://get.ai-dev.tools | iex

.EXAMPLE
    irm https://raw.githubusercontent.com/ernedo/ai-dev-tools/main/scripts/install.ps1 | iex
#>

$ErrorActionPreference = "Stop"

$Repo = "ernedo/ai-dev-tools"
$Binary = "ai-dev.exe"
$InstallDir = $env:INSTALL_DIR ?? "C:\Program Files\ai-dev-tools"

function Log-Info { Write-Host "[INFO] $($args[0])" -ForegroundColor Green }
function Log-Warn { Write-Host "[WARN] $($args[0])" -ForegroundColor Yellow }
function Log-Error { Write-Host "[ERROR] $($args[0])" -ForegroundColor Red }

# Detect architecture
$Arch = $env:PROCESSOR_ARCHITECTURE
if ($Arch -eq "AMD64") { $Arch = "x86_64" }
elseif ($Arch -eq "ARM64") { $Arch = "arm64" }
else { Log-Error "Unsupported architecture: $Arch"; exit 1 }

Log-Info "Detected platform: Windows/$Arch"

# Get latest release version
Log-Info "Fetching latest release..."
try {
    $Release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -ErrorAction Stop
    $Version = $Release.tag_name
    if (-not $Version) { throw "No version found" }
    Log-Info "Latest version: $Version"
} catch {
    Log-Error "Failed to fetch latest version: $_"
    exit 1
}

# Download and install
$AssetName = "ai-dev-tools_${Version.TrimStart('v')}_Windows_${Arch}.zip"
$Url = "https://github.com/$Repo/releases/download/$Version/$AssetName"
$TempDir = [System.IO.Path]::GetTempPath()
$TempFile = Join-Path $TempDir $AssetName

Log-Info "Downloading $AssetName..."
try {
    Invoke-WebRequest -Uri $Url -OutFile $TempFile -ErrorAction Stop
} catch {
    Log-Error "Failed to download $Url"
    Log-Warn "Try installing via: scoop bucket add ernedo https://github.com/ernedo/scoop-bucket; scoop install ai-dev-tools"
    Log-Warn "Or: npm i -g @ernedo/ai-dev-tools"
    exit 1
}

Log-Info "Extracting..."
$ExtractDir = Join-Path $TempDir "ai-dev-tools-extract"
if (Test-Path $ExtractDir) { Remove-Item $ExtractDir -Recurse -Force }
Expand-Archive -Path $TempFile -DestinationPath $ExtractDir -Force

Log-Info "Installing to $InstallDir..."
if (-not (Test-Path $InstallDir)) { New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null }
Move-Item (Join-Path $ExtractDir $Binary) (Join-Path $InstallDir $Binary) -Force

# Add to PATH if not already
$CurrentPath = [Environment]::GetEnvironmentVariable("Path", "Machine")
if ($CurrentPath -notlike "*$InstallDir*") {
    Log-Info "Adding $InstallDir to system PATH..."
    [Environment]::SetEnvironmentVariable("Path", "$CurrentPath;$InstallDir", "Machine")
    $env:Path = "$env:Path;$InstallDir"
}

# Cleanup
Remove-Item $TempFile -Force
Remove-Item $ExtractDir -Recurse -Force

# Verify
if (Get-Command $Binary -ErrorAction SilentlyContinue) {
    $Ver = & $Binary version 2>$null | Select-Object -First 1
    Log-Info "Successfully installed: $Ver"
} else {
    Log-Error "Installation failed: $Binary not found in PATH"
    exit 1
}

Write-Host ""
Log-Info "Next steps:"
Write-Host "  1. Configure your LLM provider: $Binary config init"
Write-Host "  2. Generate a commit: $Binary commit --dry-run"
Write-Host "  3. Review code: $Binary review --help"
Write-Host "  4. Generate tests: $Binary testgen --help"
Write-Host "  5. Create changelog: $Binary changelog --help"
Write-Host ""
Log-Warn "Note: You may need to restart your terminal for PATH changes to take effect."
param(
  [Parameter(Mandatory = $true)]
  [ValidateSet("codex", "claude")]
  [string]$Client,
  [string]$DbPath = ""
)

$cmd = Get-Command vcontext -ErrorAction SilentlyContinue
if (-not $cmd) {
  Write-Error "vcontext not found in PATH. Install first (scripts/install.ps1) or add it to PATH."
  exit 1
}

$args = @()
if ($DbPath -ne "") {
  $args += @("-db", $DbPath)
}

if ($Client -eq "codex") {
  & codex mcp add vcontext -- $cmd.Source @args
} else {
  & claude mcp add --transport stdio vcontext -- $cmd.Source @args
}

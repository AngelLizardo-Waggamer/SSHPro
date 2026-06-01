# build.ps1

param(
    [Parameter(Mandatory = $true)]
    [string]$Version
)

$AppName = "sshpro"
$OutputDir = ".\local-builds"

$targets = @(
    @{GOOS="windows"; GOARCH="amd64"; EXT=".exe"},
    @{GOOS="linux"; GOARCH="amd64"; EXT=""},
    @{GOOS="linux"; GOARCH="arm64"; EXT=""}
)

New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null

foreach ($target in $targets)
{
    $env:GOOS = $target.GOOS
    $env:GOARCH = $target.GOARCH

    $output = "$OutputDir/$AppName-$Version-$($target.GOOS)-$($target.GOARCH)$($target.EXT)"

    Write-Host "Building $output"

    go build -ldflags="-s -w" -o $output .
}

Remove-Item Env:GOOS -ErrorAction SilentlyContinue
Remove-Item Env:GOARCH -ErrorAction SilentlyContinue

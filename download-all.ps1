# Download solar usage data for 2020 through the current year into data.tsv
# Usage: .\download-all.ps1 -Token "your-bearer-token"
# Or:    $env:SOLAR_TOKEN = "your-bearer-token"; .\download-all.ps1

param(
    [string]$Token = $env:SOLAR_TOKEN
)
if (-not $Token) {
    Write-Error "Provide -Token 'your-bearer-token' or set env SOLAR_TOKEN"
}

$ErrorActionPreference = "Stop"
$exe = ".\solarscrape.exe"
if (-not (Test-Path $exe)) {
    Write-Error "solarscrape.exe not found. Run 'go build ./...' first."
}

$startYear = 2020
$endYear = (Get-Date).Year
$outFile = "data.tsv"

# Start with a header line
"Date`tWh_sum" | Out-File -FilePath $outFile -Encoding utf8

foreach ($year in $startYear..$endYear) {
    Write-Host "Fetching $year..."
    & $exe -token $Token -year $year | Out-File -FilePath $outFile -Append -Encoding utf8
}

Write-Host "Done. Output in $outFile"

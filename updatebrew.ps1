param(
    [Parameter(Mandatory = $true)]
    [string]$Version
)

$repo = "yetanotherchris/keepassview"
$platforms = @("darwin-amd64", "darwin-arm64", "linux-amd64", "linux-arm64")
$templatePath = "$PSScriptRoot/Formula/keepassview.rb.tmpl"
$formulaPath = "$PSScriptRoot/Formula/keepassview.rb"

# Always regenerate from the template so placeholders are never exhausted
$formula = Get-Content -Path $templatePath -Raw

$formula = $formula -replace '\{\{VERSION\}\}', $Version

foreach ($platform in $platforms) {
    $url = "https://github.com/$repo/releases/download/v$Version/keepassview-$platform.tar.gz"
    $tempFile = Join-Path ([System.IO.Path]::GetTempPath()) "keepassview-$platform.tar.gz"

    Write-Host "Downloading $url ..."
    Invoke-WebRequest -Uri $url -OutFile $tempFile

    $hash = (Get-FileHash -Path $tempFile -Algorithm SHA256).Hash.ToLower()
    Write-Host "SHA256 for ${platform}: $hash"

    Remove-Item $tempFile

    $placeholderKey = $platform.ToUpper() -replace '-', '_'
    $formula = $formula -replace "\{\{SHA256_$placeholderKey\}\}", $hash
}

Set-Content -Path $formulaPath -Value $formula -NoNewline
Write-Host "Updated $formulaPath with version $Version"
